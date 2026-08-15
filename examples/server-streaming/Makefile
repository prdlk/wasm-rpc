GOROOT    := $(shell cd ../.. && go env GOROOT)
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/misc/wasm/wasm_exec.js $(GOROOT)/lib/wasm/wasm_exec.js))
WASMLD    := -trimpath -ldflags="-s -w"

.PHONY: wasm wasm-tinygo wasip1 wasip1-tinygo test clean

wasm:
	mkdir -p bin
	cd ../.. && GOOS=js GOARCH=wasm go build $(WASMLD) \
		-o examples/server-streaming/bin/server.wasm ./examples/server-streaming/wasm
	cp "$(WASM_EXEC)" bin/wasm_exec.js

wasm-tinygo:
	mkdir -p bin
	cd ../.. && tinygo build -target=wasm -no-debug \
		-o examples/server-streaming/bin/server.wasm ./examples/server-streaming/wasm
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" bin/wasm_exec.js

wasip1:
	mkdir -p bin
	cd ../.. && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared $(WASMLD) \
		-o examples/server-streaming/bin/server.wasip1.wasm ./examples/server-streaming/wasip1

wasip1-tinygo:
	mkdir -p bin
	cd ../.. && tinygo build -target=wasip1 -buildmode=c-shared -no-debug \
		-o examples/server-streaming/bin/server.wasip1.wasm ./examples/server-streaming/wasip1

test:
	cd ../.. && go test ./examples/server-streaming/

clean:
	rm -rf bin
