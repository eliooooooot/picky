package generate_test

import (
	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/eliotsamuelmiller/picky/internal/fs"
	"github.com/eliotsamuelmiller/picky/internal/generate"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	// Create test filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddDir("/root")
	memfs.AddFile("/root/file1.txt", "Content of file 1")
	memfs.AddDir("/root/subdir")
	memfs.AddFile("/root/subdir/file2.txt", "Content of file 2\nWith multiple lines")
	memfs.AddFile("/root/file3.txt", "Content of file 3")
	
	// Build tree
	tree, err := domain.BuildTree(memfs, "/root")
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Create state and select some files
	state := domain.NewViewState(tree.Root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetOpen("/root/subdir", true)
	state = state.SetSelected("/root/file1.txt", true)
	state = state.SetSelected("/root/subdir/file2.txt", true)
	
	// Generate output
	err = generate.Generate("/output.txt", "", tree, state, memfs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Check output
	content, err := memfs.GetContent("/output.txt")
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	
	// Verify content includes directory structure
	if !strings.Contains(content, "# Directory Structure") {
		t.Error("Output should contain directory structure header")
	}
	
	// Verify content includes selected files
	if !strings.Contains(content, "# Selected Files") {
		t.Error("Output should contain selected files header")
	}
	
	if !strings.Contains(content, "file1.txt") {
		t.Error("Output should contain file1.txt")
	}
	
	if !strings.Contains(content, "Content of file 1") {
		t.Error("Output should contain content of file1.txt")
	}
	
	if !strings.Contains(content, "file2.txt") {
		t.Error("Output should contain file2.txt")
	}
	
	if !strings.Contains(content, "Content of file 2") {
		t.Error("Output should contain content of file2.txt")
	}
	
	// Should not contain unselected file content (but may appear in structure)
	if strings.Contains(content, "Content of file 3") {
		t.Error("Output should not contain content of unselected file3.txt")
	}
}

func TestGenerateNoSelection(t *testing.T) {
	// Create test filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddDir("/root")
	memfs.AddFile("/root/file1.txt", "Content")
	
	// Build tree without selections
	tree, err := domain.BuildTree(memfs, "/root")
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Create empty state
	state := domain.NewViewState(tree.Root.Path)
	
	// Generate output
	err = generate.Generate("/output.txt", "", tree, state, memfs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Check output
	content, err := memfs.GetContent("/output.txt")
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	
	if !strings.Contains(content, "No files selected") {
		t.Error("Output should indicate no files selected")
	}
}

func TestGenerateWithPrompt(t *testing.T) {
	// Create test filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddDir("/root")
	memfs.AddFile("/root/file1.txt", "Content of file 1")
	
	// Build tree
	tree, err := domain.BuildTree(memfs, "/root")
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Create state and select a file
	state := domain.NewViewState(tree.Root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetSelected("/root/file1.txt", true)
	
	// Generate output with prompt
	testPrompt := "Please analyze this code and provide suggestions."
	err = generate.Generate("/output.txt", testPrompt, tree, state, memfs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Check output
	content, err := memfs.GetContent("/output.txt")
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	
	// Verify prompt is included
	if !strings.Contains(content, "# Prompt") {
		t.Error("Output should contain prompt header")
	}
	
	if !strings.Contains(content, testPrompt) {
		t.Error("Output should contain the actual prompt text")
	}
	
	// Verify content still includes directory structure
	if !strings.Contains(content, "# Directory Structure") {
		t.Error("Output should still contain directory structure")
	}
	
	// Verify selected file content is included
	if !strings.Contains(content, "Content of file 1") {
		t.Error("Output should contain selected file content")
	}
}

func TestTextWriter(t *testing.T) {
	// Create test filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddFile("/test.txt", "test content")
	
	// Create a simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file.txt", Name: "file.txt"},
			{
				Path:  "/root/dir",
				Name:  "dir",
				IsDir: true,
				Children: []*domain.Node{
					{Path: "/root/dir/nested.txt", Name: "nested.txt"},
				},
			},
		},
	}
	
	// Set parent pointers
	for _, child := range root.Children {
		child.Parent = root
		for _, grandchild := range child.Children {
			grandchild.Parent = child
		}
	}
	
	// Create state with selections
	state := domain.NewViewState(root.Path)
	state = state.SetSelected("/root/file.txt", true)
	state = state.SetSelected("/root/dir/nested.txt", true)
	
	// Test text writer directly
	writer := generate.NewTextWriter()
	var buf strings.Builder
	
	// Test structure writing
	err := writer.WriteStructure(&buf, root, state)
	if err != nil {
		t.Fatalf("WriteStructure failed: %v", err)
	}
	
	structure := buf.String()
	if !strings.Contains(structure, "# Directory Structure") {
		t.Error("Structure should contain header")
	}
	
	if !strings.Contains(structure, "file.txt *") {
		t.Error("Structure should mark selected files")
	}
	
	// Test content writing
	buf.Reset()
	paths := []string{"/test.txt"}
	err = writer.WriteContent(&buf, paths, memfs)
	if err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	
	content := buf.String()
	if !strings.Contains(content, "# Selected Files") {
		t.Error("Content should contain header")
	}
	
	if !strings.Contains(content, "test content") {
		t.Error("Content should contain file content")
	}
}