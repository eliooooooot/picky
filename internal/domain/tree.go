package domain

import "path/filepath"

// Node represents a file or directory in the tree
type Node struct {
	Path     string
	Name     string
	IsDir    bool
	Parent   *Node
	Children []*Node
}

// Tree represents the file tree
type Tree struct {
	Root *Node
}

// NewTree creates a new tree
func NewTree(root *Node) *Tree {
	return &Tree{
		Root: root,
	}
}

// Flatten returns all nodes in the tree in depth-first order
func (t *Tree) Flatten() []*Node {
	var result []*Node
	flattenNode(t.Root, &result)
	return result
}

func flattenNode(node *Node, result *[]*Node) {
	*result = append(*result, node)
	for _, child := range node.Children {
		flattenNode(child, result)
	}
}

// ExcludeNode removes the given node from the tree
// Returns the relative path of the excluded node and the removed node itself
// Returns empty string and nil if it was the root or not found
func (t *Tree) ExcludeNode(nodePath string) (string, *Node) {
	if nodePath == t.Root.Path {
		return "", nil
	}
	
	node := FindNodeByPath(t.Root, nodePath)
	if node == nil || node.Parent == nil {
		return "", nil
	}
	
	// Calculate relative path
	relPath, err := filepath.Rel(t.Root.Path, node.Path)
	if err != nil {
		return "", nil
	}
	
	// Remove from parent's children
	parent := node.Parent
	for i, child := range parent.Children {
		if child == node {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}
	
	// Clean up parent reference to help GC
	node.Parent = nil
	
	return relPath, node
}