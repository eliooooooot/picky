package tui

import (
	"testing"

	"github.com/eliotsamuelmiller/picky/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSettingsModalToggle(t *testing.T) {
	// Create a small test tree
	root := &domain.Node{
		Path:  "/test",
		Name:  "test",
		IsDir: true,
	}
	
	file1 := &domain.Node{
		Path:   "/test/file1.txt",
		Name:   "file1.txt",
		IsDir:  false,
		Parent: root,
	}
	
	dir1 := &domain.Node{
		Path:   "/test/dir1",
		Name:   "dir1",
		IsDir:  true,
		Parent: root,
	}
	
	root.Children = []*domain.Node{file1, dir1}
	tree := &domain.Tree{Root: root}
	
	// Create model
	ignores := make(map[string]struct{})
	model := NewModel(tree, &ignores)
	
	// Test opening settings modal
	t.Run("open settings modal", func(t *testing.T) {
		assert.False(t, model.isSettingsOpen)
		
		// Send 's' key
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m := updatedModel.(*Model)
		
		assert.True(t, m.isSettingsOpen)
		assert.Equal(t, 0, m.settingsCursorIdx)
	})
	
	// Test toggling emoji setting
	t.Run("toggle emoji setting", func(t *testing.T) {
		model.isSettingsOpen = true
		model.settingsCursorIdx = 0
		assert.False(t, model.settings.Emoji)
		
		// Send space to toggle
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		m := updatedModel.(*Model)
		
		assert.True(t, m.settings.Emoji)
	})
	
	// Test closing modal and emoji rendering
	t.Run("close modal and check emoji rendering", func(t *testing.T) {
		model.isSettingsOpen = true
		model.settings.Emoji = true
		
		// Send esc to close
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m := updatedModel.(*Model)
		
		assert.False(t, m.isSettingsOpen)
		
		// Check that formatNodeLabel now includes emojis
		fileLabel := m.formatNodeLabel(file1)
		dirLabel := m.formatNodeLabel(dir1)
		
		assert.Contains(t, fileLabel, "📄")
		assert.Contains(t, dirLabel, "📁")
	})
	
	// Test that keys are ignored while modal is open
	t.Run("keys ignored while modal open", func(t *testing.T) {
		model.isSettingsOpen = true
		originalCursor := model.state.CursorPath
		
		// Try to navigate (should be ignored)
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		m := updatedModel.(*Model)
		
		// Cursor should not have moved
		assert.Equal(t, originalCursor, m.state.CursorPath)
	})
	
	// Test navigation wrapping in settings
	t.Run("settings navigation wraps", func(t *testing.T) {
		model.isSettingsOpen = true
		model.settingsCursorIdx = 0
		
		// Test up wraps to bottom (which is still 0 since we only have 1 item)
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
		m := updatedModel.(*Model)
		assert.Equal(t, 0, m.settingsCursorIdx)
		
		// Test down wraps to top (which is still 0 since we only have 1 item)
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(*Model)
		assert.Equal(t, 0, m.settingsCursorIdx)
	})
}

func TestEmojiRendering(t *testing.T) {
	// Create test nodes
	root := &domain.Node{
		Path:  "/test",
		Name:  "test",
		IsDir: true,
	}
	
	file := &domain.Node{
		Path:   "/test/file.txt",
		Name:   "file.txt",
		IsDir:  false,
		Parent: root,
	}
	
	dir := &domain.Node{
		Path:   "/test/subdir",
		Name:   "subdir",
		IsDir:  true,
		Parent: root,
	}
	
	root.Children = []*domain.Node{file, dir}
	tree := &domain.Tree{Root: root}
	
	t.Run("emoji disabled by default", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		
		fileLabel := model.formatNodeLabel(file)
		dirLabel := model.formatNodeLabel(dir)
		
		assert.NotContains(t, fileLabel, "📄")
		assert.NotContains(t, dirLabel, "📁")
	})
	
	t.Run("emoji enabled shows icons", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.settings = Settings{Emoji: true}
		
		// Open the directory to test arrow + emoji
		model.state = model.state.SetOpen(dir.Path, true)
		
		fileLabel := model.formatNodeLabel(file)
		dirLabel := model.formatNodeLabel(dir)
		
		// Check emojis are present
		assert.Contains(t, fileLabel, "📄")
		assert.Contains(t, dirLabel, "📂")  // Open folder since we opened it above
		
		// Check that emojis replace arrows for directories
		assert.Contains(t, dirLabel, "📂 ")  // Open folder emoji
		assert.NotContains(t, dirLabel, "▼")  // No arrow when emoji mode is on
		
		// Test with closed directory
		model.state = model.state.SetOpen(dir.Path, false)
		dirLabel = model.formatNodeLabel(dir)
		assert.Contains(t, dirLabel, "📁 ")  // Closed folder emoji
		assert.NotContains(t, dirLabel, "▶")  // No arrow when emoji mode is on
	})
	
	t.Run("view contains emoji setting in modal", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.isSettingsOpen = true
		model.settings.Emoji = true
		
		view := model.View()
		
		// Check that settings modal is shown
		assert.Contains(t, view, "Settings")
		assert.Contains(t, view, "[x] Emoji icons")
		assert.Contains(t, view, "space/enter toggle")
		
		// Toggle off and check
		model.settings.Emoji = false
		view = model.View()
		assert.Contains(t, view, "[ ] Emoji icons")
	})
}

func TestSettingsModalKeyHandling(t *testing.T) {
	// Simple tree for testing
	root := &domain.Node{Path: "/", Name: "/", IsDir: true}
	tree := &domain.Tree{Root: root}
	ignores := make(map[string]struct{})
	
	tests := []struct {
		name         string
		key          string
		expectOpen   bool
		expectQuit   bool
		expectToggle bool
	}{
		{
			name:       "enter toggles setting",
			key:        "enter",
			expectOpen: true,
			expectToggle: true,
		},
		{
			name:       "s closes modal",
			key:        "s",
			expectOpen: false,
		},
		{
			name:       "q quits app even with modal open",
			key:        "q",
			expectOpen: true,
			expectQuit: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(tree, &ignores)
			model.isSettingsOpen = true
			originalEmoji := model.settings.Emoji
			
			updatedModel, cmd := model.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{rune(tt.key[0])},
			})
			
			if tt.key == "enter" {
				updatedModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			}
			
			m := updatedModel.(*Model)
			
			assert.Equal(t, tt.expectOpen, m.isSettingsOpen)
			
			if tt.expectQuit {
				assert.NotNil(t, cmd)
			}
			
			if tt.expectToggle {
				assert.NotEqual(t, originalEmoji, m.settings.Emoji)
			}
		})
	}
}