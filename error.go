package wasi

import (
	"fmt"
	"syscall"
)

// Errno are the error codes returned by functions.
//
// Not all of these error codes are returned by the functions provided by this
// API; some are used in higher-level library layers, and others are provided
// merely for alignment with POSIX.
type Errno uint16

const (
	// ESUCCESS indicates that no error occurred (system call completed
	// successfully).
	ESUCCESS Errno = iota

	// E2BIG means an argument list is too long.
	E2BIG

	// EACCES means permission is denied.
	EACCES

	// EADDRINUSE means an address is already in use.
	EADDRINUSE

	// EADDRNOTAVAIL means an address is not available.
	EADDRNOTAVAIL

	// EAFNOSUPPORT means an address family is not supported by a protocol family.
	EAFNOSUPPORT

	// EAGAIN means the caller should try again.
	EAGAIN

	// EALREADY means a socket already connected
	EALREADY

	// EBADF means bad file number.
	EBADF

	// EBADMSG indicates that the caller is trying to read an unreadable message.
	EBADMSG

	// EBUSY means a device or resource busy.
	EBUSY

	// ECANCELED means an operation was canceled.
	ECANCELED

	// ECHILD means no child processes.
	ECHILD

	// ECONNABORTED means a connection was aborted.
	ECONNABORTED

	// ECONNREFUSED means connection was refused.
	ECONNREFUSED

	// ECONNRESET means a connection was reset by peer.
	ECONNRESET

	// EDEADLK indicates a deadlock condition.
	EDEADLK

	// EDESTADDRREQ means a destination address is required.
	EDESTADDRREQ

	// EDOM means a math argument is out of domain of func.
	EDOM

	// EDQUOT means a quota was exceeded.
	EDQUOT

	// EEXIST means a file exists.
	EEXIST

	// EFAULT means bad address.
	EFAULT

	// EFBIG indicates a file is too large.
	EFBIG

	// EHOSTUNREACH means a host is unreachable.
	EHOSTUNREACH

	// EIDRM means identifier removed.
	EIDRM

	// EILSEQ indicates an illegal byte sequence
	EILSEQ

	// EINPROGRESS means a connection is already in progress.
	EINPROGRESS

	// EINTR means a system call was interrupted.
	EINTR

	// EINVAL means an argument was invalid.
	EINVAL

	// EIO means an I/O error occurred.
	EIO

	// EISCONN means a socket is already connected.
	EISCONN

	// EISDIR means a file is a directory.
	EISDIR

	// ELOOP indicates that there are too many symbolic links.
	ELOOP

	// EMFILE indicates that there are too many open files
	EMFILE

	// EMLINK indicates that there are too many links
	EMLINK

	// EMSGSIZE means a message is too long.
	EMSGSIZE

	// EMULTIHOP means a multihop was attempted.
	EMULTIHOP

	// ENAMETOOLONG means a file name is too long.
	ENAMETOOLONG

	// ENETDOWN means a network interface is not configured.
	ENETDOWN

	// ENETRESET means a network was dropped connection on reset.
	ENETRESET

	// ENETUNREACH means a network is unreachable.
	ENETUNREACH

	// ENFILE means a file table overflow occurred.
	ENFILE

	// ENOBUFS means that no buffer space is available.
	ENOBUFS

	// ENODEV means no such device.
	ENODEV

	// ENOENT means no such file or directory.
	ENOENT

	// ENOEXEC means an exec format error.
	ENOEXEC

	// ENOLCK means that there are no record locks available
	ENOLCK

	// ENOLINK means the link has been severed.
	ENOLINK

	// ENOMEM means out of memory.
	ENOMEM

	// ENOMSG means that there is no message of desired type.
	ENOMSG

	// ENOPROTOOPT means a protocol is not available.
	ENOPROTOOPT

	// ENOSPC means that there is no space left on a device.
	ENOSPC

	// ENOSYS means not implemented.
	ENOSYS

	// ENOTCONN means a socket is not connected.
	ENOTCONN

	// ENOTDIR means a file is not a directory
	ENOTDIR

	// ENOTEMPTY means a directory is not empty.
	ENOTEMPTY

	// ENOTRECOVERABLE means state is not recoverable.
	ENOTRECOVERABLE

	// ENOTSOCK means a socket operation was attempted on a non-socket.
	ENOTSOCK

	// ENOTSUP means not supported.
	ENOTSUP

	// ENOTTY means not a typewriter.
	ENOTTY

	// ENXIO means no such device or address.
	ENXIO

	// EOVERFLOW means the value is too large for defined data type.
	EOVERFLOW

	// EOWNERDEAD means an owner died.
	EOWNERDEAD

	// EPERM means an operation is not permitted.
	EPERM

	// EPIPE means broken pipe.
	EPIPE

	// EPROTO means a protocol error ocurred.
	EPROTO

	// EPROTONOSUPPORT means a protocol is not supported.
	EPROTONOSUPPORT

	// EPROTOTYPE means that a protocol is the wrong type for socket.
	EPROTOTYPE

	// ERANGE means a math result is not representable.
	ERANGE

	// EROFS means a file system is read-only.
	EROFS

	// ESPIPE means a seek is illegal.
	ESPIPE

	// ESRCH means no such process.
	ESRCH

	// ESTALE means a file handle is stale.
	ESTALE

	// ETIMEDOUT means a connection timed out.
	ETIMEDOUT

	// ETXTBSY means text file busy.
	ETXTBSY

	// EXDEV means cross-device link.
	EXDEV

	// ENOTCAPABLE means capabilities are insufficient.
	ENOTCAPABLE
)

// MakeErrno converts a Go error to a WASI errno value in a way that is portable
// across platforms.
func MakeErrno(err error) Errno { return makeErrno(err) }

func (e Errno) Error() string {
	if i := int(e); i >= 0 && i < len(errorStrings) {
		return errorStrings[i]
	}
	return fmt.Sprintf("Unknown Error (%d)", int(e))
}

func (e Errno) Name() string {
	if i := int(e); i >= 0 && i < len(errorNames) {
		return errorNames[i]
	}
	return fmt.Sprintf("errno(%d)", int(e))
}

// Syscall convers the error to a native error number of the host platform.
//
// The method does the inverse of passing the syscall.Errno value to
// wasi.MakeErrno tho some error numbers may differ on the host platform,
// therefore there is no guarantee that the wasi.Errno value obtained by a
// call to wasi.MakeErrno will yield the same syscall.Errno returned by this
// method.
func (e Errno) Syscall() syscall.Errno { return errnoToSyscall(e) }

var errorStrings = [...]string{
	ESUCCESS:        "OK",
	E2BIG:           "Argument list too long",
	EACCES:          "Permission denied",
	EADDRINUSE:      "Address already in use",
	EADDRNOTAVAIL:   "Address not available",
	EAFNOSUPPORT:    "Address family not supported by protocol family",
	EAGAIN:          "Try again",
	EALREADY:        "Socket already connected",
	EBADF:           "Bad file number",
	EBADMSG:         "Trying to read unreadable message",
	EBUSY:           "Device or resource busy",
	ECANCELED:       "Operation canceled.",
	ECHILD:          "No child processes",
	ECONNABORTED:    "Connection aborted",
	ECONNREFUSED:    "Connection refused",
	ECONNRESET:      "Connection reset by peer",
	EDEADLK:         "Deadlock condition",
	EDESTADDRREQ:    "Destination address required",
	EDOM:            "Math arg out of domain of func",
	EDQUOT:          "Quota exceeded",
	EEXIST:          "File exists",
	EFAULT:          "Bad address",
	EFBIG:           "File too large",
	EHOSTUNREACH:    "Host is unreachable",
	EIDRM:           "Identifier removed",
	EILSEQ:          "Illegal byte sequence",
	EINPROGRESS:     "Connection already in progress",
	EINTR:           "Interrupted system call",
	EINVAL:          "Invalid argument",
	EIO:             "I/O error",
	EISCONN:         "Socket is already connected",
	EISDIR:          "Is a directory",
	ELOOP:           "Too many symbolic links",
	EMFILE:          "Too many open files",
	EMLINK:          "Too many links",
	EMSGSIZE:        "Message too long",
	EMULTIHOP:       "Multihop attempted",
	ENAMETOOLONG:    "File name too long",
	ENETDOWN:        "Network interface is not configured",
	ENETRESET:       "Network dropped connection on reset",
	ENETUNREACH:     "Network is unreachable",
	ENFILE:          "File table overflow",
	ENOBUFS:         "No buffer space available",
	ENODEV:          "No such device",
	ENOENT:          "No such file or directory",
	ENOEXEC:         "Exec format error",
	ENOLCK:          "No record locks available",
	ENOLINK:         "The link has been severed",
	ENOMEM:          "Out of memory",
	ENOMSG:          "No message of desired type",
	ENOPROTOOPT:     "Protocol not available",
	ENOSPC:          "No space left on device",
	ENOSYS:          "Not implemented",
	ENOTCONN:        "Socket is not connected",
	ENOTDIR:         "Not a directory",
	ENOTEMPTY:       "Directory not empty",
	ENOTRECOVERABLE: "State not recoverable",
	ENOTSOCK:        "Socket operation on non-socket",
	ENOTSUP:         "Not supported",
	ENOTTY:          "Not a typewriter",
	ENXIO:           "No such device or address",
	EOVERFLOW:       "Value too large for defined data type",
	EOWNERDEAD:      "Owner died",
	EPERM:           "Operation not permitted",
	EPIPE:           "Broken pipe",
	EPROTO:          "Protocol error",
	EPROTONOSUPPORT: "Unknown protocol",
	EPROTOTYPE:      "Protocol wrong type for socket",
	ERANGE:          "Math result not representable",
	EROFS:           "Read-only file system",
	ESPIPE:          "Illegal seek",
	ESRCH:           "No such process",
	ESTALE:          "Stale file handle",
	ETIMEDOUT:       "Connection timed out",
	ETXTBSY:         "Text file busy",
	EXDEV:           "Cross-device link",
	ENOTCAPABLE:     "Capabilities insufficient",
}

var errorNames = [...]string{
	ESUCCESS:        "ESUCCESS",
	E2BIG:           "E2BIG",
	EACCES:          "EACCES",
	EADDRINUSE:      "EADDRINUSE",
	EADDRNOTAVAIL:   "EADDRNOTAVAIL",
	EAFNOSUPPORT:    "EAFNOSUPPORT",
	EAGAIN:          "EAGAIN",
	EALREADY:        "EALREADY",
	EBADF:           "EBADF",
	EBADMSG:         "EBADMSG",
	EBUSY:           "EBUSY",
	ECANCELED:       "ECANCELED",
	ECHILD:          "ECHILD",
	ECONNABORTED:    "ECONNABORTED",
	ECONNREFUSED:    "ECONNREFUSED",
	ECONNRESET:      "ECONNRESET",
	EDEADLK:         "EDEADLK",
	EDESTADDRREQ:    "EDESTADDRREQ",
	EDOM:            "EDOM",
	EDQUOT:          "EDQUOT",
	EEXIST:          "EEXIST",
	EFAULT:          "EFAULT",
	EFBIG:           "EFBIG",
	EHOSTUNREACH:    "EHOSTUNREACH",
	EIDRM:           "EIDRM",
	EILSEQ:          "EILSEQ",
	EINPROGRESS:     "EINPROGRESS",
	EINTR:           "EINTR",
	EINVAL:          "EINVAL",
	EIO:             "EIO",
	EISCONN:         "EISCONN",
	EISDIR:          "EISDIR",
	ELOOP:           "ELOOP",
	EMFILE:          "EMFILE",
	EMLINK:          "EMLINK",
	EMSGSIZE:        "EMSGSIZE",
	EMULTIHOP:       "EMULTIHOP",
	ENAMETOOLONG:    "ENAMETOOLONG",
	ENETDOWN:        "ENETDOWN",
	ENETRESET:       "ENETRESET",
	ENETUNREACH:     "ENETUNREACH",
	ENFILE:          "ENFILE",
	ENOBUFS:         "ENOBUFS",
	ENODEV:          "ENODEV",
	ENOENT:          "ENOENT",
	ENOEXEC:         "ENOEXEC",
	ENOLCK:          "ENOLCK",
	ENOLINK:         "ENOLINK",
	ENOMEM:          "ENOMEM",
	ENOMSG:          "ENOMSG",
	ENOPROTOOPT:     "ENOPROTOOPT",
	ENOSPC:          "ENOSPC",
	ENOSYS:          "ENOSYS",
	ENOTCONN:        "ENOTCONN",
	ENOTDIR:         "ENOTDIR",
	ENOTEMPTY:       "ENOTEMPTY",
	ENOTRECOVERABLE: "ENOTRECOVERABLE",
	ENOTSOCK:        "ENOTSOCK",
	ENOTSUP:         "ENOTSUP",
	ENOTTY:          "ENOTTY",
	ENXIO:           "ENXIO",
	EOVERFLOW:       "EOVERFLOW",
	EOWNERDEAD:      "EOWNERDEAD",
	EPERM:           "EPERM",
	EPIPE:           "EPIPE",
	EPROTO:          "EPROTO",
	EPROTONOSUPPORT: "EPROTONOSUPPORT",
	EPROTOTYPE:      "EPROTOTYPE",
	ERANGE:          "ERANGE",
	EROFS:           "EROFS",
	ESPIPE:          "ESPIPE",
	ESRCH:           "ESRCH",
	ESTALE:          "ESTALE",
	ETIMEDOUT:       "ETIMEDOUT",
	ETXTBSY:         "ETXTBSY",
	EXDEV:           "EXDEV",
	ENOTCAPABLE:     "ENOTCAPABLE",
}
