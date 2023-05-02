test: wasiunix submodules
	python3 wasi-testsuite/test-runner/wasi_test_runner.py \
		-t wasi-testsuite/tests/assemblyscript/testsuite \
		   wasi-testsuite/tests/c/testsuite \
		   wasi-testsuite/tests/rust/testsuite \
		-r adapter.py

wasiunix: ../cmd/wasiunix/main.go
	go build -o $@ $<

.PHONY: submodules
submodules:
	git submodule update --init --recursive

.PHONY: clean
clean:
	$(RM) wasiunix
