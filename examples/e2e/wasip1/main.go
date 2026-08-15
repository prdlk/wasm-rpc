//go:build wasip1

// Command wasip1 is the combined test reactor for the wazero-hosted
// wasip1 E2E (both example services).
package main

import (
	basic "github.com/prdlk/wasm-rpc/examples/server-basic"
	streaming "github.com/prdlk/wasm-rpc/examples/server-streaming"
	echov1 "github.com/prdlk/wasm-rpc/gen/go/echo/v1"
	tickerv1 "github.com/prdlk/wasm-rpc/gen/go/ticker/v1"
	"github.com/prdlk/wasm-rpc/server"
)

func init() {
	r := server.NewRouter()
	echov1.RegisterEchoServiceServer(r, basic.New())
	tickerv1.RegisterTickerServiceServer(r, streaming.New())
	server.Attach(r)
}

func main() {}
