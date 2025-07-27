package domain

// NavigateUp moves the cursor up in the flattened view
func NavigateUp(root *Node, state ViewState) ViewState {
	flat := Flatten(root, state)
	cursor := FindNodeByPath(root, state.CursorPath)
	if cursor == nil {
		return state
	}
	
	currentIdx := findNodeIndex(flat, cursor)
	if currentIdx > 0 {
		return state.SetCursor(flat[currentIdx-1].Path)
	}
	
	return state
}

// NavigateDown moves the cursor down in the flattened view
func NavigateDown(root *Node, state ViewState) ViewState {
	flat := Flatten(root, state)
	cursor := FindNodeByPath(root, state.CursorPath)
	if cursor == nil {
		return state
	}
	
	currentIdx := findNodeIndex(flat, cursor)
	if currentIdx < len(flat)-1 {
		return state.SetCursor(flat[currentIdx+1].Path)
	}
	
	return state
}

// NavigateIn expands the current directory or moves into first child
func NavigateIn(root *Node, state ViewState) ViewState {
	cursor := FindNodeByPath(root, state.CursorPath)
	if cursor == nil || !cursor.IsDir {
		return state
	}
	
	if !state.IsOpen(cursor.Path) {
		// Expand the directory
		return state.SetOpen(cursor.Path, true)
	} else if len(cursor.Children) > 0 {
		// Move to first child
		return state.SetCursor(cursor.Children[0].Path)
	}
	
	return state
}

// NavigateOut collapses the current directory or moves to parent
func NavigateOut(root *Node, state ViewState) ViewState {
	cursor := FindNodeByPath(root, state.CursorPath)
	if cursor == nil {
		return state
	}
	
	if cursor.IsDir && state.IsOpen(cursor.Path) {
		// Collapse the directory
		return state.SetOpen(cursor.Path, false)
	} else if cursor.Parent != nil {
		// Move to parent
		return state.SetCursor(cursor.Parent.Path)
	}
	
	return state
}

// FindNodeByPath recursively searches for a node by its path
func FindNodeByPath(root *Node, path string) *Node {
	if root.Path == path {
		return root
	}
	
	for _, child := range root.Children {
		if found := FindNodeByPath(child, path); found != nil {
			return found
		}
	}
	
	return nil
}

func findNodeIndex(nodes []*Node, target *Node) int {
	for i, n := range nodes {
		if n == target {
			return i
		}
	}
	return -1
}