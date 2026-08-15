//go:build wasip1

package server

// wasip1 bridge
// -------------
// Transport-agnostic ABI over exported functions + linear memory, for
// hosts like wazero/wasmtime and for TinyGo (-target=wasip1) builds.
// Build the guest as a reactor:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o app.wasm ./...
//	tinygo build -target=wasip1 -buildmode=c-shared -no-debug -o app.wasm ./...
//
// Register the router from an init() (reactors run inits, not main) via
// server.Attach(r).
//
// ABI (all pointers are guest-linear-memory offsets):
//
//	wasmrpc_alloc(n u32) -> ptr u32          host allocates an input buffer
//	wasmrpc_free(ptr u32)                    release any guest-returned buffer
//	wasmrpc_invoke(mPtr,mLen,pPtr,pLen) u64  packed ptr<<32|len result FRAME
//	wasmrpc_stream_open(mPtr,mLen,pPtr,pLen) u32   nonzero stream id
//	wasmrpc_stream_recv(id u32) u64          blocks; packed ptr<<32|len FRAME
//	wasmrpc_stream_cancel(id u32)
//
// FRAME layout: [status u8][payload]. Status: 0=ok message, 1=end (empty
// payload), 2=error (payload = code "\n" method "\n" message). Hosts
// must wasmrpc_free every returned frame and every alloc'd input after
// the call that consumes it.

import (
	"context"
	"sync"
	"unsafe"
)

const (
	frameOK    byte = 0
	frameEnd   byte = 1
	frameError byte = 2
)

var wasiRouter *Router

// Attach installs the router served by the wasip1 exports. Call from an
// init() in the guest's main package.
func Attach(r *Router) { wasiRouter = r }

// keepAlive pins buffers whose pointers have been handed to the host so
// the GC cannot collect or move them (Go's wasm GC is non-moving today,
// but pinning via reachability is the documented-safe contract).
var keepAlive = struct {
	sync.Mutex
	m map[uintptr][]byte
}{m: make(map[uintptr][]byte)}

func pin(b []byte) uint32 {
	if len(b) == 0 {
		return 0
	}
	p := uintptr(unsafe.Pointer(&b[0]))
	keepAlive.Lock()
	keepAlive.m[p] = b
	keepAlive.Unlock()
	return uint32(p)
}

//go:wasmexport wasmrpc_alloc
func wasmrpcAlloc(n uint32) uint32 {
	return pin(make([]byte, n))
}

//go:wasmexport wasmrpc_free
func wasmrpcFree(ptr uint32) {
	keepAlive.Lock()
	delete(keepAlive.m, uintptr(ptr))
	keepAlive.Unlock()
}

func pinned(ptr, n uint32) []byte {
	if n == 0 {
		return nil
	}
	keepAlive.Lock()
	b, ok := keepAlive.m[uintptr(ptr)]
	keepAlive.Unlock()
	if !ok || uint32(len(b)) < n {
		return nil
	}
	return b[:n]
}

func packFrame(status byte, payload []byte) uint64 {
	frame := make([]byte, 1+len(payload))
	frame[0] = status
	copy(frame[1:], payload)
	ptr := pin(frame)
	return uint64(ptr)<<32 | uint64(len(frame))
}

func errFrame(e *Error) uint64 {
	return packFrame(frameError, []byte(string(e.Code)+"\n"+e.Method+"\n"+e.Message))
}

//go:wasmexport wasmrpc_invoke
func wasmrpcInvoke(mPtr, mLen, pPtr, pLen uint32) uint64 {
	if wasiRouter == nil {
		return errFrame(&Error{Code: CodeInternal, Message: "no router attached: call server.Attach in init()"})
	}
	m := pinned(mPtr, mLen)
	if m == nil {
		return errFrame(&Error{Code: CodeInvalidArgument, Message: "method buffer not allocated via wasmrpc_alloc"})
	}
	payload := pinned(pPtr, pLen)
	out, rpcErr := wasiRouter.Dispatch(context.Background(), string(m), payload)
	if rpcErr != nil {
		return errFrame(rpcErr)
	}
	packed := packFrame(frameOK, out)
	RecycleResponse(out)
	return packed
}

// Streams: pull-based. open starts the handler on a goroutine feeding a
// bounded channel (natural backpressure); recv blocks until the next
// frame — the Go scheduler keeps handler goroutines running while an
// export call blocks.

type wasiStream struct {
	frames chan uint64
	cancel context.CancelFunc
}

var wasiStreams = struct {
	sync.Mutex
	next uint32
	m    map[uint32]*wasiStream
}{m: make(map[uint32]*wasiStream)}

//go:wasmexport wasmrpc_stream_open
func wasmrpcStreamOpen(mPtr, mLen, pPtr, pLen uint32) uint32 {
	method := ""
	if m := pinned(mPtr, mLen); m != nil {
		method = string(m)
	}
	// Copy the payload: the host may free its input right after open,
	// but the handler runs asynchronously.
	var payload []byte
	if p := pinned(pPtr, pLen); p != nil {
		payload = append([]byte(nil), p...)
	}

	ctx, cancel := context.WithCancel(context.Background())
	st := &wasiStream{frames: make(chan uint64, 16), cancel: cancel}
	wasiStreams.Lock()
	wasiStreams.next++
	id := wasiStreams.next
	wasiStreams.m[id] = st
	wasiStreams.Unlock()

	go func() {
		defer cancel()
		var terminal uint64
		if wasiRouter == nil {
			terminal = errFrame(&Error{Code: CodeInternal, Message: "no router attached"})
		} else {
			send := func(frame []byte) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				select {
				case st.frames <- packFrame(frameOK, frame):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if rpcErr := wasiRouter.DispatchStream(ctx, method, payload, send); rpcErr != nil {
				terminal = errFrame(rpcErr)
			} else {
				terminal = packFrame(frameEnd, nil)
			}
		}
		select {
		case st.frames <- terminal:
		case <-ctx.Done():
		}
		close(st.frames)
	}()
	return id
}

//go:wasmexport wasmrpc_stream_recv
func wasmrpcStreamRecv(id uint32) uint64 {
	wasiStreams.Lock()
	st := wasiStreams.m[id]
	wasiStreams.Unlock()
	if st == nil {
		return errFrame(&Error{Code: CodeInvalidArgument, Message: "unknown stream id"})
	}
	packed, ok := <-st.frames
	if !ok {
		packed = packFrame(frameEnd, nil)
	}
	// Terminal frame -> drop the registry entry.
	if status := frameStatus(packed); status != frameOK {
		wasiStreams.Lock()
		delete(wasiStreams.m, id)
		wasiStreams.Unlock()
	}
	return packed
}

//go:wasmexport wasmrpc_stream_cancel
func wasmrpcStreamCancel(id uint32) {
	wasiStreams.Lock()
	st := wasiStreams.m[id]
	delete(wasiStreams.m, id)
	wasiStreams.Unlock()
	if st != nil {
		st.cancel()
	}
}

func frameStatus(packed uint64) byte {
	ptr := uint32(packed >> 32)
	if b := pinned(ptr, 1); b != nil {
		return b[0]
	}
	return frameEnd
}
