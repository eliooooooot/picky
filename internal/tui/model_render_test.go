package tui

import (
	"os"
	"github.com/eliooooooot/picky/internal/domain"
	"strings"
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderDirectoryWithAllChildrenSelected(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Build test tree
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
					{Path: "/root/dir1/file2.txt", Name: "file2.txt"},
				},
			},
			{Path: "/root/file3.txt", Name: "file3.txt"},
		},
	}
	
	// Set parent pointers
	setParents(root)
	
	tree := domain.NewTree(root)
	model := NewModel(tree, nil)
	
	// Initialize state
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetOpen("/root/dir1", true)
	
	// Select all files in dir1
	state = state.SetSelected("/root/dir1/file1.txt", true)
	state = state.SetSelected("/root/dir1/file2.txt", true)
	
	model.state = state
	
	// Initialize the model to set up viewport
	model.Init()
	
	// Render and check
	view := model.View()
	lines := strings.Split(view, "\n")
	
	// Find the dir1 line
	var dir1Line string
	for _, line := range lines {
		if strings.Contains(line, "✓ ▼ dir1") {
			dir1Line = line
			break
		}
	}
	
	if dir1Line == "" {
		t.Fatal("Could not find dir1 line with checkmark")
	}
	
	// Check if dir1 has green color (ANSI code \x1b[32m)
	if !strings.Contains(dir1Line, "\x1b[32m") {
		t.Errorf("Directory with all children selected should be rendered in green. Line: %q", dir1Line)
		t.Logf("Full view:\n%s", view)
	}
}

func TestRenderAfterParentSelection(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Build test tree
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
					{Path: "/root/dir1/file2.txt", Name: "file2.txt"},
					{Path: "/root/dir1/file3.txt", Name: "file3.txt"},
				},
			},
		},
	}
	
	// Set parent pointers
	setParents(root)
	
	tree := domain.NewTree(root)
	model := NewModel(tree, nil)
	
	// Initialize state
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetOpen("/root/dir1", true)
	
	// First select only file1
	state = state.SetSelected("/root/dir1/file1.txt", true)
	
	// Then select the parent directory (which should select all children)
	state = state.SetCursor("/root/dir1")
	state = domain.ToggleSelection(root, state)
	
	model.state = state
	
	// Initialize the model to set up viewport
	model.Init()
	
	// Render and check
	view := model.View()
	lines := strings.Split(view, "\n")
	
	// Count green file lines
	greenFileCount := 0
	for _, line := range lines {
		if (strings.Contains(line, "✓ file1.txt") ||
			strings.Contains(line, "✓ file2.txt") ||
			strings.Contains(line, "✓ file3.txt")) &&
			strings.Contains(line, "\x1b[32m") {
			greenFileCount++
		}
	}
	
	if greenFileCount != 3 {
		t.Errorf("After selecting parent, all 3 files should be green, but only %d are", greenFileCount)
	}
}

func TestRenderPartialSelection(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Build test tree
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
					{Path: "/root/dir1/file2.txt", Name: "file2.txt"},
				},
			},
		},
	}
	
	// Set parent pointers
	setParents(root)
	
	tree := domain.NewTree(root)
	model := NewModel(tree, nil)
	
	// Initialize state with only one file selected
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	state = state.SetOpen("/root/dir1", true)
	state = state.SetSelected("/root/dir1/file1.txt", true)
	
	model.state = state
	
	// Initialize the model to set up viewport
	model.Init()
	
	// Render and check
	view := model.View()
	lines := strings.Split(view, "\n")
	
	// Find the dir1 line
	var dir1Line string
	for _, line := range lines {
		if strings.Contains(line, "- ▼ dir1") {
			dir1Line = line
			break
		}
	}
	
	if dir1Line == "" {
		t.Fatal("Could not find dir1 line with partial selection indicator")
	}
	
	// Check that dir1 is NOT green (partial selection should not be green)
	if strings.Contains(dir1Line, "\x1b[32m") {
		t.Error("Directory with partial selection should not be rendered in green")
	}
}

// Helper to set parent pointers
func setParents(node *domain.Node) {
	for _, child := range node.Children {
		child.Parent = node
		setParents(child)
	}
}

func TestPromptModeDimmedTree(t *testing.T) {
	// Force color output for tests
	os.Setenv("CLICOLOR_FORCE", "1")
	
	// Build test tree
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
		Children: []*domain.Node{
			{Path: "/root/file1.txt", Name: "file1.txt"},
			{Path: "/root/file2.txt", Name: "file2.txt"},
		},
	}
	
	// Set parent pointers
	setParents(root)
	
	tree := domain.NewTree(root)
	ignores := make(map[string]struct{})
	model := NewModel(tree, &ignores)
	
	// Initialize state
	state := domain.NewViewState(root.Path)
	state = state.SetOpen("/root", true)
	model.state = state
	
	// Initialize the model to set up viewport
	model.Init()
	
	// Get initial view (not in prompt mode)
	view := model.View()
	
	// Find a file line
	var fileLine string
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "file1.txt") {
			fileLine = line
			break
		}
	}
	
	if fileLine == "" {
		t.Fatal("Could not find file1.txt in view")
	}
	
	// Check that the file line is not dimmed
	if strings.Contains(fileLine, "\x1b[2m") {
		t.Error("File line should not be dimmed when not in prompt mode")
	}
	
	// Enter prompt mode
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updatedModel.(*Model)
	
	// Get view in prompt mode
	promptView := m.View()
	
	// Find the same file line
	var promptFileLine string
	for _, line := range strings.Split(promptView, "\n") {
		if strings.Contains(line, "file1.txt") {
			promptFileLine = line
			break
		}
	}
	
	if promptFileLine == "" {
		t.Fatal("Could not find file1.txt in prompt mode view")
	}
	
	// Check that the file line is now dimmed
	if !strings.Contains(promptFileLine, "\x1b[2m") {
		t.Error("File line should be dimmed when in prompt mode")
	}
}