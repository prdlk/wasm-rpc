module github.com/hyperauth/wasm-rpc/examples/e2e/wasip1host

go 1.24

require (
	github.com/hyperauth/wasm-rpc v0.0.0
	github.com/tetratelabs/wazero v1.9.0
	google.golang.org/protobuf v1.36.5
)

require github.com/planetscale/vtprotobuf v0.6.0 // indirect

replace github.com/hyperauth/wasm-rpc => ../../..
