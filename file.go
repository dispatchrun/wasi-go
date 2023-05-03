package wasi

import "fmt"

// FD is a file descriptor handle.
type FD int32

// INode is a file serial number that is unique within its file system.
type INode uint64

// Device is an identifier for a device containing a file system.
//
// It can be used in combination with INode to uniquely identify a file or
// directory in the filesystem.
type Device uint64

// FileSize is a non-negative file size or length of a region within a file.
type FileSize uint64

// FileDelta is a relative offset within a file.
type FileDelta int64

// LinkCount is the number of hard links to an INode.
type LinkCount uint64

// FileStat are file attributes.
type FileStat struct {
	// Device is the ID of the device containing the file.
	Device Device

	// INode is the file serial number.
	INode INode

	// FileType is the file type.
	FileType FileType

	// NLink is the number of hard links to the file.
	NLink LinkCount

	// Size is the size. For regular files, it's the file size in bytes. For
	// symbolic links, the length in bytes of the pathname contained in the
	// symbolic link.
	Size FileSize

	// AccessTime is the last data access timestamp.
	AccessTime Timestamp

	// ModifyTime is the last data modification timestamp.
	ModifyTime Timestamp

	// ChangeTime is the last file status change timestamp.
	ChangeTime Timestamp
}

// Whence is the position relative to which to set the offset of the file
// descriptor.
type Whence uint8

const (
	// SeekStart seeks relative to the origin of the file.
	SeekStart Whence = iota

	// SeekCurrent seeks relative to current offset.
	SeekCurrent

	// SeekEnd seeks relative to end of the file.
	SeekEnd
)

func (w Whence) String() string {
	switch w {
	case SeekStart:
		return "SeekStart"
	case SeekCurrent:
		return "SeekCurrent"
	case SeekEnd:
		return "SeekEnd"
	default:
		return fmt.Sprintf("Whence(%d)", w)
	}
}

// FileType is the type of a file descriptor or file.
type FileType uint8

const (
	// UnknownType indicates that the type of the file descriptor or file is
	// unknown or is different from any of the other types specified.
	UnknownType FileType = iota

	// BlockDeviceType is indicates that the file descriptor or file refers to
	// a block device inode.
	BlockDeviceType

	// CharacterDeviceType indicates that the file descriptor or file refers to
	// a character device inode.
	CharacterDeviceType

	// DirectoryType indicates that the file descriptor or file refers to a
	// directory inode.
	DirectoryType

	// RegularFileType indicates that the file descriptor or file refers to a
	// regular file inode.
	RegularFileType

	// SocketDGramType indicates that the file descriptor or file refers to a
	// datagram socket.
	SocketDGramType

	// SocketStreamType indicates that the file descriptor or file refers to a
	// byte-stream socket.
	SocketStreamType

	// SymbolicLinkType indicates that the file refers to a symbolic link
	// inode.
	SymbolicLinkType
)

func (f FileType) String() string {
	switch f {
	case UnknownType:
		return "UnknownType"
	case BlockDeviceType:
		return "BlockDeviceType"
	case CharacterDeviceType:
		return "CharacterDeviceType"
	case DirectoryType:
		return "DirectoryType"
	case RegularFileType:
		return "RegularFileType"
	case SocketDGramType:
		return "SocketDGramType"
	case SocketStreamType:
		return "SocketStreamType"
	case SymbolicLinkType:
		return "SymbolicLinkType"
	default:
		return fmt.Sprintf("FileType(%d)", f)
	}
}

// FDFlags are file descriptor flags.
type FDFlags uint16

const (
	// Append indicates append mode; data written to the file is always
	// appended to the file's end.
	Append FDFlags = 1 << iota

	// DSync means write according to synchronized I/O data integrity
	// completion.
	//
	// Only the data stored in the file is synchronized.
	DSync

	// NonBlock indicates non-blocking mode.
	NonBlock

	// RSync indicates synchronized read I/O operations.
	RSync

	// Sync means write according to synchronized I/O data integrity
	// completion.
	//
	// In addition to synchronizing the data stored in the file, the
	// implementation may also synchronously update the file's metadata.
	Sync
)

// Has is true if the flag is set.
func (flags FDFlags) Has(f FDFlags) bool {
	return (flags & f) == f
}

var fdflagsStrings = [...]string{
	"Append",
	"DSync",
	"NonBlock",
	"RSync",
	"Sync",
}

func (flags FDFlags) String() (s string) {
	if flags == 0 {
		return "FDFlags(0)"
	}
	for i, name := range fdflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("FDFlags(%d)", flags)
	}
	return
}

// FDStat is file descriptor attributes.
type FDStat struct {
	// FileType is the file type.
	FileType FileType

	// Flags are the file descriptor flags.
	Flags FDFlags

	// RightsBase are rights that apply to this file descriptor.
	RightsBase Rights

	// RightsInheriting are the maximum set of rights that may be installed on
	// new file descriptors that are created through this file descriptor,
	// e.g. through PathOpen.
	RightsInheriting Rights
}

// SizeOfDirent is the size in bytes of directory entries when serialized to the
// output buffer of fd_readdir.
const SizeOfDirent = 24

// DirEntry is a directory entry.
type DirEntry struct {
	// Next is the offset of the next directory entry stored in this directory.
	Next DirCookie

	// INode is the serial number of the file referred to by this directory
	// entry.
	INode INode

	// Type is the type of the file referred to by this directory entry.
	Type FileType

	// Name of the directory entry. When the directory entry is retrieved by a
	// call to FDReadDir, the name may point to an internal buffer and therefore
	// remains valid only until the next call to FDReadDir.
	Name []byte
}

// DirCookie is a reference to the offset of a directory entry.
//
// The value 0 signifies the start of the directory.
type DirCookie uint64

// DirNameLength is the type for the DirEntry.NameLength field.
type DirNameLength uint32

// Advice is file or memory access pattern advisory information.
type Advice uint8

const (
	// Normal indicates that the application has no advice to give on its
	// behavior with respect to the specified data.
	Normal Advice = iota

	// Sequential indicates that the application expects to access the
	// specified data sequentially from lower offsets to higher offsets.
	Sequential

	// Random indicates that the application expects to access the specified
	// data in a random order.
	Random

	// WillNeed indicates that the application expects to access the specified
	// data in the near future.
	WillNeed

	// DontNeed indicates that the application expects that it will not access
	// the specified data in the near future.
	DontNeed

	// NoReuse indicates that the application expects to access the specified
	// data once and then not reuse it thereafter.
	NoReuse
)

func (a Advice) String() string {
	switch a {
	case Normal:
		return "Normal"
	case Sequential:
		return "Sequential"
	case Random:
		return "Random"
	case WillNeed:
		return "WillNeed"
	case DontNeed:
		return "DontNeed"
	case NoReuse:
		return "NoReuse"
	default:
		return fmt.Sprintf("Advice(%d)", a)
	}
}

// FSTFlags indicate which file time attributes to adjust.
type FSTFlags uint16

const (
	// AccessTime means adjust the last data access timestamp to the value
	// stored in FileStat.AccessTime.
	AccessTime FSTFlags = 1 << iota

	// AccessTimeNow means adjust the last data access timestamp to the time
	// of clock Realtime.
	AccessTimeNow

	// ModifyTime means adjust the last data modification timestamp to the value
	// stored in FileStat.ModifyTime.
	ModifyTime

	// ModifyTimeNow means adjust the last data modification timestamp to the time
	// of clock Realtime.
	ModifyTimeNow
)

// Has is true if the flag is set.
func (flags FSTFlags) Has(f FSTFlags) bool {
	return (flags & f) == f
}

var fstflagsStrings = [...]string{
	"AccessTime",
	"AccessTimeNow",
	"ModifyTime",
	"ModifyTimeNow",
}

func (flags FSTFlags) String() (s string) {
	if flags == 0 {
		return "FSTFlags(0)"
	}
	for i, name := range fstflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("FSTFlags(%d)", flags)
	}
	return
}

// LookupFlags determine the method of how paths are resolved.
type LookupFlags uint32

const (
	// SymlinkFollow means that as long as the resolved path corresponds to a
	// symbolic link, it is expanded.
	SymlinkFollow LookupFlags = 1 << iota
)

// Has is true if the flag is set.
func (flags LookupFlags) Has(f LookupFlags) bool {
	return (flags & f) == f
}

func (flags LookupFlags) String() string {
	switch flags {
	case SymlinkFollow:
		return "SymlinkFollow"
	default:
		return fmt.Sprintf("LookupFlags(%d)", flags)
	}
}

// OpenFlags are flags used by PathOpen.
type OpenFlags uint16

const (
	// OpenCreate means create a file if it does not exist.
	OpenCreate OpenFlags = 1 << iota

	// OpenDirectory means fail if the path is not a directory.
	OpenDirectory

	// OpenExclusive means fail if the file already exists.
	OpenExclusive

	// OpenTruncate means truncate file to size 0.
	OpenTruncate
)

// Has is true if the flag is set.
func (flags OpenFlags) Has(f OpenFlags) bool {
	return (flags & f) == f
}

var openflagsStrings = [...]string{
	"OpenCreate",
	"OpenDirectory",
	"OpenExclusive",
	"OpenTruncate",
}

func (flags OpenFlags) String() (s string) {
	if flags == 0 {
		return "OpenFlags(0)"
	}
	for i, name := range openflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("OpenFlags(%d)", flags)
	}
	return
}

// PreOpenType are identifiers for pre-opened capabilities.
type PreOpenType uint8

const (
	// PreOpenDir is a pre-opened directory.
	PreOpenDir PreOpenType = iota
)

func (p PreOpenType) String() string {
	switch p {
	case PreOpenDir:
		return "PreOpenDir"
	default:
		return fmt.Sprintf("PreOpenType(%d)", p)
	}
}

// PreStatDir is the contents of a PreStat when type is PreOpenDir.
type PreStatDir struct {
	// NameLength is the length of the directory name for use with
	// FDPreStatDirName.
	NameLength Size
}

// PreStat is information about a pre-opened capability.
type PreStat struct {
	// Type is the type of pre-open.
	Type PreOpenType

	// PreStatDir is directory information when type is PreOpenDir.
	PreStatDir PreStatDir
}

// Size represents a size.
type Size uint32

// IOVec is a slice of bytes.
type IOVec []byte
