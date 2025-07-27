package domain

import "io"

// StructureWriter writes the directory structure to an output
type StructureWriter interface {
	WriteStructure(w io.Writer, root *Node, state ViewState) error
}

// ContentWriter writes file contents to an output
type ContentWriter interface {
	WriteContent(w io.Writer, paths []string, fs FileSystem) error
}

// OutputWriter combines structure and content writing
type OutputWriter interface {
	StructureWriter
	ContentWriter
}