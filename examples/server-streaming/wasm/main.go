//go:build js && wasm

// Command wasm builds the reusable browser-resident ticker server:
//
//	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o server.wasm ./examples/server-streaming/wasm
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
