package tui

import (
	"strings"
	"testing"

	"github.com/eliooooooot/picky/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func TestResizeNoDuplicates(t *testing.T) {
	// Build a deep tree structure
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	
	dir1 := &domain.Node{
		Path:   "/root/dir1",
		Name:   "dir1",
		IsDir:  true,
		Parent: root,
	}
	
	dir2 := &domain.Node{
		Path:   "/root/dir1/dir2",
		Name:   "dir2",
		IsDir:  true,
		Parent: dir1,
	}
	
	file1 := &domain.Node{
		Path:   "/root/dir1/file1.txt",
		Name:   "file1.txt",
		IsDir:  false,
		Parent: dir1,
	}
	
	file2 := &domain.Node{
		Path:   "/root/dir1/dir2/file2.txt",
		Name:   "file2.txt",
		IsDir:  false,
		Parent: dir2,
	}
	
	file3 := &domain.Node{
		Path:   "/root/dir1/dir2/file3.txt",
		Name:   "file3.txt",
		IsDir:  false,
		Parent: dir2,
	}
	
	// Build tree structure
	root.Children = []*domain.Node{dir1}
	dir1.Children = []*domain.Node{dir2, file1}
	dir2.Children = []*domain.Node{file2, file3}
	
	tree := &domain.Tree{Root: root}
	ignores := make(map[string]struct{})
	
	// Create model and initialize
	model := NewModel(tree, &ignores)
	model.Init()
	
	// Open all directories
	model.state = model.state.SetOpen(dir1.Path, true)
	model.state = model.state.SetOpen(dir2.Path, true)
	
	// Simulate window size changes
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = m.(*Model)
	m, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 10}) // Shrink
	model = m.(*Model)
	m, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30}) // Grow back
	model = m.(*Model)
	
	// Check for duplicates in the rendered output
	view := model.View()
	lines := strings.Split(view, "\n")
	
	// Count occurrences of each line
	seen := make(map[string]int)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			seen[trimmed]++
		}
	}
	
	// Check for duplicates
	for line, count := range seen {
		// Skip header lines and empty lines
		if strings.Contains(line, "Picky - File Selector") ||
			strings.Contains(line, "navigate") ||
			strings.Contains(line, "esc:") ||
			line == "" {
			continue
		}
		
		if count > 1 {
			t.Errorf("Line appears %d times (expected 1): %q", count, line)
		}
	}
}

func TestResizeMaintainsCursorVisibility(t *testing.T) {
	// Build a tree with many items
	root := &domain.Node{
		Path:  "/root",
		Name:  "root",
		IsDir: true,
	}
	
	// Create many files to exceed viewport
	for i := 0; i < 50; i++ {
		file := &domain.Node{
			Path:   "/root/file" + string(rune('0'+i)) + ".txt",
			Name:   "file" + string(rune('0'+i)) + ".txt",
			IsDir:  false,
			Parent: root,
		}
		root.Children = append(root.Children, file)
	}
	
	tree := &domain.Tree{Root: root}
	ignores := make(map[string]struct{})
	
	// Create model
	model := NewModel(tree, &ignores)
	model.Init()
	
	// Navigate to middle of list
	for i := 0; i < 25; i++ {
		model.state = domain.NavigateDown(model.tree.Root, model.state)
	}
	
	// Resize window
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	model = m.(*Model)
	m, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 10}) // Smaller
	model = m.(*Model)
	
	// Check that cursor is still visible in the view
	view := model.View()
	if !strings.Contains(view, "→") && !strings.Contains(view, "▶") {
		t.Error("Cursor indicator not found in view after resize")
	}
}