package file

import (
	"errors"
	"fmt"
	"os"
)

// Lockfile manages an atomic lock implementation for a file path.
type Lockfile struct {
	filePath string
	lockPath string
	file     *os.File
}

// New creates a lockfile wrapper for the given path.
func NewLockfile(path string) *Lockfile {
	return &Lockfile{
		filePath: path,
		lockPath: path + ".lock",
	}
}

// HoldForUpdate tries to acquire the lock for writing.
// It returns false when another process already holds the lock.
func (l *Lockfile) HoldForUpdate() (bool, error) {
	if l.file != nil {
		return true, nil
	}

	flags := os.O_RDWR | os.O_CREATE | os.O_EXCL
	f, err := os.OpenFile(l.lockPath, flags, ModeFile)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrExist):
			return false, nil
		case errors.Is(err, os.ErrNotExist):
			return false, fmt.Errorf("missing parent directory for lock file %s: %w", l.lockPath, err)
		case errors.Is(err, os.ErrPermission):
			return false, fmt.Errorf("permission denied for lock file %s: %w", l.lockPath, err)
		default:
			return false, err
		}
	}

	l.file = f
	return true, nil
}

// Write writes the supplied string to the locked file.
func (l *Lockfile) Write(content string) error {
	if err := l.ensureLock(); err != nil {
		return err
	}

	if _, err := l.file.WriteString(content); err != nil {
		return err
	}

	return nil
}

// WriteBytes writes the supplied bytes to the locked file.
func (l *Lockfile) WriteBytes(content []byte) error {
	if err := l.ensureLock(); err != nil {
		return err
	}

	if _, err := l.file.Write(content); err != nil {
		return err
	}

	return nil
}

// Commit finalizes the lock by closing the handle and renaming the lock file.
func (l *Lockfile) Commit() error {
	if err := l.ensureLock(); err != nil {
		return err
	}

	if err := l.file.Close(); err != nil {
		return err
	}

	if err := os.Rename(l.lockPath, l.filePath); err != nil {
		return err
	}

	l.file = nil
	return nil
}

func (l *Lockfile) ensureLock() error {
	if l.file == nil {
		return fmt.Errorf("not holding lock on file: %s", l.lockPath)
	}
	return nil
}
