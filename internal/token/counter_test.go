package token

import (
	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/eliotsamuelmiller/picky/internal/fs"
	"testing"
)

// readTrackingFS wraps a FileSystem to track ReadFile calls
type readTrackingFS struct {
	domain.FileSystem
	readCount map[string]int
}

func (r *readTrackingFS) ReadFile(path string) ([]byte, error) {
	r.readCount[path]++
	return r.FileSystem.ReadFile(path)
}

func TestCounter_BuildTreeTokenMap(t *testing.T) {
	// Create mock filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddFile("/root/file1.txt", "hello world")     // 11 chars = 3 tokens
	memfs.AddFile("/root/file2.txt", "test")            // 4 chars = 1 token
	memfs.AddFile("/root/dir/file3.txt", "1234567890")  // 10 chars = 3 tokens

	// Create a simple tree
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
	file2 := &domain.Node{
		Path:   "/root/file2.txt",
		Name:   "file2.txt",
		IsDir:  false,
		Parent: root,
	}
	dir := &domain.Node{
		Path:   "/root/dir",
		Name:   "dir",
		IsDir:  true,
		Parent: root,
	}
	file3 := &domain.Node{
		Path:   "/root/dir/file3.txt",
		Name:   "file3.txt",
		IsDir:  false,
		Parent: dir,
	}

	root.Children = []*domain.Node{file1, file2, dir}
	dir.Children = []*domain.Node{file3}

	tree := &domain.Tree{Root: root}

	// Create counter and build token map
	counter := NewCounter(memfs, NaiveTokenizer{})
	tokenMap, err := counter.BuildTreeTokenMap(tree)

	if err != nil {
		t.Fatalf("BuildTreeTokenMap failed: %v", err)
	}

	// Check results
	tests := []struct {
		path     string
		expected int
	}{
		{"/root/file1.txt", 3},
		{"/root/file2.txt", 1},
		{"/root/dir/file3.txt", 3},
	}

	for _, tt := range tests {
		got, ok := tokenMap[tt.path]
		if !ok {
			t.Errorf("Path %s not found in token map", tt.path)
			continue
		}
		if got != tt.expected {
			t.Errorf("Token count for %s = %d, want %d", tt.path, got, tt.expected)
		}
	}

	// Verify directories are not in the map
	if _, ok := tokenMap["/root"]; ok {
		t.Error("Directory /root should not be in token map")
	}
	if _, ok := tokenMap["/root/dir"]; ok {
		t.Error("Directory /root/dir should not be in token map")
	}
}

func TestCounter_Caching(t *testing.T) {
	// Create mock filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddFile("/file.txt", "test content")

	// Track read counts by wrapping the filesystem
	readCount := make(map[string]int)
	trackingFS := &readTrackingFS{
		FileSystem: memfs,
		readCount:  readCount,
	}

	counter := NewCounter(trackingFS, NaiveTokenizer{})

	// First call should read the file
	tokens1, err := counter.tokensForFile("/file.txt")
	if err != nil {
		t.Fatalf("First tokensForFile call failed: %v", err)
	}

	// Second call should use cache
	tokens2, err := counter.tokensForFile("/file.txt")
	if err != nil {
		t.Fatalf("Second tokensForFile call failed: %v", err)
	}

	// Check tokens are the same
	if tokens1 != tokens2 {
		t.Errorf("Token counts differ: %d vs %d", tokens1, tokens2)
	}

	// Check file was only read once
	if readCount["/file.txt"] != 1 {
		t.Errorf("File was read %d times, expected 1", readCount["/file.txt"])
	}
}