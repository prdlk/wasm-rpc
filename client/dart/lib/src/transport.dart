import 'dart:typed_data';

/// Moves opaque protobuf frames to a wasm-rpc router.
abstract interface class WasmRpcTransport {
  Future<Uint8List> invoke(String method, Uint8List payload);
}

/// Standardized rejection payload produced by the Go bridge.
class WasmRpcException implements Exception {
  final String code;
  final String? method;
  final String message;

  const WasmRpcException(this.code, this.method, this.message);

  @override
  String toString() => 'WasmRpcException($code, $method): $message';
}

/// Server-streaming extension of [WasmRpcTransport].
abstract interface class WasmRpcStreamTransport implements WasmRpcTransport {
  /// Emits one [Uint8List] frame per server message; completes on end,
  /// errors with [WasmRpcException]. Cancelling the subscription cancels
  /// the underlying stream.
  Stream<Uint8List> listen(String method, Uint8List payload);
}
