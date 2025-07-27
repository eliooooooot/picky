package domain_test

import (
	"github.com/eliooooooot/picky/internal/domain"
	"testing"
)

func TestNavigation(t *testing.T) {
	// Build test tree
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
	
	// Create initial state with root open
	state := domain.NewViewState(root.Path)
	state = state.SetOpen(root.Path, true)
	
	// Test Down navigation
	state = domain.NavigateDown(root, state) // Move to file1.txt
	if state.CursorPath != "/root/file1.txt" {
		t.Errorf("After Down, cursor = %s, want /root/file1.txt", state.CursorPath)
	}
	
	state = domain.NavigateDown(root, state) // Move to dir1
	if state.CursorPath != "/root/dir1" {
		t.Errorf("After second Down, cursor = %s, want /root/dir1", state.CursorPath)
	}
	
	// Test In (expand directory)
	state = domain.NavigateIn(root, state) // Should expand dir1
	if !state.IsOpen("/root/dir1") {
		t.Error("After In on collapsed dir, it should be expanded")
	}
	
	state = domain.NavigateIn(root, state) // Should move into dir1 (to file2.txt)
	if state.CursorPath != "/root/dir1/file2.txt" {
		t.Errorf("After In on expanded dir, cursor = %s, want /root/dir1/file2.txt", state.CursorPath)
	}
	
	// Test Out (move to parent)
	state = domain.NavigateOut(root, state) // Should move back to dir1
	if state.CursorPath != "/root/dir1" {
		t.Errorf("After Out, cursor = %s, want /root/dir1", state.CursorPath)
	}
	
	// Test Out (collapse directory)
	state = domain.NavigateOut(root, state) // Should collapse dir1
	if state.IsOpen("/root/dir1") {
		t.Error("After Out on expanded dir, it should be collapsed")
	}
	
	// Test Up navigation
	state = domain.NavigateUp(root, state) // Move to file1.txt
	if state.CursorPath != "/root/file1.txt" {
		t.Errorf("After Up, cursor = %s, want /root/file1.txt", state.CursorPath)
	}
	
	state = domain.NavigateUp(root, state) // Move to root
	if state.CursorPath != "/root" {
		t.Errorf("After second Up, cursor = %s, want /root", state.CursorPath)
	}
	
	// Test boundary conditions
	state = domain.NavigateUp(root, state) // Should stay at root
	if state.CursorPath != "/root" {
		t.Errorf("Up at root should stay at root, but cursor = %s", state.CursorPath)
	}
}

func TestNavigationBoundaries(t *testing.T) {
	// Single node tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	
	state := domain.NewViewState(root.Path)
	state = state.SetOpen(root.Path, true)
	
	// All navigation should be safe
	originalState := state
	state = domain.NavigateUp(root, state)
	state = domain.NavigateDown(root, state)
	state = domain.NavigateIn(root, state)
	state = domain.NavigateOut(root, state)
	
	if state.CursorPath != originalState.CursorPath {
		t.Errorf("Cursor should remain at root for single-node tree, but moved to %s", state.CursorPath)
	}
}