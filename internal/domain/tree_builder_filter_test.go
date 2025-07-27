package domain_test

import (
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	pickyfs "github.com/eliooooooot/picky/internal/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTreeWithFilter(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	
	// Create test filesystem structure
	require.NoError(t, fs.MkdirAll("/root/dir1", 0755))
	require.NoError(t, fs.MkdirAll("/root/dir2", 0755))
	require.NoError(t, fs.MkdirAll("/root/dir3", 0755))
	require.NoError(t, fs.WriteFile("/root/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile("/root/dir1/file2.txt", []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile("/root/dir2/file3.txt", []byte("content3"), 0644))
	
	// Filter that excludes dir2
	keep := func(path string, isDir bool) bool {
		return path != "/root/dir2"
	}
	
	tree, err := domain.BuildTreeWithFilter(fs, "/root", keep)
	require.NoError(t, err)
	
	// Flatten tree to check results
	nodes := tree.Flatten()
	paths := make([]string, len(nodes))
	for i, node := range nodes {
		paths[i] = node.Path
	}
	
	// dir2 and its contents should be excluded
	assert.Contains(t, paths, "/root")
	assert.Contains(t, paths, "/root/dir1")
	assert.Contains(t, paths, "/root/dir3")
	assert.Contains(t, paths, "/root/file1.txt")
	assert.Contains(t, paths, "/root/dir1/file2.txt")
	assert.NotContains(t, paths, "/root/dir2")
	assert.NotContains(t, paths, "/root/dir2/file3.txt")
}

func TestBuildTreeWithNilFilter(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	
	require.NoError(t, fs.MkdirAll("/root/dir1", 0755))
	require.NoError(t, fs.WriteFile("/root/file1.txt", []byte("content1"), 0644))
	
	// BuildTreeWithFilter with nil filter should behave like BuildTree
	tree1, err := domain.BuildTree(fs, "/root")
	require.NoError(t, err)
	
	tree2, err := domain.BuildTreeWithFilter(fs, "/root", nil)
	require.NoError(t, err)
	
	// Both trees should have the same structure
	nodes1 := tree1.Flatten()
	nodes2 := tree2.Flatten()
	
	assert.Equal(t, len(nodes1), len(nodes2))
}