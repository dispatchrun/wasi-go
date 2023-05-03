[![Build](https://github.com/stealthrocket/wasi/actions/workflows/go.yml/badge.svg)](https://github.com/stealthrocket/wasi/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/wasi.svg)](https://pkg.go.dev/github.com/stealthrocket/wasi)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# WASI

A Go implementation of the WebAssembly System Interface ([WASI][wasi]).

## Package Layout

- `.`: types, constants and an [interface][provider] for [WASI preview 1][preview1]
- [`wasiunix`](wasiunix): a Unix implementation
- [`imports/wasi_snapshot_preview1`](imports/wasi_snapshot_preview1): a host module for the [wazero][wazero] runtime


[wasi]: https://github.com/WebAssembly/WASI
[provider]: https://github.com/stealthrocket/wasi/blob/main/provider.go
[preview1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wazero]: https://wazero.io
