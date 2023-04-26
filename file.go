package wasip1

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
	// WhenceStart seeks relative to start-of-file.
	WhenceStart Whence = iota

	// WhenceCurrent seeks relative to current position.
	WhenceCurrent

	// WhenceEnd seeks relative to end-of-file.
	WhenceEnd
)

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

// Rights are file descriptor rights, determining which actions may be performed.
type Rights uint64

// Has is true if the flag is set.
func (flags Rights) Has(f Rights) bool {
	return (flags & f) != 0
}

const (
	// FDDataSyncRight is the right to invoke FDDataSync.
	//
	// If PathOpenRight is set, it includes the right to invoke PathOpen with
	// the DSync flag.
	FDDataSyncRight Rights = 1 << iota

	// FDReadRight is the right to invoke FDRead and SockRecv.
	//
	// If FDSeekRight is set, it includes the right to invoke FDPread.
	FDReadRight

	// FDSeekRight is the right to invoke FDSeek. This flag implies FDTellRight.
	FDSeekRight

	// FDStatSetFlagsRight is the right to invoke FDStatSetFlags.
	FDStatSetFlagsRight

	// FDSyncRight is the right to invoke FDSync.
	//
	// If PathOpenRight is set, it includes the right to invoke PathOpen with
	// flags RSync and DSync.
	FDSyncRight

	// FDTellRight is the right to invoke FDTell, and the right to invoke
	// FDSeek in such a way that the file offset remains unaltered (i.e.
	// WhenceCurrent with offset zero).
	FDTellRight

	// FDWriteRight is the right to invoke FDWrite and SockSend.
	//
	// If FDSeekRight is set, it includes the right to invoke FDPwrite.
	FDWriteRight

	// FDAdviseRight is the right to invoke FDAdvise.
	FDAdviseRight

	// FDAllocateRight is the right to invoke FDAllocate.
	FDAllocateRight

	// PathCreateDirectoryRight is the right to invoke PathCreateDirectory.
	PathCreateDirectoryRight

	// PathCreateFileRight is (along with PathOpenRight) the right to invoke
	// PathOpen with the OpenCreate flag.
	PathCreateFileRight

	// PathLinkSourceRight is the right to invoke PathLink with the file
	// descriptor as the source directory.
	PathLinkSourceRight

	// PathLinkTargetRight is the right to invoke PathLink with the file
	// descriptor as the target directory.
	PathLinkTargetRight

	// PathOpenRight is the right to invoke PathOpen.
	PathOpenRight

	// FDReadDirRight is the right to invoke FDReadDir.
	FDReadDirRight

	// PathReadLinkRight is the right to invoke PathReadLink.
	PathReadLinkRight

	// PathRenameSourceRight is the right to invoke PathRename with the file
	// descriptor as the source directory.
	PathRenameSourceRight

	// PathRenameTargetRight is the right to invoke PathRename with the file
	// descriptor as the target directory.
	PathRenameTargetRight

	// PathFileStatGetRight is the right to invoke PathFileStatGet.
	PathFileStatGetRight

	// PathFileStatSetSizeRight is the right to change a file's size.
	//
	// If PathOpenRight is set, it includes the right to invoke PathOpen with
	// the OpenTruncate flag.
	//
	// Note: there is no function named PathFileStatSetSize. This follows POSIX
	// design, which only has ftruncate and does not provide ftruncateat. While
	// such function would be desirable from the API design perspective, there
	// are virtually no use cases for it since no code written for POSIX
	// systems would use it. Moreover, implementing it would require multiple
	// syscalls, leading to inferior performance.
	PathFileStatSetSizeRight

	// PathFileStatSetTimesRight is the right to invoke PathFileStatSetTimes.
	PathFileStatSetTimesRight

	// FDFileStatGetRight is the right to invoke FDFileStatGet.
	FDFileStatGetRight

	// FDFileStatSetSizeRight is the right to invoke FDFileStatSetSize.
	FDFileStatSetSizeRight

	// FDFileStatSetTimesRight is the right to invoke FDFileStatSetTimes.
	FDFileStatSetTimesRight

	// PathSymlinkRight is the right to invoke PathSymlink.
	PathSymlinkRight

	// PathRemoveDirectoryRight is the right to invoke PathRemoveDirectory.
	PathRemoveDirectoryRight

	// PathUnlinkFileRight is the right to invoke PathUnlinkFile.
	PathUnlinkFileRight

	// PollFDReadWriteRight is the right to invoke PollOneOff.
	//
	// If FDReadWrite is set, it includes the right to invoke PollOneOff with a
	// FDReadEvent subscription. If FDWriteWrite is set, it includes the right
	// to invoke PollOneOff with a FDWriteEvent subscription.
	PollFDReadWriteRight

	// SockShutdownRight is the right to invoke SockShutdown
	SockShutdownRight

	// SockAccessRight is the right to invoke SockAccept
	SockAcceptRight

	// AllRights is the set of all available rights
	AllRights Rights = (1 << 30) - 1
)

// DirEntry is a directory entry.
type DirEntry struct {
	// Next is the offset of the next directory entry stored in this directory.
	Next DirCookie

	// INode is the serial number of the file referred to by this directory
	// entry.
	INode INode

	// NameLength is the length of the name of the directory entry.
	NameLength DirNameLength

	// Type is the type of the file referred to by this directory entry.
	Type FileType
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

// LookupFlags determine the method of how paths are resolved.
type LookupFlags uint32

const (
	// SymlinkFollow means that as long as the resolved path corresponds to a
	// symbolic link, it is expanded.
	SymlinkFollow LookupFlags = 1 << iota
)

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

// PreOpenType are identifiers for pre-opened capabilities.
type PreOpenType uint8

const (
	// PreOpenDir is a pre-opened directory.
	PreOpenDir PreOpenType = iota
)

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
