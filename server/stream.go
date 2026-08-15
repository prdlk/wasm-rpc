package server

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/protobuf/proto"
)

// Server-streaming
// ----------------
// A stream handler receives the request payload once and emits zero or
// more response frames through send. Frame bytes are only valid for the
// duration of the send call (bridges copy them across the boundary
// immediately); handlers must not retain them.

// SendFunc emits one response frame to the client. It returns a non-nil
// error once the stream is cancelled; handlers should stop promptly.
type SendFunc func(msg []byte) error

// StreamHandler is the raw byte-level server-streaming handler.
type StreamHandler func(ctx context.Context, payload []byte, send SendFunc) error

// RegisterRawStream installs a byte-level server-streaming handler.
// Prefer RegisterServerStream for typed Protobuf handlers.
func (r *Router) RegisterRawStream(method string, h StreamHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.streams[method]; dup {
		panic(fmt.Sprintf("wasmrpc: duplicate stream registration for %q", method))
	}
	if _, dup := r.handlers[method]; dup {
		panic(fmt.Sprintf("wasmrpc: %q already registered as unary", method))
	}
	r.streams[method] = h
}

// DispatchStream routes payload to the registered stream handler,
// converting panics and errors into *Error values. It returns nil on
// normal completion (including client cancellation surfaced as a nil
// send error contract — bridges translate ctx cancellation themselves).
func (r *Router) DispatchStream(ctx context.Context, method string, payload []byte, send SendFunc) (rpcErr *Error) {
	defer func() {
		if rec := recover(); rec != nil {
			rpcErr = &Error{
				Code:    CodeInternal,
				Method:  method,
				Message: fmt.Sprintf("panic in stream handler: %v", rec),
				detail:  string(debug.Stack()),
			}
		}
	}()

	r.mu.RLock()
	h, ok := r.streams[method]
	_, isUnary := r.handlers[method]
	r.mu.RUnlock()
	if !ok {
		if isUnary {
			return &Error{Code: CodeInvalidArgument, Method: method, Message: "unary method: use invoke, not listen"}
		}
		return &Error{Code: CodeUnimplemented, Method: method, Message: "method not registered"}
	}

	if err := h(ctx, payload, send); err != nil {
		// Client cancellation is a normal termination, not an error.
		if ctx.Err() != nil {
			return nil
		}
		return asError(err, method)
	}
	return nil
}

// ServerStream is the typed sender handed to streaming handlers.
type ServerStream[T proto.Message] struct {
	ctx  context.Context
	send SendFunc
}

// Context returns the stream context; it is cancelled when the client
// cancels the stream.
func (s *ServerStream[T]) Context() context.Context { return s.ctx }

// Send marshals and emits one message. Marshal uses a pooled buffer;
// the frame is copied across the boundary before Send returns.
func (s *ServerStream[T]) Send(m T) error {
	if err := s.ctx.Err(); err != nil {
		return err
	}
	out, err := marshalMessage(m)
	if err != nil {
		return err
	}
	err = s.send(out)
	RecycleResponse(out)
	return err
}

// RegisterServerStream installs a strongly-typed server-streaming
// handler; type parameters are inferred from the handler signature:
//
//	server.RegisterServerStream(r, "echo.v1.EchoService/Watch", svc.Watch)
func RegisterServerStream[Req any, Resp proto.Message, PReq interface {
	proto.Message
	*Req
}](r *Router, method string, h func(context.Context, PReq, *ServerStream[Resp]) error) {
	r.RegisterRawStream(method, func(ctx context.Context, payload []byte, send SendFunc) error {
		req := PReq(new(Req))
		if err := unmarshalMessage(payload, req); err != nil {
			return &Error{Code: CodeInvalidArgument, Method: method, Message: "unmarshal request: " + err.Error()}
		}
		return h(ctx, req, &ServerStream[Resp]{ctx: ctx, send: send})
	})
}
