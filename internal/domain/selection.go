package domain

// ToggleSelection toggles selection on the current node
// For directories, it recursively selects/deselects all files within
func ToggleSelection(root *Node, state ViewState) ViewState {
	cursor := FindNodeByPath(root, state.CursorPath)
	if cursor == nil {
		return state
	}
	
	newState := state
	isSelected := state.IsSelected(cursor.Path)
	
	// Toggle the current node
	newState = newState.SetSelected(cursor.Path, !isSelected)
	
	// If it's a directory, recursively update all descendant files
	if cursor.IsDir {
		newState = setSelectionRecursive(cursor, newState, !isSelected)
	}
	
	return newState
}

// GetSelectedPaths returns all selected file paths in depth-first order
func GetSelectedPaths(root *Node, state ViewState) []string {
	var paths []string
	collectSelectedPaths(root, state, &paths)
	return paths
}

func collectSelectedPaths(node *Node, state ViewState, paths *[]string) {
	if state.IsSelected(node.Path) && !node.IsDir {
		*paths = append(*paths, node.Path)
	}
	
	for _, child := range node.Children {
		collectSelectedPaths(child, state, paths)
	}
}

// setSelectionRecursive recursively sets selection state for all descendants
func setSelectionRecursive(node *Node, state ViewState, selected bool) ViewState {
	newState := state
	
	for _, child := range node.Children {
		// Select/deselect both files and directories
		newState = newState.SetSelected(child.Path, selected)
		
		// If it's a directory, recurse into it
		if child.IsDir {
			newState = setSelectionRecursive(child, newState, selected)
		}
	}
	
	return newState
}

// HasPartialSelection returns true if a directory has some but not all files selected
func HasPartialSelection(node *Node, state ViewState) bool {
	if !node.IsDir {
		return false
	}
	
	selected, total := countSelectedFiles(node, state)
	return selected > 0 && selected < total
}

// HasFullSelection returns true if a directory has all files selected
func HasFullSelection(node *Node, state ViewState) bool {
	if !node.IsDir {
		return false
	}
	
	selected, total := countSelectedFiles(node, state)
	return selected > 0 && selected == total
}

// countSelectedFiles returns the number of selected files and total files in a directory tree
func countSelectedFiles(node *Node, state ViewState) (selected, total int) {
	if !node.IsDir {
		if state.IsSelected(node.Path) {
			return 1, 1
		}
		return 0, 1
	}
	
	for _, child := range node.Children {
		s, t := countSelectedFiles(child, state)
		selected += s
		total += t
	}
	
	return selected, total
}