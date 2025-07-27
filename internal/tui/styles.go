package tui

import "github.com/charmbracelet/lipgloss"

// Color scheme using ANSI 256 color codes for better compatibility
var (
	// Selection colors
	FullySelectedColor    = lipgloss.Color("40")  // Bright green for fully selected
	PartiallySelectedColor = lipgloss.Color("220") // Amber/yellow for partially selected
	NotSelectedColor      = lipgloss.Color("245")  // Light gray for not selected
	
	// Cursor/focus colors
	CursorBgColor        = lipgloss.Color("237")  // Dark gray background
	CursorSelectedBg     = lipgloss.Color("22")   // Dark green background for selected+cursor
	CursorPartialBg      = lipgloss.Color("94")   // Dark amber background for partial+cursor
	
	// Text colors
	FileNameColor        = lipgloss.Color("252")  // Very light gray
	DirectoryNameColor   = lipgloss.Color("33")   // Blue for directories
	CursorTextColor      = lipgloss.Color("255")  // White for cursor text
	
	// UI element colors
	HeaderColor          = lipgloss.Color("39")   // Bright blue
	HelpTextColor        = lipgloss.Color("241")  // Gray
	StatusMessageColor   = lipgloss.Color("214")  // Orange for status
	
	// Indicator colors
	CheckmarkColor       = lipgloss.Color("40")   // Green checkmark
	PartialCheckColor    = lipgloss.Color("220")  // Yellow dash
	DirectoryArrowColor  = lipgloss.Color("245")  // Gray arrows
)

// Styles
var (
	// Base styles
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(HeaderColor)
	
	HelpStyle = lipgloss.NewStyle().
		Foreground(HelpTextColor)
	
	StatusStyle = lipgloss.NewStyle().
		Foreground(StatusMessageColor).
		Bold(true)
	
	// File/Directory name styles
	FileStyle = lipgloss.NewStyle().
		Foreground(FileNameColor)
	
	DirectoryStyle = lipgloss.NewStyle().
		Foreground(DirectoryNameColor).
		Bold(true)
	
	// Selection indicator styles
	SelectedIndicatorStyle = lipgloss.NewStyle().
		Foreground(CheckmarkColor).
		Bold(true)
	
	PartialIndicatorStyle = lipgloss.NewStyle().
		Foreground(PartialCheckColor).
		Bold(true)
	
	// Cursor styles (without selection)
	CursorStyle = lipgloss.NewStyle().
		Background(CursorBgColor).
		Foreground(CursorTextColor)
	
	CursorFileStyle = CursorStyle.Copy()
	
	CursorDirectoryStyle = CursorStyle.Copy().
		Foreground(DirectoryNameColor).
		Bold(true)
	
	// Cursor + selection combination styles
	CursorSelectedStyle = lipgloss.NewStyle().
		Background(CursorSelectedBg).
		Foreground(CursorTextColor).
		Bold(true)
	
	CursorPartialStyle = lipgloss.NewStyle().
		Background(CursorPartialBg).
		Foreground(CursorTextColor)
	
	// Selected item styles (without cursor)
	SelectedFileStyle = lipgloss.NewStyle().
		Foreground(FullySelectedColor)
	
	SelectedDirectoryStyle = lipgloss.NewStyle().
		Foreground(FullySelectedColor).
		Bold(true)
	
	PartialDirectoryStyle = lipgloss.NewStyle().
		Foreground(PartiallySelectedColor).
		Bold(true)
	
	// Directory arrow style
	ArrowStyle = lipgloss.NewStyle().
		Foreground(DirectoryArrowColor)
	
	// Prompt mode styles
	DimmedStyle     = lipgloss.NewStyle().Faint(true)
	PromptBorderDim = lipgloss.Color("240")
	PromptBorderLit = lipgloss.Color("33") // cyan, matches header
)