package object

import (
	"os"
	"path/filepath"
)

const (
	RegularMode    = "100644"
	ExecutableMode = "100755"
	DirectoryMode  = "40000"
)

type Entry struct {
	Name string
	OID  string
	stat os.FileInfo
}

func NewEntry(name string, oid string, stat os.FileInfo) *Entry {
	return &Entry{
		Name: name,
		OID:  oid,
		stat: stat,
	}
}

// Basename returns the base name of the entry path.
func (e *Entry) Basename() string {
	return filepath.Base(e.Name)
}

// GetOID returns the object ID (SHA-1 hash) of the entry.
func (e *Entry) GetOID() string {
	return e.OID
}

// Mode returns the file mode as a string (e.g., "100644" or "100755").
func (e *Entry) Mode() string {
	if e.stat.Mode()&0111 != 0 {
		return ExecutableMode
	}
	return RegularMode
}

// ParentDirectories returns a slice of parent directories for this entry.
// For example, "a/b/c.txt" would return ["a", "a/b"].
func (e *Entry) ParentDirectories() []string {
	dir := filepath.Dir(e.Name)
	if dir == "." || dir == "/" {
		return []string{}
	}

	var parents []string
	current := dir
	for current != "." && current != "/" {
		parents = append([]string{current}, parents...)
		current = filepath.Dir(current)
	}

	return parents
}
