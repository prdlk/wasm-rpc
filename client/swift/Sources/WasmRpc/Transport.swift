import Foundation

/// Moves opaque protobuf frames to a wasm-rpc router.
public protocol WasmRpcTransport: Sendable {
    func invoke(method: String, payload: Data) async throws -> Data
}

/// Server-streaming extension of ``WasmRpcTransport``.
public protocol WasmRpcStreamTransport: WasmRpcTransport {
    /// Yields one `Data` frame per server message; finishes on end,
    /// throws ``WasmRpcError`` on error. Terminating the sequence
    /// cancels the underlying stream.
    func listen(method: String, payload: Data) -> AsyncThrowingStream<Data, Error>
}

/// Standardized rejection payload produced by the Go bridge.
public struct WasmRpcError: Error, Sendable {
    public let code: String
    public let method: String?
    public let message: String

    public init(code: String, method: String? = nil, message: String) {
        self.code = code
        self.method = method
        self.message = message
    }
}
