// Package wasmrpcclient provides the native Go client runtime for
// wasm-rpc: a Transport abstraction plus an in-process loopback used to
// exercise generated clients against a Router without a Wasm boundary.
package wasmrpcclient

import (
	"context"

	"github.com/prdlk/wasm-rpc/server"
)

// Transport moves opaque protobuf frames to a wasm-rpc router.
type Transport interface {
	Invoke(ctx context.Context, method string, payload []byte) ([]byte, error)
}

// Error mirrors the standardized rejection payload seen by every client
// language: {code, method, message}.
type Error struct {
	Code    string
	Method  string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Loopback adapts a Router into a StreamTransport (superset of
// Transport) for tests and native hosts.
func Loopback(r *server.Router) StreamTransport { return loopback{r} }

type loopback struct{ r *server.Router }

func (l loopback) Invoke(ctx context.Context, method string, payload []byte) ([]byte, error) {
	out, rpcErr := l.r.Dispatch(ctx, method, payload)
	if rpcErr != nil {
		return nil, &Error{Code: string(rpcErr.Code), Method: rpcErr.Method, Message: rpcErr.Message}
	}
	// Copy: Dispatch responses may use pooled buffers on the wasm side;
	// keep loopback semantics identical (caller owns the slice).
	dup := make([]byte, len(out))
	copy(dup, out)
	server.RecycleResponse(out)
	return dup, nil
}

// StreamTransport extends Transport with server-streaming. Generated
// clients for services containing streaming methods require it.
type StreamTransport interface {
	Transport
	// Listen blocks until the stream ends, invoking onMsg per frame.
	// Frame bytes are only valid during the callback. A non-nil error
	// from onMsg cancels the stream and is returned; ctx cancellation
	// returns ctx.Err().
	Listen(ctx context.Context, method string, payload []byte, onMsg func([]byte) error) error
}

func (l loopback) Listen(ctx context.Context, method string, payload []byte, onMsg func([]byte) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var cbErr error
	rpcErr := l.r.DispatchStream(ctx, method, payload, func(frame []byte) error {
		if err := onMsg(frame); err != nil {
			cbErr = err
			cancel()
			return err
		}
		return nil
	})
	switch {
	case cbErr != nil:
		return cbErr
	case rpcErr != nil:
		return &Error{Code: string(rpcErr.Code), Method: rpcErr.Method, Message: rpcErr.Message}
	case ctx.Err() != nil && cbErr == nil && rpcErr == nil:
		return ctx.Err()
	}
	return nil
}
