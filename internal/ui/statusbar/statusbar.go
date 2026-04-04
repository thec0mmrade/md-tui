package statusbar

import (
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/lipgloss"
)

type Model struct {
    DeviceName string
    Connected  bool
    Message    string
    Width      int
}

func New(width int) Model {
    return Model{Width: width}
}

func (m Model) View() string {
    left := ""
    if m.Connected {
        left = lipgloss.NewStyle().Foreground(theme.AccentColor).Render("● ") +
            theme.InfoValueStyle.Render(m.DeviceName)
    } else {
        left = lipgloss.NewStyle().Foreground(theme.SubtleColor).Render("○ ") +
            theme.DimStyle.Render("disconnected")
    }

    right := ""
    if m.Message != "" {
        right = theme.DimStyle.Render(m.Message)
    }

    gap := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
    if gap < 1 {
        gap = 1
    }

    padding := lipgloss.NewStyle().Width(gap).Render("")
    return lipgloss.JoinHorizontal(lipgloss.Top, left, padding, right)
}
