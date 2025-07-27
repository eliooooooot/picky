package tui_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eliooooooot/picky/internal/domain"
	"github.com/eliooooooot/picky/internal/fs"
	"github.com/eliooooooot/picky/internal/tui"
	"github.com/stretchr/testify/require"
)

func TestViewportScrollsWhenCursorMovesAboveTop(t *testing.T) {
	mem := fs.NewMemFileSystem()
	// 30 files => plenty to scroll
	for i := 0; i < 30; i++ {
		path := fmt.Sprintf("/root/f%02d.txt", i)
		require.NoError(t, mem.WriteFile(path, []byte("x"), 0644))
	}
	tree, _ := domain.BuildTree(mem, "/root")

	ign := make(map[string]struct{})
	m := tui.NewModel(tree, &ign)
	m.Init() // Initialize the model first
	
	// fake window size: height = 8 rows of tree (after headers etc.)
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 14})
	mod := model.(*tui.Model)
	
	// Move down until we can scroll
	moveCount := 0
	for moveCount < 20 {  // safety limit
		model, _ := mod.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		mod = model.(*tui.Model)
		moveCount++
		if mod.VP().YOffset > 0 {
			break
		}
	}
	
	if mod.VP().YOffset == 0 {
		t.Fatal("Failed to scroll the viewport after moving down")
	}
	
	// Now position cursor at top of visible area
	// Move up until we're at the first visible line
	for {
		currentOffset := mod.VP().YOffset
		flat := domain.Flatten(mod.Tree().Root, mod.State())
		cursor := domain.FindNodeByPath(mod.Tree().Root, mod.State().CursorPath)
		var cursorIdx int
		for i, n := range flat {
			if n == cursor {
				cursorIdx = i
				break
			}
		}
		
		// If cursor is at the top visible line, we're ready
		if cursorIdx == currentOffset {
			break
		}
		
		// Move up
		model, _ := mod.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		mod = model.(*tui.Model)
	}

	startOff := mod.VP().YOffset
	
	// Now the actual test: move up one more time
	model, _ = mod.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	mod = model.(*tui.Model)

	if got := mod.VP().YOffset; got != startOff-1 {
		t.Fatalf("viewport should have scrolled up by 1 (from %d to %d), got %d",
			startOff, startOff-1, got)
	}
}