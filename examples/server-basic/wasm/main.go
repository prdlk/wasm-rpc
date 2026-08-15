//go:build js && wasm

// Command wasm builds the reusable browser-resident echo server:
//
//	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o server.wasm ./examples/server-basic/wasm
//
// Drop the resulting server.wasm (plus Go's wasm_exec.js) into any web
// app and drive it through the generated clients.
package main

import (
	"syscall/js"

	basic "github.com/prdlk/wasm-rpc/examples/server-basic"
	echov1 "github.com/prdlk/wasm-rpc/gen/go/echo/v1"
	"github.com/prdlk/wasm-rpc/server"
)

func main() {
	r := server.NewRouter()
	echov1.RegisterEchoServiceServer(r, basic.New())
	r.Mount("wasmRPC")

	if ready := js.Global().Get("__wasmRPCReady"); !ready.IsUndefined() {
		ready.Invoke()
	}
	select {}
}
