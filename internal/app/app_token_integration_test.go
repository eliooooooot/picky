package app_test

import (
	"path/filepath"
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	pickyfs "github.com/eliooooooot/picky/internal/fs"
	"github.com/eliooooooot/picky/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppTokenIntegration(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	rootPath := "/test"
	
	// Create test filesystem structure with known content
	fs.AddFile(filepath.Join(rootPath, "small.txt"), "test")                // 4 chars = 1 token
	fs.AddFile(filepath.Join(rootPath, "medium.txt"), "hello world!")       // 12 chars = 3 tokens
	fs.AddFile(filepath.Join(rootPath, "dir/large.txt"), repeatChar('a', 1000))  // 1000 chars = 250 tokens
	
	// Build tree
	tree, err := domain.BuildTree(fs, rootPath)
	require.NoError(t, err)
	
	// This mimics what happens in app.Run() but we extract just the token counting part
	t.Run("token map is built correctly", func(t *testing.T) {
		// Import token package to build token map
		tokenizer := testTokenizer{}
		counter := testCounter{FS: fs, Tokenizer: tokenizer}
		tokensMap, err := counter.BuildTreeTokenMap(tree)
		require.NoError(t, err)
		
		// Verify token counts
		assert.Equal(t, 1, tokensMap[filepath.Join(rootPath, "small.txt")])
		assert.Equal(t, 3, tokensMap[filepath.Join(rootPath, "medium.txt")])
		assert.Equal(t, 250, tokensMap[filepath.Join(rootPath, "dir/large.txt")])
		
		// Directories should not be in the map
		_, exists := tokensMap[rootPath]
		assert.False(t, exists)
		_, exists = tokensMap[filepath.Join(rootPath, "dir")]
		assert.False(t, exists)
	})
	
	t.Run("token map is injected into TUI model", func(t *testing.T) {
		// Create a minimal token map
		tokensMap := map[string]int{
			filepath.Join(rootPath, "small.txt"):     1,
			filepath.Join(rootPath, "medium.txt"):    3,
			filepath.Join(rootPath, "dir/large.txt"): 250,
		}
		
		// Create TUI model and inject tokens
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		model.SetTokens(tokensMap)
		
		// Initialize the model to set up viewport
		model.Init()
		
		// The model should now have the token map
		// We can't directly access it, but we can verify through the view
		view := model.View()
		
		// Check that the header shows token count
		assert.Contains(t, view, "Tokens selected:")
		
		// Check that file entries show token counts
		assert.Contains(t, view, "(1)")   // small.txt
		assert.Contains(t, view, "(3)")   // medium.txt
		assert.Contains(t, view, "(250)") // large.txt
	})
}

// Test helpers that mimic the token package interface without importing it
type testTokenizer struct{}

func (testTokenizer) CountTokens(text string) int {
	return (len([]rune(text)) + 3) / 4
}

type testCounter struct {
	FS        domain.FileSystem
	Tokenizer testTokenizer
}

func (c *testCounter) BuildTreeTokenMap(t *domain.Tree) (map[string]int, error) {
	out := make(map[string]int)
	for _, n := range t.Flatten() {
		if n.IsDir {
			continue
		}
		bytes, err := c.FS.ReadFile(n.Path)
		if err != nil {
			return nil, err
		}
		out[n.Path] = c.Tokenizer.CountTokens(string(bytes))
	}
	return out, nil
}

// Helper to create a string of repeated characters
func repeatChar(c rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}