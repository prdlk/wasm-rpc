//go:build wasip1

// Command wasip1 builds the reusable WASI reactor for the echo server:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o server.wasip1.wasm ./examples/server-basic/wasip1
//	tinygo build -target=wasip1 -buildmode=c-shared -no-debug -o server.wasip1.wasm ./examples/server-basic/wasip1
//
// Reactors run inits (not main), so registration happens in init().
package main

import (
	basic "github.com/hyperauth/wasm-rpc/examples/server-basic"
	echov1 "github.com/hyperauth/wasm-rpc/gen/go/echo/v1"
	"github.com/hyperauth/wasm-rpc/server"
)

func init() {
	r := server.NewRouter()
	echov1.RegisterEchoServiceServer(r, basic.New())
	server.Attach(r)
}

func main() {}
