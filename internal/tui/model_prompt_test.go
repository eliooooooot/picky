package tui_test

import (
	"os"
	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/eliotsamuelmiller/picky/internal/tui"
	"strings"
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptMode(t *testing.T) {
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
		},
	}
	root.Children[0].Parent = root
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := tui.NewModel(tree, &ignores)
	
	// Test entering prompt mode
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*tui.Model)
	
	// Check that we're in prompt mode (can't directly test inPromptMode as it's private)
	// Instead, we'll test behavior - sending text should update the prompt
	
	// Type some text
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updatedModel.(*tui.Model)
	
	// Check that prompt contains our text
	if m.Prompt() != "Hello" {
		t.Errorf("Expected prompt to be 'Hello', got '%s'", m.Prompt())
	}
	
	// Test exiting prompt mode with ESC
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*tui.Model)
	
	// Verify prompt text is still retained
	if m.Prompt() != "Hello" {
		t.Errorf("Prompt text should be retained after exiting prompt mode")
	}
	
	// Test that 'p' key enters prompt mode again
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(*tui.Model)
	
	// Add more text
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedModel.(*tui.Model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updatedModel.(*tui.Model)
	
	if m.Prompt() != "Hello world" {
		t.Errorf("Expected prompt to be 'Hello world', got '%s'", m.Prompt())
	}
}

func TestPromptDoesNotInterfereWithNavigation(t *testing.T) {
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{Path: "/root/file2.txt", Name: "file2.txt"},
		},
	}
	root.Children[0].Parent = root
	root.Children[1].Parent = root
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := tui.NewModel(tree, &ignores)
	
	// Init the model to open root
	_ = model.Init()
	
	// Navigate down to first file
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(*tui.Model)
	
	initialCursor := m.State().CursorPath
	
	// Enter prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(*tui.Model)
	
	// Try to navigate (should not work in prompt mode)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(*tui.Model)
	
	// Cursor should not have moved
	if m.State().CursorPath != initialCursor {
		t.Error("Navigation should not work in prompt mode")
	}
	
	// Exit prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*tui.Model)
	
	// Now navigation should work - navigate down to second file
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(*tui.Model)
	
	if m.State().CursorPath == initialCursor {
		t.Errorf("Navigation should work after exiting prompt mode. Cursor was at %s, still at %s", initialCursor, m.State().CursorPath)
	}
}

func TestPromptModeFaintRendering(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
		},
	}
	root.Children[0].Parent = root
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := tui.NewModel(tree, &ignores)
	
	// Initialize the model
	_ = model.Init()
	
	// Get initial view (not in prompt mode)
	view := model.View()
	
	// Count faint sequences in the view (should only be in the collapsed prompt)
	normalFaintCount := countFaintSequences(view)
	
	// Enter prompt mode
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*tui.Model)
	
	// Get view in prompt mode
	promptView := m.View()
	
	// Count faint sequences in prompt mode (should be more due to dimmed tree/header)
	promptFaintCount := countFaintSequences(promptView)
	
	// Should have more faint sequences when in prompt mode
	if promptFaintCount <= normalFaintCount {
		t.Errorf("Expected more faint sequences in prompt mode. Normal: %d, Prompt: %d", normalFaintCount, promptFaintCount)
	}
	
	// Check that the header is dimmed
	headerLine := ""
	for _, line := range strings.Split(promptView, "\n") {
		if strings.Contains(line, "Picky - File Selector") {
			headerLine = line
			break
		}
	}
	if !strings.Contains(headerLine, "\x1b[2m") {
		t.Error("Header should be dimmed in prompt mode")
	}
	
	// Exit prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*tui.Model)
	
	// Get view after exiting prompt mode
	normalView := m.View()
	
	// Should be back to normal faint count
	finalFaintCount := countFaintSequences(normalView)
	if finalFaintCount != normalFaintCount {
		t.Errorf("Faint count should return to normal after exiting prompt mode. Expected: %d, Got: %d", normalFaintCount, finalFaintCount)
	}
}

func countFaintSequences(s string) int {
	count := 0
	for i := 0; i < len(s)-3; i++ {
		if s[i] == '\x1b' && s[i+1] == '[' && s[i+2] == '2' && s[i+3] == 'm' {
			count++
		}
	}
	return count
}

func TestPromptCollapsedViewEllipsis(t *testing.T) {
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := tui.NewModel(tree, &ignores)
	
	// Enter prompt mode
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*tui.Model)
	
	// Type a long prompt (more than 60 characters)
	longText := "This is a very long prompt that should be truncated with an ellipsis when displayed in collapsed view"
	for _, r := range longText {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updatedModel.(*tui.Model)
	}
	
	// Exit prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*tui.Model)
	
	// Get view and check for ellipsis
	view := m.View()
	if !strings.Contains(view, "…") {
		t.Error("Collapsed prompt view should contain ellipsis for long prompts")
	}
	
	// Check that the prompt itself wasn't truncated
	if m.Prompt() != longText {
		t.Error("The actual prompt value should not be truncated, only its display")
	}
}

func TestNoCursorInPromptMode(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{Path: "/root/file2.txt", Name: "file2.txt"},
		},
	}
	root.Children[0].Parent = root
	root.Children[1].Parent = root
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := tui.NewModel(tree, &ignores)
	
	// Initialize and navigate to first file
	_ = model.Init()
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(*tui.Model)
	
	// Get view (cursor should be visible)
	normalView := m.View()
	
	// Check for cursor background - lipgloss renders background colors
	// Color 237 might be rendered as 48;5;237 or other formats
	hasCursor := false
	for _, line := range strings.Split(normalView, "\n") {
		if strings.Contains(line, "file1.txt") {
			// Check for various background color formats
			// Look for any background color (40-47, 100-107, or 48;...)
			if strings.Contains(line, ";40m") || 
			   strings.Contains(line, "\x1b[40m") ||
			   strings.Contains(line, "\x1b[100m") ||
			   strings.Contains(line, "[48;5;237m") ||
			   strings.Contains(line, "48;5;237") {
				hasCursor = true
				break
			}
		}
	}
	if !hasCursor {
		t.Error("Cursor background should be visible in normal mode")
		// Debug: print lines containing file1.txt
		for _, line := range strings.Split(normalView, "\n") {
			if strings.Contains(line, "file1.txt") {
				t.Logf("file1.txt line: %q", line)
			}
		}
	}
	
	// Enter prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(*tui.Model)
	
	// Get view in prompt mode
	promptView := m.View()
	
	// Check that cursor background is NOT present
	hasCursorInPrompt := false
	for _, line := range strings.Split(promptView, "\n") {
		if strings.Contains(line, "file1.txt") {
			if strings.Contains(line, ";40m") || 
			   strings.Contains(line, "\x1b[40m") ||
			   strings.Contains(line, "\x1b[100m") ||
			   strings.Contains(line, "[48;5;237m") ||
			   strings.Contains(line, "48;5;237") {
				hasCursorInPrompt = true
				break
			}
		}
	}
	if hasCursorInPrompt {
		t.Error("Cursor background should not be visible in prompt mode")
	}
	
	// Exit prompt mode
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(*tui.Model)
	
	// Cursor should be back
	exitView := m.View()
	hasCursorAfterExit := false
	for _, line := range strings.Split(exitView, "\n") {
		if strings.Contains(line, "file1.txt") {
			if strings.Contains(line, ";40m") || 
			   strings.Contains(line, "\x1b[40m") ||
			   strings.Contains(line, "\x1b[100m") ||
			   strings.Contains(line, "[48;5;237m") ||
			   strings.Contains(line, "48;5;237") {
				hasCursorAfterExit = true
				break
			}
		}
	}
	if !hasCursorAfterExit {
		t.Error("Cursor background should be visible again after exiting prompt mode")
	}
}