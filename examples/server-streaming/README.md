# server-streaming

The reference server-streaming wasm-rpc service:
`ticker.v1.TickerService/Subscribe` — one request in, a paced,
cancellable sequence of responses out. Same deployment matrix as
`server-basic` (browser + WASI, big Go + TinyGo); only the handler
shape differs.

```
service.go        streaming handler over server.ServerStream[T]
service_test.go   stream loopback: collect, cancel, typed error, panic
wasm/main.go      browser entrypoint
wasip1/main.go    WASI reactor entrypoint
```

## The streaming handler shape

```go
func (s *Service) Subscribe(ctx context.Context, req *tickerv1.SubscribeRequest,
	stream *server.ServerStream[*tickerv1.SubscribeResponse]) error {
	for i := uint32(1); i <= count; i++ {
		if err := stream.Send(&tickerv1.SubscribeResponse{Seq: i}); err != nil {
			return err // client cancelled
		}
		time.Sleep(interval)
	}
	return nil
}
```

`Send` returns an error once the client cancels; returning it ends the
stream cleanly. Client cancellation also cancels `ctx`. On the wire:
browsers get `listen`/`cancel` callbacks, WASI hosts get pull-based
`wasmrpc_stream_recv` with a 16-frame buffer for backpressure.

## Build & test

```sh
make wasm | wasm-tinygo | wasip1 | wasip1-tinygo   # bin/server*.wasm
make test                                          # 4 loopback stream tests
```

`examples/web-streaming` consumes this exact service from React as an
`AsyncIterable` with live cancellation.

## Deliberate behaviors for testing

- topic `"explode"` panics the handler: big Go recovers it into a typed
  `internal` error with the stack on console/stderr only; TinyGo aborts
  (recover() unsupported on wasm) — so don't panic in TinyGo services.
- `count > 10000` returns a typed `invalid_argument` before streaming.
- Pacing uses `time.Sleep`, not `select`+`time.After` — the latter
  breaks under TinyGo's asyncify scheduler in reactor mode.
