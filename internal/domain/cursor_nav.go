package domain

// NextCursorAfterRemoval determines the next cursor position after a node is removed
// Parameters:
//   - flatBefore: flattened list of nodes before removal
//   - removedIdx: index of the removed node in flatBefore
//   - flatAfter: flattened list of nodes after removal
// Returns the path of the node that should become the new cursor
func NextCursorAfterRemoval(flatBefore []*Node, removedIdx int, flatAfter []*Node) string {
	if len(flatAfter) == 0 {
		// No nodes left, this shouldn't happen in practice
		return ""
	}
	
	// If we removed the last item, move to the new last item
	if removedIdx >= len(flatAfter) {
		return flatAfter[len(flatAfter)-1].Path
	}
	
	// If we removed the first item or something at the beginning
	if removedIdx == 0 {
		return flatAfter[0].Path
	}
	
	// For items in the middle, prefer the previous sibling
	// The item that was at removedIdx-1 should still be at removedIdx-1
	if removedIdx-1 < len(flatAfter) {
		return flatAfter[removedIdx-1].Path
	}
	
	// Fallback to first item
	return flatAfter[0].Path
}