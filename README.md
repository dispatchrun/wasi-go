[![Build](https://github.com/stealthrocket/wasi-go/actions/workflows/go.yml/badge.svg)](https://github.com/stealthrocket/wasi-go/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/wasi-go.svg)](https://pkg.go.dev/github.com/stealthrocket/wasi-go)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# WASI

A Go implementation of the WebAssembly System Interface ([WASI][wasi]).

## Package Layout

- `.`: types, constants and an [interface][system] for [WASI preview 1][preview1]
- [`systems/unix`](systems/unix): a Unix implementation
- [`imports/wasi_snapshot_preview1`](imports/wasi_snapshot_preview1): a host module for the [wazero][wazero] runtime
- [`cmd/wasirun`][wasirun]: a command to run WASM modules that use WASI system calls

[wasi]: https://github.com/WebAssembly/WASI
[system]: https://github.com/stealthrocket/wasi-go/blob/main/system.go
[preview1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wazero]: https://wazero.io
[wasirun]: https://github.com/stealthrocket/wasi-go/tree/main/cmd/wasirun
