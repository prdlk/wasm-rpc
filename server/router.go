// Package wasmrpc implements a minimal Protobuf-over-Wasm RPC runtime.
//
// The router itself is platform-independent (compiles and unit-tests on
// native GOOS) — only bridge.go is gated behind js/wasm build tags.
package server

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Handler is the raw byte-level handler signature. Payload is only valid
// for the duration of the call: the bridge recycles the backing buffer,
// so handlers must not retain it. (proto.Unmarshal copies, so typed
// handlers registered via Register are always safe.)
type Handler func(ctx context.Context, payload []byte) ([]byte, error)

// Router maps fully-qualified method names ("pkg.v1.Service/Method")
// to handlers. Registration is expected at startup; dispatch is
// concurrency-safe afterwards.
type Router struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	streams  map[string]StreamHandler
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
		streams:  make(map[string]StreamHandler),
	}
}

// RegisterRaw installs a byte-level handler. Prefer Register for typed
// Protobuf handlers.
func (r *Router) RegisterRaw(method string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.handlers[method]; dup {
		panic(fmt.Sprintf("wasmrpc: duplicate registration for %q", method))
	}
	if _, dup := r.streams[method]; dup {
		panic(fmt.Sprintf("wasmrpc: %q already registered as stream", method))
	}
	r.handlers[method] = h
}

// Methods returns the registered method names (for introspection/tests).
func (r *Router) Methods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.handlers)+len(r.streams))
	for m := range r.handlers {
		out = append(out, m)
	}
	for m := range r.streams {
		out = append(out, m)
	}
	return out
}

// Dispatch decodes nothing itself — it routes payload bytes to the
// handler and converts panics and plain errors into *Error values so the
// bridge can reject the JS Promise with a standardized payload instead
// of crashing the Wasm module.
func (r *Router) Dispatch(ctx context.Context, method string, payload []byte) (out []byte, rpcErr *Error) {
	defer func() {
		if rec := recover(); rec != nil {
			rpcErr = &Error{
				Code:    CodeInternal,
				Method:  method,
				Message: fmt.Sprintf("panic in handler: %v", rec),
				detail:  string(debug.Stack()),
			}
		}
	}()

	r.mu.RLock()
	h, ok := r.handlers[method]
	_, isStream := r.streams[method]
	r.mu.RUnlock()
	if !ok {
		if isStream {
			return nil, &Error{Code: CodeInvalidArgument, Method: method, Message: "server-streaming method: use listen, not invoke"}
		}
		return nil, &Error{Code: CodeUnimplemented, Method: method, Message: "method not registered"}
	}

	resp, err := h(ctx, payload)
	if err != nil {
		return nil, asError(err, method)
	}
	return resp, nil
}

// Register installs a strongly-typed Protobuf handler. Type parameters
// are inferred from the handler signature:
//
//	wasmrpc.Register(r, "echo.v1.EchoService/Echo", svc.Echo)
//
// Unmarshal failures reject with CodeInvalidArgument; handler errors are
// passed through asError; marshal uses MarshalAppend into a pooled
// buffer to keep steady-state allocations flat.
func Register[Req any, Resp any, PReq interface {
	proto.Message
	*Req
}, PResp interface {
	proto.Message
	*Resp
}](r *Router, method string, h func(context.Context, PReq) (PResp, error)) {
	r.RegisterRaw(method, func(ctx context.Context, payload []byte) ([]byte, error) {
		req := PReq(new(Req))
		if err := unmarshalMessage(payload, req); err != nil {
			return nil, &Error{Code: CodeInvalidArgument, Method: method, Message: "unmarshal request: " + err.Error()}
		}
		resp, err := h(ctx, req)
		if err != nil {
			return nil, err
		}
		out, err := marshalMessage(resp)
		if err != nil {
			return nil, &Error{Code: CodeInternal, Method: method, Message: "marshal response: " + err.Error()}
		}
		// Note: out may alias buf (or a grown copy). The bridge copies it
		// to JS and then calls RecycleResponse, returning it to the pool.
		return out, nil
	})
}

// TinyGo compatibility: google.golang.org/protobuf's reflection path
// (reflect.NewAt) is unsupported by TinyGo. When messages carry
// vtprotobuf's generated fast-path methods, the runtime uses them and
// never touches protobuf reflection — making TinyGo wasip1/js builds
// viable and speeding up big-Go builds as a bonus.

type vtMarshaler interface{ MarshalVT() ([]byte, error) }
type vtUnmarshaler interface{ UnmarshalVT([]byte) error }

func marshalMessage(m proto.Message) ([]byte, error) {
	if vt, ok := any(m).(vtMarshaler); ok {
		return vt.MarshalVT()
	}
	buf := marshalPool.Get().([]byte)
	out, err := proto.MarshalOptions{}.MarshalAppend(buf[:0], m)
	if err != nil {
		marshalPool.Put(buf) //nolint:staticcheck
		return nil, err
	}
	return out, nil
}

func unmarshalMessage(payload []byte, m proto.Message) error {
	if vt, ok := any(m).(vtUnmarshaler); ok {
		return vt.UnmarshalVT(payload)
	}
	return proto.Unmarshal(payload, m)
}

var marshalPool = sync.Pool{
	New: func() any { return make([]byte, 0, 4096) },
}

// RecycleResponse returns a response buffer produced by a typed handler
// to the marshal pool. Called by the bridge after CopyBytesToJS.
func RecycleResponse(b []byte) {
	if cap(b) == 0 {
		return
	}
	marshalPool.Put(b[:0]) //nolint:staticcheck
}
