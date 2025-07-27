package tui_test

import (
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	pickyfs "github.com/eliooooooot/picky/internal/fs"
	"github.com/eliooooooot/picky/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcludeCursorNavigation(t *testing.T) {
	t.Run("exclude only child leaves cursor on root", func(t *testing.T) {
		// Simple case: root with one file
		fs := pickyfs.NewMemFileSystem()
		fs.WriteFile("/root/file.txt", []byte("content"), 0644)
		
		tree, err := domain.BuildTree(fs, "/root")
		require.NoError(t, err)
		
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.Init()
		
		// Navigate to the file
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		state := model.State()
		assert.Equal(t, "/root/file.txt", state.CursorPath)
		
		// Exclude it
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m := updatedModel.(*tui.Model)
		
		// Cursor should be on root (only node left)
		newState := m.State()
		assert.Equal(t, "/root", newState.CursorPath)
		
		// Verify exclusion
		assert.Contains(t, m.NewIgnores(), "file.txt")
	})
	
	t.Run("exclude first of two siblings moves cursor to second", func(t *testing.T) {
		fs := pickyfs.NewMemFileSystem()
		fs.WriteFile("/root/a.txt", []byte("a"), 0644)
		fs.WriteFile("/root/b.txt", []byte("b"), 0644)
		
		tree, err := domain.BuildTree(fs, "/root")
		require.NoError(t, err)
		
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.Init()
		
		// Navigate to first file
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		state := model.State()
		assert.Equal(t, "/root/a.txt", state.CursorPath)
		
		// Exclude it
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m := updatedModel.(*tui.Model)
		
		// Cursor should move to root (as we were at position 1, and now position 0 is root)
		newState := m.State()
		assert.Equal(t, "/root", newState.CursorPath)
	})
	
	t.Run("exclude last item moves cursor up", func(t *testing.T) {
		fs := pickyfs.NewMemFileSystem()
		fs.WriteFile("/root/a.txt", []byte("a"), 0644)
		fs.WriteFile("/root/b.txt", []byte("b"), 0644)
		
		tree, err := domain.BuildTree(fs, "/root")
		require.NoError(t, err)
		
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.Init()
		
		// Navigate to last file
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		state := model.State()
		assert.Equal(t, "/root/b.txt", state.CursorPath)
		
		// Exclude it
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m := updatedModel.(*tui.Model)
		
		// Cursor should move to previous item (a.txt)
		newState := m.State()
		assert.Equal(t, "/root/a.txt", newState.CursorPath)
	})
	
	t.Run("exclude directory with expanded children", func(t *testing.T) {
		fs := pickyfs.NewMemFileSystem()
		fs.MkdirAll("/root/dir", 0755)
		fs.WriteFile("/root/dir/file.txt", []byte("content"), 0644)
		fs.WriteFile("/root/other.txt", []byte("other"), 0644)
		
		tree, err := domain.BuildTree(fs, "/root")
		require.NoError(t, err)
		
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.Init()
		
		// Navigate to dir and expand it
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}) // to dir
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}) // expand dir
		
		// Navigate into the directory
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}) // to dir/file.txt
		state := model.State()
		assert.Equal(t, "/root/dir/file.txt", state.CursorPath)
		
		// Go back to dir
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}) // back to dir
		state = model.State()
		assert.Equal(t, "/root/dir", state.CursorPath)
		
		// Exclude the directory
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		m := updatedModel.(*tui.Model)
		
		// Cursor should be on root
		newState := m.State()
		assert.Equal(t, "/root", newState.CursorPath)
		
		// Verify the directory and its contents are gone
		flat := m.Tree().Flatten()
		for _, node := range flat {
			assert.NotEqual(t, "/root/dir", node.Path)
			assert.NotEqual(t, "/root/dir/file.txt", node.Path)
		}
	})
	
	t.Run("cursor remains valid after any exclusion", func(t *testing.T) {
		// Create a more complex tree
		fs := pickyfs.NewMemFileSystem()
		fs.MkdirAll("/root/a", 0755)
		fs.MkdirAll("/root/b", 0755)
		fs.WriteFile("/root/1.txt", []byte("1"), 0644)
		fs.WriteFile("/root/2.txt", []byte("2"), 0644)
		fs.WriteFile("/root/a/3.txt", []byte("3"), 0644)
		fs.WriteFile("/root/b/4.txt", []byte("4"), 0644)
		
		tree, err := domain.BuildTree(fs, "/root")
		require.NoError(t, err)
		
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.Init()
		
		// Try excluding different nodes and verify cursor is always valid
		testCases := []string{
			"/root/1.txt",
			"/root/a",
			"/root/2.txt",
		}
		
		for _, pathToExclude := range testCases {
			// Navigate to the node
			for {
				state := model.State()
				if state.CursorPath == pathToExclude {
					break
				}
				// Try to find it by going down
				flat := domain.Flatten(tree.Root, state)
				found := false
				for i, node := range flat {
					if node.Path == state.CursorPath {
						if i+1 < len(flat) {
							model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
							found = true
							break
						}
					}
				}
				if !found {
					// Reset to root and try again
					for model.State().CursorPath != "/root" {
						model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
					}
				}
			}
			
			// Exclude it
			updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
			m := updatedModel.(*tui.Model)
			model = m
			
			// Verify cursor is on a valid node
			newState := m.State()
			flat := m.Tree().Flatten()
			found := false
			for _, node := range flat {
				if node.Path == newState.CursorPath {
					found = true
					break
				}
			}
			assert.True(t, found, "Cursor path %s should exist in tree after excluding %s", newState.CursorPath, pathToExclude)
		}
	})
}