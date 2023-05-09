[![Build](https://github.com/stealthrocket/wasi-go/actions/workflows/wasi-testuite.yml/badge.svg)](https://github.com/stealthrocket/wasi-go/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/wasi-go.svg)](https://pkg.go.dev/github.com/stealthrocket/wasi-go)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# WASI

The WebAssembly System Interface ([WASI][wasi]) is a set of system calls that
allow WebAssembly modules to interact with the outside world (e.g. perform
I/O, read clocks).

The WASI [standard][wasi] is still under development. This repository provides
a Go implementation of [WASI preview 1][preview1] for Unix systems, and a command
to run WebAssembly modules that use WASI system calls.

## Goals

:zap: **Performance**

The provided implementation of WASI is a thin zero-allocation layer around OS
system calls.

Non-blocking I/O is fully supported, allowing WebAssembly modules with an embedded
scheduler (e.g. the Go runtime, or Rust Tokio scheduler) to schedule green
threads while waiting for I/O.

:battery: **Extensibility**

The library separates the implementation of WASI from the WebAssembly runtime host
module, so that implementations of the provided [WASI interface][system] don't
have to worry about ABI concerns.

The design makes it easy to wrap, augment and extend WASI, for example see the
provided [tracer][tracer] and a basic [sockets extension][path_open_sockets].

## Package Layout

- `.`: types, constants and an [interface][system] for [WASI preview 1][preview1]
- [`systems/unix`][unix-system]: a Unix implementation
- [`imports/wasi_snapshot_preview1`][host-module]: a host module for the [wazero][wazero] runtime
- [`cmd/wasirun`][wasirun]: a command to run WebAssembly modules that use WASI system calls

[wasi]: https://github.com/WebAssembly/WASI
[system]: https://github.com/stealthrocket/wasi-go/blob/main/system.go
[unix-system]: https://github.com/stealthrocket/wasi-go/blob/main/systems/unix/system.go
[host-module]: https://github.com/stealthrocket/wasi-go/blob/main/imports/wasi_snapshot_preview1/module.go
[preview1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wazero]: https://wazero.io
[wasirun]: https://github.com/stealthrocket/wasi-go/tree/main/cmd/wasirun
[tracer]: https://github.com/stealthrocket/wasi-go/blob/main/tracer.go
[path_open_sockets]: https://github.com/stealthrocket/wasi-go/blob/main/systems/unix/path_open_sockets.go
