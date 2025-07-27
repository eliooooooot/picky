package domain

import (
	"io"
	"io/fs"
)

// FileSystem abstracts file system operations
type FileSystem interface {
	ReadDir(path string) ([]fs.DirEntry, error)
	Open(path string) (io.ReadCloser, error)
	Stat(path string) (fs.FileInfo, error)
	Create(path string) (io.WriteCloser, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}