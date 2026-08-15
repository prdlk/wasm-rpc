package dev.hyperauth.wasmrpc

/**
 * Reference sketch for Android: bridge into a WebView hosting the Go
 * Wasm module. Wire evaluateJavascript against `wasmRPC.invoke`,
 * shuttling payloads as base64, and resume the continuation from the
 * Promise callbacks. Host apps own the WebView lifecycle, so this stays
 * an interface contract here rather than a concrete dependency.
 */
