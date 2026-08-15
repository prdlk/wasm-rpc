// Transport for Dart web / Flutter web: calls the global exported by the
// Go module (`Router.Mount("wasmRPC")`) via dart:js_interop.
import 'dart:js_interop';
import 'dart:async';
import 'dart:typed_data';

import 'dart:js_interop_unsafe';

import 'transport.dart';

@JS('wasmRPC.invoke')
external JSPromise<JSUint8Array> _invoke(String method, JSUint8Array payload);

@JS('wasmRPC.listen')
external int _listen(String method, JSUint8Array payload, JSObject handlers);

@JS('wasmRPC.cancel')
external void _cancelStream(int id);

/// Transport bound to `globalThis.wasmRPC`. The Go Wasm module must be
/// instantiated (wasm_exec.js + WebAssembly.instantiateStreaming) before
/// the first call.
final class JsGlobalTransport implements WasmRpcStreamTransport {
  const JsGlobalTransport();

  @override
  Future<Uint8List> invoke(String method, Uint8List payload) async {
    try {
      final raw = await _invoke(method, payload.toJS).toDart;
      return raw.toDart;
    } catch (e) {
      final o = e as JSObject?;
      throw WasmRpcException(
        _str(o, 'code') ?? 'unknown',
        _str(o, 'method'),
        _str(o, 'message') ?? e.toString(),
      );
    }
  }

  @override
  Stream<Uint8List> listen(String method, Uint8List payload) {
    late final StreamController<Uint8List> controller;
    var id = 0;
    controller = StreamController<Uint8List>(
      onListen: () {
        final handlers = JSObject()
          ..setProperty(
            'onMessage'.toJS,
            ((JSUint8Array m) => controller.add(m.toDart)).toJS,
          )
          ..setProperty(
            'onError'.toJS,
            ((JSObject e) {
              controller.addError(WasmRpcException(
                _str(e, 'code') ?? 'unknown',
                _str(e, 'method') ?? method,
                _str(e, 'message') ?? 'stream error',
              ));
              controller.close();
            }).toJS,
          )
          ..setProperty('onEnd'.toJS, controller.close.toJS);
        id = _listen(method, payload.toJS, handlers);
      },
      onCancel: () => _cancelStream(id),
    );
    return controller.stream;
  }

  static String? _str(JSObject? o, String key) {
    if (o == null) return null;
    final v = o.getProperty(key.toJS);
    return v.isA<JSString>() ? (v as JSString).toDart : null;
  }
}
