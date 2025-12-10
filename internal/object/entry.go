package object

type Entry struct {
	Name string
	OID  string
}

func NewEntry(name string, oid string) *Entry {
	return &Entry{
		Name: name,
		OID:  oid,
	}
}
