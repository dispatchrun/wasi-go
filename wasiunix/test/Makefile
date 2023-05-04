test: wasiunix submodules
	python3 wasi-testsuite/test-runner/wasi_test_runner.py \
		-t wasi-testsuite/tests/assemblyscript/testsuite \
		   wasi-testsuite/tests/c/testsuite \
		   wasi-testsuite/tests/rust/testsuite \
		-r adapter.py

wasi.src.go = $(wildcard ../../*.go)
wasiunix.src.go = $(wildcard ../*.go)
wasi_snapshot_preview1.src.go = $(wildcard ../../imports/wasi_snapshot_preview1/*.go)

wasiunix: ../cmd/wasiunix/main.go $(wasi.src.go) $(wasiunix.src.go) $(wasi_snapshot_preview1.src.go)
	go build -o $@ $<

.PHONY: submodules
submodules:
	git submodule update --init --recursive

.PHONY: clean
clean:
	$(RM) wasiunix
