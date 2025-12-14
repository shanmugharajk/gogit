package index

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"syscall"
)

const (
	// RegularMode is the mode for regular files (octal 0100644 = decimal 33188)
	RegularMode = 0o100644
	// ExecutableMode is the mode for executable files (octal 0100755 = decimal 33261)
	ExecutableMode = 0o100755
	// MaxPathSize is the maximum path size (4095 bytes)
	MaxPathSize = 0xfff
	// EntryBlock is the block size for entry padding (8 bytes)
	EntryBlock = 8
)

// Entry represents an entry in the git index file.
type Entry struct {
	CTime     int64
	CTimeNsec int32
	MTime     int64
	MTimeNsec int32
	Dev       uint32
	Ino       uint32
	Mode      uint32
	UID       uint32
	GID       uint32
	Size      uint32
	OID       string
	Flags     uint16
	Path      string
}

// CreateEntry creates a new Entry from a pathname, OID, and file stat.
func CreateEntry(pathname string, oid string, stat os.FileInfo) *Entry {
	path := pathname
	mode := RegularMode
	if stat.Mode()&0111 != 0 {
		mode = ExecutableMode
	}

	flags := uint16(len(path))
	if flags > MaxPathSize {
		flags = MaxPathSize
	}

	var ctime, mtime int64
	var ctimeNsec, mtimeNsec int32
	var dev, ino, uid, gid uint32

	// Try to extract system-specific stat info
	sys := stat.Sys()
	if sysStat, ok := sys.(*syscall.Stat_t); ok {
		// Use syscall.Stat_t for Unix systems
		ctime = sysStat.Ctimespec.Sec
		ctimeNsec = int32(sysStat.Ctimespec.Nsec)
		mtime = sysStat.Mtimespec.Sec
		mtimeNsec = int32(sysStat.Mtimespec.Nsec)
		dev = uint32(sysStat.Dev)
		ino = uint32(sysStat.Ino)
		uid = sysStat.Uid
		gid = sysStat.Gid
	} else {
		// Fallback to ModTime if syscall.Stat_t is not available
		modTime := stat.ModTime()
		ctime = modTime.Unix()
		mtime = modTime.Unix()
		ctimeNsec = int32(modTime.Nanosecond())
		mtimeNsec = int32(modTime.Nanosecond())
	}

	return &Entry{
		CTime:     ctime,
		CTimeNsec: ctimeNsec,
		MTime:     mtime,
		MTimeNsec: mtimeNsec,
		Dev:       dev,
		Ino:       ino,
		Mode:      uint32(mode),
		UID:       uid,
		GID:       gid,
		Size:      uint32(stat.Size()),
		OID:       oid,
		Flags:     flags,
		Path:      path,
	}
}

// Bytes returns the binary representation of the entry.
// Format: N10H40nZ* (10 uint32s, 40 hex bytes, 1 uint16, null-terminated string)
// Padded to multiples of EntryBlock (8) bytes.
func (e *Entry) Bytes() []byte {
	buf := make([]byte, 0, 64+len(e.Path)+8)

	// Write 10 uint32 values (big-endian)
	buf = binary.BigEndian.AppendUint32(buf, uint32(e.CTime))
	buf = binary.BigEndian.AppendUint32(buf, uint32(e.CTimeNsec))
	buf = binary.BigEndian.AppendUint32(buf, uint32(e.MTime))
	buf = binary.BigEndian.AppendUint32(buf, uint32(e.MTimeNsec))
	buf = binary.BigEndian.AppendUint32(buf, e.Dev)
	buf = binary.BigEndian.AppendUint32(buf, e.Ino)
	buf = binary.BigEndian.AppendUint32(buf, e.Mode)
	buf = binary.BigEndian.AppendUint32(buf, e.UID)
	buf = binary.BigEndian.AppendUint32(buf, e.GID)
	buf = binary.BigEndian.AppendUint32(buf, e.Size)

	// Write OID as 40 hex bytes
	oidBytes, err := hex.DecodeString(e.OID)
	if err != nil {
		// If OID is not valid hex, pad with zeros
		oidBytes = make([]byte, 20)
	}
	if len(oidBytes) < 20 {
		// Pad to 20 bytes if needed
		padded := make([]byte, 20)
		copy(padded, oidBytes)
		oidBytes = padded
	}
	buf = append(buf, oidBytes...)

	// Write flags as uint16 (big-endian)
	buf = binary.BigEndian.AppendUint16(buf, e.Flags)

	// Write path as null-terminated string
	buf = append(buf, []byte(e.Path)...)
	buf = append(buf, 0) // null terminator

	// Pad to multiple of EntryBlock (8) bytes
	for len(buf)%EntryBlock != 0 {
		buf = append(buf, 0)
	}

	return buf
}
