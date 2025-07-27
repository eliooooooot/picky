package domain_test

import (
	"testing"

	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNextCursorAfterRemoval(t *testing.T) {
	// Helper to create test nodes
	createNode := func(path string) *domain.Node {
		return &domain.Node{Path: path}
	}
	
	t.Run("remove last item", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/b.txt"),
			createNode("/root/c.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/b.txt"),
		}
		
		// Removed c.txt at index 3
		next := domain.NextCursorAfterRemoval(flatBefore, 3, flatAfter)
		assert.Equal(t, "/root/b.txt", next, "Should move to new last item")
	})
	
	t.Run("remove first item after root", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/b.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root"),
			createNode("/root/b.txt"),
		}
		
		// Removed a.txt at index 1
		next := domain.NextCursorAfterRemoval(flatBefore, 1, flatAfter)
		assert.Equal(t, "/root", next, "Should move to root (previous item)")
	})
	
	t.Run("remove middle item", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/b.txt"),
			createNode("/root/c.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/c.txt"),
		}
		
		// Removed b.txt at index 2
		next := domain.NextCursorAfterRemoval(flatBefore, 2, flatAfter)
		assert.Equal(t, "/root/a.txt", next, "Should move to previous sibling")
	})
	
	t.Run("remove only child", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/only.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root"),
		}
		
		// Removed only.txt at index 1
		next := domain.NextCursorAfterRemoval(flatBefore, 1, flatAfter)
		assert.Equal(t, "/root", next, "Should move to root")
	})
	
	t.Run("remove from beginning", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/a.txt"),
			createNode("/root/b.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root/b.txt"),
		}
		
		// Removed root and a.txt (unusual case)
		next := domain.NextCursorAfterRemoval(flatBefore, 0, flatAfter)
		assert.Equal(t, "/root/b.txt", next, "Should move to first remaining item")
	})
	
	t.Run("empty after list", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
		}
		
		flatAfter := []*domain.Node{}
		
		// Everything removed (shouldn't happen in practice)
		next := domain.NextCursorAfterRemoval(flatBefore, 0, flatAfter)
		assert.Equal(t, "", next, "Should return empty string for empty list")
	})
	
	t.Run("remove expanded directory", func(t *testing.T) {
		flatBefore := []*domain.Node{
			createNode("/root"),
			createNode("/root/dir"),
			createNode("/root/dir/file1.txt"),
			createNode("/root/dir/file2.txt"),
			createNode("/root/other.txt"),
		}
		
		flatAfter := []*domain.Node{
			createNode("/root"),
			createNode("/root/other.txt"),
		}
		
		// Removed dir at index 1 (and its children)
		next := domain.NextCursorAfterRemoval(flatBefore, 1, flatAfter)
		assert.Equal(t, "/root", next, "Should move to parent")
	})
}