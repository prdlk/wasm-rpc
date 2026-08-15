# web-streaming

A Vite + React app consuming a **Go server-stream running inside the
page**. `server/main.go` (reusing `examples/server-streaming`'s
service) compiles to `public/server.wasm`; React subscribes through the
generated `TickerServiceClient` and renders ticks live.

```
server/main.go    Go entrypoint reusing examples/server-streaming
src/App.tsx       AsyncIterable consumption + cancellation UI
public/           server.wasm + wasm_exec.js (built artifacts)
Makefile          orchestrates Go build + client dist + npm
```

## Run

```sh
make dev        # builds server.wasm, the TS client package, then vite dev
make build      # production bundle in dist/
make server-tinygo && npm run dev   # ~1.2 MB TinyGo module (see caveat below)
```

## What it demonstrates

- **Server-streaming as an `AsyncIterable`**: `for await (const tick of
  client.subscribe({topic, count, intervalMs}))` — each frame is one
  protobuf message, one byte-copy out of the Go module.
- **Cancellation**: `stream.cancel()` resolves the iterator; inside Go
  the handler observes it as a `Send` error / `ctx.Done()`.
- **Panic recovery**: the "Trigger handler panic" button uses topic
  `"explode"` — the Go panic is recovered into a typed `internal`
  error (stack on the browser console only). With a TinyGo-built
  server this instead aborts the module: `recover()` is unsupported on
  wasm, which is exactly why TinyGo services should return
  `server.Errorf` rather than panic.
