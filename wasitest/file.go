package wasitest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

var file = testSuite{
	"exceeding the limit of open files":       testMaxOpenFiles,
	"exceeding the limit of open directories": testMaxOpenDirs,
}

func testMaxOpenFiles(t *testing.T, ctx context.Context, newSystem newSystem) {
	tmp := t.TempDir()
	sys := newSystem(TestConfig{
		RootFS:       tmp,
		MaxOpenFiles: 10,
	})

	for i := 0; i < 10; i++ {
		_, errno := sys.PathOpen(ctx, 3, 0, ".", wasi.OpenDirectory, wasi.AllRights, wasi.AllRights, 0)
		if errno == wasi.ENFILE {
			break
		}
		assertEqual(t, errno, wasi.ESUCCESS)
	}

	for i := 0; i < 10; i++ {
		_, errno := sys.PathOpen(ctx, 3, 0, ".", wasi.OpenDirectory, wasi.AllRights, wasi.AllRights, 0)
		assertEqual(t, errno, wasi.ENFILE)
	}
}

func testMaxOpenDirs(t *testing.T, ctx context.Context, newSystem newSystem) {
	tmp := t.TempDir()
	sys := newSystem(TestConfig{
		RootFS:      tmp,
		MaxOpenDirs: 10,
	})

	assertOK(t, os.WriteFile(filepath.Join(tmp, "file-1"), []byte("1"), 0666))
	assertOK(t, os.WriteFile(filepath.Join(tmp, "file-2"), []byte("2"), 0666))
	assertOK(t, os.WriteFile(filepath.Join(tmp, "file-3"), []byte("3"), 0666))

	for i := 0; i < 10; i++ {
		d, errno := sys.PathOpen(ctx, 3, 0, ".", wasi.OpenDirectory, wasi.AllRights, wasi.AllRights, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		dirEntry := [1]wasi.DirEntry{}
		n, errno := sys.FDReadDir(ctx, d, dirEntry[:], 0, 1024)
		assertEqual(t, n, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
	}

	for i := 0; i < 10; i++ {
		d, errno := sys.PathOpen(ctx, 3, 0, ".", wasi.OpenDirectory, wasi.AllRights, wasi.AllRights, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		dirEntry := [1]wasi.DirEntry{}
		n, errno := sys.FDReadDir(ctx, d, dirEntry[:], 0, 1024)
		assertEqual(t, n, 0)
		assertEqual(t, errno, wasi.ENFILE)
		assertEqual(t, sys.FDClose(ctx, d), wasi.ESUCCESS)
	}
}
