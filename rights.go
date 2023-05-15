package wasi

import "fmt"

// Rights are file descriptor rights, determining which actions may be performed.
type Rights uint64

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

	// ReadRights are rights related to reads.
	ReadRights Rights = FDReadRight | FDReadDirRight

	// WriteRights are rights related to writes.
	WriteRights Rights = FDWriteRight | FDAllocateRight | PathFileStatSetSizeRight | FDDataSyncRight

	syncRights     Rights = FDSyncRight | FDDataSyncRight
	seekRights     Rights = FDSeekRight | FDTellRight
	fileStatRights Rights = FDFileStatGetRight | FDFileStatSetSizeRight | FDFileStatSetTimesRight
	pathRights     Rights = PathCreateDirectoryRight | PathCreateFileRight | PathLinkSourceRight | PathLinkTargetRight | PathOpenRight | PathReadLinkRight | PathRenameSourceRight | PathRenameTargetRight | PathFileStatGetRight | PathFileStatSetSizeRight | PathFileStatSetTimesRight | PathSymlinkRight | PathRemoveDirectoryRight | PathUnlinkFileRight

	// FileRights are rights related to files.
	FileRights Rights = syncRights | seekRights | fileStatRights | FDReadRight | FDStatSetFlagsRight | FDWriteRight | FDAdviseRight | FDAllocateRight | PollFDReadWriteRight

	// DirectoryRights are rights related to directories.
	// See https://github.com/WebAssembly/wasi-testsuite/blob/1b1d4a5/tests/rust/src/bin/directory_seek.rs
	DirectoryRights Rights = pathRights | syncRights | fileStatRights | FDStatSetFlagsRight | FDReadDirRight

	// TTYRights are rights related to terminals.
	// See https://github.com/WebAssembly/wasi-libc/blob/a6f871343/libc-bottom-half/sources/isatty.c
	TTYRights = FileRights &^ seekRights

	// SockListenRights are rights for listener sockets.
	SockListenRights = SockAcceptRight | PollFDReadWriteRight | FDFileStatGetRight | FDStatSetFlagsRight

	// SockConnectionRights are rights for connection sockets.
	SockConnectionRights = FDReadRight | FDWriteRight | PollFDReadWriteRight | SockShutdownRight | FDFileStatGetRight | FDStatSetFlagsRight
)

// Has is true if the flag is set. If multiple flags are specified, Has returns
// true if all flags are set.
func (flags Rights) Has(f Rights) bool {
	return (flags & f) == f
}

// HasAny is true if any flag in a set of flags is set.
func (flags Rights) HasAny(f Rights) bool {
	return (flags & f) != 0
}

var rightsStrings = [...]string{
	"FDDataSyncRight",
	"FDReadRight",
	"FDSeekRight",
	"FDStatSetFlagsRight",
	"FDSyncRight",
	"FDTellRight",
	"FDWriteRight",
	"FDAdviseRight",
	"FDAllocateRight",
	"PathCreateDirectoryRight",
	"PathCreateFileRight",
	"PathLinkSourceRight",
	"PathLinkTargetRight",
	"PathOpenRight",
	"FDReadDirRight",
	"PathReadLinkRight",
	"PathRenameSourceRight",
	"PathRenameTargetRight",
	"PathFileStatGetRight",
	"PathFileStatSetSizeRight",
	"PathFileStatSetTimesRight",
	"FDFileStatGetRight",
	"FDFileStatSetSizeRight",
	"FDFileStatSetTimesRight",
	"PathSymlinkRight",
	"PathRemoveDirectoryRight",
	"PathUnlinkFileRight",
	"PollFDReadWriteRight",
	"SockShutdownRight",
	"SockAcceptRight",
}

func (flags Rights) String() (s string) {
	switch {
	case flags == 0:
		return "Rights(0)"
	case flags.Has(AllRights):
		return "AllRights"
	case flags == FileRights:
		return "FileRights"
	case flags == DirectoryRights:
		return "DirectoryRights"
	case flags == DirectoryRights|FileRights:
		return "DirectoryRights|FileRights"
	case flags == TTYRights:
		return "TTYRights"
	case flags == SockListenRights:
		return "SockListenRights"
	case flags == SockConnectionRights:
		return "SockConnectionRights"
	case flags == SockConnectionRights|SockListenRights:
		return "SockConnectionRights|SockListenRights"
	}
	for i, name := range rightsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("Rights(%d)", flags)
	}
	return
}
