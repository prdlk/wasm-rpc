package basic_test

import (
	"context"
	"errors"
	"testing"

	wasmrpcclient "github.com/prdlk/wasm-rpc/client/go"
	basic "github.com/prdlk/wasm-rpc/examples/server-basic"
	echov1 "github.com/prdlk/wasm-rpc/gen/go/echo/v1"
	"github.com/prdlk/wasm-rpc/server"
)

// Exercises the generated Go client + Register glue over the loopback
// transport — no Wasm boundary, pure protocol round-trip.
func TestGeneratedClientLoopback(t *testing.T) {
	r := server.NewRouter()
	echov1.RegisterEchoServiceServer(r, basic.New())
	c := echov1.NewEchoServiceClient(wasmrpcclient.Loopback(r))

	resp, err := c.Echo(context.Background(), &echov1.EchoRequest{Message: "hi", Repeat: 2})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetMessage() != "hihi" {
		t.Fatalf("got %q", resp.GetMessage())
	}

	rev, err := c.Reverse(context.Background(), &echov1.ReverseRequest{Message: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if rev.GetMessage() != "cba" {
		t.Fatalf("got %q", rev.GetMessage())
	}

	_, err = c.Echo(context.Background(), &echov1.EchoRequest{Repeat: 9999})
	var rpcErr *wasmrpcclient.Error
	if !errors.As(err, &rpcErr) || rpcErr.Code != "invalid_argument" {
		t.Fatalf("want invalid_argument, got %v", err)
	}
}
