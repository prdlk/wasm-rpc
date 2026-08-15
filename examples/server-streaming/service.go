// Package streaming implements ticker.v1.TickerService — the reference
// server-streaming wasm-rpc service. One request in, a paced,
// cancellable sequence of responses out.
package streaming

import (
	"context"
	"time"

	tickerv1 "github.com/prdlk/wasm-rpc/gen/go/ticker/v1"
	"github.com/prdlk/wasm-rpc/server"
)

// Service implements tickerv1.TickerServiceServer.
type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Subscribe(ctx context.Context, req *tickerv1.SubscribeRequest, stream *server.ServerStream[*tickerv1.SubscribeResponse]) error {
	if req.GetTopic() == "explode" {
		panic("subscribe handler exploded") // exercises stream panic recovery (big Go)
	}
	count := req.GetCount()
	if count == 0 {
		count = 3
	}
	if count > 10000 {
		return server.Errorf(server.CodeInvalidArgument, "count must be <= 10000")
	}
	interval := time.Duration(req.GetIntervalMs()) * time.Millisecond
	if interval == 0 {
		interval = time.Millisecond
	}
	if interval > 10*time.Second {
		return server.Errorf(server.CodeInvalidArgument, "interval_ms must be <= 10000")
	}

	for i := uint32(1); i <= count; i++ {
		if err := stream.Send(&tickerv1.SubscribeResponse{
			Topic:      req.GetTopic(),
			Seq:        i,
			UnixMillis: time.Now().UnixMilli(),
		}); err != nil {
			return err // cancelled
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		// time.Sleep (not select+time.After) keeps pacing portable to
		// TinyGo's asyncify scheduler in wasip1 reactor mode.
		time.Sleep(interval)
	}
	return nil
}
