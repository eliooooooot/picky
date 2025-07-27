package domain_test

import (
	"github.com/eliooooooot/picky/internal/domain"
	"github.com/eliooooooot/picky/internal/fs"
	"testing"
)

func TestBuildTree(t *testing.T) {
	// Create test filesystem
	memfs := fs.NewMemFileSystem()
	memfs.AddDir("/root")
	memfs.AddDir("/root/dir1")
	memfs.AddFile("/root/dir1/file1.txt", "content1")
	memfs.AddFile("/root/dir1/file2.txt", "content2")
	memfs.AddDir("/root/dir2")
	memfs.AddFile("/root/file3.txt", "content3")
	
	// Build tree
	tree, err := domain.BuildTree(memfs, "/root")
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Verify root
	if tree.Root.Path != "/root" {
		t.Errorf("Root path = %s, want /root", tree.Root.Path)
	}
	if !tree.Root.IsDir {
		t.Error("Root should be a directory")
	}
	
	// Verify children count
	if len(tree.Root.Children) != 3 {
		t.Errorf("Root children = %d, want 3", len(tree.Root.Children))
	}
	
	// Verify directories come first
	if !tree.Root.Children[0].IsDir || tree.Root.Children[0].Name != "dir1" {
		t.Error("First child should be dir1")
	}
	if !tree.Root.Children[1].IsDir || tree.Root.Children[1].Name != "dir2" {
		t.Error("Second child should be dir2")
	}
	if tree.Root.Children[2].IsDir || tree.Root.Children[2].Name != "file3.txt" {
		t.Error("Third child should be file3.txt")
	}
	
	// Verify nested structure
	dir1 := tree.Root.Children[0]
	if len(dir1.Children) != 2 {
		t.Errorf("dir1 children = %d, want 2", len(dir1.Children))
	}
}

func TestFlatten(t *testing.T) {
	// Build simple tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{
				Path:  "/root/dir1",
				Name:  "dir1", 
				IsDir: true,
				Children: []*domain.Node{
					{Path: "/root/dir1/file1.txt", Name: "file1.txt"},
				},
			},
			{
				Path:  "/root/dir2",
				Name:  "dir2",
				IsDir: true,
				Children: []*domain.Node{
					{Path: "/root/dir2/file2.txt", Name: "file2.txt"},
				},
			},
			{Path: "/root/file3.txt", Name: "file3.txt"},
		},
	}
	
	// Set parent pointers
	for _, child := range root.Children {
		child.Parent = root
		for _, grandchild := range child.Children {
			grandchild.Parent = child
		}
	}
	
	// Create view state with different open states
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetOpen("/root/dir1", true)
	// dir2 is not open (collapsed)
	
	flat := domain.Flatten(root, state)
	
	// With dir2 collapsed, we should see: root, dir1, file1.txt, dir2, file3.txt
	expected := []string{"root", "dir1", "file1.txt", "dir2", "file3.txt"}
	if len(flat) != len(expected) {
		t.Errorf("Flattened length = %d, want %d", len(flat), len(expected))
	}
	
	for i, name := range expected {
		if i < len(flat) && flat[i].Name != name {
			t.Errorf("flat[%d].Name = %s, want %s", i, flat[i].Name, name)
		}
	}
}