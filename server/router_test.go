package server_test

import (
	"context"
	"strings"
	"testing"

	echov1 "github.com/hyperauth/wasm-rpc/gen/go/echo/v1"
	"github.com/hyperauth/wasm-rpc/server"
	"google.golang.org/protobuf/proto"
)

func newTestRouter(t *testing.T) *server.Router {
	t.Helper()
	r := server.NewRouter()
	server.Register(r, "echo.v1.EchoService/Echo",
		func(_ context.Context, req *echov1.EchoRequest) (*echov1.EchoResponse, error) {
			return &echov1.EchoResponse{Message: req.GetMessage()}, nil
		})
	server.Register(r, "test/Panic",
		func(_ context.Context, _ *echov1.EchoRequest) (*echov1.EchoResponse, error) {
			panic("handler exploded")
		})
	server.Register(r, "test/TypedErr",
		func(_ context.Context, _ *echov1.EchoRequest) (*echov1.EchoResponse, error) {
			return nil, server.Errorf(server.CodeInvalidArgument, "bad input")
		})
	return r
}

func TestRoundTrip(t *testing.T) {
	r := newTestRouter(t)
	in, _ := proto.Marshal(&echov1.EchoRequest{Message: "hyperauth"})

	out, rpcErr := r.Dispatch(context.Background(), "echo.v1.EchoService/Echo", in)
	if rpcErr != nil {
		t.Fatalf("unexpected error: %v", rpcErr)
	}
	var resp echov1.EchoResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.GetMessage() != "hyperauth" {
		t.Fatalf("got %q", resp.GetMessage())
	}
}

func TestPanicRecovery(t *testing.T) {
	r := newTestRouter(t)
	_, rpcErr := r.Dispatch(context.Background(), "test/Panic", nil)
	if rpcErr == nil || rpcErr.Code != server.CodeInternal {
		t.Fatalf("expected internal error, got %+v", rpcErr)
	}
	if !strings.Contains(rpcErr.Message, "handler exploded") {
		t.Fatalf("panic value lost: %q", rpcErr.Message)
	}
}

func TestTypedErrorPassthrough(t *testing.T) {
	r := newTestRouter(t)
	_, rpcErr := r.Dispatch(context.Background(), "test/TypedErr", nil)
	if rpcErr == nil || rpcErr.Code != server.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %+v", rpcErr)
	}
}

func TestUnimplemented(t *testing.T) {
	r := newTestRouter(t)
	_, rpcErr := r.Dispatch(context.Background(), "nope/Nope", nil)
	if rpcErr == nil || rpcErr.Code != server.CodeUnimplemented {
		t.Fatalf("expected unimplemented, got %+v", rpcErr)
	}
}

func TestMalformedPayload(t *testing.T) {
	r := newTestRouter(t)
	_, rpcErr := r.Dispatch(context.Background(), "echo.v1.EchoService/Echo", []byte{0xff, 0xff, 0xff})
	if rpcErr == nil || rpcErr.Code != server.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument for malformed proto, got %+v", rpcErr)
	}
}
