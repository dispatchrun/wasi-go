.PHONY: clean test testdata wasi-libc

CFLAGS = -Wall -Os

testdata.c.src = $(wildcard testdata/c/*.c)
testdata.c.wasm = $(testdata.c.src:.c=.wasm)

testdata.go.src = $(wildcard testdata/go/*.go)
testdata.go.wasm = $(testdata.go.src:.go=.wasm)

testdata.tinygo.src = $(wildcard testdata/tinygo/*.go)
testdata.tinygo.wasm = $(testdata.tinygo.src:.go=.wasm)

testdata.files = \
	$(testdata.c.wasm) \
	$(testdata.go.wasm) \
	$(testdata.tinygo.wasm)

test: testdata
	go test ./...

testdata: $(testdata.files)

testdata/.deps:
	mkdir testdata/.deps

testdata/.wasi-libc: .gitmodules
	git submodule update -- testdata/.deps/wasi-libc
	cd testdata/.deps/wasi-libc && make -j4 install INSTALL_DIR=../../.wasi-libc

clean:
	rm -f $(testdata.files)

wasi-libc: testdata/.wasi-libc

testdata/c/%.c: wasi-libc
testdata/c/%.wasm: testdata/c/%.c
	clang $< -o $@ $(CFLAGS) -target wasm32-unknown-wasi --sysroot testdata/.wasi-libc -Wl,--allow-undefined

testdata/go/%.wasm: testdata/go/%.go
	GOARCH=wasm GOOS=wasip1 gotip build -o $@ $<

testdata/tinygo/%.wasm: testdata/tinygo/%.go
	tinygo build -target=wasi -o $@ $<

.gitmodules: testdata/.deps
	git submodule add 'https://github.com/WebAssembly/wasi-libc' testdata/.deps/wasi-libc
