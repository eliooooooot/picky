package tui

import "github.com/charmbracelet/lipgloss"

// ColorScheme defines the three base colors for the tree UI
type ColorScheme struct {
	Name               string
	Selected           lipgloss.Color
	PartiallySelected  lipgloss.Color
	Unselected         lipgloss.Color
}

// Settings stores user preferences for the TUI
type Settings struct {
	Emoji       bool
	ColorScheme ColorScheme
}

// defaultSettings returns Settings with sane defaults
func defaultSettings() Settings {
	return Settings{
		Emoji:       false,
		ColorScheme: colorSchemes[0], // Default to first scheme
	}
}

// ToggleEmoji returns a copy with Emoji toggled
func (s Settings) ToggleEmoji() Settings {
	s.Emoji = !s.Emoji
	return s
}

// NextColorScheme cycles to the next color scheme
func (s Settings) NextColorScheme() Settings {
	for i, scheme := range colorSchemes {
		if scheme.Name == s.ColorScheme.Name {
			s.ColorScheme = colorSchemes[(i+1)%len(colorSchemes)]
			break
		}
	}
	return s
}

// PrevColorScheme cycles to the previous color scheme
func (s Settings) PrevColorScheme() Settings {
	for i, scheme := range colorSchemes {
		if scheme.Name == s.ColorScheme.Name {
			idx := i - 1
			if idx < 0 {
				idx = len(colorSchemes) - 1
			}
			s.ColorScheme = colorSchemes[idx]
			break
		}
	}
	return s
}

// Available color schemes
var colorSchemes = []ColorScheme{
	{
		Name:              "Ocean",
		Selected:          lipgloss.Color("39"),  // Bright blue
		PartiallySelected: lipgloss.Color("45"),  // Cyan
		Unselected:        lipgloss.Color("243"), // Gray
	},
	{
		Name:              "Classic",
		Selected:          lipgloss.Color("237"), // Dark gray (original cursor bg)
		PartiallySelected: lipgloss.Color("240"), // Medium gray
		Unselected:        lipgloss.Color("245"), // Light gray
	},
	{
		Name:              "Forest",
		Selected:          lipgloss.Color("64"),  // Medium sage green
		PartiallySelected: lipgloss.Color("107"), // Soft green
		Unselected:        lipgloss.Color("243"), // Medium gray
	},
	{
		Name:              "Sunset",
		Selected:          lipgloss.Color("202"), // Orange
		PartiallySelected: lipgloss.Color("214"), // Light orange
		Unselected:        lipgloss.Color("244"), // Medium gray
	},
	{
		Name:              "Monochrome",
		Selected:          lipgloss.Color("255"), // White
		PartiallySelected: lipgloss.Color("250"), // Light gray
		Unselected:        lipgloss.Color("240"), // Dark gray
	},
	{
		Name:              "Neon",
		Selected:          lipgloss.Color("201"), // Bright magenta
		PartiallySelected: lipgloss.Color("99"),  // Purple
		Unselected:        lipgloss.Color("242"), // Gray
	},
}