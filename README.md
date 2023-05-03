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

## Troubleshooting

### wasm-ld: error: cannot open */libclang_rt.builtins-wasm32.a

This error may occur is the local clang installation does not have a version of
`libclang_rt` built for WebAssembly, for example:

e```
wasm-ld: error: cannot open {...}/lib/wasi/libclang_rt.builtins-wasm32.a: No such file or directory
clang-9: error: linker command failed with exit code 1 (use -v to see invocation)
```

This article describes how to solve the issue https://depth-first.com/articles/2019/10/16/compiling-c-to-webassembly-and-running-it-without-emscripten/
which instructs to download precompiled versions of the library distributed at https://github.com/jedisct1/libclang_rt.builtins-wasm32.a
and install them at the location where clang expects to find them (the path in the error message).

Here is a [direct link](https://raw.githubusercontent.com/jedisct1/libclang_rt.builtins-wasm32.a/master/precompiled/libclang_rt.builtins-wasm32.a)
to download the library.
