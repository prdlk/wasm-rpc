//go:build js && wasm

// Command wasm is the combined test guest for the Node E2E: it mounts
// both example services (echo unary + ticker streaming) so one module
// exercises every code path.
package main

import (
	"syscall/js"

	basic "github.com/prdlk/wasm-rpc/examples/server-basic"
	streaming "github.com/prdlk/wasm-rpc/examples/server-streaming"
	echov1 "github.com/prdlk/wasm-rpc/gen/go/echo/v1"
	tickerv1 "github.com/prdlk/wasm-rpc/gen/go/ticker/v1"
	"github.com/prdlk/wasm-rpc/server"
)

func main() {
	r := server.NewRouter()
	echov1.RegisterEchoServiceServer(r, basic.New())
	tickerv1.RegisterTickerServiceServer(r, streaming.New())
	r.Mount("wasmRPC")

	if ready := js.Global().Get("__wasmRPCReady"); !ready.IsUndefined() {
		ready.Invoke()
	}
	select {}
}
