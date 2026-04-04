package discview

import (
    "fmt"
    "strings"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type moveState int

const (
    moveNone moveState = iota
    moveActive
)

type MoveTrackMsg struct {
    From, To int
}

type Model struct {
    table    table.Model
    disc     *device.Disc
    width    int
    height   int
    moveMode moveState
    moveFrom int
    showHelp bool
}

func New(width, height int) Model {
    cols := []table.Column{
        {Title: "#", Width: 4},
        {Title: "Title", Width: 30},
        {Title: "Duration", Width: 10},
        {Title: "Fmt", Width: 5},
        {Title: "Ch", Width: 4},
    }

    t := table.New(
        table.WithColumns(cols),
        table.WithFocused(true),
    )

    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(theme.SubtleColor).
        BorderBottom(true).
        Bold(true)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("#000000")).
        Background(theme.AccentColor).
        Bold(true)
    t.SetStyles(s)

    m := Model{
        table:  t,
        width:  width,
        height: height,
    }
    m.updateTableSize()
    return m
}

func (m *Model) SetDisc(d *device.Disc) {
    m.disc = d
    m.rebuildRows()
}

func (m *Model) SetSize(w, h int) {
    m.width = w
    m.height = h
    m.updateTableSize()
}

func (m *Model) SelectedTrackIndex() int {
    return m.table.Cursor()
}

func (m *Model) HasTracks() bool {
    return m.disc != nil && len(m.disc.Tracks) > 0
}

func (m *Model) InMoveMode() bool {
    return m.moveMode == moveActive
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.moveMode == moveActive {
            return m.updateMoveMode(msg)
        }

        switch {
        case key.Matches(msg, theme.Keys.Move):
            if m.HasTracks() && m.disc != nil && !m.disc.WriteProtected {
                m.moveMode = moveActive
                m.moveFrom = m.table.Cursor()
            }
            return m, nil
        case key.Matches(msg, theme.Keys.Help):
            m.showHelp = !m.showHelp
            return m, nil
        }
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m Model) updateMoveMode(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, theme.Keys.Confirm):
        to := m.table.Cursor()
        m.moveMode = moveNone
        if m.moveFrom != to {
            from := m.moveFrom
            return m, func() tea.Msg {
                return MoveTrackMsg{From: from, To: to}
            }
        }
        return m, nil
    case key.Matches(msg, theme.Keys.Cancel):
        m.moveMode = moveNone
        return m, nil
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if m.disc == nil {
        return "  Loading disc content..."
    }

    tableWidth := m.tableWidth()
    infoWidth := m.width - tableWidth - 5

    // Left pane: track table
    tableView := m.table.View()
    leftPane := theme.PanelStyle.Width(tableWidth).Render(tableView)

    // Right pane: disc info
    infoView := m.renderDiscInfo(infoWidth)
    rightPane := theme.PanelStyle.Width(infoWidth).Render(infoView)

    // Join panes
    mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

    // Footer
    footer := m.renderFooter()

    return lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
}

func (m Model) renderDiscInfo(width int) string {
    if m.disc == nil {
        return ""
    }

    d := m.disc
    var b strings.Builder

    label := theme.InfoLabelStyle.Render
    value := theme.InfoValueStyle.Render

    title := d.Title
    if title == "" {
        title = theme.DimStyle.Render("(untitled)")
    }
    b.WriteString(label("Title:  ") + value(title) + "\n")
    b.WriteString(label("Tracks: ") + value(fmt.Sprintf("%d", len(d.Tracks))) + "\n")
    b.WriteString("\n")
    b.WriteString(label("Used:   ") + value(formatTime(d.UsedSeconds)) + "\n")
    b.WriteString(label("Free:   ") + value(formatTime(d.FreeSeconds)) + "\n")
    b.WriteString(label("Total:  ") + value(formatTime(d.TotalSeconds)) + "\n")
    b.WriteString("\n")

    if d.WriteProtected {
        b.WriteString(lipgloss.NewStyle().Foreground(theme.WarningColor).Render("Read Only"))
    } else {
        b.WriteString(lipgloss.NewStyle().Foreground(theme.SuccessColor).Render("Read/Write"))
    }

    return b.String()
}

func (m Model) renderFooter() string {
    if m.moveMode == moveActive {
        return lipgloss.NewStyle().Foreground(theme.WarningColor).Render(
            fmt.Sprintf("  Moving track %d → use ↑/↓ to pick destination, Enter to confirm, Esc to cancel",
                m.moveFrom+1))
    }

    if m.showHelp {
        return m.renderFullHelp()
    }

    return m.renderShortHelp()
}

func (m Model) renderShortHelp() string {
    k := theme.KeyStyle.Render
    d := theme.KeyDescStyle.Render
    return d("  ") +
        k("u") + d("pload ") +
        k("x") + d("tract ") +
        k("r") + d("ename ") +
        k("d") + d("elete ") +
        k("m") + d("ove ") +
        k("R") + d("ename disc ") +
        k("W") + d("ipe  ") +
        k("?") + d("help ") +
        k("q") + d("uit")
}

func (m Model) renderFullHelp() string {
    k := theme.KeyStyle.Render
    d := theme.KeyDescStyle.Render
    lines := []string{
        d("  ") + k("↑/k") + d(" up  ") + k("↓/j") + d(" down  ") + k("g") + d(" top  ") + k("G") + d(" bottom"),
        d("  ") + k("u") + d(" upload  ") + k("x") + d(" extract/download  ") + k("r") + d(" rename track  ") + k("R") + d(" rename disc"),
        d("  ") + k("d") + d(" delete  ") + k("m") + d(" move  ") + k("W") + d(" wipe disc  ") + k("?") + d(" close help  ") + k("q") + d(" quit"),
    }
    return strings.Join(lines, "\n")
}

func (m *Model) rebuildRows() {
    if m.disc == nil {
        m.table.SetRows(nil)
        return
    }

    rows := make([]table.Row, len(m.disc.Tracks))
    for i, t := range m.disc.Tracks {
        title := t.Title
        if title == "" {
            title = "(untitled)"
        }
        ch := "St"
        if t.Channels == 1 {
            ch = "Mo"
        }
        rows[i] = table.Row{
            fmt.Sprintf("%d", i+1),
            title,
            formatDuration(t.Duration),
            t.Encoding.String(),
            ch,
        }
    }
    m.table.SetRows(rows)
}

func (m *Model) updateTableSize() {
    tableW := m.tableWidth()
    tableH := m.height - 6
    if tableH < 3 {
        tableH = 3
    }
    m.table.SetWidth(tableW)
    m.table.SetHeight(tableH)

    titleW := tableW - 4 - 10 - 5 - 4 - 8
    if titleW < 10 {
        titleW = 10
    }
    m.table.SetColumns([]table.Column{
        {Title: "#", Width: 4},
        {Title: "Title", Width: titleW},
        {Title: "Duration", Width: 10},
        {Title: "Fmt", Width: 5},
        {Title: "Ch", Width: 4},
    })
}

func (m Model) tableWidth() int {
    w := m.width * 68 / 100
    if w < 40 {
        w = 40
    }
    return w
}

func formatDuration(d interface{ Minutes() float64 }) string {
    type hasSec interface{ Seconds() float64 }
    if ds, ok := d.(hasSec); ok {
        total := int(ds.Seconds())
        m := total / 60
        s := total % 60
        return fmt.Sprintf("%d:%02d", m, s)
    }
    return "?"
}

func formatTime(seconds int) string {
    m := seconds / 60
    s := seconds % 60
    return fmt.Sprintf("%d:%02d", m, s)
}
