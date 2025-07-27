package domain_test

import (
	"github.com/eliooooooot/picky/internal/domain"
	"testing"
)

func TestToggleSelect(t *testing.T) {
	node := &domain.Node{
		Path: "/test/file.txt",
		Name: "file.txt",
	}
	
	state := domain.NewViewState(node.Path)
	
	// Toggle on
	state = domain.ToggleSelection(node, state)
	if !state.IsSelected(node.Path) {
		t.Error("After toggle, node should be selected")
	}
	
	// Toggle off
	state = domain.ToggleSelection(node, state)
	if state.IsSelected(node.Path) {
		t.Error("After second toggle, node should not be selected")
	}
}

func TestSelectedPaths(t *testing.T) {
	// Build test tree with some selections
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{
				Path:  "/root/dir1",
				Name:  "dir1",
				IsDir: true,
				Children: []*domain.Node{
					{Path: "/root/dir1/file2.txt", Name: "file2.txt"},
					{Path: "/root/dir1/file3.txt", Name: "file3.txt"},
				},
			},
			{Path: "/root/file4.txt", Name: "file4.txt"},
			{Path: "/root/dir2", Name: "dir2", IsDir: true},
		},
	}
	
	state := domain.NewViewState(root.Path)
	state = state.SetSelected("/root/file1.txt", true)
	state = state.SetSelected("/root/dir1/file3.txt", true)
	state = state.SetSelected("/root/dir2", true) // Should be ignored
	
	paths := domain.GetSelectedPaths(root, state)
	
	// Should only include selected files, not directories
	expected := []string{"/root/file1.txt", "/root/dir1/file3.txt"}
	if len(paths) != len(expected) {
		t.Errorf("Selected paths = %d, want %d", len(paths), len(expected))
	}
	
	for i, path := range expected {
		if i < len(paths) && paths[i] != path {
			t.Errorf("paths[%d] = %s, want %s", i, paths[i], path)
		}
	}
}

func TestNoSelections(t *testing.T) {
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{Path: "/root/file2.txt", Name: "file2.txt"},
		},
	}
	
	state := domain.NewViewState(root.Path)
	paths := domain.GetSelectedPaths(root, state)
	
	if len(paths) != 0 {
		t.Errorf("With no selections, paths = %d, want 0", len(paths))
	}
}

func TestRecursiveDirectorySelection(t *testing.T) {
	// Build test tree with nested structure
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			// Directories first
			{
				Path:  "/root/dir1",
				Name:  "dir1",
				IsDir: true,
				Children: []*domain.Node{
					{Path: "/root/dir1/file2.txt", Name: "file2.txt"},
					{Path: "/root/dir1/file3.txt", Name: "file3.txt"},
					{
						Path:  "/root/dir1/subdir",
						Name:  "subdir",
						IsDir: true,
						Children: []*domain.Node{
							{Path: "/root/dir1/subdir/file4.txt", Name: "file4.txt"},
						},
					},
				},
			},
			// Then files
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{Path: "/root/file5.txt", Name: "file5.txt"},
		},
	}
	
	// Set parent pointers
	setParents(root)
	
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetCursor("/root/dir1")
	
	// Select dir1 (should recursively select all files)
	state = domain.ToggleSelection(root, state)
	
	// Check that all files under dir1 are selected
	paths := domain.GetSelectedPaths(root, state)
	expected := []string{
		"/root/dir1/file2.txt",
		"/root/dir1/file3.txt",
		"/root/dir1/subdir/file4.txt",
	}
	
	if len(paths) != len(expected) {
		t.Errorf("After selecting dir1, got %d paths, want %d", len(paths), len(expected))
	}
	
	for i, path := range expected {
		if i < len(paths) && paths[i] != path {
			t.Errorf("paths[%d] = %s, want %s", i, paths[i], path)
		}
	}
	
	// Deselect dir1
	state = domain.ToggleSelection(root, state)
	paths = domain.GetSelectedPaths(root, state)
	
	if len(paths) != 0 {
		t.Errorf("After deselecting dir1, got %d paths, want 0", len(paths))
	}
}

func TestPartialDirectorySelection(t *testing.T) {
	// Build test tree
	dir := &domain.Node{
		Path:  "/root/dir",
		Name:  "dir",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/dir/file1.txt", Name: "file1.txt"},
			{Path: "/root/dir/file2.txt", Name: "file2.txt"},
			{Path: "/root/dir/file3.txt", Name: "file3.txt"},
		},
	}
	
	state := domain.NewViewState("/root")
	state = state.SetSelected("/root/dir/file1.txt", true)
	state = state.SetSelected("/root/dir/file3.txt", true)
	
	// Test partial selection
	if !domain.HasPartialSelection(dir, state) {
		t.Error("Directory with some selected files should have partial selection")
	}
	
	if domain.HasFullSelection(dir, state) {
		t.Error("Directory with some unselected files should not have full selection")
	}
	
	// Select all files
	state = state.SetSelected("/root/dir/file2.txt", true)
	
	if domain.HasPartialSelection(dir, state) {
		t.Error("Directory with all files selected should not have partial selection")
	}
	
	if !domain.HasFullSelection(dir, state) {
		t.Error("Directory with all files selected should have full selection")
	}
}

// Helper to set parent pointers
func setParents(node *domain.Node) {
	for _, child := range node.Children {
		child.Parent = node
		setParents(child)
	}
}