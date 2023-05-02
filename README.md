# WASI

This is a Go implementation of the WebAssembly System Interface ([WASI][wasi]).

## Package Layout

- `.`: types and constants from the [WASI preview 1 specification][preview1]
- [`wasiunix`][wasiunix]: a Unix implementation
- [`wasizero`][wasizero]: a host module for the [wazero][wazero] runtime


[wasi]: https://github.com/WebAssembly/WASI
[preview1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wasiunix]: https://github.com/stealthrocket/wasi/tree/main/wasiunix
[wasizero]: https://github.com/stealthrocket/wasi/tree/main/wasizero
[wazero]: https://wazero.io
