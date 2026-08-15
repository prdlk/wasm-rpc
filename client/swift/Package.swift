// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "wasm-rpc-client",
    platforms: [.iOS(.v15), .macOS(.v12)],
    products: [
        .library(name: "WasmRpc", targets: ["WasmRpc"]),
        .library(name: "WasmRpcGen", targets: ["WasmRpcGen"]),
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.28.0"),
    ],
    targets: [
        .target(name: "WasmRpc"),
        .target(
            name: "WasmRpcGen",
            dependencies: [
                "WasmRpc",
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ]
        ),
    ]
)
