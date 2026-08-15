# web-echo

A Vite + React app whose backend is a **Go RPC server running inside
the page**: `server/main.go` compiles to `public/server.wasm`, React
talks to it through the generated `EchoServiceClient` — typed unary
calls, typed errors, zero HTTP.

```
server/main.go    Go entrypoint reusing examples/server-basic's service
src/App.tsx       React UI over the generated TS client
public/           server.wasm + wasm_exec.js land here (built artifacts)
Makefile          orchestrates Go build + client dist + npm
```

## Run

```sh
make dev        # builds server.wasm, the TS client package, then vite dev
make build      # production bundle in dist/ (server.wasm included)
make server-tinygo && npm run dev   # ~1.2 MB module instead of ~6.4 MB
```

## How the pieces connect

1. `index.html` loads `/wasm_exec.js` (Go's JS glue) via a plain script
   tag — it defines `globalThis.Go`.
2. `App.tsx` calls `loadWasmRpc("/server.wasm")`, which instantiates
   the module and resolves once the Go router mounts
   `globalThis.wasmRPC`.
3. The generated `EchoServiceClient` wraps the transport: protobuf
   encode → one byte-copy into Go → handler → one byte-copy out →
   decode. Handler errors reject as `WasmRpcError{code, method,
   message}` — the "Trigger typed error" button demonstrates it.

The npm dependency `@hyperauth/wasm-rpc-client` is a `file:` link to
`client/ts`, which contains both the runtime and the buf-generated
clients — regenerating the schema (`buf generate` at the repo root)
flows into this app on the next build.
