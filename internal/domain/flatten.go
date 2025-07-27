package domain

// Flatten returns a depth-first slice of all visible nodes
func Flatten(root *Node, state ViewState) []*Node {
	var result []*Node
	flatten(root, state, &result)
	return result
}

func flatten(node *Node, state ViewState, result *[]*Node) {
	*result = append(*result, node)
	
	if node.IsDir && state.IsOpen(node.Path) {
		for _, child := range node.Children {
			flatten(child, state, result)
		}
	}
}