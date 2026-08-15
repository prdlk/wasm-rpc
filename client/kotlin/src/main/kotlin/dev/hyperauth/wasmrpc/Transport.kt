package dev.hyperauth.wasmrpc

import kotlinx.coroutines.flow.Flow

/** Moves opaque protobuf frames to a wasm-rpc router. */
interface WasmRpcTransport {
    suspend fun invoke(method: String, payload: ByteArray): ByteArray
}

/** Server-streaming extension of [WasmRpcTransport]. */
interface WasmRpcStreamTransport : WasmRpcTransport {
    /**
     * Emits one frame per server message; completes on end, throws
     * [WasmRpcException] on error. Collector cancellation cancels the
     * underlying stream.
     */
    fun listen(method: String, payload: ByteArray): Flow<ByteArray>
}

/** Standardized rejection payload produced by the Go bridge. */
class WasmRpcException(
    val code: String,
    val method: String?,
    message: String,
) : Exception("$code: $message")
