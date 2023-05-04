test: wasirun submodules
	python3 wasi-testsuite/test-runner/wasi_test_runner.py \
		-t wasi-testsuite/tests/assemblyscript/testsuite \
		   wasi-testsuite/tests/c/testsuite \
		   wasi-testsuite/tests/rust/testsuite \
		-r adapter.py

wasi.src.go = $(wildcard ../../*.go)
wasiunix.src.go = $(wildcard ../*.go)
wasirun.src.go = $(wildcard ../../cmd/wasirun/*.go)
wasi_snapshot_preview1.src.go = $(wildcard ../../imports/wasi_snapshot_preview1/*.go)

wasirun: $(wasirun.src.go) $(wasi.src.go) $(wasiunix.src.go) $(wasi_snapshot_preview1.src.go)
	go build -o $@ $(wasirun.src.go)

.PHONY: submodules
submodules:
	git submodule update --init --recursive

.PHONY: clean
clean:
	$(RM) wasiunix
