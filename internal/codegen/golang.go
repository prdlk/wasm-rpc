package codegen

import (
	"strconv"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	serverPkg  = protogen.GoImportPath("github.com/prdlk/wasm-rpc/server")
	clientPkg  = protogen.GoImportPath("github.com/prdlk/wasm-rpc/client/go")
	contextPkg = protogen.GoImportPath("context")
	protoPkg   = protogen.GoImportPath("google.golang.org/protobuf/proto")
)

func hasServerStream(f *protogen.File) bool {
	for _, s := range f.Services {
		for _, m := range s.Methods {
			if isServerStream(m) {
				return true
			}
		}
	}
	return false
}

// genGo emits {file}_wasm.pb.go into the same package as the messages:
// method constants, a Server interface, RegisterXServer(router, impl),
// and a native XClient over the client/go Transport (loopback-testable).
func genGo(p *protogen.Plugin, f *protogen.File) error {
	g := p.NewGeneratedFile(f.GeneratedFilenamePrefix+"_wasmrpc.pb.go", f.GoImportPath)
	g.P(header)
	g.P("// source: ", f.Desc.Path(), " (lang=go)")
	g.P()
	g.P("package ", f.GoPackageName)
	g.P()

	for _, s := range f.Services {
		g.P("// Fully-qualified wire identifiers for ", s.GoName, " methods.")
		g.P("const (")
		for _, m := range s.Methods {
			g.P(s.GoName, m.GoName, "Method = ", strconv.Quote(fullMethod(s, m)))
		}
		g.P(")")
		g.P()

		// Server interface.
		g.P("// ", s.GoName, "Server is the handler contract implemented by services.")
		g.P("type ", s.GoName, "Server interface {")
		for _, m := range s.Methods {
			switch {
			case isServerStream(m):
				g.P(m.GoName, "(", contextPkg.Ident("Context"), ", *", m.Input.GoIdent, ", *", serverPkg.Ident("ServerStream"), "[*", m.Output.GoIdent, "]) error")
			case isStreaming(m):
				g.P("// ", m.GoName, ": client/bidi streaming is not supported by wasm-rpc; skipped.")
			default:
				g.P(m.GoName, "(", contextPkg.Ident("Context"), ", *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error)")
			}
		}
		g.P("}")
		g.P()

		// Router registration.
		g.P("// Register", s.GoName, "Server wires every unary method to the router.")
		g.P("func Register", s.GoName, "Server(r *", serverPkg.Ident("Router"), ", s ", s.GoName, "Server) {")
		for _, m := range s.Methods {
			switch {
			case isServerStream(m):
				g.P(serverPkg.Ident("RegisterServerStream"), "(r, ", s.GoName, m.GoName, "Method, s.", m.GoName, ")")
			case isStreaming(m):
			default:
				g.P(serverPkg.Ident("Register"), "(r, ", s.GoName, m.GoName, "Method, s.", m.GoName, ")")
			}
		}
		g.P("}")
		g.P()

		// Native client.
		g.P("// ", s.GoName, "Client is a typed client over any wasm-rpc Transport")
		g.P("// (in-process loopback for tests, or a host-embedded module).")
		transportIdent := clientPkg.Ident("Transport")
		if svcHasServerStream(s) {
			transportIdent = clientPkg.Ident("StreamTransport")
		}
		g.P("type ", s.GoName, "Client struct {")
		g.P("t ", transportIdent)
		g.P("}")
		g.P()
		g.P("func New", s.GoName, "Client(t ", transportIdent, ") *", s.GoName, "Client {")
		g.P("return &", s.GoName, "Client{t: t}")
		g.P("}")
		g.P()
		for _, m := range s.Methods {
			if isServerStream(m) {
				g.P("// ", m.GoName, " server-streams; fn is invoked once per message and a")
				g.P("// non-nil return cancels the stream.")
				g.P("func (c *", s.GoName, "Client) ", m.GoName, "(ctx ", contextPkg.Ident("Context"), ", req *", m.Input.GoIdent, ", fn func(*", m.Output.GoIdent, ") error) error {")
				g.P("payload, err := ", protoPkg.Ident("Marshal"), "(req)")
				g.P("if err != nil { return err }")
				g.P("return c.t.Listen(ctx, ", s.GoName, m.GoName, "Method, payload, func(raw []byte) error {")
				g.P("msg := new(", m.Output.GoIdent, ")")
				g.P("if err := ", protoPkg.Ident("Unmarshal"), "(raw, msg); err != nil { return err }")
				g.P("return fn(msg)")
				g.P("})")
				g.P("}")
				g.P()
				continue
			}
			if isStreaming(m) {
				continue
			}
			g.P("func (c *", s.GoName, "Client) ", m.GoName, "(ctx ", contextPkg.Ident("Context"), ", req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error) {")
			g.P("payload, err := ", protoPkg.Ident("Marshal"), "(req)")
			g.P("if err != nil { return nil, err }")
			g.P("raw, err := c.t.Invoke(ctx, ", s.GoName, m.GoName, "Method, payload)")
			g.P("if err != nil { return nil, err }")
			g.P("resp := new(", m.Output.GoIdent, ")")
			g.P("if err := ", protoPkg.Ident("Unmarshal"), "(raw, resp); err != nil { return nil, err }")
			g.P("return resp, nil")
			g.P("}")
			g.P()
		}
	}
	return nil
}
