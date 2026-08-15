//go:build wasip1

// Command wasip1 builds the reusable WASI reactor for the ticker
// server (see server-basic/wasip1 for build commands).
package main

import (
	streaming "github.com/hyperauth/wasm-rpc/examples/server-streaming"
	tickerv1 "github.com/hyperauth/wasm-rpc/gen/go/ticker/v1"
	"github.com/hyperauth/wasm-rpc/server"
)

func init() {
	r := server.NewRouter()
	tickerv1.RegisterTickerServiceServer(r, streaming.New())
	server.Attach(r)
}

func main() {}
