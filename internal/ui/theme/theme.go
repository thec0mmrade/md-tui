package theme

import (
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"
)

// Colors — retro MiniDisc aesthetic
var (
    AccentColor  = lipgloss.Color("#FF6B35")
    SubtleColor  = lipgloss.Color("#666666")
    TextColor    = lipgloss.Color("#FAFAFA")
    DimTextColor = lipgloss.Color("#999999")
    ErrorColor   = lipgloss.Color("#FF4444")
    SuccessColor = lipgloss.Color("#44FF44")
    WarningColor = lipgloss.Color("#FFAA00")
)

// Styles
var (
    PanelStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(AccentColor)

    TitleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(TextColor).
            Background(AccentColor).
            Padding(0, 1)

    InfoLabelStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(AccentColor)

    InfoValueStyle = lipgloss.NewStyle().
            Foreground(TextColor)

    StatusBarStyle = lipgloss.NewStyle().
            Foreground(DimTextColor)

    KeyStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(AccentColor)

    KeyDescStyle = lipgloss.NewStyle().
            Foreground(DimTextColor)

    ModalStyle = lipgloss.NewStyle().
            Border(lipgloss.DoubleBorder()).
            BorderForeground(AccentColor).
            Padding(1, 2)

    ErrorBannerStyle = lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("#FFFFFF")).
                Background(ErrorColor).
                Padding(0, 1)

    SuccessBannerStyle = lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("#000000")).
                Background(SuccessColor).
                Padding(0, 1)

    DimStyle = lipgloss.NewStyle().
            Foreground(SubtleColor).
            Italic(true)
)

// Keybindings
type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    Top      key.Binding
    Bottom   key.Binding
    Upload   key.Binding
    Download key.Binding
    Rename   key.Binding
    DiscName key.Binding
    Delete   key.Binding
    Move     key.Binding
    Wipe     key.Binding
    Confirm  key.Binding
    Cancel   key.Binding
    Rescan   key.Binding
    Help     key.Binding
    Quit     key.Binding
}

var Keys = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "down"),
    ),
    Top: key.NewBinding(
        key.WithKeys("home", "g"),
        key.WithHelp("g/home", "top"),
    ),
    Bottom: key.NewBinding(
        key.WithKeys("end", "G"),
        key.WithHelp("G/end", "bottom"),
    ),
    Upload: key.NewBinding(
        key.WithKeys("u"),
        key.WithHelp("u", "upload"),
    ),
    Download: key.NewBinding(
        key.WithKeys("x"),
        key.WithHelp("x", "extract"),
    ),
    Rename: key.NewBinding(
        key.WithKeys("r"),
        key.WithHelp("r", "rename"),
    ),
    DiscName: key.NewBinding(
        key.WithKeys("R"),
        key.WithHelp("R", "disc name"),
    ),
    Delete: key.NewBinding(
        key.WithKeys("d"),
        key.WithHelp("d", "delete"),
    ),
    Move: key.NewBinding(
        key.WithKeys("m"),
        key.WithHelp("m", "move"),
    ),
    Wipe: key.NewBinding(
        key.WithKeys("W"),
        key.WithHelp("W", "wipe disc"),
    ),
    Confirm: key.NewBinding(
        key.WithKeys("enter"),
        key.WithHelp("enter", "confirm"),
    ),
    Cancel: key.NewBinding(
        key.WithKeys("esc"),
        key.WithHelp("esc", "cancel"),
    ),
    Rescan: key.NewBinding(
        key.WithKeys("r"),
        key.WithHelp("r", "rescan"),
    ),
    Help: key.NewBinding(
        key.WithKeys("?"),
        key.WithHelp("?", "help"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
}
