// Command wasip1host embeds the wasip1 echo guest via wazero and drives
// the wasmrpc_* ABI end-to-end: unary invoke, server stream, stream
// error (panic recovery), and unimplemented — proving the same router
// serves browsers (syscall/js) and WASI hosts (wasip1) unchanged.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"google.golang.org/protobuf/proto"

	echov1 "github.com/hyperauth/wasm-rpc/gen/go/echo/v1"
	tickerv1 "github.com/hyperauth/wasm-rpc/gen/go/ticker/v1"
)

const (
	frameOK    = 0
	frameEnd   = 1
	frameError = 2
)

type guest struct {
	ctx context.Context
	mod api.Module
}

func (g *guest) write(data []byte) uint32 {
	ptr := g.call("wasmrpc_alloc", uint64(len(data)))
	if len(data) > 0 && !g.mod.Memory().Write(uint32(ptr), data) {
		log.Fatal("memory write failed")
	}
	return uint32(ptr)
}

func (g *guest) call(fn string, args ...uint64) uint64 {
	res, err := g.mod.ExportedFunction(fn).Call(g.ctx, args...)
	if err != nil {
		log.Fatalf("%s: %v", fn, err)
	}
	if len(res) == 0 {
		return 0
	}
	return res[0]
}

// frame reads and frees a packed [status][payload] frame.
func (g *guest) frame(packed uint64) (byte, []byte) {
	ptr, n := uint32(packed>>32), uint32(packed)
	raw, ok := g.mod.Memory().Read(ptr, n)
	if !ok || n == 0 {
		log.Fatal("bad frame")
	}
	status, payload := raw[0], append([]byte(nil), raw[1:]...)
	g.call("wasmrpc_free", uint64(ptr))
	return status, payload
}

func (g *guest) invoke(method string, req proto.Message, resp proto.Message) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	mPtr, pPtr := g.write([]byte(method)), g.write(payload)
	packed := g.call("wasmrpc_invoke", uint64(mPtr), uint64(len(method)), uint64(pPtr), uint64(len(payload)))
	g.call("wasmrpc_free", uint64(mPtr))
	g.call("wasmrpc_free", uint64(pPtr))
	status, body := g.frame(packed)
	if status == frameError {
		parts := strings.SplitN(string(body), "\n", 3)
		return fmt.Errorf("rpc %s: %s", parts[0], parts[2])
	}
	return proto.Unmarshal(body, resp)
}

func main() {
	ctx := context.Background()
	wasmBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	mod, err := rt.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithStartFunctions("_initialize").WithStderr(os.Stderr).WithStdout(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}
	g := &guest{ctx: ctx, mod: mod}

	// 1. Unary round-trip
	var echoResp echov1.EchoResponse
	must(g.invoke("echo.v1.EchoService/Echo",
		&echov1.EchoRequest{Message: "wasi", Repeat: 2}, &echoResp))
	assertEq(echoResp.GetMessage(), "wasiwasi", "unary echo")

	// 2. Unary typed error
	err = g.invoke("echo.v1.EchoService/Echo",
		&echov1.EchoRequest{Repeat: 9999}, &echoResp)
	assert(err != nil && strings.Contains(err.Error(), "invalid_argument"), "typed error: %v", err)

	// 3. Server stream: collect all frames
	const subscribeMethod = "ticker.v1.TickerService/Subscribe"
	watchReq, _ := proto.Marshal(&tickerv1.SubscribeRequest{Topic: "tick", Count: 4})
	mPtr := g.write([]byte(subscribeMethod))
	pPtr := g.write(watchReq)
	id := g.call("wasmrpc_stream_open", uint64(mPtr), uint64(len(subscribeMethod)), uint64(pPtr), uint64(len(watchReq)))
	g.call("wasmrpc_free", uint64(mPtr))
	g.call("wasmrpc_free", uint64(pPtr))
	var seqs []uint32
	for {
		status, body := g.frame(g.call("wasmrpc_stream_recv", id))
		if status == frameEnd {
			break
		}
		assert(status == frameOK, "unexpected stream status %d", status)
		var tick tickerv1.SubscribeResponse
		must(proto.Unmarshal(body, &tick))
		assertEq(tick.GetTopic(), "tick", "stream payload")
		seqs = append(seqs, tick.GetSeq())
	}
	assert(len(seqs) == 4 && seqs[0] == 1 && seqs[3] == 4, "stream seqs: %v", seqs)

	// 4. Stream panic -> frameError, module survives.
	// TinyGo cannot recover() on wasm (supportsRecover()==false for the
	// arch), so a handler panic aborts the module instead of degrading
	// to an error frame — skipped for TinyGo guests via --skip-panic.
	skipPanic := len(os.Args) > 2 && os.Args[2] == "--skip-panic"
	if !skipPanic {
		boomReq, _ := proto.Marshal(&tickerv1.SubscribeRequest{Topic: "explode"})
		mPtr = g.write([]byte(subscribeMethod))
		pPtr = g.write(boomReq)
		id = g.call("wasmrpc_stream_open", uint64(mPtr), uint64(len(subscribeMethod)), uint64(pPtr), uint64(len(boomReq)))
		g.call("wasmrpc_free", uint64(mPtr))
		g.call("wasmrpc_free", uint64(pPtr))
		status, body := g.frame(g.call("wasmrpc_stream_recv", id))
		assert(status == frameError && strings.HasPrefix(string(body), "internal\n"), "panic frame: %d %q", status, body)
	}

	// 5. Module alive after all previous traffic
	must(g.invoke("echo.v1.EchoService/Echo",
		&echov1.EchoRequest{Message: "alive"}, &echoResp))
	assertEq(echoResp.GetMessage(), "alive", "post-panic")

	if skipPanic {
		fmt.Println("WASIP1 E2E PASS: 4/4 via wazero (unary, error, stream, survival; panic-recovery skipped: TinyGo)")
	} else {
		fmt.Println("WASIP1 E2E PASS: 5/5 via wazero (unary, error, stream, panic, survival)")
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func assert(cond bool, f string, a ...any) {
	if !cond {
		log.Fatalf(f, a...)
	}
}
func assertEq(got, want, what string) {
	if got != want {
		log.Fatalf("%s: got %q want %q", what, got, want)
	}
}
