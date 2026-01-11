// Package filesystem provides file system adapters.
package filesystem

import (
	"fmt"
	"io"
	"os"

	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// LocalFileSystem implements FileSystem using the local file system.
type LocalFileSystem struct {
	dryRun   bool
	reporter secondary.ProgressReporter
}

// NewLocalFileSystem creates a new LocalFileSystem instance.
func NewLocalFileSystem(dryRun bool, reporter secondary.ProgressReporter) *LocalFileSystem {
	return &LocalFileSystem{
		dryRun:   dryRun,
		reporter: reporter,
	}
}

// Exists checks if a path exists.
func (fs *LocalFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory.
func (fs *LocalFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFile reads a file and returns its contents.
func (fs *LocalFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes content to a file.
func (fs *LocalFileSystem) WriteFile(path string, content []byte, perm os.FileMode) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] write file %s", path))
		}
		return nil
	}
	return os.WriteFile(path, content, perm)
}

// MkdirAll creates a directory and all parents.
func (fs *LocalFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] mkdir -p %s", path))
		}
		return nil
	}
	return os.MkdirAll(path, perm)
}

// RemoveAll removes a path and all children.
func (fs *LocalFileSystem) RemoveAll(path string) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] rm -rf %s", path))
		}
		return nil
	}
	return os.RemoveAll(path)
}

// Copy copies a file from src to dst.
func (fs *LocalFileSystem) Copy(src, dst string) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] cp %s %s", src, dst))
		}
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Rename moves/renames a file.
func (fs *LocalFileSystem) Rename(oldpath, newpath string) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] mv %s %s", oldpath, newpath))
		}
		return nil
	}
	return os.Rename(oldpath, newpath)
}

// Open opens a file for reading.
func (fs *LocalFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Create creates a file for writing.
func (fs *LocalFileSystem) Create(path string) (io.WriteCloser, error) {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] create %s", path))
		}
		return &nopWriteCloser{}, nil
	}
	return os.Create(path)
}

// ReadDir reads a directory and returns entries.
func (fs *LocalFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// Chmod changes the mode of a file.
func (fs *LocalFileSystem) Chmod(path string, mode os.FileMode) error {
	if fs.dryRun {
		if fs.reporter != nil {
			fs.reporter.Info(fmt.Sprintf("[DRY RUN] chmod %o %s", mode, path))
		}
		return nil
	}
	return os.Chmod(path, mode)
}

// Stat returns file info.
func (fs *LocalFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// nopWriteCloser is a no-op WriteCloser for dry-run mode.
type nopWriteCloser struct{}

func (n *nopWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (n *nopWriteCloser) Close() error {
	return nil
}
