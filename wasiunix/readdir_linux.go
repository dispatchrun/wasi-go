package wasiunix

import (
	"bytes"
	"unsafe"

	"github.com/stealthrocket/wasi"
	"golang.org/x/sys/unix"
)

const sizeOfDirent = 19

type dirent struct {
	ino    uint64
	off    int64
	reclen uint16
	typ    uint8
}

const maxNameLen = 1024
const bufferSize = 4 * maxNameLen // must be greater than sizeOfDirent

type dirbuf struct {
	buffer *[bufferSize]byte
	offset int
	length int
	cookie wasi.DirCookie
}

func (d *dirbuf) readDirEntries(fd int, entries []wasi.DirEntry, cookie wasi.DirCookie, bufferSizeBytes int) (int, error) {
	if d.buffer == nil {
		d.buffer = new([bufferSize]byte)
	}

	if cookie < d.cookie {
		if _, err := unix.Seek(fd, 0, unix.SEEK_SET); err != nil {
			return 0, err
		}
		d.offset = 0
		d.length = 0
		d.cookie = 0
	}

	numEntries := 0
	for {
		if numEntries == len(entries) {
			return numEntries, nil
		}

		if (d.length - d.offset) < sizeOfDirent {
			if numEntries > 0 {
				return numEntries, nil
			}
			n, err := unix.Getdents(fd, d.buffer[:])
			if err != nil {
				return numEntries, err
			}
			if n == 0 {
				return numEntries, nil
			}
			d.offset = 0
			d.length = n
		}

		dirent := (*dirent)(unsafe.Pointer(&d.buffer[d.offset]))

		if (d.offset + int(dirent.reclen)) > d.length {
			d.offset = d.length
			continue
		}

		if dirent.ino == 0 {
			d.offset += int(dirent.reclen)
			continue
		}

		if d.cookie >= cookie {
			dirEntry := wasi.DirEntry{
				Next:  d.cookie + 1,
				INode: wasi.INode(dirent.ino),
			}

			switch dirent.typ {
			case unix.DT_BLK:
				dirEntry.Type = wasi.BlockDeviceType
			case unix.DT_CHR:
				dirEntry.Type = wasi.CharacterDeviceType
			case unix.DT_DIR:
				dirEntry.Type = wasi.DirectoryType
			case unix.DT_LNK:
				dirEntry.Type = wasi.SymbolicLinkType
			case unix.DT_REG:
				dirEntry.Type = wasi.RegularFileType
			case unix.DT_SOCK:
				dirEntry.Type = wasi.SocketStreamType
			default: // DT_FIFO, DT_UNKNOWN
				dirEntry.Type = wasi.UnknownType
			}

			i := d.offset + sizeOfDirent
			j := d.offset + int(dirent.reclen)
			dirEntry.Name = d.buffer[i:j:j]

			n := bytes.IndexByte(dirEntry.Name, 0)
			if n >= 0 {
				dirEntry.Name = dirEntry.Name[:n:n]
			}

			entries[numEntries] = dirEntry
			numEntries++

			bufferSizeBytes -= wasi.SizeOfDirent
			bufferSizeBytes -= len(dirEntry.Name)

			if bufferSizeBytes <= 0 {
				return numEntries, nil
			}
		}

		d.offset += int(dirent.reclen)
		d.cookie += 1
	}
}
