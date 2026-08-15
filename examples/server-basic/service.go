// Package basic implements echo.v1.EchoService — the minimal unary
// wasm-rpc service. The same implementation serves the browser
// (syscall/js bridge), any WASI host (go:wasmexport ABI), and native
// tests (loopback), unchanged.
package basic

import (
	"context"
	"strings"
	"time"

	echov1 "github.com/hyperauth/wasm-rpc/gen/go/echo/v1"
	"github.com/hyperauth/wasm-rpc/server"
)

// Service implements echov1.EchoServiceServer.
type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Echo(_ context.Context, req *echov1.EchoRequest) (*echov1.EchoResponse, error) {
	if req.GetRepeat() > 1000 {
		return nil, server.Errorf(server.CodeInvalidArgument, "repeat must be <= 1000")
	}
	msg := req.GetMessage()
	if n := int(req.GetRepeat()); n > 1 {
		msg = strings.Repeat(msg, n)
	}
	return &echov1.EchoResponse{Message: msg, UnixMillis: time.Now().UnixMilli()}, nil
}

func (s *Service) Reverse(_ context.Context, req *echov1.ReverseRequest) (*echov1.ReverseResponse, error) {
	runes := []rune(req.GetMessage())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return &echov1.ReverseResponse{Message: string(runes), UnixMillis: time.Now().UnixMilli()}, nil
}
