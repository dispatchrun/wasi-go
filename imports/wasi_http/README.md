# WASI-HTTP
This module implements the [wasi-http](https://github.com/WebAssembly/wasi-http) specification.
The specification is in active development/flux as is the [`wit-bindgen`](https://github.com/bytecodealliance/wit-bindgen) tool which is used to generate client libraries.

You should expect a degree of instability in these interfaces for the foreseeable future.

## Example guest code
There are existing examples of working guest code in the following languages
* [Golang](https://github.com/dev-wasm/dev-wasm-go/tree/main/http)
* [C](https://github.com/dev-wasm/dev-wasm-c/tree/main/http)
* [AssemblyScript](https://github.com/dev-wasm/dev-wasm-ts/tree/main/http)
* [Dotnet](https://github.com/dev-wasm/dev-wasm-dotnet/tree/main/http)
* [Rust](https://github.com/bytecodealliance/wasmtime/blob/main/crates/test-programs/wasi-http-tests/src/bin/outbound_request.rs)

## Example server code
There is an example server in the following languages (more to come):
* [Golang](https://github.com/dev-wasm/dev-wasm-go/blob/main/http/server.go)
* [C](https://github.com/brendandburns/wasi-go/blob/server/testdata/c/http/server.c)

