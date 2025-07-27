package ignore_test

import (
	"testing"

	pickyfs "github.com/eliooooooot/picky/internal/fs"
	"github.com/eliooooooot/picky/internal/ignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	root := "/test"

	err := fs.MkdirAll(root, 0755)
	require.NoError(t, err)

	expected := map[string]struct{}{
		"dir1":        {},
		"file1.txt":   {},
		"dir2/file.go": {},
	}

	err = ignore.Save(fs, root, expected)
	require.NoError(t, err)

	loaded, err := ignore.Load(fs, root)
	require.NoError(t, err)

	assert.Equal(t, expected, loaded)
	
	// Verify the file is sorted
	data, err := fs.ReadFile("/test/.pickyignore")
	require.NoError(t, err)
	content := string(data)
	assert.Equal(t, "dir1\ndir2/file.go\nfile1.txt\n", content)
}

func TestLoadWithComments(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	root := "/test"

	err := fs.MkdirAll(root, 0755)
	require.NoError(t, err)

	content := `# This is a comment
dir1
# Another comment
file1.txt

# Empty line above should be ignored
dir2/file.go
`
	err = fs.WriteFile("/test/.pickyignore", []byte(content), 0644)
	require.NoError(t, err)

	loaded, err := ignore.Load(fs, root)
	require.NoError(t, err)

	expected := map[string]struct{}{
		"dir1":        {},
		"file1.txt":   {},
		"dir2/file.go": {},
	}
	assert.Equal(t, expected, loaded)
}

func TestLoadNonExistent(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	root := "/test"

	loaded, err := ignore.Load(fs, root)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestSaveEmptySet(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	root := "/test"

	err := fs.MkdirAll(root, 0755)
	require.NoError(t, err)

	err = ignore.Save(fs, root, map[string]struct{}{})
	require.NoError(t, err)

	_, err = fs.Stat("/test/.pickyignore")
	assert.Error(t, err, "Empty set should not create file")
}

func TestGoldenFile(t *testing.T) {
	fs := pickyfs.NewMemFileSystem()
	root := "/test"

	err := fs.MkdirAll(root, 0755)
	require.NoError(t, err)

	// Load expected ignores from a complex file
	testContent := `# Sample .pickyignore file for testing
# This is a comment

dir1/
dir2/subdir/
file1.txt

# Another comment
file2.go
path/to/nested/file.js

# Duplicate entries should be removed
file1.txt`

	err = fs.WriteFile("/test/.pickyignore", []byte(testContent), 0644)
	require.NoError(t, err)

	// Load the file
	loaded, err := ignore.Load(fs, root)
	require.NoError(t, err)

	// Should have 5 unique entries (file1.txt appears twice)
	assert.Len(t, loaded, 5)

	// Save it back
	err = ignore.Save(fs, root, loaded)
	require.NoError(t, err)

	// Read the result
	data, err := fs.ReadFile("/test/.pickyignore")
	require.NoError(t, err)
	result := string(data)

	// Expected output: sorted, deduplicated, no comments
	expected := `dir1/
dir2/subdir/
file1.txt
file2.go
path/to/nested/file.js
`
	assert.Equal(t, expected, result)
}