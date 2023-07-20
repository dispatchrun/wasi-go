# WASI-HTTP
This module implements the [wasi-http](https://github.com/WebAssembly/wasi-http) specification.
The specification is in active development/flux as is the [`wit-bindgen`](https://github.com/bytecodealliance/wit-bindgen) tool which is used to generate client libraries.

You should expect a degree of instability in these interfaces for the foreseeable future.

## Versioning & Clients
WASI-http doesn't have a great versioning story currently, so explicitly tying to git commit SHA and
`wit-bindgen` tooling is the best we can do.

Currently, the `wasi-go` codebase has the following versions:
* `v1` corresponds to [244e068c2d](https://github.com/WebAssembly/wasi-http/tree/244e068c2de43088bda308fcdf51ed2479d885f5) and [`wasmtime` 9.0.0](https://github.com/bytecodealliance/wasmtime/tree/release-9.0.0)

If you want to generate guest code for your particular language, you will need to use [`wit-bindgen` 0.4.0](https://github.com/bytecodealliance/wit-bindgen/releases/tag/wit-bindgen-cli-0.4.0)

Here is an example for generating the Golang guest bindings:

```sh
# Setup
git clone https://github.com/WebAssembly/wasi-http
cd wasi-http
# Sync to the definitions for the v1
git checkout 244e068
cd ..

# Generate
wit-bindgen tiny-go wasi-http/wit -w proxy
```

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

