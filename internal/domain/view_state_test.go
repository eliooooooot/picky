package domain_test

import (
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestViewStatePrune(t *testing.T) {
	t.Run("prune removes entries with matching prefix", func(t *testing.T) {
		state := domain.NewViewState("/root")
		
		// Set up some state
		state = state.SetOpen("/root/dir1", true)
		state = state.SetOpen("/root/dir1/subdir", true)
		state = state.SetOpen("/root/dir2", true)
		state = state.SetSelected("/root/dir1/file.txt", true)
		state = state.SetSelected("/root/dir2/file.txt", true)
		
		// Prune dir1 and all its children
		prunedState := state.Prune("/root/dir1")
		
		// Verify dir1 entries are removed
		assert.False(t, prunedState.IsOpen("/root/dir1"))
		assert.False(t, prunedState.IsOpen("/root/dir1/subdir"))
		assert.False(t, prunedState.IsSelected("/root/dir1/file.txt"))
		
		// Verify other entries remain
		assert.True(t, prunedState.IsOpen("/root/dir2"))
		assert.True(t, prunedState.IsSelected("/root/dir2/file.txt"))
	})
	
	t.Run("prune with exact match", func(t *testing.T) {
		state := domain.NewViewState("/root")
		state = state.SetOpen("/root/file.txt", true)
		state = state.SetSelected("/root/file.txt", true)
		
		// Prune exact file
		prunedState := state.Prune("/root/file.txt")
		
		assert.False(t, prunedState.IsOpen("/root/file.txt"))
		assert.False(t, prunedState.IsSelected("/root/file.txt"))
	})
	
	t.Run("prune preserves non-matching paths", func(t *testing.T) {
		state := domain.NewViewState("/root")
		state = state.SetOpen("/root/dir1", true)
		state = state.SetOpen("/root/dir2", true)
		
		// Prune a non-existent path
		prunedState := state.Prune("/root/dir3")
		
		// Everything should remain
		assert.True(t, prunedState.IsOpen("/root/dir1"))
		assert.True(t, prunedState.IsOpen("/root/dir2"))
	})
	
	t.Run("prune is immutable", func(t *testing.T) {
		state := domain.NewViewState("/root")
		state = state.SetOpen("/root/dir1", true)
		
		// Prune creates new state
		prunedState := state.Prune("/root/dir1")
		
		// Original state unchanged
		assert.True(t, state.IsOpen("/root/dir1"))
		assert.False(t, prunedState.IsOpen("/root/dir1"))
	})
}