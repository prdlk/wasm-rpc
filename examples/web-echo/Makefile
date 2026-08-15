# web-echo build orchestration. `make dev` is the one-command path.

GOROOT    := $(shell go env GOROOT)
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/misc/wasm/wasm_exec.js $(GOROOT)/lib/wasm/wasm_exec.js))

.PHONY: dev build preview server server-tinygo client install clean

# Build the Go RPC server into public/ so Vite serves it at /server.wasm.
server:
	cd ../.. && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" \
		-o examples/web-echo/public/server.wasm ./examples/web-echo/server
	cp "$(WASM_EXEC)" public/wasm_exec.js

# ~1.2 MB module instead of ~6.4 MB; needs tinygo's wasm_exec.js.
# Services must not panic under TinyGo (recover() unsupported on wasm).
server-tinygo:
	cd ../.. && tinygo build -target=wasm -no-debug \
		-o examples/web-echo/public/server.wasm ./examples/web-echo/server
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" public/wasm_exec.js

# The file: dependency needs the client package's dist/ built first.
client:
	cd ../../client/ts && npm install && npm run build

install: client
	npm install

dev: server install
	npm run dev

build: server install
	npm run build

preview: build
	npm run preview

clean:
	rm -rf node_modules dist public/server.wasm public/wasm_exec.js
