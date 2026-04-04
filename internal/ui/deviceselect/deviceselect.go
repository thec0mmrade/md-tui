package deviceselect

import (
    "fmt"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type SelectMsg struct {
    Index int
}

type Model struct {
    devices  []device.DeviceInfo
    cursor   int
    scanning bool
    noDevice bool
    width    int
}

func New(width int) Model {
    return Model{width: width, scanning: true}
}

func (m *Model) SetDevices(devices []device.DeviceInfo) {
    m.devices = devices
    m.scanning = false
    m.noDevice = len(devices) == 0
    m.cursor = 0
}

func (m *Model) SetScanning() {
    m.scanning = true
    m.noDevice = false
    m.devices = nil
}

func (m *Model) SetWidth(w int) {
    m.width = w
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if m.scanning || m.noDevice {
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, theme.Keys.Up):
            if m.cursor > 0 {
                m.cursor--
            }
        case key.Matches(msg, theme.Keys.Down):
            if m.cursor < len(m.devices)-1 {
                m.cursor++
            }
        case key.Matches(msg, theme.Keys.Confirm):
            idx := m.devices[m.cursor].Index
            return m, func() tea.Msg { return SelectMsg{Index: idx} }
        }
    }

    return m, nil
}

func (m Model) View() string {
    if m.scanning {
        return "" // spinner handled by parent
    }

    if m.noDevice {
        return lipgloss.NewStyle().Foreground(theme.WarningColor).Render(
            "  No NetMD device found.\n\n") +
            theme.KeyDescStyle.Render("  Connect a device and press ") +
            theme.KeyStyle.Render("r") +
            theme.KeyDescStyle.Render(" to rescan.")
    }

    if len(m.devices) == 1 {
        return lipgloss.NewStyle().Foreground(theme.TextColor).Render(
            fmt.Sprintf("  Found: %s — connecting...", m.devices[0].Name))
    }

    // Multiple devices — show selection list
    s := lipgloss.NewStyle().Foreground(theme.TextColor).Render("  Select a device:\n\n")
    for i, d := range m.devices {
        cursor := "  "
        style := theme.KeyDescStyle
        if i == m.cursor {
            cursor = theme.KeyStyle.Render("> ")
            style = lipgloss.NewStyle().Foreground(theme.TextColor).Bold(true)
        }
        s += fmt.Sprintf("  %s%s\n", cursor, style.Render(d.Name))
    }
    s += "\n" + theme.KeyDescStyle.Render("  ↑/↓ to navigate, Enter to connect")
    return s
}
