//go:build js && wasm

package server

import (
	"context"
	"sync"
	"syscall/js"
)

// Streaming over the JS boundary
// ------------------------------
// listen() copies the request once (CopyBytesToGo), runs the handler on
// a goroutine, and pushes each frame through one CopyBytesToJS into the
// caller's onMessage. Streams are identified by numeric IDs so the
// exported cancel() is a single persistent js.Func — no per-stream
// js.Func allocations to leak or release.

var streamReg = struct {
	sync.Mutex
	next    uint32
	cancels map[uint32]context.CancelFunc
}{cancels: make(map[uint32]context.CancelFunc)}

// jsListen(method string, payload Uint8Array, handlers {onMessage,onError,onEnd}) -> id
func (r *Router) jsListen(_ js.Value, args []js.Value) any {
	fail := func(e *Error) any {
		// Deliver asynchronously if we can, so callers always observe
		// callback-style errors; fall back to a thrown JS error.
		if len(args) == 3 && args[2].Type() == js.TypeObject {
			if onErr := args[2].Get("onError"); onErr.Type() == js.TypeFunction {
				onErr.Invoke(errToJS(e))
				return js.ValueOf(0)
			}
		}
		panic(js.Error{Value: errToJS(e)})
	}

	if len(args) != 3 {
		return fail(&Error{Code: CodeInvalidArgument, Message: "listen(method, payload, handlers) requires exactly 3 arguments"})
	}
	if args[0].Type() != js.TypeString {
		return fail(&Error{Code: CodeInvalidArgument, Message: "method must be a string"})
	}
	method := args[0].String()
	payload := args[1]
	if payload.Type() != js.TypeObject || payload.Get("byteLength").IsUndefined() {
		return fail(&Error{Code: CodeInvalidArgument, Method: method, Message: "payload must be a Uint8Array"})
	}
	handlers := args[2]
	onMessage := handlers.Get("onMessage")
	onError := handlers.Get("onError")
	onEnd := handlers.Get("onEnd")
	if onMessage.Type() != js.TypeFunction {
		return fail(&Error{Code: CodeInvalidArgument, Method: method, Message: "handlers.onMessage must be a function"})
	}

	n := payload.Get("byteLength").Int()
	buf := reqPool.Get().([]byte)
	if cap(buf) < n {
		buf = make([]byte, n)
	} else {
		buf = buf[:n]
	}
	if copied := js.CopyBytesToGo(buf, payload); copied != n {
		reqPool.Put(buf[:0]) //nolint:staticcheck
		return fail(&Error{Code: CodeInternal, Method: method, Message: "short copy across Wasm boundary"})
	}

	ctx, cancel := context.WithCancel(context.Background())
	streamReg.Lock()
	streamReg.next++
	id := streamReg.next
	streamReg.cancels[id] = cancel
	streamReg.Unlock()

	go func() {
		defer func() {
			streamReg.Lock()
			delete(streamReg.cancels, id)
			streamReg.Unlock()
			cancel()
		}()

		send := func(frame []byte) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			u8 := uint8ArrayCtor.New(len(frame))
			js.CopyBytesToJS(u8, frame)
			onMessage.Invoke(u8)
			return nil
		}

		rpcErr := r.DispatchStream(ctx, method, buf, send)
		reqPool.Put(buf[:0]) //nolint:staticcheck

		switch {
		case rpcErr != nil:
			if rpcErr.detail != "" {
				console.Call("error", "[wasm-rpc] "+method+" panicked:\n"+rpcErr.detail)
			}
			if onError.Type() == js.TypeFunction {
				onError.Invoke(errToJS(rpcErr))
			}
		default:
			if onEnd.Type() == js.TypeFunction {
				onEnd.Invoke()
			}
		}
	}()

	return js.ValueOf(int(id))
}

// jsCancel(id number) — idempotent; the stream terminates with onEnd.
func (r *Router) jsCancel(_ js.Value, args []js.Value) any {
	if len(args) != 1 || args[0].Type() != js.TypeNumber {
		return js.Undefined()
	}
	id := uint32(args[0].Int())
	streamReg.Lock()
	cancel, ok := streamReg.cancels[id]
	streamReg.Unlock()
	if ok {
		cancel()
	}
	return js.Undefined()
}
