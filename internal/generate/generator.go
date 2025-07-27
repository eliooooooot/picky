package generate

import (
	"fmt"
	"github.com/eliotsamuelmiller/picky/internal/domain"
)

// Generate creates the output file with selected files using TextWriter
func Generate(outPath, prompt string, tree *domain.Tree, state domain.ViewState, fs domain.FileSystem) error {
	w, err := fs.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer w.Close()
	
	// Write prompt first if non-empty
	if prompt != "" {
		fmt.Fprintln(w, "# Prompt")
		fmt.Fprintln(w)
		fmt.Fprintln(w, prompt)
		fmt.Fprintln(w) // extra blank line
	}
	
	// Get all selected paths
	paths := domain.GetSelectedPaths(tree.Root, state)
	if len(paths) == 0 {
		_, err = w.Write([]byte("No files selected\n"))
		return err
	}
	
	// Use TextWriter for output
	writer := NewTextWriter()
	
	// Write directory structure
	if err := writer.WriteStructure(w, tree.Root, state); err != nil {
		return err
	}
	
	// Write file contents
	if err := writer.WriteContent(w, paths, fs); err != nil {
		return err
	}
	
	return nil
}