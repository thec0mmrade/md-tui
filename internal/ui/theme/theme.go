package theme

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Palette defines a color theme.
type Palette struct {
	Name    string
	Accent  lipgloss.Color
	Subtle  lipgloss.Color
	Text    lipgloss.Color
	DimText lipgloss.Color
	Error   lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color
}

// Built-in palettes.
var Palettes = []Palette{
	{Name: "Default", Accent: "#FF6B35", Subtle: "#666666", Text: "#FAFAFA", DimText: "#999999", Error: "#FF4444", Success: "#44FF44", Warning: "#FFAA00"},
	{Name: "OneDark Pro", Accent: "#61AFEF", Subtle: "#5C6370", Text: "#ABB2BF", DimText: "#636D83", Error: "#E06C75", Success: "#98C379", Warning: "#E5C07B"},
	{Name: "Tokyo Night", Accent: "#7AA2F7", Subtle: "#565F89", Text: "#C0CAF5", DimText: "#737AA2", Error: "#F7768E", Success: "#9ECE6A", Warning: "#E0AF68"},
	{Name: "Catppuccin", Accent: "#CBA6F7", Subtle: "#585B70", Text: "#CDD6F4", DimText: "#7F849C", Error: "#F38BA8", Success: "#A6E3A1", Warning: "#F9E2AF"},
	{Name: "Gruvbox", Accent: "#FE8019", Subtle: "#665C54", Text: "#EBDBB2", DimText: "#928374", Error: "#FB4934", Success: "#B8BB26", Warning: "#FABD2F"},
	{Name: "Dracula", Accent: "#BD93F9", Subtle: "#6272A4", Text: "#F8F8F2", DimText: "#6272A4", Error: "#FF5555", Success: "#50FA7B", Warning: "#F1FA8C"},
	{Name: "Nord", Accent: "#88C0D0", Subtle: "#4C566A", Text: "#ECEFF4", DimText: "#7B88A1", Error: "#BF616A", Success: "#A3BE8C", Warning: "#EBCB8B"},
}

var currentPaletteIndex = 0

// LogoCacheBuster is incremented on theme change to invalidate cached renders.
var LogoCacheBuster int

// Colors
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

// Apply updates all colors and recomputes all styles from a palette.
func Apply(p Palette) {
	AccentColor = p.Accent
	SubtleColor = p.Subtle
	TextColor = p.Text
	DimTextColor = p.DimText
	ErrorColor = p.Error
	SuccessColor = p.Success
	WarningColor = p.Warning

	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor)
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor).
		Background(AccentColor).
		Padding(0, 1)
	InfoLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	InfoValueStyle = lipgloss.NewStyle().Foreground(TextColor)
	StatusBarStyle = lipgloss.NewStyle().Foreground(DimTextColor)
	KeyStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	KeyDescStyle = lipgloss.NewStyle().Foreground(DimTextColor)
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

	LogoCacheBuster++
}

// CycleTheme switches to the next (forward=true) or previous theme.
// Returns the new theme name.
func CycleTheme(forward bool) string {
	if forward {
		currentPaletteIndex = (currentPaletteIndex + 1) % len(Palettes)
	} else {
		currentPaletteIndex = (currentPaletteIndex - 1 + len(Palettes)) % len(Palettes)
	}
	Apply(Palettes[currentPaletteIndex])
	return Palettes[currentPaletteIndex].Name
}

// CurrentPaletteName returns the active theme name.
func CurrentPaletteName() string {
	return Palettes[currentPaletteIndex].Name
}

// Keybindings
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Top       key.Binding
	Bottom    key.Binding
	Upload    key.Binding
	Download  key.Binding
	Rename    key.Binding
	DiscName  key.Binding
	Delete    key.Binding
	Move      key.Binding
	Wipe      key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
	Rescan    key.Binding
	Help      key.Binding
	Quit      key.Binding
	ThemeNext key.Binding
	ThemePrev key.Binding
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
	ThemeNext: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "next theme"),
	),
	ThemePrev: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "prev theme"),
	),
}
