package download

import (
    "fmt"
    "path/filepath"
    "strings"
    "time"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type phaseTickMsg struct{}

type state int

const (
    stateForm state = iota
    stateDownloading
    stateDone
)

type DoneMsg struct {
    Err error
}

type CancelMsg struct{}

type ProgressMsg struct {
    Progress device.TransferProgress
}

type StartDownloadMsg struct{}

type Model struct {
    active     bool
    state      state
    trackIndex int
    trackTitle string
    pathInput  textinput.Model
    progress   progress.Model
    phase      string
    pct        float64
    err        error
    width      int
    dotFrame   int
}

func New() Model {
    pi := textinput.New()
    pi.Placeholder = "/path/to/output.raw"
    pi.CharLimit = 256
    pi.Width = 50

    p := progress.New(
        progress.WithDefaultGradient(),
        progress.WithWidth(50),
    )

    return Model{
        pathInput: pi,
        progress:  p,
    }
}

func (m *Model) Open(trackIndex int, trackTitle string, width int) {
    m.active = true
    m.state = stateForm
    m.trackIndex = trackIndex
    m.trackTitle = trackTitle
    m.width = width
    m.err = nil
    m.pct = 0
    m.phase = ""

    // Default output path from track title
    safe := strings.Map(func(r rune) rune {
        if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
            return '_'
        }
        return r
    }, trackTitle)
    if safe == "" {
        safe = fmt.Sprintf("track_%d", trackIndex+1)
    }
    m.pathInput.SetValue(filepath.Join(".", safe+".raw"))
    m.pathInput.Focus()
    m.pathInput.CursorEnd()
}

func (m *Model) IsActive() bool {
    return m.active
}

func (m *Model) Close() {
    m.active = false
}

func (m *Model) GetDownloadParams() (trackIndex int, destPath string) {
    return m.trackIndex, m.pathInput.Value()
}

func (m *Model) SetDownloading() tea.Cmd {
    m.state = stateDownloading
    m.pathInput.Blur()
    m.dotFrame = 0
    return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg {
        return phaseTickMsg{}
    })
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if !m.active {
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.state == stateForm {
            return m.updateForm(msg)
        }
        if m.state == stateDone {
            m.active = false
            if m.err != nil {
                return m, func() tea.Msg { return CancelMsg{} }
            }
            return m, func() tea.Msg { return DoneMsg{} }
        }

    case phaseTickMsg:
        if m.state == stateDownloading {
            m.dotFrame = (m.dotFrame + 1) % 3
            return m, tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg {
                return phaseTickMsg{}
            })
        }
        return m, nil

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
    }

    return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, theme.Keys.Cancel):
        m.active = false
        return m, func() tea.Msg { return CancelMsg{} }
    case key.Matches(msg, theme.Keys.Confirm):
        if m.pathInput.Value() != "" {
            m.state = stateDownloading
            m.pathInput.Blur()
            return m, func() tea.Msg { return StartDownloadMsg{} }
        }
        return m, nil
    }

    var cmd tea.Cmd
    m.pathInput, cmd = m.pathInput.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if !m.active {
        return ""
    }

    modalWidth := 60
    if m.width > 0 && m.width < 70 {
        modalWidth = m.width - 10
    }

    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.AccentColor).
        Render("Download Track")

    trackInfo := fmt.Sprintf("Track %d: %s", m.trackIndex+1, m.trackTitle)
    if m.trackTitle == "" {
        trackInfo = fmt.Sprintf("Track %d: (untitled)", m.trackIndex+1)
    }
    trackLine := lipgloss.NewStyle().Foreground(theme.TextColor).Render(trackInfo)

    var body string
    switch m.state {
    case stateForm:
        body = m.viewForm()
    case stateDownloading:
        body = m.viewProgress()
    case stateDone:
        if m.err != nil {
            body = theme.ErrorBannerStyle.Render("Download failed: "+m.err.Error()) + "\n\n" +
                theme.KeyDescStyle.Render("Press any key to go back")
        } else {
            body = theme.SuccessBannerStyle.Render("Download complete!") + "\n\n" +
                theme.KeyDescStyle.Render("Press any key to continue")
        }
    }

    content := lipgloss.JoinVertical(lipgloss.Left, header, "", trackLine, "", body)
    return theme.ModalStyle.Width(modalWidth).Render(content)
}

func (m Model) viewForm() string {
    label := lipgloss.NewStyle().Foreground(theme.AccentColor).Bold(true).Render("Save to:")
    hints := theme.KeyDescStyle.Render("Enter: download  Esc: cancel")

    return lipgloss.JoinVertical(lipgloss.Left,
        label,
        m.pathInput.View(),
        "",
        hints,
    )
}

func (m Model) viewProgress() string {
    phase := m.phase
    if phase == "" {
        phase = "reading"
    }
    dots := strings.Repeat(".", m.dotFrame+1) + strings.Repeat(" ", 2-m.dotFrame)
    return fmt.Sprintf("  %s%s %.0f%%\n\n", phase, dots, m.pct*100) +
        "  " + m.progress.ViewAs(m.pct)
}
