//go:build js && wasm

// Command server is the Wasm RPC server bundled with the web-echo Vite
// app. It reuses the server-basic service implementation unchanged —
// the point of the example: services are plain Go packages, entrypoints
// are three lines of wiring.
package main

import (
	"syscall/js"

	basic "github.com/hyperauth/wasm-rpc/examples/server-basic"
	echov1 "github.com/hyperauth/wasm-rpc/gen/go/echo/v1"
	"github.com/hyperauth/wasm-rpc/server"
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
