package storage

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shanmugharajk/gogit/internal/file"
	"github.com/shanmugharajk/gogit/internal/object"
)

// Database manages the storage of git objects on disk.
type Database struct {
	pathname string
}

// New creates a new Database at the specified pathname.
func New(pathname string) *Database {
	return &Database{
		pathname: pathname,
	}
}

// Store persists a git object to the database and sets its OID.
// The OID is computed as SHA1("type size\0data").
func (db *Database) Store(obj object.Object) error {
	data := obj.Bytes()

	// Create the content with null terminator: "type size\0data"
	header := fmt.Sprintf("%s %d\x00", obj.Type(), len(data))
	content := append([]byte(header), data...)

	// Compute SHA1 hash and set OID
	hash := sha1.Sum(content)
	oid := fmt.Sprintf("%x", hash)
	obj.SetOID(oid)

	// Write object to disk
	return db.writeObject(oid, content)
}

// writeObject writes a git object to disk with atomic writes.
// The object is stored at pathname/XX/YYYYYYY where XX are the first 2 hex chars of the OID
// and YYYYYYY are the remaining hex chars.
func (db *Database) writeObject(oid string, content []byte) error {
	// Compute object path: first 2 chars as directory, rest as filename
	objDir := filepath.Join(db.pathname, oid[:2])
	objPath := filepath.Join(objDir, oid[2:])

	// Ensure directory exists, create if necessary
	if err := os.MkdirAll(objDir, file.ModeDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temporary file in the object directory
	tempFile, err := os.CreateTemp(objDir, "tmp_obj_")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Compress content and write to temporary file
	if err := db.compressAndWrite(tempFile, content); err != nil {
		return err
	}

	// Sync to ensure data is written to disk before renaming
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the file before renaming
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomic rename: move temp file to final location
	if err := os.Rename(tempFile.Name(), objPath); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// compressAndWrite compresses the content using zlib deflate and writes to the writer.
func (db *Database) compressAndWrite(w io.Writer, content []byte) error {
	encoder := zlib.NewWriter(w)
	defer encoder.Close()

	_, err := encoder.Write(content)
	return err
}
