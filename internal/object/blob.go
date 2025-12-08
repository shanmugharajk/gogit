package object

// Object defines the interface for git objects.
type Object interface {
	// Type returns the type of the object (e.g., "blob", "commit", "tree").
	Type() string

	// Bytes returns the raw data of the object.
	Bytes() []byte

	// SetOID sets the object identifier (SHA1 hash).
	SetOID(oid string)

	// GetOID returns the object identifier.
	GetOID() string
}

// Blob represents a git blob object containing arbitrary data.
type Blob struct {
	data []byte
	oid  string
}

// NewBlob creates a new Blob with the given data.
func NewBlob(data []byte) *Blob {
	return &Blob{
		data: data,
		oid:  "",
	}
}

// Type returns the type identifier for a blob object.
func (b *Blob) Type() string {
	return "blob"
}

// Bytes returns the blob's data.
func (b *Blob) Bytes() []byte {
	return b.data
}

// SetOID sets the object identifier for this blob.
func (b *Blob) SetOID(oid string) {
	b.oid = oid
}

// GetOID returns the object identifier of this blob.
func (b *Blob) GetOID() string {
	return b.oid
}
