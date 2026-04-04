package upload

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "unicode"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui/browser"
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

var audioExts = map[string]bool{
    ".wav": true, ".mp3": true, ".flac": true, ".ogg": true,
    ".aac": true, ".m4a": true, ".wma": true, ".opus": true,
}

type state int

const (
    stateBrowse state = iota
    stateTitle
    stateBatchConfirm
    stateUploading
    stateDone
)

type titleField int

const (
    fieldTitleInput titleField = iota
    fieldTitleFormat
)

type DoneMsg struct {
    Err error
}

type CancelMsg struct{}

type ProgressMsg struct {
    Progress device.TransferProgress
}

type StartUploadMsg struct{}

type queueItem struct {
    Path  string
    Title string
}

type Model struct {
    active       bool
    state        state
    browser      browser.Model
    titleInput   textinput.Model
    titleFocus   titleField
    selectedFile string
    selectedDir  string
    format       device.UploadFormat
    progress     progress.Model
    phase        string
    pct          float64
    err          error
    width        int
    height       int

    // Batch upload queue
    queue    []queueItem
    queueIdx int
    batchMode bool
}

func New() Model {
    ti := textinput.New()
    ti.Placeholder = "Track title"
    ti.CharLimit = 120
    ti.Width = 50

    p := progress.New(
        progress.WithDefaultGradient(),
        progress.WithWidth(50),
    )

    return Model{
        titleInput: ti,
        progress:   p,
        format:     device.FormatSP,
    }
}

func (m *Model) Open(width, height int) tea.Cmd {
    m.active = true
    m.state = stateBrowse
    m.width = width
    m.height = height
    m.err = nil
    m.pct = 0
    m.phase = ""
    m.selectedFile = ""
    m.selectedDir = ""
    m.titleInput.SetValue("")
    m.titleInput.Blur()
    m.titleFocus = fieldTitleInput
    m.format = device.FormatSP
    m.queue = nil
    m.queueIdx = 0
    m.batchMode = false
    m.browser = browser.New(width, height-2)
    return m.browser.Init()
}

func (m *Model) IsActive() bool {
    return m.active
}

func (m *Model) Close() {
    m.active = false
}

func (m *Model) SetSize(w, h int) {
    m.width = w
    m.height = h
    m.browser.SetSize(w, h-2)
}

func (m *Model) GetUploadParams() (path, title string, format device.UploadFormat) {
    if m.batchMode && m.queueIdx < len(m.queue) {
        item := m.queue[m.queueIdx]
        return item.Path, item.Title, m.format
    }
    return m.selectedFile, m.titleInput.Value(), m.format
}

func (m *Model) SetUploading() {
    m.state = stateUploading
    m.titleInput.Blur()
    m.pct = 0
    m.phase = ""
}

func (m *Model) AdvanceQueue() bool {
    if !m.batchMode {
        return false
    }
    m.queueIdx++
    return m.queueIdx < len(m.queue)
}

func (m *Model) QueueStatus() string {
    if !m.batchMode || len(m.queue) == 0 {
        return ""
    }
    return fmt.Sprintf("Track %d / %d", m.queueIdx+1, len(m.queue))
}

func (m *Model) CurrentTrackName() string {
    if m.batchMode && m.queueIdx < len(m.queue) {
        return m.queue[m.queueIdx].Title
    }
    return m.titleInput.Value()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if !m.active {
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.state {
        case stateBrowse:
            return m.updateBrowse(msg)
        case stateTitle:
            return m.updateTitle(msg)
        case stateBatchConfirm:
            return m.updateBatchConfirm(msg)
        case stateDone:
            m.active = false
            if m.err != nil {
                return m, func() tea.Msg { return CancelMsg{} }
            }
            return m, func() tea.Msg { return DoneMsg{} }
        }

    case ProgressMsg:
        if msg.Progress.TotalBytes > 0 {
            m.pct = float64(msg.Progress.BytesSent) / float64(msg.Progress.TotalBytes)
        }
        m.phase = msg.Progress.Phase
        return m, nil

    case progress.FrameMsg:
        pm, cmd := m.progress.Update(msg)
        m.progress = pm.(progress.Model)
        return m, cmd

    case browser.FileSelectedMsg:
        m.selectedFile = msg.Path
        m.state = stateTitle
        m.titleFocus = fieldTitleInput
        base := filepath.Base(msg.Path)
        ext := filepath.Ext(base)
        m.titleInput.SetValue(strings.TrimSuffix(base, ext))
        m.titleInput.Focus()
        m.titleInput.CursorEnd()
        return m, nil

    case browser.DirSelectedMsg:
        files := scanAudioFiles(msg.Path)
        if len(files) > 0 {
            m.queue = files
            m.queueIdx = 0
            m.batchMode = true
            m.selectedDir = msg.Path
            m.state = stateBatchConfirm
        }
        return m, nil
    }

    // Delegate non-key messages to browser
    if m.state == stateBrowse {
        var cmd tea.Cmd
        m.browser, cmd = m.browser.Update(msg)
        return m, cmd
    }

    return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (Model, tea.Cmd) {
    if key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))) {
        m.active = false
        return m, func() tea.Msg { return CancelMsg{} }
    }

    var cmd tea.Cmd
    m.browser, cmd = m.browser.Update(msg)
    return m, cmd
}

func (m Model) updateTitle(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, theme.Keys.Cancel):
        m.state = stateBrowse
        m.titleInput.Blur()
        return m, nil

    case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
        if m.titleFocus == fieldTitleInput {
            m.titleFocus = fieldTitleFormat
            m.titleInput.Blur()
        } else {
            m.titleFocus = fieldTitleInput
            m.titleInput.Focus()
        }
        return m, nil

    case key.Matches(msg, theme.Keys.Confirm):
        if m.titleFocus == fieldTitleFormat {
            // Enter on format field just confirms
        }
        m.state = stateUploading
        m.titleInput.Blur()
        return m, func() tea.Msg { return StartUploadMsg{} }
    }

    // Format selector
    if m.titleFocus == fieldTitleFormat {
        switch msg.String() {
        case "left", "h":
            if m.format > device.FormatSP {
                m.format--
            }
            return m, nil
        case "right", "l":
            if m.format < device.FormatLP2 {
                m.format++
            }
            return m, nil
        }
        return m, nil
    }

    var cmd tea.Cmd
    m.titleInput, cmd = m.titleInput.Update(msg)
    return m, cmd
}

func (m Model) updateBatchConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, theme.Keys.Cancel):
        m.state = stateBrowse
        m.batchMode = false
        m.queue = nil
        return m, nil

    case key.Matches(msg, theme.Keys.Confirm):
        m.state = stateUploading
        return m, func() tea.Msg { return StartUploadMsg{} }

    case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
        if m.format > device.FormatSP {
            m.format--
        }
        return m, nil

    case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
        if m.format < device.FormatLP2 {
            m.format++
        }
        return m, nil
    }

    return m, nil
}

func (m Model) View() string {
    if !m.active {
        return ""
    }

    switch m.state {
    case stateBrowse:
        return m.browser.View()
    case stateTitle:
        return m.viewTitle()
    case stateBatchConfirm:
        return m.viewBatchConfirm()
    case stateUploading:
        return m.viewModal("Upload to Disc", m.viewProgress())
    case stateDone:
        if m.err != nil {
            return m.viewModal("Upload to Disc",
                theme.ErrorBannerStyle.Render("Upload failed: "+m.err.Error())+"\n\n"+
                    theme.KeyDescStyle.Render("Press any key to go back"))
        }
        return m.viewModal("Upload to Disc",
            theme.SuccessBannerStyle.Render("Upload complete!")+"\n\n"+
                theme.KeyDescStyle.Render("Press any key to continue"))
    }
    return ""
}

func (m Model) viewTitle() string {
    selectedLabel := lipgloss.NewStyle().Foreground(theme.DimTextColor).Render("File: ")
    selectedFile := lipgloss.NewStyle().Foreground(theme.TextColor).Render(filepath.Base(m.selectedFile))

    titleLabel := lipgloss.NewStyle().Foreground(theme.DimTextColor).Render("Title:")
    if m.titleFocus == fieldTitleInput {
        titleLabel = lipgloss.NewStyle().Foreground(theme.AccentColor).Bold(true).Render("Title:")
    }

    formatLine := m.renderFormatSelector(m.titleFocus == fieldTitleFormat)
    hints := theme.KeyDescStyle.Render("Tab: switch field  Enter: upload  Esc: back")

    body := lipgloss.JoinVertical(lipgloss.Left,
        selectedLabel+selectedFile,
        "",
        titleLabel,
        m.titleInput.View(),
        "",
        formatLine,
        "",
        hints,
    )

    return m.viewModal("Upload to Disc", body)
}

func (m Model) viewBatchConfirm() string {
    dirName := filepath.Base(m.selectedDir)
    trackCount := len(m.queue)

    info := lipgloss.NewStyle().Foreground(theme.TextColor).Render(
        fmt.Sprintf("Upload %d tracks from: %s", trackCount, dirName))

    formatLine := m.renderFormatSelector(true)
    hints := theme.KeyDescStyle.Render("←/→: change format  Enter: start  Esc: cancel")

    body := lipgloss.JoinVertical(lipgloss.Left,
        info,
        "",
        formatLine,
        "",
        hints,
    )

    return m.viewModal("Batch Upload", body)
}

func (m Model) renderFormatSelector(focused bool) string {
    label := lipgloss.NewStyle().Foreground(theme.DimTextColor).Render("Format: ")
    if focused {
        label = lipgloss.NewStyle().Foreground(theme.AccentColor).Bold(true).Render("Format: ")
    }

    activeStyle := lipgloss.NewStyle().Foreground(theme.AccentColor).Bold(true)
    inactiveStyle := lipgloss.NewStyle().Foreground(theme.DimTextColor)

    sp := inactiveStyle.Render(" SP ")
    lp2 := inactiveStyle.Render(" LP2 ")

    switch m.format {
    case device.FormatSP:
        sp = activeStyle.Render("[SP]")
    case device.FormatLP2:
        lp2 = activeStyle.Render("[LP2]")
    }

    return label + sp + "  " + lp2
}

func (m Model) viewModal(title, body string) string {
    modalWidth := 60
    if m.width > 0 && m.width < 70 {
        modalWidth = m.width - 10
    }

    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.AccentColor).
        Render(title)

    content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
    modal := theme.ModalStyle.Width(modalWidth).Render(content)

    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) viewProgress() string {
    phase := m.phase
    if phase == "" {
        phase = "preparing"
    }

    status := ""
    if m.batchMode {
        status = fmt.Sprintf("  %s\n", m.QueueStatus())
        status += fmt.Sprintf("  %s\n\n", lipgloss.NewStyle().Foreground(theme.TextColor).Render(m.CurrentTrackName()))
    }

    fmtLabel := "SP"
    if m.format == device.FormatLP2 {
        fmtLabel = "LP2"
    }

    return status +
        fmt.Sprintf("  %s (%s)... %.0f%%\n\n", phase, fmtLabel, m.pct*100) +
        "  " + m.progress.ViewAs(m.pct)
}

func scanAudioFiles(dir string) []queueItem {
    var items []queueItem
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil
    }
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        ext := strings.ToLower(filepath.Ext(e.Name()))
        if audioExts[ext] {
            title := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
            items = append(items, queueItem{
                Path:  filepath.Join(dir, e.Name()),
                Title: title,
            })
        }
    }
    sort.Slice(items, func(i, j int) bool {
        return naturalLess(filepath.Base(items[i].Path), filepath.Base(items[j].Path))
    })
    return items
}

func naturalLess(a, b string) bool {
    for {
        if a == "" {
            return b != ""
        }
        if b == "" {
            return false
        }
        aDigit := unicode.IsDigit(rune(a[0]))
        bDigit := unicode.IsDigit(rune(b[0]))
        if aDigit && bDigit {
            ai, aj := 0, 0
            for aj < len(a) && unicode.IsDigit(rune(a[aj])) {
                aj++
            }
            bi, bj := 0, 0
            for bj < len(b) && unicode.IsDigit(rune(b[bj])) {
                bj++
            }
            an, _ := strconv.Atoi(a[ai:aj])
            bn, _ := strconv.Atoi(b[bi:bj])
            if an != bn {
                return an < bn
            }
            a = a[aj:]
            b = b[bj:]
        } else if aDigit != bDigit {
            return aDigit
        } else {
            la := strings.ToLower(string(a[0]))
            lb := strings.ToLower(string(b[0]))
            if la != lb {
                return la < lb
            }
            a = a[1:]
            b = b[1:]
        }
    }
}
