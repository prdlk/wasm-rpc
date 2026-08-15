//go:build js && wasm

// Command server is the Wasm RPC server bundled with the web-streaming
// Vite app, reusing the server-streaming service implementation.
package main

import (
	"syscall/js"

	streaming "github.com/hyperauth/wasm-rpc/examples/server-streaming"
	tickerv1 "github.com/hyperauth/wasm-rpc/gen/go/ticker/v1"
	"github.com/hyperauth/wasm-rpc/server"
)

func main() {
	r := server.NewRouter()
	tickerv1.RegisterTickerServiceServer(r, streaming.New())
	r.Mount("wasmRPC")

	if ready := js.Global().Get("__wasmRPCReady"); !ready.IsUndefined() {
		ready.Invoke()
	}
	select {}
}
