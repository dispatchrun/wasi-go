package wasi

import (
	"context"
	"io"
	"io/fs"
	"path"
	"time"
)

// FS constructs a fs.FS file system backed by a wasi system.
func FS(ctx context.Context, sys System, root FD) fs.FS {
	return &fileSystem{ctx, sys, root}
}

type fileSystem struct {
	ctx context.Context
	System
	root FD
}

func (fsys *fileSystem) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{"open", name, fs.ErrInvalid}
	}
	const rights = PathOpenRight |
		PathFileStatGetRight |
		FDReadRight |
		FDReadDirRight |
		FDSeekRight |
		FDTellRight |
		FDFileStatGetRight
	f, errno := fsys.PathOpen(fsys.ctx, fsys.root, SymlinkFollow, name, 0, rights, rights, 0)
	if errno != ESUCCESS {
		return nil, &fs.PathError{"open", name, errno}
	}
	return &file{fsys: fsys, name: name, fd: f}, nil
}

type file struct {
	fsys *fileSystem
	name string
	fd   FD
	*dir
}

type dir struct {
	dirent []DirEntry
	cookie DirCookie
}

func (f *file) Close() error {
	if f.fd >= 0 {
		f.fsys.FDClose(f.fsys.ctx, f.fd)
		f.fd = -1
	}
	return nil
}

func (f *file) Read(b []byte) (int, error) {
	if f.fd < 0 {
		return 0, io.EOF
	}
	if len(b) == 0 {
		return 0, nil
	}
	n, errno := f.fsys.FDRead(f.fsys.ctx, f.fd, []IOVec{b})
	if errno != ESUCCESS {
		return int(n), &fs.PathError{"read", f.name, errno}
	}
	if n == 0 {
		return 0, io.EOF
	}
	return int(n), nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	if f.fd < 0 {
		return nil, &fs.PathError{"stat", f.name, fs.ErrClosed}
	}
	s, errno := f.fsys.FDFileStatGet(f.fsys.ctx, f.fd)
	if errno != ESUCCESS {
		return nil, &fs.PathError{"stat", f.name, errno}
	}
	return &fileInfo{stat: s, name: path.Base(f.name)}, nil
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if f.fd < 0 {
		return 0, &fs.PathError{"seek", f.name, fs.ErrClosed}
	}
	seek, errno := f.fsys.FDSeek(f.fsys.ctx, f.fd, FileDelta(offset), Whence(whence))
	if errno != ESUCCESS {
		return int64(seek), &fs.PathError{"seek", f.name, errno}
	}
	if f.dir != nil {
		f.dir.cookie = DirCookie(seek)
	}
	return int64(seek), nil
}

func (f *file) ReadAt(b []byte, off int64) (int, error) {
	if f.fd < 0 {
		return 0, io.EOF
	}
	n, errno := f.fsys.FDPread(f.fsys.ctx, f.fd, []IOVec{b}, FileSize(off))
	if errno != ESUCCESS {
		return int(n), &fs.PathError{"pread", f.name, errno}
	}
	if int(n) < len(b) {
		return int(n), io.EOF
	}
	return int(n), nil
}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.fd < 0 {
		return nil, io.EOF
	}
	if f.dir == nil {
		f.dir = new(dir)
	}

	capacity := n
	if n <= 0 {
		if capacity = cap(f.dirent); capacity == 0 {
			capacity = 20
		}
	}
	if cap(f.dirent) < capacity {
		f.dirent = make([]DirEntry, capacity)
	} else {
		f.dirent = f.dirent[:capacity]
	}

	var dirent []fs.DirEntry
	for {
		limit := len(f.dirent)
		if n > 0 {
			limit = n - len(dirent)
		}
		rn, errno := f.fsys.FDReadDir(f.fsys.ctx, f.fd, f.dirent[:limit], f.cookie, 4096)
		if rn > 0 {
			for _, e := range f.dirent[:rn] {
				switch string(e.Name) {
				case ".", "..":
					continue
				}
				dirent = append(dirent, &dirEntry{
					typ:  e.Type,
					name: string(e.Name),
					dir:  f,
				})
			}
			f.cookie = f.dirent[rn-1].Next
		}
		if errno != ESUCCESS {
			return dirent, &fs.PathError{"readdir", f.name, errno}
		}
		if n > 0 && n == len(dirent) {
			return dirent, nil
		}
		if rn == 0 {
			if n <= 0 {
				return dirent, nil
			} else {
				return dirent, io.EOF
			}
		}
	}
}

var (
	_ fs.ReadDirFile = (*file)(nil)
	_ io.ReaderAt    = (*file)(nil)
	_ io.Seeker      = (*file)(nil)
)

type fileInfo struct {
	stat FileStat
	name string
}

func (info *fileInfo) Name() string {
	return info.name
}

func (info *fileInfo) Size() int64 {
	return int64(info.stat.Size)
}

func (info *fileInfo) Mode() fs.FileMode {
	return makeFileMode(info.stat.FileType)
}

func (info *fileInfo) ModTime() time.Time {
	return time.Unix(0, int64(info.stat.ModifyTime))
}

func (info *fileInfo) IsDir() bool {
	return info.stat.FileType == DirectoryType
}

func (info *fileInfo) Sys() any {
	return &info.stat
}

type dirEntry struct {
	typ  FileType
	dir  *file
	name string
}

func (d *dirEntry) Name() string {
	return d.name
}

func (d *dirEntry) IsDir() bool {
	return d.typ == DirectoryType
}

func (d *dirEntry) Type() fs.FileMode {
	return makeFileMode(d.typ)
}

func (d *dirEntry) Info() (fs.FileInfo, error) {
	s, errno := d.dir.fsys.PathFileStatGet(d.dir.fsys.ctx, d.dir.fd, 0, d.name)
	if errno != ESUCCESS {
		return nil, &fs.PathError{"stat", d.name, errno}
	}
	return &fileInfo{stat: s, name: d.name}, nil
}

func makeFileMode(fileType FileType) fs.FileMode {
	switch fileType {
	case BlockDeviceType:
		return fs.ModeDevice
	case CharacterDeviceType:
		return fs.ModeDevice | fs.ModeCharDevice
	case DirectoryType:
		return fs.ModeDir
	case RegularFileType:
		return 0
	case SocketDGramType, SocketStreamType:
		return fs.ModeSocket
	case SymbolicLinkType:
		return fs.ModeSymlink
	default:
		return fs.ModeIrregular
	}
}
