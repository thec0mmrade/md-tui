package trackedit

import (
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type Mode int

const (
    ModeRenameTrack Mode = iota
    ModeRenameDisc
)

type DoneMsg struct {
    Mode     Mode
    Index    int
    NewTitle string
}

type CancelMsg struct{}

type Model struct {
    mode      Mode
    index     int
    input     textinput.Model
    title     string // header label
    active    bool
    width     int
}

func New() Model {
    ti := textinput.New()
    ti.CharLimit = 120
    ti.Width = 40
    return Model{input: ti}
}

func (m *Model) Open(mode Mode, index int, currentTitle string, width int) {
    m.mode = mode
    m.index = index
    m.active = true
    m.width = width
    m.input.SetValue(currentTitle)
    m.input.Focus()
    m.input.CursorEnd()

    switch mode {
    case ModeRenameTrack:
        m.title = "Rename Track"
    case ModeRenameDisc:
        m.title = "Rename Disc"
    }
}

func (m *Model) IsActive() bool {
    return m.active
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if !m.active {
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, theme.Keys.Confirm):
            m.active = false
            return m, func() tea.Msg {
                return DoneMsg{
                    Mode:     m.mode,
                    Index:    m.index,
                    NewTitle: m.input.Value(),
                }
            }
        case key.Matches(msg, theme.Keys.Cancel):
            m.active = false
            return m, func() tea.Msg { return CancelMsg{} }
        }
    }

    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if !m.active {
        return ""
    }

    modalWidth := 50
    if m.width > 0 && m.width < 60 {
        modalWidth = m.width - 10
    }

    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.AccentColor).
        Render(m.title)

    inputView := m.input.View()

    hints := theme.KeyDescStyle.Render("Enter") +
        theme.DimStyle.Render(": save  ") +
        theme.KeyDescStyle.Render("Esc") +
        theme.DimStyle.Render(": cancel")

    content := lipgloss.JoinVertical(lipgloss.Left,
        header,
        "",
        inputView,
        "",
        hints,
    )

    return theme.ModalStyle.Width(modalWidth).Render(content)
}
