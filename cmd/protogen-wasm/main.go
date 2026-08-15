// Command protogen-wasm is a protoc/buf plugin that generates wasm-rpc
// bindings from service definitions:
//
//	lang=go      server interface + Register glue + native client
//	lang=ts      typed client over @bufbuild/protobuf schemas
//	lang=dart    typed client over protoc-gen-dart messages
//	lang=kotlin  typed client over protoc java/kotlin messages (JVM)
//	lang=swift   typed client over swift-protobuf messages
//
// Wire it via buf.gen.yaml v2:
//
//   - local: [go, run, ./cmd/protogen-wasm]
//     out: client/ts/gen
//     opt: lang=ts
package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/prdlk/wasm-rpc/internal/codegen"
)

func main() {
	var flags flag.FlagSet
	lang := flags.String("lang", "go", "target language: go|ts|dart|kotlin|swift")
	opts := protogen.Options{ParamFunc: flags.Set}
	opts.Run(func(p *protogen.Plugin) error {
		p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		gen, ok := codegen.Languages[*lang]
		if !ok {
			return fmt.Errorf("protogen-wasm: unknown lang %q", *lang)
		}
		for _, f := range p.Files {
			if !f.Generate || len(f.Services) == 0 {
				continue
			}
			if err := gen(p, f); err != nil {
				return fmt.Errorf("%s: %w", f.Desc.Path(), err)
			}
		}
		return nil
	})
}
