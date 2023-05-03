# Contibution Guidelines

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
