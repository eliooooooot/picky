package app_test

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/eliooooooot/picky/internal/app"
	"github.com/eliooooooot/picky/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestAppRunResolvesRelativePaths(t *testing.T) {
	// This test verifies that relative paths like "." are resolved to absolute paths
	// We can't easily test the actual TUI, but we can verify the tree building works
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "picky-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create some test files
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "src", "main.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test"), 0644))
	
	// Change to the temp directory to test "."
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	
	require.NoError(t, os.Chdir(tempDir))
	
	// Create app
	testApp := &app.App{
		FS:         fs.NewOSFileSystem(),
		OutputPath: "test-output.txt",
	}
	
	// Test with "." - should work without error
	// Note: We can't fully test the TUI part, but BuildTree should work
	// The actual Run method will fail because we can't test the TUI
	// So we'll just verify that the path resolution doesn't error
	err = testApp.Run(".")
	// We expect an error from the TUI, but not from path resolution
	// If path resolution failed, we'd get "resolve path: ..." error
	if err != nil {
		require.NotContains(t, err.Error(), "resolve path:")
	}
}