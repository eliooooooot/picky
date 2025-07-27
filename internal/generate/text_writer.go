package generate

import (
	"fmt"
	"io"
	"path/filepath"
	"github.com/eliotsamuelmiller/picky/internal/domain"
	"strings"
)

// TextWriter implements domain.OutputWriter for text output
type TextWriter struct{}

// NewTextWriter creates a new text writer
func NewTextWriter() *TextWriter {
	return &TextWriter{}
}

// WriteStructure writes the directory structure in text format
func (tw *TextWriter) WriteStructure(w io.Writer, root *domain.Node, state domain.ViewState) error {
	if _, err := fmt.Fprintln(w, "# Directory Structure"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	
	// Build a simple tree representation
	if err := tw.writeNodeStructure(w, root, state, "", true); err != nil {
		return err
	}
	
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

func (tw *TextWriter) writeNodeStructure(w io.Writer, node *domain.Node, state domain.ViewState, prefix string, isLast bool) error {
	if node.Parent != nil { // Skip root node name in structure
		// Determine the prefix for this line
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		
		line := prefix + marker + node.Name
		if state.IsSelected(node.Path) && !node.IsDir {
			line += " *"
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
		
		// Update prefix for children
		if isLast {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}
	
	// Write children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		if err := tw.writeNodeStructure(w, child, state, prefix, isLastChild); err != nil {
			return err
		}
	}
	
	return nil
}

// WriteContent writes the content of selected files
func (tw *TextWriter) WriteContent(w io.Writer, paths []string, fs domain.FileSystem) error {
	if _, err := fmt.Fprintln(w, "# Selected Files"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	
	for _, path := range paths {
		// Write file header
		if _, err := fmt.Fprintf(w, "## %s\n\n", filepath.Base(path)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "```"); err != nil {
			return err
		}
		
		// Read and write file content
		f, err := fs.Open(path)
		if err != nil {
			if _, err := fmt.Fprintf(w, "Error reading file: %v\n", err); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, "```"); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			continue
		}
		
		content, err := io.ReadAll(f)
		f.Close()
		
		if err != nil {
			if _, err := fmt.Fprintf(w, "Error reading file: %v\n", err); err != nil {
				return err
			}
		} else {
			if _, err := w.Write(content); err != nil {
				return err
			}
			if !strings.HasSuffix(string(content), "\n") {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
		}
		
		if _, err := fmt.Fprintln(w, "```"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	
	return nil
}