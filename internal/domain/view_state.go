package domain

import "strings"

// ViewState represents the UI state separate from the domain model
type ViewState struct {
	// CursorPath is the path of the currently focused node
	CursorPath string
	
	// Open tracks which directories are expanded
	// Key is the node path, value is whether it's open
	Open map[string]bool
	
	// Selected tracks which nodes are selected
	// Key is the node path, value is whether it's selected
	Selected map[string]bool
}

// NewViewState creates a new ViewState with the given root path as cursor
func NewViewState(rootPath string) ViewState {
	return ViewState{
		CursorPath: rootPath,
		Open:       make(map[string]bool),
		Selected:   make(map[string]bool),
	}
}

// IsOpen returns whether a node at the given path is expanded
func (v ViewState) IsOpen(path string) bool {
	return v.Open[path]
}

// IsSelected returns whether a node at the given path is selected
func (v ViewState) IsSelected(path string) bool {
	return v.Selected[path]
}

// SetOpen sets the expanded state for a node at the given path
func (v ViewState) SetOpen(path string, open bool) ViewState {
	newState := v.copy()
	if open {
		newState.Open[path] = true
	} else {
		delete(newState.Open, path)
	}
	return newState
}

// SetSelected sets the selected state for a node at the given path
func (v ViewState) SetSelected(path string, selected bool) ViewState {
	newState := v.copy()
	if selected {
		newState.Selected[path] = true
	} else {
		delete(newState.Selected, path)
	}
	return newState
}

// SetCursor updates the cursor position
func (v ViewState) SetCursor(path string) ViewState {
	newState := v.copy()
	newState.CursorPath = path
	return newState
}

// Prune removes all entries from Open and Selected maps that have the given path as prefix
// This is useful when a node is removed from the tree
func (v ViewState) Prune(pathPrefix string) ViewState {
	newState := v.copy()
	
	// Remove from Open map
	for path := range newState.Open {
		if strings.HasPrefix(path, pathPrefix) {
			delete(newState.Open, path)
		}
	}
	
	// Remove from Selected map
	for path := range newState.Selected {
		if strings.HasPrefix(path, pathPrefix) {
			delete(newState.Selected, path)
		}
	}
	
	return newState
}

// copy creates a deep copy of the ViewState
func (v ViewState) copy() ViewState {
	newOpen := make(map[string]bool, len(v.Open))
	for k, val := range v.Open {
		newOpen[k] = val
	}
	
	newSelected := make(map[string]bool, len(v.Selected))
	for k, val := range v.Selected {
		newSelected[k] = val
	}
	
	return ViewState{
		CursorPath: v.CursorPath,
		Open:       newOpen,
		Selected:   newSelected,
	}
}