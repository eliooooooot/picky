package app_test

import (
	"path/filepath"
	"testing"

	"github.com/eliotsamuelmiller/picky/internal/domain"
	pickyfs "github.com/eliotsamuelmiller/picky/internal/fs"
	"github.com/eliotsamuelmiller/picky/internal/ignore"
	"github.com/eliotsamuelmiller/picky/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppIgnoreIntegration(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	rootPath := "/test"
	
	// Create test filesystem structure
	require.NoError(t, fs.MkdirAll(filepath.Join(rootPath, "dir1"), 0755))
	require.NoError(t, fs.MkdirAll(filepath.Join(rootPath, "dir2"), 0755))
	require.NoError(t, fs.WriteFile(filepath.Join(rootPath, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(filepath.Join(rootPath, "dir1", "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile(filepath.Join(rootPath, "dir2", "file3.txt"), []byte("content3"), 0644))
	
	// First run - exclude dir1
	t.Run("first run excludes dir1", func(t *testing.T) {
		// Build tree without filter
		tree, err := domain.BuildTree(fs, rootPath)
		require.NoError(t, err)
		
		// Simulate TUI exclusion
		ignores := make(map[string]struct{})
		model := tui.NewModel(tree, &ignores)
		
		// Exclude dir1
		relPath, removedNode := tree.ExcludeNode(filepath.Join(rootPath, "dir1"))
		assert.NotNil(t, removedNode)
		assert.Equal(t, "dir1", relPath)
		
		// Add to model's new ignores (simulating 'x' key press)
		model.NewIgnores()[relPath] = struct{}{}
		
		// Merge new ignores
		for k := range model.NewIgnores() {
			ignores[k] = struct{}{}
		}
		
		// Save ignores
		err = ignore.Save(fs, rootPath, ignores)
		require.NoError(t, err)
		
		// Verify .pickyignore was created
		data, err := fs.ReadFile(filepath.Join(rootPath, ".pickyignore"))
		require.NoError(t, err)
		assert.Contains(t, string(data), "dir1")
	})
	
	// Second run - verify dir1 is hidden
	t.Run("second run hides dir1", func(t *testing.T) {
		// Load ignores
		ignores, err := ignore.Load(fs, rootPath)
		require.NoError(t, err)
		assert.Contains(t, ignores, "dir1")
		
		// Build tree with filter
		keep := func(p string, isDir bool) bool {
			rel, err := filepath.Rel(rootPath, p)
			if err != nil {
				return true
			}
			_, skip := ignores[rel]
			return !skip
		}
		
		tree, err := domain.BuildTreeWithFilter(fs, rootPath, keep)
		require.NoError(t, err)
		
		// Verify dir1 is not in tree
		nodes := tree.Flatten()
		paths := make([]string, len(nodes))
		for i, node := range nodes {
			paths[i] = node.Path
		}
		
		assert.Contains(t, paths, rootPath)
		assert.Contains(t, paths, filepath.Join(rootPath, "file1.txt"))
		assert.Contains(t, paths, filepath.Join(rootPath, "dir2"))
		assert.Contains(t, paths, filepath.Join(rootPath, "dir2", "file3.txt"))
		assert.NotContains(t, paths, filepath.Join(rootPath, "dir1"))
		assert.NotContains(t, paths, filepath.Join(rootPath, "dir1", "file2.txt"))
	})
}

func TestAppIgnoreEmptyCase(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	rootPath := "/test"
	
	require.NoError(t, fs.MkdirAll(rootPath, 0755))
	require.NoError(t, fs.WriteFile(filepath.Join(rootPath, "file1.txt"), []byte("content"), 0644))
	
	// Since we can't run the full TUI, we simulate the flow
	ignores, err := ignore.Load(fs, rootPath)
	require.NoError(t, err)
	assert.Empty(t, ignores)
	
	// Build tree with empty filter
	keep := func(p string, isDir bool) bool {
		rel, err := filepath.Rel(rootPath, p)
		if err != nil {
			return true
		}
		_, skip := ignores[rel]
		return !skip
	}
	
	_, err = domain.BuildTreeWithFilter(fs, rootPath, keep)
	require.NoError(t, err)
	
	// No new ignores
	newIgnores := make(map[string]struct{})
	
	// Should not save anything
	if len(newIgnores) > 0 {
		for k := range newIgnores {
			ignores[k] = struct{}{}
		}
		err = ignore.Save(fs, rootPath, ignores)
		require.NoError(t, err)
	}
	
	// Verify no .pickyignore was created
	_, err = fs.Stat(filepath.Join(rootPath, ".pickyignore"))
	assert.Error(t, err, ".pickyignore should not exist when no ignores are added")
}