.PHONY: all clean fmt lint test testdata wasi-libc wasi-testsuite

GO ?= go

count ?= 1

wasi-go.src = \
	$(wildcard *.go) \
	$(wildcard */*.go) \
	$(wildcard */*/*.go)

wasirun.src = $(wasi-go.src)

testdata.c.src = $(wildcard testdata/c/*.c)
testdata.c.wasm = $(testdata.c.src:.c=.wasm)

testdata.http.src = $(wildcard testdata/c/http/http*.c)
testdata.http.wasm = $(testdata.http.src:.c=.wasm)

testdata.go.src = $(wildcard testdata/go/*.go)
testdata.go.wasm = $(testdata.go.src:.go=.wasm)

testdata.tinygo.src = $(wildcard testdata/tinygo/*.go)
testdata.tinygo.wasm = $(testdata.tinygo.src:.go=.wasm)

testdata.files = \
	$(testdata.c.wasm) \
	$(testdata.http.wasm) \
	$(testdata.go.wasm) \
	$(testdata.tinygo.wasm)

all: test wasi-testsuite

clean:
	rm -f $(testdata.files)

test: testdata
	$(GO) test -count=$(count) ./...

fmt:
	$(GO) fmt ./...

lint:
	which golangci-lint >/dev/null && golangci-lint run

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

testdata/c/http/http.wasm: testdata/c/http/http.c
	clang $< -o $@ -Wall -Os -target wasm32-unknown-wasi testdata/c/http/proxy.c testdata/c/http/proxy_component_type.o

testdata/go/%.wasm: testdata/go/%.go
	GOARCH=wasm GOOS=wasip1 $(GO) build -o $@ $<

testdata/tinygo/%.wasm: testdata/tinygo/%.go
	tinygo build -target=wasi -o $@ $<

wasirun: go.mod $(wasirun.src)
	$(GO) build -o wasirun ./cmd/wasirun

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
