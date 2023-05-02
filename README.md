# WASI

This is a Go implementation of the WebAssembly System Interface ([WASI][wasi]).

[wasi]: https://github.com/WebAssembly/WASI

## Package Layout

- `./`: types and constants from the [WASI preview 1 specification][preview-1]
- `./wasiunix`: a Unix implementation
- `./wasizero`: a host module for the [wazero][wazero] runtime

[preview-1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wazero]: https://wazero.io
