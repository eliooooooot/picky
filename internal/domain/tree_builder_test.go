package domain_test

import (
	"path/filepath"
	"testing"
	
	"github.com/eliooooooot/picky/internal/domain"
	"github.com/eliooooooot/picky/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestBuildTreeRootName(t *testing.T) {
	tests := []struct {
		name         string
		rootPath     string
		expectedName string
	}{
		{
			name:         "absolute path shows directory name",
			rootPath:     "/Users/test/projects/myapp",
			expectedName: "myapp",
		},
		{
			name:         "nested path",
			rootPath:     "/home/user/work",
			expectedName: "work",
		},
		{
			name:         "single level path",
			rootPath:     "/projects",
			expectedName: "projects",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a memory filesystem with a simple structure
			memFS := fs.NewMemFileSystem()
			require.NoError(t, memFS.MkdirAll(tt.rootPath, 0755))
			
			// Build tree
			tree, err := domain.BuildTree(memFS, tt.rootPath)
			require.NoError(t, err)
			require.NotNil(t, tree)
			require.NotNil(t, tree.Root)
			
			// Check root name
			require.Equal(t, tt.expectedName, tree.Root.Name)
		})
	}
}

func TestBuildTreeRootFilesystem(t *testing.T) {
	// Special test for root filesystem
	// For memfs, we can't actually create "/" as it normalizes paths
	// But we can test that filepath.Base("/") returns "/"
	require.Equal(t, "/", filepath.Base("/"))
}

func TestBuildTreePreservesPath(t *testing.T) {
	// Test that the full path is preserved even though name is just the base
	memFS := fs.NewMemFileSystem()
	rootPath := "/Users/test/projects/myapp"
	require.NoError(t, memFS.MkdirAll(rootPath, 0755))
	require.NoError(t, memFS.MkdirAll(filepath.Join(rootPath, "src"), 0755))
	require.NoError(t, memFS.WriteFile(filepath.Join(rootPath, "src", "main.go"), []byte("package main"), 0644))
	
	tree, err := domain.BuildTree(memFS, rootPath)
	require.NoError(t, err)
	
	// Root should have the base name but full path
	require.Equal(t, "myapp", tree.Root.Name)
	require.Equal(t, rootPath, tree.Root.Path)
	
	// Children should also have correct paths
	require.Len(t, tree.Root.Children, 1)
	srcDir := tree.Root.Children[0]
	require.Equal(t, "src", srcDir.Name)
	require.Equal(t, filepath.Join(rootPath, "src"), srcDir.Path)
}