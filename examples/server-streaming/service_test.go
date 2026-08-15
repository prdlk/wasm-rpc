package streaming_test

import (
	"context"
	"errors"
	"testing"

	wasmrpcclient "github.com/prdlk/wasm-rpc/client/go"
	streaming "github.com/prdlk/wasm-rpc/examples/server-streaming"
	tickerv1 "github.com/prdlk/wasm-rpc/gen/go/ticker/v1"
	"github.com/prdlk/wasm-rpc/server"
)

func newClient() *tickerv1.TickerServiceClient {
	r := server.NewRouter()
	tickerv1.RegisterTickerServiceServer(r, streaming.New())
	return tickerv1.NewTickerServiceClient(wasmrpcclient.Loopback(r))
}

func TestStreamLoopback(t *testing.T) {
	c := newClient()

	var seqs []uint32
	err := c.Subscribe(context.Background(), &tickerv1.SubscribeRequest{Topic: "t", Count: 3},
		func(m *tickerv1.SubscribeResponse) error {
			if m.GetTopic() != "t" {
				t.Fatalf("topic %q", m.GetTopic())
			}
			seqs = append(seqs, m.GetSeq())
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(seqs) != 3 || seqs[2] != 3 {
		t.Fatalf("seqs = %v", seqs)
	}
}

func TestStreamCancelViaCallback(t *testing.T) {
	c := newClient()
	n := 0
	sentinel := context.Canceled
	err := c.Subscribe(context.Background(), &tickerv1.SubscribeRequest{Count: 10},
		func(*tickerv1.SubscribeResponse) error {
			n++
			if n == 2 {
				return sentinel
			}
			return nil
		})
	if !errors.Is(err, sentinel) || n != 2 {
		t.Fatalf("cancel: err=%v n=%d", err, n)
	}
}

func TestStreamPanicIsInternal(t *testing.T) {
	c := newClient()
	err := c.Subscribe(context.Background(), &tickerv1.SubscribeRequest{Topic: "explode"},
		func(*tickerv1.SubscribeResponse) error { return nil })
	var rpcErr *wasmrpcclient.Error
	if !errors.As(err, &rpcErr) || rpcErr.Code != "internal" {
		t.Fatalf("panic: %v", err)
	}
}

func TestStreamTypedError(t *testing.T) {
	c := newClient()
	err := c.Subscribe(context.Background(), &tickerv1.SubscribeRequest{Count: 99999},
		func(*tickerv1.SubscribeResponse) error { return nil })
	var rpcErr *wasmrpcclient.Error
	if !errors.As(err, &rpcErr) || rpcErr.Code != "invalid_argument" {
		t.Fatalf("typed error: %v", err)
	}
}
