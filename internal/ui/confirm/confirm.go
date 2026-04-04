package confirm

import (
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type ResultMsg struct {
    Confirmed bool
    Action    string
}

type Model struct {
    prompt string
    action string // passed through in result so caller knows what was confirmed
    active bool
    width  int
}

func New() Model {
    return Model{}
}

func (m *Model) Open(prompt, action string, width int) {
    m.prompt = prompt
    m.action = action
    m.active = true
    m.width = width
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
        case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
            m.active = false
            action := m.action
            return m, func() tea.Msg { return ResultMsg{Confirmed: true, Action: action} }
        case key.Matches(msg, key.NewBinding(key.WithKeys("n"))),
            key.Matches(msg, theme.Keys.Cancel):
            m.active = false
            action := m.action
            return m, func() tea.Msg { return ResultMsg{Confirmed: false, Action: action} }
        }
    }

    return m, nil
}

func (m Model) View() string {
    if !m.active {
        return ""
    }

    modalWidth := 40
    if m.width > 0 && m.width < 50 {
        modalWidth = m.width - 10
    }

    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.AccentColor).
        Render("Confirm")

    promptText := lipgloss.NewStyle().Foreground(theme.TextColor).Render(m.prompt)

    hints := theme.KeyStyle.Render("y") +
        theme.KeyDescStyle.Render("es  ") +
        theme.KeyStyle.Render("n") +
        theme.KeyDescStyle.Render("o")

    content := lipgloss.JoinVertical(lipgloss.Left,
        header,
        "",
        promptText,
        "",
        hints,
    )

    return theme.ModalStyle.Width(modalWidth).Render(content)
}
