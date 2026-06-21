package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds all color slots used by Lip Gloss styles.
// No component file contains hardcoded hex values — everything
// references ActiveTheme.
type Theme struct {
	// Backgrounds
	Base    lipgloss.Color
	Surface lipgloss.Color
	Overlay lipgloss.Color

	// Text
	Text   lipgloss.Color
	Muted  lipgloss.Color
	Subtle lipgloss.Color

	// Semantic
	Green  lipgloss.Color
	Yellow lipgloss.Color
	Red    lipgloss.Color
	Blue   lipgloss.Color
	Purple lipgloss.Color

	// Chrome
	Accent lipgloss.Color
	Badge  lipgloss.Color
}

// ActiveTheme is the theme applied to all rendered output.
// Set this at startup based on config.defaults.theme.
var ActiveTheme = CatppuccinMocha

// CatppuccinMocha is the default theme.
var CatppuccinMocha = Theme{
	Base:    "#1e1e2e",
	Surface: "#181825",
	Overlay: "#313244",
	Text:    "#cdd6f4",
	Muted:   "#6c7086",
	Subtle:  "#45475a",
	Green:   "#a6e3a1",
	Yellow:  "#f9e2af",
	Red:     "#f38ba8",
	Blue:    "#89b4fa",
	Purple:  "#cba6f7",
	Accent:  "#89dceb",
	Badge:   "#f38ba8",
}

// CatppuccinLatte is the default light theme.
var CatppuccinLatte = Theme{
	Base:    "#eff1f5",
	Surface: "#e6e9ef",
	Overlay: "#ccd0da",
	Text:    "#4c4f69",
	Muted:   "#8c8fa1",
	Subtle:  "#bcc0cc",
	Green:   "#40a02b",
	Yellow:  "#df8e1d",
	Red:     "#d20f39",
	Blue:    "#1e66f5",
	Purple:  "#8839ef",
	Accent:  "#04a5e5",
	Badge:   "#d20f39",
}

// Dracula is a classic dark purple theme.
var Dracula = Theme{
	Base:    "#282a36",
	Surface: "#21222c",
	Overlay: "#44475a",
	Text:    "#f8f8f2",
	Muted:   "#6272a4",
	Subtle:  "#44475a",
	Green:   "#50fa7b",
	Yellow:  "#f1fa8c",
	Red:     "#ff5555",
	Blue:    "#8be9fd",
	Purple:  "#bd93f9",
	Accent:  "#ff79c6",
	Badge:   "#ff5555",
}

// ThemeRegistry maps config key strings to Theme values.
// Additional themes are added in subsequent slices.
var ThemeRegistry = map[string]Theme{
	"catppuccin-mocha": CatppuccinMocha,
	"catppuccin-latte": CatppuccinLatte,
	"dracula":          Dracula,
}

// Load sets ActiveTheme from the registry key, falling back to
// CatppuccinMocha if the key is unknown.
func Load(key string) {
	if t, ok := ThemeRegistry[key]; ok {
		ActiveTheme = t
	} else {
		ActiveTheme = CatppuccinMocha
	}
}
