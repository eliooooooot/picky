package tui

import (
	"fmt"
	"testing"

	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestModel_TokenFunctionality(t *testing.T) {
	// Create a test tree structure
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	
	file1 := &domain.Node{
		Path:   "/root/file1.txt",
		Name:   "file1.txt",
		IsDir:  false,
		Parent: root,
	}
	
	dir1 := &domain.Node{
		Path:   "/root/dir1",
		Name:   "dir1",
		IsDir:  true,
		Parent: root,
	}
	
	file2 := &domain.Node{
		Path:   "/root/dir1/file2.txt",
		Name:   "file2.txt",
		IsDir:  false,
		Parent: dir1,
	}
	
	file3 := &domain.Node{
		Path:   "/root/dir1/file3.txt",
		Name:   "file3.txt",
		IsDir:  false,
		Parent: dir1,
	}
	
	root.Children = []*domain.Node{file1, dir1}
	dir1.Children = []*domain.Node{file2, file3}
	
	tree := &domain.Tree{Root: root}
	
	// Create token map
	tokens := map[string]int{
		"/root/file1.txt":      100,
		"/root/dir1/file2.txt": 200,
		"/root/dir1/file3.txt": 300,
	}
	
	t.Run("tokenCount returns correct values", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.SetTokens(tokens)
		
		// Test file token counts
		assert.Equal(t, 100, model.tokenCount(file1))
		assert.Equal(t, 200, model.tokenCount(file2))
		assert.Equal(t, 300, model.tokenCount(file3))
		
		// Test directory aggregation
		assert.Equal(t, 500, model.tokenCount(dir1)) // 200 + 300
		assert.Equal(t, 600, model.tokenCount(root)) // 100 + 200 + 300
	})
	
	t.Run("selectedTokens counts only selected files", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.SetTokens(tokens)
		
		// Initially nothing selected
		assert.Equal(t, 0, model.selectedTokens())
		
		// Select file1
		model.state = model.state.SetSelected("/root/file1.txt", true)
		assert.Equal(t, 100, model.selectedTokens())
		
		// Select file2 as well
		model.state = model.state.SetSelected("/root/dir1/file2.txt", true)
		assert.Equal(t, 300, model.selectedTokens()) // 100 + 200
		
		// Select all files in dir1 (by selecting the directory)
		model.state = model.state.SetSelected("/root/dir1/file3.txt", true)
		assert.Equal(t, 600, model.selectedTokens()) // 100 + 200 + 300
	})
	
	t.Run("formatTokenCount handles large numbers", func(t *testing.T) {
		assert.Equal(t, "999", formatTokenCount(999))
		assert.Equal(t, "1.0k", formatTokenCount(1000))
		assert.Equal(t, "1.5k", formatTokenCount(1500))
		assert.Equal(t, "999.9k", formatTokenCount(999900))
		assert.Equal(t, "1.0M", formatTokenCount(1000000))
		assert.Equal(t, "2.5M", formatTokenCount(2500000))
	})
	
	t.Run("view displays token counts correctly", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.SetTokens(tokens)
		
		// Open the directory
		model.state = model.state.SetOpen("/root/dir1", true)
		
		// Get the view
		view := model.View()
		
		// Check header shows selected tokens
		assert.Contains(t, view, "Tokens selected: ~0")
		
		// Check individual file token counts are shown
		assert.Contains(t, view, "file1.txt (100)")
		assert.Contains(t, view, "file2.txt (200)")
		assert.Contains(t, view, "file3.txt (300)")
		
		// Check directory shows aggregated count
		assert.Contains(t, view, "dir1 (500)")
		
		// Select some files and check the header updates
		model.state = model.state.SetSelected("/root/file1.txt", true)
		model.state = model.state.SetSelected("/root/dir1/file2.txt", true)
		view = model.View()
		assert.Contains(t, view, "Tokens selected: ~300")
	})
	
	t.Run("excluded nodes don't affect token counts", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		model.SetTokens(tokens)
		
		// Initially dir1 has 500 tokens
		assert.Equal(t, 500, model.tokenCount(dir1))
		
		// Exclude file2
		tree.ExcludeNode("/root/dir1/file2.txt")
		
		// dir1 should now only count file3
		assert.Equal(t, 300, model.tokenCount(dir1))
		
		// Root should also be updated
		assert.Equal(t, 400, model.tokenCount(root)) // 100 + 300
	})
	
	t.Run("handles nil token map gracefully", func(t *testing.T) {
		ignores := make(map[string]struct{})
		model := NewModel(tree, &ignores)
		// Don't set tokens
		
		// Should return 0 for all queries
		assert.Equal(t, 0, model.tokenCount(file1))
		assert.Equal(t, 0, model.tokenCount(dir1))
		assert.Equal(t, 0, model.selectedTokens())
		
		// View should still work
		view := model.View()
		assert.Contains(t, view, "(0)") // All files show 0 tokens
	})
}

func TestModel_TokensWithLargeNumbers(t *testing.T) {
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	
	bigFile := &domain.Node{
		Path:   "/root/big.txt",
		Name:   "big.txt",
		IsDir:  false,
		Parent: root,
	}
	
	hugeFile := &domain.Node{
		Path:   "/root/huge.txt",
		Name:   "huge.txt",
		IsDir:  false,
		Parent: root,
	}
	
	root.Children = []*domain.Node{bigFile, hugeFile}
	tree := &domain.Tree{Root: root}
	
	// Create token map with large numbers
	tokens := map[string]int{
		"/root/big.txt":  12345,
		"/root/huge.txt": 1234567,
	}
	
	ignores := make(map[string]struct{})
	model := NewModel(tree, &ignores)
	model.SetTokens(tokens)
	
	view := model.View()
	
	// Check formatting
	assert.Contains(t, view, "big.txt (12.3k)")
	assert.Contains(t, view, "huge.txt (1.2M)")
	
	// Select all and check header
	model.state = model.state.SetSelected("/root/big.txt", true)
	model.state = model.state.SetSelected("/root/huge.txt", true)
	view = model.View()
	
	// Total should be formatted correctly
	totalTokens := 12345 + 1234567 // = 1246912
	expectedHeader := fmt.Sprintf("Tokens selected: ~%.1fM", float64(totalTokens)/1000000)
	assert.Contains(t, view, expectedHeader)
}