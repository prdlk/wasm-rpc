GOROOT    := $(shell cd ../.. && go env GOROOT)
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/misc/wasm/wasm_exec.js $(GOROOT)/lib/wasm/wasm_exec.js))
WASMLD    := -trimpath -ldflags="-s -w"

.PHONY: wasm wasm-tinygo wasip1 wasip1-tinygo test clean

wasm:
	mkdir -p bin
	cd ../.. && GOOS=js GOARCH=wasm go build $(WASMLD) \
		-o examples/server-basic/bin/server.wasm ./examples/server-basic/wasm
	cp "$(WASM_EXEC)" bin/wasm_exec.js

wasm-tinygo:
	mkdir -p bin
	cd ../.. && tinygo build -target=wasm -no-debug \
		-o examples/server-basic/bin/server.wasm ./examples/server-basic/wasm
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" bin/wasm_exec.js

wasip1:
	mkdir -p bin
	cd ../.. && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared $(WASMLD) \
		-o examples/server-basic/bin/server.wasip1.wasm ./examples/server-basic/wasip1

wasip1-tinygo:
	mkdir -p bin
	cd ../.. && tinygo build -target=wasip1 -buildmode=c-shared -no-debug \
		-o examples/server-basic/bin/server.wasip1.wasm ./examples/server-basic/wasip1

test:
	cd ../.. && go test ./examples/server-basic/

clean:
	rm -rf bin
