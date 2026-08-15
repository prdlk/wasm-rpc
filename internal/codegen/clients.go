package codegen

import (
	"fmt"
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// ---------------------------------------------------------------------------
// TypeScript — clients over @bufbuild/protobuf (protoc-gen-es, target=ts)
// ---------------------------------------------------------------------------

// genTS emits {file}_wasm.pb.ts next to protoc-gen-es output under
// client/ts/gen. Runtime is imported relatively from client/ts/src.
func genTS(p *protogen.Plugin, f *protogen.File) error {
	prefix := filePrefix(f)
	dir := path.Dir(prefix)
	g := p.NewGeneratedFile(prefix+"_wasmrpc.pb.ts", "")

	depth := 0
	if dir != "." {
		depth = strings.Count(dir, "/") + 1
	}
	runtime := strings.Repeat("../", depth+1) + "src/client.js"

	g.P(header)
	g.P("// source: ", f.Desc.Path(), " (lang=ts)")
	g.P()
	runtimeImports := "unary, type WasmRpcTransport"
	if hasServerStream(f) {
		runtimeImports = "serverStream, unary, type WasmRpcTransport"
	}
	g.P(`import { `, runtimeImports, ` } from "`, runtime, `";`)

	// Collect schema imports grouped by defining file.
	type imp struct{ names []string }
	byFile := map[string]*imp{}
	var order []string
	addSchema := func(m *protogen.Message) {
		fp := m.Desc.ParentFile().Path()
		e, ok := byFile[fp]
		if !ok {
			e = &imp{}
			byFile[fp] = e
			order = append(order, fp)
		}
		name := string(m.Desc.Name()) + "Schema"
		for _, n := range e.names {
			if n == name {
				return
			}
		}
		e.names = append(e.names, name)
	}
	for _, s := range f.Services {
		for _, m := range s.Methods {
			if isStreaming(m) && !isServerStream(m) {
				continue
			}
			addSchema(m.Input)
			addSchema(m.Output)
		}
	}
	for _, fp := range order {
		g.P(`import { `, strings.Join(byFile[fp].names, ", "), ` } from "`, relImport(dir, fp, "_pb.js"), `";`)
	}
	g.P()

	for _, s := range f.Services {
		g.P("/** Typed wasm-rpc client for ", s.Desc.FullName(), ". */")
		g.P("export class ", s.GoName, "Client {")
		for _, m := range s.Methods {
			if isStreaming(m) && !isServerStream(m) {
				g.P("  // ", lowerCamel(m.GoName), ": client/bidi streaming is not supported by wasm-rpc; skipped.")
				continue
			}
			g.P("  readonly ", lowerCamel(m.GoName), ";")
		}
		g.P()
		g.P("  constructor(transport: WasmRpcTransport) {")
		for _, m := range s.Methods {
			if isStreaming(m) && !isServerStream(m) {
				continue
			}
			factory := "unary"
			if isServerStream(m) {
				factory = "serverStream"
			}
			g.P("    this.", lowerCamel(m.GoName), " = ", factory, "(")
			g.P("      transport,")
			g.P("      ", fmt.Sprintf("%q", fullMethod(s, m)), ",")
			g.P("      ", string(m.Input.Desc.Name()), "Schema,")
			g.P("      ", string(m.Output.Desc.Name()), "Schema,")
			g.P("    );")
		}
		g.P("  }")
		g.P("}")
		g.P()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Dart — clients over protoc-gen-dart messages
// ---------------------------------------------------------------------------

// genDart emits {file}_wasm.pb.dart next to protoc-gen-dart output under
// client/dart/lib/gen. Runtime comes from package:wasm_rpc_client.
func genDart(p *protogen.Plugin, f *protogen.File) error {
	prefix := filePrefix(f)
	dir := path.Dir(prefix)
	g := p.NewGeneratedFile(prefix+"_wasmrpc.pb.dart", "")

	g.P(header)
	g.P("// source: ", f.Desc.Path(), " (lang=dart)")
	g.P()
	g.P("import 'package:wasm_rpc_client/wasm_rpc_client.dart';")
	seen := map[string]bool{}
	for _, s := range f.Services {
		for _, m := range s.Methods {
			if isStreaming(m) && !isServerStream(m) {
				continue
			}
			for _, msg := range []*protogen.Message{m.Input, m.Output} {
				imp := relImport(dir, msg.Desc.ParentFile().Path(), ".pb.dart")
				imp = strings.TrimPrefix(imp, "./")
				if !seen[imp] {
					seen[imp] = true
					g.P("import '", imp, "';")
				}
			}
		}
	}
	g.P()

	for _, s := range f.Services {
		dartTransport := "WasmRpcTransport"
		if svcHasServerStream(s) {
			dartTransport = "WasmRpcStreamTransport"
		}
		g.P("/// Typed wasm-rpc client for `", s.Desc.FullName(), "`.")
		g.P("class ", s.GoName, "Client {")
		g.P("  final ", dartTransport, " _transport;")
		g.P()
		g.P("  const ", s.GoName, "Client(this._transport);")
		g.P()
		for _, m := range s.Methods {
			in := string(m.Input.Desc.Name())
			out := string(m.Output.Desc.Name())
			if isServerStream(m) {
				g.P("  Stream<", out, "> ", lowerCamel(m.GoName), "(", in, " request) => _transport")
				g.P("      .listen('", fullMethod(s, m), "', request.writeToBuffer())")
				g.P("      .map(", out, ".fromBuffer);")
				g.P()
				continue
			}
			if isStreaming(m) {
				g.P("  // ", lowerCamel(m.GoName), ": client/bidi streaming is not supported by wasm-rpc; skipped.")
				continue
			}
			g.P("  Future<", out, "> ", lowerCamel(m.GoName), "(", in, " request) async {")
			g.P("    final raw = await _transport.invoke(")
			g.P("      '", fullMethod(s, m), "',")
			g.P("      request.writeToBuffer(),")
			g.P("    );")
			g.P("    return ", out, ".fromBuffer(raw);")
			g.P("  }")
			g.P()
		}
		g.P("}")
		g.P()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Kotlin — JVM clients over protoc java/kotlin messages
// ---------------------------------------------------------------------------

// genKotlin emits {dir}/{File}WasmClient.kt next to kotlin_out output
// under client/kotlin/gen/kotlin. Message classes are the Java protos
// (outer-class-qualified when java_multiple_files is unset).
func genKotlin(p *protogen.Plugin, f *protogen.File) error {
	prefix := filePrefix(f)
	dir := path.Dir(prefix)
	base := upperCamel(path.Base(prefix))
	name := base + "WasmClient.kt"
	if dir != "." {
		name = dir + "/" + name
	}
	g := p.NewGeneratedFile(name, "")

	pkg := javaPackage(f)
	multi := f.Proto.GetOptions().GetJavaMultipleFiles()
	outer := javaOuterClass(f)
	msgRef := func(m *protogen.Message) string {
		if multi {
			return string(m.Desc.Name())
		}
		return outer + "." + string(m.Desc.Name())
	}

	g.P(header)
	g.P("// source: ", f.Desc.Path(), " (lang=kotlin)")
	g.P()
	if pkg != "" {
		g.P("package ", pkg)
		g.P()
	}
	if hasServerStream(f) {
		g.P("import dev.prdlk.wasmrpc.WasmRpcStreamTransport")
		g.P("import dev.prdlk.wasmrpc.WasmRpcTransport")
		g.P("import kotlinx.coroutines.flow.Flow")
		g.P("import kotlinx.coroutines.flow.map")
	} else {
		g.P("import dev.prdlk.wasmrpc.WasmRpcTransport")
	}
	g.P()

	for _, s := range f.Services {
		ktTransport := "WasmRpcTransport"
		if svcHasServerStream(s) {
			ktTransport = "WasmRpcStreamTransport"
		}
		g.P("/** Typed wasm-rpc client for `", s.Desc.FullName(), "`. */")
		g.P("class ", s.GoName, "Client(private val transport: ", ktTransport, ") {")
		for _, m := range s.Methods {
			in, out := msgRef(m.Input), msgRef(m.Output)
			if isServerStream(m) {
				g.P("    fun ", lowerCamel(m.GoName), "(request: ", in, "): Flow<", out, "> =")
				g.P("        transport")
				g.P("            .listen(\"", fullMethod(s, m), "\", request.toByteArray())")
				g.P("            .map(", out, "::parseFrom)")
				g.P()
				continue
			}
			if isStreaming(m) {
				g.P("    // ", lowerCamel(m.GoName), ": client/bidi streaming is not supported by wasm-rpc; skipped.")
				continue
			}
			g.P("    suspend fun ", lowerCamel(m.GoName), "(request: ", in, "): ", out, " =")
			g.P("        ", out, ".parseFrom(")
			g.P("            transport.invoke(\"", fullMethod(s, m), "\", request.toByteArray()),")
			g.P("        )")
			g.P()
		}
		g.P("}")
		g.P()
	}
	return nil
}

func javaPackage(f *protogen.File) string {
	if opts := f.Proto.GetOptions(); opts != nil && opts.GetJavaPackage() != "" {
		return opts.GetJavaPackage()
	}
	return string(f.Desc.Package())
}

// ---------------------------------------------------------------------------
// Swift — clients over swift-protobuf messages
// ---------------------------------------------------------------------------

// genSwift emits {file}.wasm.swift next to protoc-gen-swift output under
// client/swift/Sources/WasmRpcGen. Types follow swift-protobuf naming
// (package "echo.v1" -> Echo_V1_ prefix).
func genSwift(p *protogen.Plugin, f *protogen.File) error {
	g := p.NewGeneratedFile(filePrefix(f)+".wasmrpc.swift", "")

	g.P(header)
	g.P("// source: ", f.Desc.Path(), " (lang=swift)")
	g.P()
	g.P("import Foundation")
	g.P("import WasmRpc")
	g.P()

	for _, s := range f.Services {
		g.P("/// Typed wasm-rpc client for `", s.Desc.FullName(), "`.")
		swiftTransport := "WasmRpcTransport"
		if svcHasServerStream(s) {
			swiftTransport = "WasmRpcStreamTransport"
		}
		g.P("public struct ", s.GoName, "Client: Sendable {")
		g.P("    private let transport: any ", swiftTransport)
		g.P()
		g.P("    public init(transport: any ", swiftTransport, ") {")
		g.P("        self.transport = transport")
		g.P("    }")
		g.P()
		for _, m := range s.Methods {
			in := swiftTypeName(m.Input.Desc)
			out := swiftTypeName(m.Output.Desc)
			if isServerStream(m) {
				g.P("    public func ", lowerCamel(m.GoName), "(_ request: ", in, ") -> AsyncThrowingStream<", out, ", Error> {")
				g.P("        AsyncThrowingStream { continuation in")
				g.P("            let task = Task {")
				g.P("                do {")
				g.P("                    let payload = try request.serializedData()")
				g.P("                    for try await frame in transport.listen(")
				g.P("                        method: \"", fullMethod(s, m), "\",")
				g.P("                        payload: payload")
				g.P("                    ) {")
				g.P("                        continuation.yield(try ", out, "(serializedBytes: frame))")
				g.P("                    }")
				g.P("                    continuation.finish()")
				g.P("                } catch {")
				g.P("                    continuation.finish(throwing: error)")
				g.P("                }")
				g.P("            }")
				g.P("            continuation.onTermination = { _ in task.cancel() }")
				g.P("        }")
				g.P("    }")
				g.P()
				continue
			}
			if isStreaming(m) {
				g.P("    // ", lowerCamel(m.GoName), ": client/bidi streaming is not supported by wasm-rpc; skipped.")
				continue
			}
			g.P("    public func ", lowerCamel(m.GoName), "(_ request: ", in, ") async throws -> ", out, " {")
			g.P("        let payload = try request.serializedData()")
			g.P("        let raw = try await transport.invoke(")
			g.P("            method: \"", fullMethod(s, m), "\",")
			g.P("            payload: payload")
			g.P("        )")
			g.P("        return try ", out, "(serializedBytes: raw)")
			g.P("    }")
			g.P()
		}
		g.P("}")
		g.P()
	}
	return nil
}
