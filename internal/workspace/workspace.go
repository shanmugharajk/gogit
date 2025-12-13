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

// ListFiles recursively returns all files in the workspace, excluding ignored entries.
// Returns paths relative to the workspace root.
func (w *Workspace) ListFiles() ([]string, error) {
	return w.listFilesRecursive(w.pathname)
}

// listFilesRecursive is a helper function that recursively lists files in a directory.
func (w *Workspace) listFilesRecursive(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if ignoredEntries[entry.Name()] {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			subFiles, err := w.listFilesRecursive(fullPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			// Get relative path from workspace root.
			// e.g. ["/a/b", "/a/b/c.txt"] gives back c.txt
			relPath, err := filepath.Rel(w.pathname, fullPath)
			if err != nil {
				return nil, err
			}
			files = append(files, relPath)
		}
	}

	return files, nil
}

// ReadFile reads the contents of a file at the specified path relative to the workspace root.
func (w *Workspace) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(w.pathname, path)
	return os.ReadFile(fullPath)
}

// StatFile returns file information for a file at the specified path relative to the workspace root.
func (w *Workspace) StatFile(path string) (os.FileInfo, error) {
	fullPath := filepath.Join(w.pathname, path)
	return os.Stat(fullPath)
}
