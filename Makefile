# wasm-rpc build automation (buf v2 workflow)
#
# Prereqs: go >= 1.24, buf >= 1.50, node >= 20; tinygo >= 0.34 optional.
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   (cd client/ts && npm install)   # provides protoc-gen-es + tsc
# Dart/Kotlin/Swift message stubs use buf remote plugins (BSR) — no
# local toolchains required for codegen. vtprotobuf runs via the go.mod
# tool directive (go tool protoc-gen-go-vtproto).

SHELL     := /bin/bash
BUILD     := build
WASM_OUT  := examples/e2e/app.wasm
WASMLD    := -trimpath -ldflags="-s -w"
GOROOT    := $(shell go env GOROOT)
# wasm_exec.js lives in misc/wasm (<= go1.23) or lib/wasm (>= go1.24)
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/misc/wasm/wasm_exec.js $(GOROOT)/lib/wasm/wasm_exec.js))

.PHONY: all lint generate breaking wasm wasm-tinygo wasip1 wasip1-tinygo ts \
	test test-e2e test-e2e-tinygo test-wasip1 test-wasip1-tinygo serve clean

all: lint generate wasm ts

## Schema ----------------------------------------------------------------

lint:
	buf lint

breaking:
	buf breaking --against '.git#branch=main'

# One command, twelve outputs: message stubs (go + vtprotobuf fast
# paths, ts, dart, kotlin+java, swift) plus wasm-rpc bindings for all
# five languages via cmd/protogen-wasm.
generate:
	buf generate

## Build -----------------------------------------------------------------

wasm:
	GOOS=js GOARCH=wasm go build $(WASMLD) -o $(WASM_OUT) ./examples/e2e/wasm
	cp "$(WASM_EXEC)" examples/e2e/wasm_exec.js

# TinyGo browser module: ~1.2 MB vs ~6.4 MB big Go. Pairs with tinygo's
# own wasm_exec.js. TinyGo cannot recover() on wasm, so handler panics
# abort the module — return server.Errorf instead of panicking in
# TinyGo-targeted services.
wasm-tinygo:
	mkdir -p $(BUILD)
	tinygo build -target=wasm -no-debug -o $(BUILD)/app.tinygo.wasm ./examples/e2e/wasm

# wasip1 reactor modules exposing the wasmrpc_* go:wasmexport ABI.
wasip1:
	mkdir -p $(BUILD)
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared $(WASMLD) \
		-o $(BUILD)/e2e.wasip1.wasm ./examples/e2e/wasip1

wasip1-tinygo:
	mkdir -p $(BUILD)
	tinygo build -target=wasip1 -buildmode=c-shared -no-debug \
		-o $(BUILD)/e2e.tinygo.wasip1.wasm ./examples/e2e/wasip1

ts:
	cd client/ts && npx tsc -p tsconfig.json

## Verify ----------------------------------------------------------------

test:
	go vet ./...
	go test ./...
	cd client/ts && npx tsc -p tsconfig.json --noEmit

# Full pipeline: generated TS client -> Wasm boundary -> Go router -> back.
test-e2e: wasm ts
	node examples/e2e/e2e.mjs

# Same suite against the TinyGo browser module (panic-recovery skipped).
test-e2e-tinygo: wasm-tinygo ts
	WASM_EXEC_JS=$$(tinygo env TINYGOROOT)/targets/wasm_exec.js \
	WASM_MODULE=$(BUILD)/app.tinygo.wasm SKIP_PANIC=1 \
	node examples/e2e/e2e.mjs

# wasip1 ABI end-to-end via an embedded wazero host (separate module).
test-wasip1: wasip1
	cd examples/e2e/wasip1host && go run . ../../../$(BUILD)/e2e.wasip1.wasm

test-wasip1-tinygo: wasip1-tinygo
	cd examples/e2e/wasip1host && go run . ../../../$(BUILD)/e2e.tinygo.wasip1.wasm --skip-panic

## Dev -------------------------------------------------------------------

serve:
	@echo "Use the Vite examples:"
	@echo "  make -C examples/web-echo dev"
	@echo "  make -C examples/web-streaming dev"

clean:
	rm -rf gen client/ts/gen client/ts/dist client/dart/lib/gen \
		client/kotlin/gen client/swift/Sources/WasmRpcGen/echo \
		client/swift/Sources/WasmRpcGen/ticker $(BUILD) \
		$(WASM_OUT) examples/e2e/wasm_exec.js
