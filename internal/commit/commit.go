package commit

import "fmt"

type Commit struct {
	oid string

	ParentOID string
	TreeOID   string
	Author    *Author
	Message   string
}

func (c *Commit) SetOID(oid string) {
	c.oid = oid
}

func (c *Commit) GetOID() string {
	return c.oid
}

func (c *Commit) Type() string {
	return "commit"
}

func (c *Commit) Bytes() []byte {
	authorBytes := c.Author.Bytes()

	var result []byte
	result = fmt.Appendf(result, "tree %s\n", c.TreeOID)

	if c.ParentOID != "" {
		result = fmt.Appendf(result, "parent %s\n", c.ParentOID)
	}

	result = fmt.Appendf(result, "author %s\ncommitter %s\n\n%s",
		string(authorBytes),
		string(authorBytes),
		c.Message)

	return result
}

func NewCommit(parentOID string, treeOID string, author *Author, message string) *Commit {
	return &Commit{
		ParentOID: parentOID,
		TreeOID:   treeOID,
		Author:    author,
		Message:   message,
	}
}
