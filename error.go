package wasip1

// Errno are the error codes returned by functions.
//
// Not all of these error codes are returned by the functions provided by this
// API; some are used in higher-level library layers, and others are provided
// merely for alignment with POSIX.
type Errno uint16

const (
	E2BIG Errno = iota + 1
	EACCES
	EADDRINUSE
	EADDRNOTAVAIL
	EAFNOSUPPORT
	EAGAIN
	EALREADY
	EBADF
	EBADMSG
	EBUSY
	ECANCELED
	ECHILD
	ECONNABORTED
	ECONNREFUSED
	ECONNRESET
	EDEADLK
	EDESTADDRREQ
	EDOM
	EDQUOT
	EEXIST
	EFAULT
	EFBIG
	EHOSTUNREACH
	EIDRM
	EILSEQ
	EINPROGRESS
	EINTR
	EINVAL
	EIO
	EISCONN
	EISDIR
	ELOOP
	EMFILE
	EMLINK
	EMSGSIZE
	EMULTIHOP
	ENAMETOOLONG
	ENETDOWN
	ENETRESET
	ENETUNREACH
	ENFILE
	ENOBUFS
	ENODEV
	ENOENT
	ENOEXEC
	ENOLCK
	ENOLINK
	ENOMEM
	ENOMSG
	ENOPROTOOPT
	ENOSPC
	ENOSYS
	ENOTCONN
	ENOTDIR
	ENOTEMPTY
	ENOTRECOVERABLE
	ENOTSOCK
	ENOTSUP
	ENOTTY
	ENXIO
	EOVERFLOW
	EOWNERDEAD
	EPERM
	EPIPE
	EPROTO
	EPROTONOSUPPORT
	EPROTOTYPE
	ERANGE
	EROFS
	ESPIPE
	ESRCH
	ESTALE
	ETIMEDOUT
	ETXTBSY
	EXDEV
	ENOTCAPABLE

	// SUCCESS indicates that no error occurred (system call completed
	// successfully).
	SUCCESS = 0
)

func (e Errno) Error() string {
	if e >= 1 && int(e) <= len(errorStrings) {
		return errorStrings[e]
	}
	return ""
}

func (e Errno) Name() string {
	if e >= 1 && int(e) <= len(errorNames) {
		return errorNames[e]
	}
	return ""
}

var errorStrings = [...]string{
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
	EILSEQ:          "EILSEQ",
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
	EPROTONOSUPPORT: "UnknownType protocol",
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
