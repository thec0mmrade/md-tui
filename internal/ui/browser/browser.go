package browser

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Messages emitted to parent
type FileSelectedMsg struct{ Path string }
type DirSelectedMsg struct{ Path string }

type Item struct {
    Name  string
    IsDir bool
    Size  int64
}

type paneTarget int

const (
    paneParent paneTarget = iota
    paneCurrent
    panePreview
)

type dirReadMsg struct {
    path   string
    items  []Item
    target paneTarget
}

type Model struct {
    parentDir    string
    currentDir   string
    parentItems  []Item
    items        []Item
    previewItems []Item
    cursor       int
    parentHL     int // highlighted index in parent pane
    width        int
    height       int
    ready        bool
}

func New(width, height int) Model {
    cwd, _ := os.Getwd()
    return Model{
        currentDir: cwd,
        parentDir:  filepath.Dir(cwd),
        width:      width,
        height:     height,
    }
}

func (m *Model) SetSize(w, h int) {
    m.width = w
    m.height = h
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        readDir(m.currentDir, paneCurrent),
        readDir(m.parentDir, paneParent),
    )
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dirReadMsg:
        switch msg.target {
        case paneParent:
            m.parentItems = msg.items
            m.parentHL = findIndex(m.parentItems, filepath.Base(m.currentDir))
        case paneCurrent:
            m.items = msg.items
            m.cursor = 0
            m.ready = true
            return m, m.loadPreview()
        case panePreview:
            m.previewItems = msg.items
        }
        return m, nil

    case tea.KeyMsg:
        return m.handleKey(msg)
    }

    return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    // Navigate down
    case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
        if m.cursor < len(m.items)-1 {
            m.cursor++
            return m, m.loadPreview()
        }

    // Navigate up
    case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
        if m.cursor > 0 {
            m.cursor--
            return m, m.loadPreview()
        }

    // Jump to top
    case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
        if m.cursor != 0 {
            m.cursor = 0
            return m, m.loadPreview()
        }

    // Jump to bottom
    case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
        if m.cursor != len(m.items)-1 && len(m.items) > 0 {
            m.cursor = len(m.items) - 1
            return m, m.loadPreview()
        }

    // Enter directory (l / right)
    case key.Matches(msg, key.NewBinding(key.WithKeys("l", "right"))):
        if m.cursor < len(m.items) && m.items[m.cursor].IsDir {
            return m.enterDir()
        }

    // Go to parent (h / left)
    case key.Matches(msg, key.NewBinding(key.WithKeys("h", "left"))):
        return m.goUp()

    // Upload / select (Enter / u)
    case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "u"))):
        if m.cursor < len(m.items) {
            item := m.items[m.cursor]
            path := filepath.Join(m.currentDir, item.Name)
            if item.IsDir {
                return m, func() tea.Msg { return DirSelectedMsg{Path: path} }
            }
            return m, func() tea.Msg { return FileSelectedMsg{Path: path} }
        }
    }

    return m, nil
}

func (m Model) enterDir() (Model, tea.Cmd) {
    if m.cursor >= len(m.items) {
        return m, nil
    }
    target := m.items[m.cursor]
    if !target.IsDir {
        return m, nil
    }

    newDir := filepath.Join(m.currentDir, target.Name)

    // Shift panes: parent ← current, current ← preview (if available)
    m.parentDir = m.currentDir
    m.parentItems = m.items
    m.parentHL = m.cursor
    m.currentDir = newDir
    m.items = m.previewItems
    m.previewItems = nil
    m.cursor = 0

    if len(m.items) == 0 {
        // Preview wasn't loaded, read fresh
        return m, readDir(newDir, paneCurrent)
    }

    return m, m.loadPreview()
}

func (m Model) goUp() (Model, tea.Cmd) {
    parent := filepath.Dir(m.currentDir)
    if parent == m.currentDir {
        return m, nil // already at root
    }

    // Shift panes: preview ← current, current ← parent
    m.previewItems = m.items
    m.items = m.parentItems
    m.cursor = m.parentHL

    m.currentDir = parent
    m.parentDir = filepath.Dir(parent)

    return m, readDir(m.parentDir, paneParent)
}

func (m Model) loadPreview() tea.Cmd {
    if m.cursor >= len(m.items) {
        return nil
    }
    item := m.items[m.cursor]
    if item.IsDir {
        return readDir(filepath.Join(m.currentDir, item.Name), panePreview)
    }
    // For files, no preview directory to read
    m.previewItems = nil
    return nil
}

func (m Model) View() string {
    if !m.ready {
        return "  Loading..."
    }

    listH := m.height - 4 // room for borders + footer

    parentW := m.width * 20 / 100
    currentW := m.width * 45 / 100
    previewW := m.width - parentW - currentW - 6

    if parentW < 10 {
        parentW = 10
    }
    if previewW < 10 {
        previewW = 10
    }

    // Render panes
    parentPane := m.renderPane(m.parentItems, m.parentHL, false, parentW, listH,
        shortenPath(m.parentDir))
    currentPane := m.renderPane(m.items, m.cursor, true, currentW, listH,
        filepath.Base(m.currentDir))
    previewPane := m.renderPreview(previewW, listH)

    panes := lipgloss.JoinHorizontal(lipgloss.Top, parentPane, currentPane, previewPane)

    // Path bar
    pathBar := theme.DimStyle.Render("  " + m.currentDir)

    // Hints
    hints := theme.KeyDescStyle.Render("  ") +
        theme.KeyStyle.Render("u") + theme.KeyDescStyle.Render("/") +
        theme.KeyStyle.Render("Enter") + theme.KeyDescStyle.Render(": upload  ") +
        theme.KeyStyle.Render("h") + theme.KeyDescStyle.Render("/") +
        theme.KeyStyle.Render("←") + theme.KeyDescStyle.Render(": back  ") +
        theme.KeyStyle.Render("l") + theme.KeyDescStyle.Render("/") +
        theme.KeyStyle.Render("→") + theme.KeyDescStyle.Render(": open  ") +
        theme.KeyStyle.Render("j") + theme.KeyDescStyle.Render("/") +
        theme.KeyStyle.Render("k") + theme.KeyDescStyle.Render(": move  ") +
        theme.KeyStyle.Render("q") + theme.KeyDescStyle.Render(": cancel")

    return lipgloss.JoinVertical(lipgloss.Left, panes, pathBar, hints)
}

func (m Model) renderPane(items []Item, highlight int, showCursor bool, width, height int, title string) string {
    titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.AccentColor)
    cursorStyle := lipgloss.NewStyle().
        Background(theme.AccentColor).
        Foreground(lipgloss.Color("#000000")).
        Bold(true)
    dirStyle := lipgloss.NewStyle().Foreground(theme.AccentColor)
    fileStyle := lipgloss.NewStyle().Foreground(theme.TextColor)
    hlStyle := lipgloss.NewStyle().Foreground(theme.AccentColor).Bold(true)

    innerW := width - 4 // padding/borders

    var lines []string

    // Scroll window
    start := 0
    if highlight >= height {
        start = highlight - height + 1
    }
    end := start + height
    if end > len(items) {
        end = len(items)
    }

    for i := start; i < end; i++ {
        item := items[i]
        name := item.Name
        suffix := ""
        if item.IsDir {
            suffix = "/"
        }

        display := truncate(name+suffix, innerW)

        if showCursor && i == highlight {
            line := cursorStyle.Render(padRight(display, innerW))
            lines = append(lines, line)
        } else if !showCursor && i == highlight {
            line := hlStyle.Render(padRight(display, innerW))
            lines = append(lines, line)
        } else if item.IsDir {
            lines = append(lines, dirStyle.Render(display))
        } else {
            lines = append(lines, fileStyle.Render(display))
        }
    }

    // Pad to fill height
    for len(lines) < height {
        lines = append(lines, "")
    }

    content := strings.Join(lines, "\n")

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(theme.SubtleColor).
        Width(width).
        Height(height).
        Render(titleStyle.Render(truncate(title, innerW)) + "\n" + content)
}

func (m Model) renderPreview(width, height int) string {
    if m.cursor < len(m.items) && !m.items[m.cursor].IsDir {
        // File preview: show file info
        item := m.items[m.cursor]
        path := filepath.Join(m.currentDir, item.Name)
        info := []string{
            theme.InfoLabelStyle.Render("Name:  ") + theme.InfoValueStyle.Render(item.Name),
            theme.InfoLabelStyle.Render("Size:  ") + theme.InfoValueStyle.Render(humanSize(item.Size)),
            theme.InfoLabelStyle.Render("Ext:   ") + theme.InfoValueStyle.Render(filepath.Ext(item.Name)),
            theme.InfoLabelStyle.Render("Path:  ") + theme.DimStyle.Render(truncate(path, width-8)),
        }
        content := strings.Join(info, "\n")
        return lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(theme.SubtleColor).
            Width(width).
            Height(height).
            Render(theme.InfoLabelStyle.Render("File Info") + "\n\n" + content)
    }

    // Directory preview
    return m.renderPane(m.previewItems, -1, false, width, height, "Preview")
}

// Helpers

func readDir(path string, target paneTarget) tea.Cmd {
    return func() tea.Msg {
        entries, err := os.ReadDir(path)
        if err != nil {
            return dirReadMsg{path: path, target: target}
        }

        var items []Item
        for _, e := range entries {
            // Skip hidden files
            if strings.HasPrefix(e.Name(), ".") {
                continue
            }
            info, err := e.Info()
            size := int64(0)
            if err == nil {
                size = info.Size()
            }
            items = append(items, Item{
                Name:  e.Name(),
                IsDir: e.IsDir(),
                Size:  size,
            })
        }

        // Sort: directories first, then alphabetically
        sort.Slice(items, func(i, j int) bool {
            if items[i].IsDir != items[j].IsDir {
                return items[i].IsDir
            }
            return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
        })

        return dirReadMsg{path: path, items: items, target: target}
    }
}

func findIndex(items []Item, name string) int {
    for i, item := range items {
        if item.Name == name {
            return i
        }
    }
    return 0
}

func truncate(s string, maxLen int) string {
    if maxLen <= 0 {
        return ""
    }
    if len(s) <= maxLen {
        return s
    }
    if maxLen <= 2 {
        return s[:maxLen]
    }
    return s[:maxLen-2] + ".."
}

func padRight(s string, width int) string {
    if len(s) >= width {
        return s[:width]
    }
    return s + strings.Repeat(" ", width-len(s))
}

func shortenPath(path string) string {
    home, err := os.UserHomeDir()
    if err != nil {
        return path
    }
    if strings.HasPrefix(path, home) {
        return "~" + path[len(home):]
    }
    return path
}

func humanSize(b int64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
