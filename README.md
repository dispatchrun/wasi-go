[![Build](https://github.com/stealthrocket/wasi-go/actions/workflows/wasi-testuite.yml/badge.svg)](https://github.com/stealthrocket/wasi-go/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/wasi-go.svg)](https://pkg.go.dev/github.com/stealthrocket/wasi-go)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# WASI

The [WebAssembly][wasm] System Interface ([WASI][wasi]) is a set of system calls
that allow WebAssembly modules to interact with the outside world (e.g. perform
I/O, read clocks).

The WASI standard is under development. This repository provides a Go
implementation of WASI [preview 1][preview1] for Unix systems, and a command
to run WebAssembly modules that use WASI system calls.

## Goals

:zap: **Performance**

The provided implementation of WASI is a thin zero-allocation layer around OS
system calls. Non-blocking I/O is fully supported, allowing WebAssembly modules
with an embedded scheduler (e.g. the Go runtime, or Rust Tokio scheduler) to
schedule goroutines / green threads while waiting for I/O.

:battery: **Extensibility**

The library separates the implementation of WASI from the WebAssembly runtime host
module, so that implementations of the provided WASI interface don't have to
worry about ABI concerns. The design makes it easy to wrap, augment and
extend WASI.

:electric_plug: **Sockets**

WASI preview 1 was unfortunately sealed before sockets support was complete.
Many WebAssembly runtimes extend WASI with system calls that allow the module
to create sockets, bind them to an address, listen for incoming connections
and make outbound connections. This library supports a few of these sockets
extensions.

## Usage

### As a Command

A `wasirun` command is provided for running WebAssembly modules that use WASI system calls.
It bundles the WASI implementation from this library with the [wazero][wazero] runtime.

```
$ go install github.com/stealthrocket/wasi-go/cmd/wasirun@latest
```

The `wasirun` command has many options for controlling the capabilities of the WebAssembly
module, and for tracing and profiling execution. See `wasirun --help` for details.

### As a Library

The package layout is as follows:

- `.` types, constants and an [interface][system] for WASI preview 1
- [`systems/unix`][unix-system] a Unix implementation (tested on Linux and macOS)
- [`imports/wasi_snapshot_preview1`][host-module] a host module for the [wazero][wazero] runtime
- [`cmd/wasirun`][wasirun] a command to run WebAssembly modules
- [`testwasi`][testwasi] a test suite against the WASI interface

To run a WebAssembly module, it's also necessary to prepare clocks and "preopens"
(files/directories that the WebAssembly module can access). To see how it all fits
together, see the implementation of the [wasirun][wasirun] command.

### With Go

As a Go implementation of WASI, we're naturally interested in Go's support for
WebAssembly and WASI, and are championing the efforts to make Go a first class
citizen in the ecosystem (along with Rust and Zig).

In Go v1.21 (scheduled for release in August 2023), Go will natively
compile WebAssembly modules with WASI system calls:

```go
$ GOOS=wasip1 GOARCH=wasm go build ...
```

To play around with this feature before release, [use `gotip`][gotip].

This repository provides [a script][go-script] that you can use to skip the
`go build` step and directly `go run` WebAssembly modules.


[wasm]: https://webassembly.org
[wasi]: https://github.com/WebAssembly/WASI
[system]: https://github.com/stealthrocket/wasi-go/blob/main/system.go
[unix-system]: https://github.com/stealthrocket/wasi-go/blob/main/systems/unix/system.go
[host-module]: https://github.com/stealthrocket/wasi-go/blob/main/imports/wasi_snapshot_preview1/module.go
[preview1]: https://github.com/WebAssembly/WASI/blob/e324ce3/legacy/preview1/docs.md
[wazero]: https://wazero.io
[wasirun]: https://github.com/stealthrocket/wasi-go/blob/main/cmd/wasirun/main.go
[testwasi]: https://github.com/stealthrocket/wasi-go/tree/main/testwasi
[tracer]: https://github.com/stealthrocket/wasi-go/blob/main/tracer.go
[sockets-extension]: https://github.com/stealthrocket/wasi-go/blob/main/sockets_extension.go
[gotip]: https://pkg.go.dev/golang.org/dl/gotip
[go-script]: https://github.com/stealthrocket/wasi-go/blob/main/share/go_wasip1_wasm_exec
