package object

import (
	"encoding/hex"
	"fmt"
	"sort"
)

const (
	Mode = "100644"
)

type Tree struct {
	oid     string
	Entries []Entry
}

func NewTree(entries []Entry) *Tree {
	return &Tree{
		Entries: entries,
	}
}

func (t *Tree) Type() string {
	return "tree"
}

func (t *Tree) SetOID(oid string) {
	t.oid = oid
}

func (t *Tree) GetOID() string {
	return t.oid
}

func (t *Tree) Bytes() []byte {
	sorted := t.sortedEntries()

	var result []byte

	for _, e := range sorted {
		// "<mode> <name>\0"
		result = fmt.Appendf(result, "%s %s\x00", Mode, e.Name)

		// raw sha1 bytes (20 bytes)
		oidBytes, _ := hex.DecodeString(e.OID)
		result = append(result, oidBytes...)
	}

	return result
}

func (t *Tree) sortedEntries() []Entry {
	sorted := make([]Entry, len(t.Entries))
	copy(sorted, t.Entries)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}
