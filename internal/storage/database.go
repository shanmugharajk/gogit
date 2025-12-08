package storage

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"

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
	// Build the content string: "type size\0data"
	objType := obj.Type()
	data := obj.Bytes()
	size := len(data)

	// Create the content with null terminator
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s %d", objType, size)
	buf.WriteByte(0)
	buf.Write(data)
	content := buf.Bytes()

	// Compute SHA1 hash
	hash := sha1.Sum(content)
	oid := fmt.Sprintf("%x", hash)

	// Write object to disk
	if err := db.writeObject(oid, content); err != nil {
		return err
	}

	// Set the OID on the object
	obj.SetOID(oid)

	return nil
}

// writeObject writes a git object to disk with atomic writes.
// The object is stored at pathname/XX/YYYYYYY where XX are the first 2 hex chars of the OID
// and YYYYYYY are the remaining hex chars.
func (db *Database) writeObject(oid string, content []byte) error {
	// Compute object path: first 2 chars as directory, rest as filename
	objDir := filepath.Join(db.pathname, oid[:2])
	objPath := filepath.Join(objDir, oid[2:])

	// Create a temporary file path in the target directory
	tempName := db.generateTempName()
	tempPath := filepath.Join(objDir, tempName)

	// Ensure directory exists, create if necessary
	if err := os.MkdirAll(objDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temporary file with exclusive creation flag
	flags := os.O_RDWR | os.O_CREATE | os.O_EXCL
	tempFile, err := os.OpenFile(tempPath, flags, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()

	// Compress content and write to temporary file
	if err := db.compressAndWrite(tempFile, content); err != nil {
		// Clean up the temporary file on error
		_ = os.Remove(tempPath)
		return err
	}

	// Atomic rename: move temp file to final location
	if err := os.Rename(tempPath, objPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// compressAndWrite compresses the content using zlib deflate and writes to the writer.
func (db *Database) compressAndWrite(w io.Writer, content []byte) error {
	// Use zlib/deflate compression matching Ruby's Zlib::Deflate.deflate
	encoder := zlib.NewWriter(w)
	defer encoder.Close()

	if _, err := encoder.Write(content); err != nil {
		return fmt.Errorf("failed to compress content: %w", err)
	}

	return nil
}

// generateTempName generates a random temporary filename.
// Format: tmp_obj_XXXXXX where X is a random alphanumeric character.
func (db *Database) generateTempName() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tempName := "tmp_obj_"
	for i := 0; i < 6; i++ {
		tempName += string(chars[rand.Intn(len(chars))])
	}
	return tempName
}
