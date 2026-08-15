# server-basic

The minimal wasm-rpc server: `echo.v1.EchoService`, two unary methods,
one plain Go package. The point of this example is the shape of the
code — a service is an ordinary struct implementing the generated
`EchoServiceServer` interface, and each deployment target is three lines
of wiring.

```
service.go        the implementation (no wasm, no js — plain Go)
service_test.go   loopback test through the generated Go client
wasm/main.go      browser entrypoint (syscall/js bridge)
wasip1/main.go    WASI reactor entrypoint (go:wasmexport ABI)
```

## Build the reusable binaries

```sh
make wasm            # bin/server.wasm + wasm_exec.js  (browser, big Go)
make wasm-tinygo     # ~1.2 MB browser build via TinyGo
make wasip1          # bin/server.wasip1.wasm  (any WASI host: wazero, wasmtime…)
make wasip1-tinygo   # ~1.1 MB WASI build via TinyGo
make test            # loopback round-trip via the generated Go client
```

Drop `bin/server.wasm` + the matching `wasm_exec.js` into any web app —
`examples/web-echo` does exactly that via its own `server/` wrapper.
For WASI hosts, `examples/e2e/wasip1host` shows the complete embedding
(alloc/invoke/free over linear memory).

## Error contract demonstrated here

`Echo` rejects `repeat > 1000` with
`server.Errorf(server.CodeInvalidArgument, …)` — every client language
receives it as a typed `{code, method, message}` rejection rather than
a string. `Reverse` shows rune-safe UTF-8 across the boundary.

## TinyGo note

Return `server.Errorf` instead of panicking in services you intend to
build with TinyGo: `recover()` is unsupported on wasm, so a panic
aborts the module instead of degrading to an `internal` error.
