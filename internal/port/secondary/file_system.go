package secondary

import (
	"io"
	"os"
)

// FileSystem defines the interface for file operations.
type FileSystem interface {
	// Exists checks if a path exists.
	Exists(path string) bool

	// IsDir checks if a path is a directory.
	IsDir(path string) bool

	// ReadFile reads a file and returns its contents.
	ReadFile(path string) ([]byte, error)

	// WriteFile writes content to a file.
	WriteFile(path string, content []byte, perm os.FileMode) error

	// MkdirAll creates a directory and all parents.
	MkdirAll(path string, perm os.FileMode) error

	// RemoveAll removes a path and all children.
	RemoveAll(path string) error

	// Copy copies a file from src to dst.
	Copy(src, dst string) error

	// Rename moves/renames a file.
	Rename(oldpath, newpath string) error

	// Open opens a file for reading.
	Open(path string) (io.ReadCloser, error)

	// Create creates a file for writing.
	Create(path string) (io.WriteCloser, error)

	// ReadDir reads a directory and returns entries.
	ReadDir(path string) ([]os.DirEntry, error)

	// Chmod changes the mode of a file.
	Chmod(path string, mode os.FileMode) error

	// Stat returns file info.
	Stat(path string) (os.FileInfo, error)
}
