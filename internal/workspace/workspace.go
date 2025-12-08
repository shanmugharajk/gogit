package workspace

import (
	"os"
	"path/filepath"
)

// ignoredEntries are the entries that should not be included when listing files.
var ignoredEntries = map[string]bool{
	".":    true,
	"..":   true,
	".git": true,
}

// Workspace represents the working directory of a git repository.
type Workspace struct {
	pathname string
}

// New creates a new Workspace at the specified path.
func New(pathname string) *Workspace {
	return &Workspace{
		pathname: pathname,
	}
}

// ListFiles returns all files in the workspace, excluding ignored entries.
func (w *Workspace) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(w.pathname)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !ignoredEntries[entry.Name()] {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ReadFile reads the contents of a file at the specified path relative to the workspace root.
func (w *Workspace) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(w.pathname, path)
	return os.ReadFile(fullPath)
}
