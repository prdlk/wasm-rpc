//go:build js && wasm

package server

import (
	"context"
	"sync"
	"syscall/js"
)

// Memory-boundary contract
// ------------------------
// Exactly one copy per direction, both via the sanctioned bulk-copy
// intrinsics — never element-wise iteration over a js.Value:
//
//	request:  js.CopyBytesToGo(pooled []byte, Uint8Array)   (1 copy)
//	response: js.CopyBytesToJS(new Uint8Array, []byte)      (1 copy)
//
// Request buffers come from a sync.Pool and are recycled after Dispatch
// returns (proto.Unmarshal copies field data, so reuse is safe).
// Response buffers are recycled via RecycleResponse after CopyBytesToJS.

var (
	uint8ArrayCtor = js.Global().Get("Uint8Array")
	promiseCtor    = js.Global().Get("Promise")
	errorCtor      = js.Global().Get("Error")
	console        = js.Global().Get("console")
)

var reqPool = sync.Pool{
	New: func() any { return make([]byte, 0, 4096) },
}

// Mount exports the router to the JS global scope under globalName:
//
//	globalThis.wasmRPC = {
//	  invoke(method: string, payload: Uint8Array): Promise<Uint8Array>,
//	  listen(method, payload, {onMessage,onError,onEnd}): number, // stream id
//	  cancel(id: number): void,
//	  methods(): string[],
//	}
func (r *Router) Mount(globalName string) {
	api := map[string]any{
		"invoke": js.FuncOf(r.jsInvoke),
		"listen": js.FuncOf(r.jsListen),
		"cancel": js.FuncOf(r.jsCancel),
		"methods": js.FuncOf(func(js.Value, []js.Value) any {
			ms := r.Methods()
			arr := make([]any, len(ms))
			for i, m := range ms {
				arr[i] = m
			}
			return arr
		}),
	}
	js.Global().Set(globalName, js.ValueOf(api))
}

// jsInvoke is the deterministic entry point. It validates arguments,
// performs the single CopyBytesToGo, and returns a Promise whose
// executor runs the handler on a goroutine so long-running handlers
// never block the JS event loop re-entrantly.
func (r *Router) jsInvoke(_ js.Value, args []js.Value) any {
	if len(args) != 2 {
		return rejectedPromise(&Error{Code: CodeInvalidArgument, Message: "invoke(method, payload) requires exactly 2 arguments"})
	}
	if args[0].Type() != js.TypeString {
		return rejectedPromise(&Error{Code: CodeInvalidArgument, Message: "method must be a string"})
	}
	method := args[0].String()

	payload := args[1]
	if payload.Type() != js.TypeObject || payload.Get("byteLength").IsUndefined() {
		return rejectedPromise(&Error{Code: CodeInvalidArgument, Method: method, Message: "payload must be a Uint8Array"})
	}

	// Single bulk copy JS -> Go into a pooled buffer.
	n := payload.Get("byteLength").Int()
	buf := reqPool.Get().([]byte)
	if cap(buf) < n {
		buf = make([]byte, n)
	} else {
		buf = buf[:n]
	}
	if copied := js.CopyBytesToGo(buf, payload); copied != n {
		reqPool.Put(buf[:0]) //nolint:staticcheck
		return rejectedPromise(&Error{Code: CodeInternal, Method: method, Message: "short copy across Wasm boundary"})
	}

	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, pArgs []js.Value) any {
		resolve, reject := pArgs[0], pArgs[1]
		go func() {
			defer executor.Release()

			out, rpcErr := r.Dispatch(context.Background(), method, buf)
			reqPool.Put(buf[:0]) //nolint:staticcheck

			if rpcErr != nil {
				if rpcErr.detail != "" {
					console.Call("error", "[wasm-rpc] "+method+" panicked:\n"+rpcErr.detail)
				}
				reject.Invoke(errToJS(rpcErr))
				return
			}

			// Single bulk copy Go -> JS.
			u8 := uint8ArrayCtor.New(len(out))
			js.CopyBytesToJS(u8, out)
			RecycleResponse(out)
			resolve.Invoke(u8)
		}()
		return nil
	})
	return promiseCtor.New(executor)
}

func errToJS(e *Error) js.Value {
	jsErr := errorCtor.New(e.Message)
	jsErr.Set("code", string(e.Code))
	jsErr.Set("method", e.Method)
	return jsErr
}

func rejectedPromise(e *Error) js.Value {
	return promiseCtor.Call("reject", errToJS(e))
}
