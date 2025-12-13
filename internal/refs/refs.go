package refs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shanmugharajk/gogit/internal/file"
)

// Refs manages reference updates such as HEAD.
type Refs struct {
	gitPath string
}

// New creates a new Refs manager scoped to the given .git directory.
func New(gitPath string) *Refs {
	return &Refs{gitPath: gitPath}
}

// UpdateHead updates the HEAD reference with the supplied commit OID.
func (r *Refs) UpdateHead(oid string) error {
	headPath := r.headPath()
	lock := file.NewLockfile(headPath)

	acquired, err := lock.HoldForUpdate()
	if err != nil {
		return err
	}
	if !acquired {
		return &LockDeniedError{Path: headPath}
	}

	if err := lock.Write(oid + "\n"); err != nil {
		return err
	}

	return lock.Commit()
}

// ReadHead returns the currently stored HEAD OID, or empty string when it does not exist.
func (r *Refs) ReadHead() (string, error) {
	headPath := r.headPath()
	data, err := os.ReadFile(headPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (r *Refs) headPath() string {
	return filepath.Join(r.gitPath, "HEAD")
}

// LockDeniedError indicates that the HEAD lock could not be acquired.
type LockDeniedError struct {
	Path string
}

func (e *LockDeniedError) Error() string {
	return fmt.Sprintf("could not acquire lock on file: %s", e.Path)
}
