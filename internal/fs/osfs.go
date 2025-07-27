package fs

import (
	"io"
	"io/fs"
	"os"
	"github.com/eliotsamuelmiller/picky/internal/domain"
)

// OSFileSystem implements domain.FileSystem using the real OS
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS-based filesystem
func NewOSFileSystem() domain.FileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

func (f *OSFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (f *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func (f *OSFileSystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}