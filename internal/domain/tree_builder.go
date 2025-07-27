package domain

import (
	"path/filepath"
	"sort"
)

// PathFilter is a predicate function that determines whether a path should be included
// Returns false to skip the path
type PathFilter func(absolutePath string, isDir bool) bool

// BuildTree creates a tree from a filesystem starting at rootPath
func BuildTree(fs FileSystem, rootPath string) (*Tree, error) {
	return BuildTreeWithFilter(fs, rootPath, nil)
}

// BuildTreeWithFilter creates a tree from a filesystem with an optional filter predicate
func BuildTreeWithFilter(fs FileSystem, rootPath string, keep PathFilter) (*Tree, error) {
	root, err := buildNodeWithFilter(fs, rootPath, nil, keep)
	if err != nil {
		return nil, err
	}
	
	return NewTree(root), nil
}

func buildNode(fs FileSystem, path string, parent *Node) (*Node, error) {
	return buildNodeWithFilter(fs, path, parent, nil)
}

func buildNodeWithFilter(fs FileSystem, path string, parent *Node, keep PathFilter) (*Node, error) {
	info, err := fs.Stat(path)
	if err != nil {
		return nil, err
	}
	
	node := &Node{
		Path:   path,
		Name:   filepath.Base(path),
		IsDir:  info.IsDir(),
		Parent: parent,
	}
	
	if node.IsDir {
		entries, err := fs.ReadDir(path)
		if err != nil {
			// Skip directories we can't read
			return node, nil
		}
		
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			
			// Apply filter if provided
			if keep != nil && !keep(childPath, entry.IsDir()) {
				continue
			}
			
			child, err := buildNodeWithFilter(fs, childPath, node, keep)
			if err != nil {
				continue // Skip files we can't access
			}
			node.Children = append(node.Children, child)
		}
		
		// Sort children: directories first, then files, both alphabetically
		sort.Slice(node.Children, func(i, j int) bool {
			a, b := node.Children[i], node.Children[j]
			if a.IsDir != b.IsDir {
				return a.IsDir
			}
			return a.Name < b.Name
		})
	}
	
	return node, nil
}