package descriptor_test

import (
	"testing"

	"github.com/stealthrocket/wasi/wasiunix/internal/descriptor"
)

type fd uint32

type file struct{ name string }

func TestTable(t *testing.T) {
	table := new(descriptor.Table[fd, file])

	if n := table.Len(); n != 0 {
		t.Errorf("new table is not empty: length=%d", n)
	}

	// The id field is used as a sentinel value.
	v0 := file{name: "1"}
	v1 := file{name: "2"}
	v2 := file{name: "3"}

	k0 := table.Insert(v0)
	k1 := table.Insert(v1)
	k2 := table.Insert(v2)

	for _, lookup := range []struct {
		key fd
		val file
	}{
		{key: k0, val: v0},
		{key: k1, val: v1},
		{key: k2, val: v2},
	} {
		if v, ok := table.Lookup(lookup.key); !ok {
			t.Errorf("value not found for key '%v'", lookup.key)
		} else if v.name != lookup.val.name {
			t.Errorf("wrong value returned for key '%v': want=%v got=%v", lookup.key, lookup.val.name, v.name)
		}
	}

	if n := table.Len(); n != 3 {
		t.Errorf("wrong table length: want=3 got=%d", n)
	}

	k0Found := false
	k1Found := false
	k2Found := false
	table.Range(func(k fd, v file) bool {
		var want file
		switch k {
		case k0:
			k0Found, want = true, v0
		case k1:
			k1Found, want = true, v1
		case k2:
			k2Found, want = true, v2
		}
		if v.name != want.name {
			t.Errorf("wrong value found ranging over '%v': want=%v got=%v", k, want.name, v.name)
		}
		return true
	})

	for _, found := range []struct {
		key fd
		ok  bool
	}{
		{key: k0, ok: k0Found},
		{key: k1, ok: k1Found},
		{key: k2, ok: k2Found},
	} {
		if !found.ok {
			t.Errorf("key not found while ranging over table: %v", found.key)
		}
	}

	for i, deletion := range []struct {
		key fd
	}{
		{key: k1},
		{key: k0},
		{key: k2},
	} {
		table.Delete(deletion.key)
		if _, ok := table.Lookup(deletion.key); ok {
			t.Errorf("item found after deletion of '%v'", deletion.key)
		}
		if n, want := table.Len(), 3-(i+1); n != want {
			t.Errorf("wrong table length after deletion: want=%d got=%d", want, n)
		}
	}
}

func BenchmarkTableInsert(b *testing.B) {
	table := new(descriptor.Table[fd, *file])
	entry := new(file)

	for i := 0; i < b.N; i++ {
		table.Insert(entry)

		if (i % 65536) == 0 {
			table.Reset() // to avoid running out of memory
		}
	}
}

func BenchmarkTableLookup(b *testing.B) {
	const sentinel = "42"
	const numFiles = 65536
	table := new(descriptor.Table[fd, *file])
	files := make([]fd, numFiles)
	entry := file{name: sentinel}

	for i := range files {
		files[i] = table.Insert(&entry)
	}

	var f *file
	for i := 0; i < b.N; i++ {
		f, _ = table.Lookup(files[i%numFiles])
	}
	if f.name != sentinel {
		b.Error("wrong file returned by lookup")
	}
}
