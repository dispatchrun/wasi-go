[![Build](https://github.com/stealthrocket/wasi-go/actions/workflows/wasi-testuite.yml/badge.svg)](https://github.com/stealthrocket/wasi-go/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/wasi-go.svg)](https://pkg.go.dev/github.com/stealthrocket/wasi-go)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# WASI

The [WebAssembly][wasm] System Interface ([WASI][wasi]) is a set of standard
system functions that allow WebAssembly modules to interact with the outside
world (e.g. perform I/O, read clocks).

The WASI standard is under development. This repository provides a Go
implementation of WASI [preview 1][preview1] for Unix systems, and a command
to run WebAssembly modules that use WASI.

## Motivation

WASI preview 1 was sealed without a complete socket API, and WASI as a standard
is still a moving target.

Some WebAssembly runtimes have taken the initiative to either extend WASI
preview 1 or provide alternative solutions for capabilities that were missing
from the core specification, enabling a wider range of applications to run as
WebAssembly modules.

This package intends to bundle WASI preview 1 and these extensions with the
[wazero][wazero] runtime, and more generally be a playground for
experimentation with cutting-edge WASI features.

:electric_plug: **Sockets**

This library provides all the socket capabilities specified in WASI preview 1,
as well as a full support for a socket API which is ABI-compatible with the
extensions implemented in the [wasmedge][wasmedge] runtime.

:zap: **Performance**

The provided implementation of WASI is a thin zero-allocation layer around OS
system calls. Non-blocking I/O is fully supported, allowing WebAssembly modules
with an embedded scheduler (e.g. the Go runtime, or Rust Tokio scheduler) to
schedule goroutines / green threads while waiting for I/O.

:battery: **Experimentation**

The library separates the implementation of WASI from the WebAssembly runtime
host module, so that implementations of the provided WASI interface don't have
to worry about ABI concerns. The design makes it easy to wrap and augment WASI,
and keep up with the evolving WASI specification.

## Non-Goals

`wasi-go` does not aim to be a drop-in replacement for the `wasi_snapshot_preview1`
package that ships with the [wazero][wazero] runtime. For example, the `wasi-go`
package does not build on Windows, nor does it allow customization of the file
systems via a `fs.FS`.

## Usage

### As a Command

A `wasirun` command is provided for running WebAssembly modules that use WASI host imports.
It bundles the WASI implementation from this repository with the [wazero][wazero] runtime.

```console
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

As the providers of a Go implementation of WASI, we're naturally interested in
Go's support for WebAssembly and WASI, and are championing the efforts to make
Go a first class citizen in the ecosystem (along with Rust and Zig).

Go v1.21, scheduled for release in August 2023, has native support for
WebAssembly and WASI. To test these features before release, use [`gotip`][gotip]
in place of `go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

```console
$ GOOS=wasip1 GOARCH=wasm gotip build -o hello.wasm hello.go
$ wasirun hello.wasm
Hello, World!
```

This repository bundles [a script][go-script] that can be used to skip the
`go build` step.

## Contributing

Pull requests are welcome! Anything that is not a simple fix would probably
benefit from being discussed in an issue first.

Remember to be respectful and open minded!

[wasm]: https://webassembly.org
[wasi]: https://wasi.dev
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
[wasmer]: https://github.com/wasmerio/wasmer
[wasmedge]: https://github.com/WasmEdge/WasmEdge
[lunatic]: https://github.com/lunatic-solutions/lunatic
