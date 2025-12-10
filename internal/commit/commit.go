package commit

import "fmt"

type Commit struct {
	oid string

	TreeOID string
	Author  *Author
	Message string
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

	return fmt.Appendf(nil, "tree %s\nauthor %s\ncommitter %s\n\n%s",
		c.TreeOID,
		string(authorBytes),
		string(authorBytes),
		c.Message)
}

func NewCommit(treeOID string, author *Author, message string) *Commit {
	return &Commit{
		TreeOID: treeOID,
		Author:  author,
		Message: message,
	}
}
