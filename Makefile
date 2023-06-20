.PHONY: all clean test testdata wasi-libc wasi-testsuite

count ?= 1

wasi-go.src = \
	$(wildcard *.go) \
	$(wildcard */*.go) \
	$(wildcard */*/*.go)

wasirun.src = $(wasi-go.src)

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

all: test wasi-testsuite

clean:
	rm -f $(testdata.files)

test: testdata
	go test -count=$(count) ./...

testdata: $(testdata.files)

testdata/.sysroot:
	mkdir -p testdata/.sysroot

testdata/.wasi-libc: testdata/.wasi-libc/.git

testdata/.wasi-libc/.git: .gitmodules
	git submodule update --init --recursive -- testdata/.wasi-libc

testdata/.wasi-testsuite: testdata/.wasi-testsuite/.git

testdata/.wasi-testsuite/.git: .gitmodules
	git submodule update --init --recursive -- testdata/.wasi-testsuite

testdata/.sysroot/lib/wasm32-wasi/libc.a: testdata/.wasi-libc
	make -j4 -C testdata/.wasi-libc install INSTALL_DIR=../.sysroot

testdata/c/%.c: wasi-libc
testdata/c/%.wasm: testdata/c/%.c
	clang $< -o $@ -Wall -Os -target wasm32-unknown-wasi --sysroot testdata/.sysroot

testdata/go/%.wasm: testdata/go/%.go
	GOARCH=wasm GOOS=wasip1 gotip build -o $@ $<

testdata/tinygo/%.wasm: testdata/tinygo/%.go
	tinygo build -target=wasi -o $@ $<

wasirun:
	go build -o wasirun ./cmd/wasirun

wasi-libc: testdata/.sysroot/lib/wasm32-wasi/libc.a

wasi-testsuite: testdata/.wasi-testsuite wasirun
	python3 testdata/.wasi-testsuite/test-runner/wasi_test_runner.py \
		-t testdata/.wasi-testsuite/tests/assemblyscript/testsuite \
		   testdata/.wasi-testsuite/tests/c/testsuite \
		   testdata/.wasi-testsuite/tests/rust/testsuite \
		-r testdata/adapter.py
	@rm -rf testdata/.wasi-testsuite/tests/rust/testsuite/fs-tests.dir/*.cleanup

.gitmodules:
	git submodule add --name wasi-libc -- \
		'https://github.com/WebAssembly/wasi-libc' testdata/.wasi-libc
	git submodule add --name wasi-testsuite -b prod/testsuite-base -- \
		"https://github.com/WebAssembly/wasi-testsuite" testdata/.wasi-testsuite
