package index

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"os"

	"github.com/shanmugharajk/gogit/internal/file"
)

const (
	// HeaderFormat: "DIRC" (4 bytes) + version (2 uint32s) + entry count (uint32)
	// "DIRC" = 4 bytes ASCII
	// Version = 2 (uint32, big-endian)
	// Entry count = uint32 (big-endian)
	headerSize = 4 + 4 + 4 // "DIRC" + version + entry count
)

// Index represents the git index file.
type Index struct {
	entries  map[string]*Entry
	lockfile *file.Lockfile
	digest   hash.Hash
}

// New creates a new Index at the specified pathname.
func New(pathname string) *Index {
	return &Index{
		entries:  make(map[string]*Entry),
		lockfile: file.NewLockfile(pathname),
	}
}

// Add adds an entry to the index.
func (idx *Index) Add(pathname string, oid string, stat os.FileInfo) {
	entry := CreateEntry(pathname, oid, stat)
	idx.entries[pathname] = entry
}

// WriteUpdates writes the index to disk.
func (idx *Index) WriteUpdates() error {
	held, err := idx.lockfile.HoldForUpdate()
	if err != nil {
		return fmt.Errorf("failed to hold lock: %w", err)
	}
	if !held {
		return fmt.Errorf("index lock is held by another process")
	}

	// Begin write - initialize SHA1 digest
	idx.digest = sha1.New()

	// Write header: "DIRC" + version (2) + entry count
	header := make([]byte, headerSize)
	copy(header[0:4], []byte("DIRC"))
	binary.BigEndian.PutUint32(header[4:8], 2) // version
	binary.BigEndian.PutUint32(header[8:12], uint32(len(idx.entries)))

	if err := idx.write(header); err != nil {
		return err
	}

	// Write entries
	for _, entry := range idx.entries {
		entryBytes := entry.Bytes()
		if err := idx.write(entryBytes); err != nil {
			return err
		}
	}

	// Finish write - append SHA1 checksum
	checksum := idx.digest.Sum(nil)
	if err := idx.lockfile.WriteBytes(checksum); err != nil {
		return fmt.Errorf("failed to write checksum: %w", err)
	}

	if err := idx.lockfile.Commit(); err != nil {
		return fmt.Errorf("failed to commit index: %w", err)
	}

	return nil
}

// write writes data to the lockfile and updates the digest.
func (idx *Index) write(data []byte) error {
	if err := idx.lockfile.WriteBytes(data); err != nil {
		return err
	}
	idx.digest.Write(data)
	return nil
}
