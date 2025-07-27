package domain_test

import (
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	pickyfs "github.com/eliooooooot/picky/internal/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcludeNode(t *testing.T) {
	// Create test tree structure
	root := &domain.Node{Path: "/root", Name: "root", IsDir: true}
	file1 := &domain.Node{Path: "/root/file1.txt", Name: "file1.txt", IsDir: false, Parent: root}
	dir1 := &domain.Node{Path: "/root/dir1", Name: "dir1", IsDir: true, Parent: root}
	file2 := &domain.Node{Path: "/root/dir1/file2.txt", Name: "file2.txt", IsDir: false, Parent: dir1}
	
	root.Children = []*domain.Node{file1, dir1}
	dir1.Children = []*domain.Node{file2}
	
	tree := domain.NewTree(root)
	
	t.Run("exclude file", func(t *testing.T) {
		initialCount := len(tree.Flatten())
		
		relPath, removedNode := tree.ExcludeNode("/root/file1.txt")
		assert.NotNil(t, removedNode)
		assert.Equal(t, "file1.txt", relPath)
		assert.Equal(t, "file1.txt", removedNode.Name)
		
		// Verify node is removed
		assert.Equal(t, 1, len(root.Children))
		assert.Equal(t, dir1, root.Children[0])
		assert.Equal(t, initialCount-1, len(tree.Flatten()))
	})
	
	t.Run("exclude directory", func(t *testing.T) {
		relPath, removedNode := tree.ExcludeNode("/root/dir1")
		assert.NotNil(t, removedNode)
		assert.Equal(t, "dir1", relPath)
		assert.True(t, removedNode.IsDir)
		
		// Verify directory and its children are removed
		assert.Equal(t, 0, len(root.Children))
		assert.Equal(t, 1, len(tree.Flatten())) // Only root remains
	})
	
	t.Run("exclude root returns nil", func(t *testing.T) {
		relPath, removedNode := tree.ExcludeNode("/root")
		assert.Nil(t, removedNode)
		assert.Empty(t, relPath)
	})
	
	t.Run("exclude non-existent node", func(t *testing.T) {
		relPath, removedNode := tree.ExcludeNode("/root/nonexistent")
		assert.Nil(t, removedNode)
		assert.Empty(t, relPath)
	})
}

func TestExcludeNodeIntegration(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	
	// Create test filesystem
	require.NoError(t, fs.MkdirAll("/root/dir1", 0755))
	require.NoError(t, fs.MkdirAll("/root/dir2", 0755))
	require.NoError(t, fs.WriteFile("/root/file1.txt", []byte("content"), 0644))
	require.NoError(t, fs.WriteFile("/root/dir1/file2.txt", []byte("content"), 0644))
	
	tree, err := domain.BuildTree(fs, "/root")
	require.NoError(t, err)
	
	// Exclude a directory
	relPath, removedNode := tree.ExcludeNode("/root/dir1")
	assert.NotNil(t, removedNode)
	assert.Equal(t, "dir1", relPath)
	
	// Verify tree structure
	nodes := tree.Flatten()
	paths := make([]string, len(nodes))
	for i, node := range nodes {
		paths[i] = node.Path
	}
	
	assert.Contains(t, paths, "/root")
	assert.Contains(t, paths, "/root/file1.txt")
	assert.Contains(t, paths, "/root/dir2")
	assert.NotContains(t, paths, "/root/dir1")
	assert.NotContains(t, paths, "/root/dir1/file2.txt")
}