package file

import "os"

const (
	ModeFile = os.FileMode(0o644)
	ModeDir  = os.FileMode(0o755)
)
