package fs

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"github.com/eliooooooot/picky/internal/domain"
	"sort"
	"time"
)

// MemFileSystem is an in-memory filesystem for testing
type MemFileSystem struct {
	files map[string]*memFile
}

type memFile struct {
	name    string
	content []byte
	isDir   bool
	modTime time.Time
}

// NewMemFileSystem creates a new in-memory filesystem
func NewMemFileSystem() *MemFileSystem {
	return &MemFileSystem{
		files: make(map[string]*memFile),
	}
}

// AddFile adds a file to the filesystem
func (m *MemFileSystem) AddFile(path string, content string) {
	m.files[path] = &memFile{
		name:    filepath.Base(path),
		content: []byte(content),
		isDir:   false,
		modTime: time.Now(),
	}
	
	// Ensure parent directories exist
	m.ensureParentDirs(path)
}

// AddDir adds a directory to the filesystem
func (m *MemFileSystem) AddDir(path string) {
	m.files[path] = &memFile{
		name:    filepath.Base(path),
		isDir:   true,
		modTime: time.Now(),
	}
	
	// Ensure parent directories exist
	m.ensureParentDirs(path)
}

func (m *MemFileSystem) ensureParentDirs(path string) {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return
	}
	
	if _, exists := m.files[dir]; !exists {
		m.AddDir(dir)
	}
}

func (m *MemFileSystem) ReadDir(path string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	
	for p, f := range m.files {
		if filepath.Dir(p) == path && p != path {
			entries = append(entries, &memDirEntry{f})
		}
	}
	
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	
	return entries, nil
}

func (m *MemFileSystem) Open(path string) (io.ReadCloser, error) {
	f, exists := m.files[path]
	if !exists {
		return nil, fs.ErrNotExist
	}
	
	return io.NopCloser(bytes.NewReader(f.content)), nil
}

func (m *MemFileSystem) Stat(path string) (fs.FileInfo, error) {
	f, exists := m.files[path]
	if !exists {
		return nil, fs.ErrNotExist
	}
	
	return &memFileInfo{f, path}, nil
}

func (m *MemFileSystem) Create(path string) (io.WriteCloser, error) {
	buf := &writeBuffer{
		fs:   m,
		path: path,
	}
	
	m.ensureParentDirs(path)
	
	return buf, nil
}

// GetContent returns the content of a file (for testing)
func (m *MemFileSystem) GetContent(path string) (string, error) {
	f, exists := m.files[path]
	if !exists {
		return "", fs.ErrNotExist
	}
	
	return string(f.content), nil
}

// memDirEntry implements fs.DirEntry
type memDirEntry struct {
	f *memFile
}

func (e *memDirEntry) Name() string               { return e.f.name }
func (e *memDirEntry) IsDir() bool                { return e.f.isDir }
func (e *memDirEntry) Type() fs.FileMode          { 
	if e.f.isDir {
		return fs.ModeDir
	}
	return 0
}
func (e *memDirEntry) Info() (fs.FileInfo, error) { return &memFileInfo{e.f, ""}, nil }

// memFileInfo implements fs.FileInfo
type memFileInfo struct {
	f    *memFile
	path string
}

func (i *memFileInfo) Name() string       { return i.f.name }
func (i *memFileInfo) Size() int64        { return int64(len(i.f.content)) }
func (i *memFileInfo) Mode() fs.FileMode  { 
	if i.f.isDir {
		return fs.ModeDir | 0755
	}
	return 0644
}
func (i *memFileInfo) ModTime() time.Time { return i.f.modTime }
func (i *memFileInfo) IsDir() bool        { return i.f.isDir }
func (i *memFileInfo) Sys() interface{}   { return nil }

// writeBuffer implements io.WriteCloser
type writeBuffer struct {
	fs   *MemFileSystem
	path string
	buf  bytes.Buffer
}

func (w *writeBuffer) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func (w *writeBuffer) Close() error {
	w.fs.files[w.path] = &memFile{
		name:    filepath.Base(w.path),
		content: w.buf.Bytes(),
		isDir:   false,
		modTime: time.Now(),
	}
	return nil
}

// ReadFile reads the entire contents of a file
func (m *MemFileSystem) ReadFile(path string) ([]byte, error) {
	f, exists := m.files[path]
	if !exists {
		return nil, fs.ErrNotExist
	}
	if f.isDir {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fs.ErrInvalid}
	}
	return f.content, nil
}

// WriteFile writes data to a file, creating it if necessary
func (m *MemFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	m.ensureParentDirs(path)
	m.files[path] = &memFile{
		name:    filepath.Base(path),
		content: data,
		isDir:   false,
		modTime: time.Now(),
	}
	return nil
}

// MkdirAll creates a directory and all necessary parents
func (m *MemFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	// Normalize the path
	path = filepath.Clean(path)
	
	// Create all parent directories
	dir := path
	var dirs []string
	for dir != "/" && dir != "." {
		dirs = append([]string{dir}, dirs...)
		dir = filepath.Dir(dir)
	}
	
	// Create each directory if it doesn't exist
	for _, d := range dirs {
		if _, exists := m.files[d]; !exists {
			m.files[d] = &memFile{
				name:    filepath.Base(d),
				isDir:   true,
				modTime: time.Now(),
			}
		}
	}
	
	return nil
}

var _ domain.FileSystem = (*MemFileSystem)(nil)