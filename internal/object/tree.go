package object

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
)

// TreeEntry represents an entry in a tree, which can be either an Entry or another Tree.
type TreeEntry interface {
	Basename() string
	Mode() string
	GetOID() string
}

type Tree struct {
	oid     string
	entries map[string]TreeEntry
}

func NewTree() *Tree {
	return &Tree{
		entries: make(map[string]TreeEntry),
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

// Basename returns the base name of the tree.
func (t *Tree) Basename() string {
	// This is needed for TreeEntry interface
	return ""
}

// Mode returns the directory mode for trees.
func (t *Tree) Mode() string {
	return DirectoryMode
}

// AddEntry adds an entry to the tree, creating intermediate trees as needed.
func (t *Tree) AddEntry(parents []string, entry *Entry) {
	if len(parents) == 0 {
		t.entries[entry.Basename()] = entry
		return
	}

	// Get or create the subtree
	firstParent := filepath.Base(parents[0])
	subtree, exists := t.entries[firstParent]
	if !exists {
		subtree = NewTree()
		t.entries[firstParent] = subtree
	}

	// Recursively add to subtree
	if tree, ok := subtree.(*Tree); ok {
		tree.AddEntry(parents[1:], entry)
	}
}

// Traverse visits all trees in depth-first order, calling the provided function for each.
func (t *Tree) Traverse(fn func(*Tree)) {
	// First traverse all subtrees
	for _, entry := range t.entries {
		if tree, ok := entry.(*Tree); ok {
			tree.Traverse(fn)
		}
	}
	// Then call the function on this tree
	fn(t)
}

func (t *Tree) Bytes() []byte {
	// Get sorted entry names
	names := make([]string, 0, len(t.entries))
	for name := range t.entries {
		names = append(names, name)
	}
	sort.Strings(names)

	var result []byte

	for _, name := range names {
		entry := t.entries[name]
		// "<mode> <name>\0"
		result = fmt.Appendf(result, "%s %s\x00", entry.Mode(), name)

		// raw sha1 bytes (20 bytes)
		oidBytes, _ := hex.DecodeString(entry.GetOID())
		result = append(result, oidBytes...)
	}

	return result
}

// Build constructs a tree hierarchy from a flat list of entries.
func Build(entries []*Entry) *Tree {
	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	root := NewTree()

	for _, entry := range entries {
		root.AddEntry(entry.ParentDirectories(), entry)
	}

	return root
}
