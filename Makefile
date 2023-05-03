.PHONY: clean test testdata wasi-libc

CFLAGS = -Wall -Os

testdata.c.src = $(wildcard testdata/c/*.c)
testdata.c.wasm = $(testdata.c.src:.c=.wasm)

testdata.files = $(testdata.c.wasm)

test: testdata
	go test ./...

testdata: $(testdata.files)

testdata/.deps:
	mkdir testdata/.deps

testdata/.wasi-libc: .gitmodules
	git submodule init -- testdata/.deps/wasi-libc
	cd testdata/.deps/wasi-libc && make -j4 install INSTALL_DIR=../../.wasi-libc

clean:
	rm -f $(testdata.files)

wasi-libc: testdata/.wasi-libc

%.wasm: %.c wasi-libc
	clang $< -o $@ $(CFLAGS) -target wasm32-unknown-wasi --sysroot testdata/.wasi-libc -Wl,--allow-undefined

.gitmodules: testdata/.deps
	git submodule add 'https://github.com/WebAssembly/wasi-libc' testdata/.deps/wasi-libc
