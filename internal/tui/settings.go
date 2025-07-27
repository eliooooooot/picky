package tui

// Settings stores user preferences for the TUI
type Settings struct {
	Emoji bool
}

// defaultSettings returns Settings with sane defaults
func defaultSettings() Settings {
	return Settings{Emoji: false}
}

// ToggleEmoji returns a copy with Emoji toggled
func (s Settings) ToggleEmoji() Settings {
	s.Emoji = !s.Emoji
	return s
}