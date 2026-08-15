/**
 * wasm-rpc-client — generic transport over the Go Wasm host.
 *
 * The Go side exports (via wasmrpc.Router.Mount):
 *   globalThis.wasmRPC.invoke(method: string, payload: Uint8Array): Promise<Uint8Array>
 *
 * Serialization is fully abstracted: application code sees only typed
 * async methods (see unary() and the generated service clients).
 */
import {
  create,
  fromBinary,
  toBinary,
  type DescMessage,
  type MessageInitShape,
  type MessageShape,
} from "@bufbuild/protobuf";

/** Standardized rejection payload produced by the Go bridge. */
export interface WasmRpcErrorShape {
  code: "invalid_argument" | "unimplemented" | "internal" | "unknown" | string;
  method?: string;
  message: string;
}

export class WasmRpcError extends Error implements WasmRpcErrorShape {
  readonly code: string;
  readonly method?: string;
  constructor(shape: WasmRpcErrorShape) {
    super(shape.message);
    this.name = "WasmRpcError";
    this.code = shape.code;
    this.method = shape.method;
  }
}

interface WasmRpcHost {
  invoke(method: string, payload: Uint8Array): Promise<Uint8Array>;
  methods(): string[];
  // Present when the module was built with streaming support:
  listen?(method: string, payload: Uint8Array, handlers: unknown): number;
  cancel?(id: number): void;
}

declare global {
  // Installed by Router.Mount("wasmRPC") inside the Go module.
  // eslint-disable-next-line no-var
  var wasmRPC: WasmRpcHost | undefined;
  // Readiness resolver awaited by loadWasmRpc, invoked by cmd/wasm.
  // eslint-disable-next-line no-var
  var __wasmRPCReady: (() => void) | undefined;
}

/**
 * Instantiates the Go Wasm module and resolves once the router is
 * mounted. Requires wasm_exec.js (the Go JS glue) to be loaded first —
 * it defines the global `Go` class.
 */
export async function loadWasmRpc(wasmUrl: string): Promise<WasmRpcTransport> {
  const GoCtor = (globalThis as Record<string, unknown>)["Go"] as
    | (new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> })
    | undefined;
  if (!GoCtor) {
    throw new Error("wasm_exec.js not loaded: global Go class is missing");
  }

  const ready = new Promise<void>((resolve) => {
    globalThis.__wasmRPCReady = resolve;
  });

  const go = new GoCtor();
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch(wasmUrl),
    go.importObject,
  );
  void go.run(instance); // runs until the module exits; do not await
  await ready;

  if (!globalThis.wasmRPC) {
    throw new Error("Go module ran but did not mount globalThis.wasmRPC");
  }
  return new WasmRpcTransport(globalThis.wasmRPC);
}

export class WasmRpcTransport {
  constructor(private readonly host: WasmRpcHost) {}

  /** Registered fully-qualified method names, from the Go router. */
  methods(): string[] {
    return this.host.methods();
  }

  /** Raw byte-level call. Prefer the typed clients. */
  async invoke(method: string, payload: Uint8Array): Promise<Uint8Array> {
    try {
      return await this.host.invoke(method, payload);
    } catch (e) {
      throw toWasmRpcError(e);
    }
  }

  /** Byte-level server stream. Prefer the typed clients. */
  listenRaw(method: string, payload: Uint8Array): WasmRpcStream<Uint8Array> {
    return listenRaw(this.host, method, payload);
  }

  /**
   * Typed unary call: encode with protobuf-es, cross the boundary once
   * per direction, decode the response.
   */
  async call<Req extends DescMessage, Resp extends DescMessage>(
    method: string,
    reqSchema: Req,
    respSchema: Resp,
    init: MessageInitShape<Req>,
  ): Promise<MessageShape<Resp>> {
    const payload = toBinary(reqSchema, create(reqSchema, init));
    const raw = await this.invoke(method, payload);
    return fromBinary(respSchema, raw);
  }
}

/** Factory producing a strongly-typed async method bound to a transport. */
export function unary<Req extends DescMessage, Resp extends DescMessage>(
  transport: WasmRpcTransport,
  method: string,
  reqSchema: Req,
  respSchema: Resp,
): (init: MessageInitShape<Req>) => Promise<MessageShape<Resp>> {
  return (init) => transport.call(method, reqSchema, respSchema, init);
}

function toWasmRpcError(e: unknown): WasmRpcError {
  if (e instanceof WasmRpcError) return e;
  if (e && typeof e === "object" && "message" in e) {
    const o = e as Partial<WasmRpcErrorShape>;
    return new WasmRpcError({
      code: typeof o.code === "string" ? o.code : "unknown",
      method: typeof o.method === "string" ? o.method : undefined,
      message: String(o.message),
    });
  }
  return new WasmRpcError({ code: "unknown", message: String(e) });
}

// ---------------------------------------------------------------------------
// Server-streaming
// ---------------------------------------------------------------------------

/** Async-iterable server stream with explicit cancellation. */
export interface WasmRpcStream<T> extends AsyncIterable<T> {
  cancel(): void;
}

interface StreamHandlers {
  onMessage(payload: Uint8Array): void;
  onError(err: unknown): void;
  onEnd(): void;
}

interface WasmRpcStreamHost {
  listen(method: string, payload: Uint8Array, handlers: StreamHandlers): number;
  cancel(id: number): void;
}

function streamHost(host: unknown): WasmRpcStreamHost {
  const h = host as Partial<WasmRpcStreamHost>;
  if (typeof h.listen !== "function" || typeof h.cancel !== "function") {
    throw new WasmRpcError({
      code: "unimplemented",
      message: "wasm host does not export listen/cancel (rebuild the module)",
    });
  }
  return h as WasmRpcStreamHost;
}

/**
 * Byte-level server stream over the host's listen/cancel exports,
 * exposed as an AsyncIterable with an unbounded ready-queue (frames are
 * small protobuf messages; apply backpressure at the protocol level if
 * needed).
 */
export function listenRaw(
  host: unknown,
  method: string,
  payload: Uint8Array,
): WasmRpcStream<Uint8Array> {
  type Slot =
    | { kind: "value"; value: Uint8Array }
    | { kind: "error"; error: WasmRpcError }
    | { kind: "end" };

  const queue: Slot[] = [];
  let wake: (() => void) | null = null;
  let finished = false;
  const push = (s: Slot) => {
    if (finished && s.kind === "value") return;
    if (s.kind !== "value") finished = true;
    queue.push(s);
    wake?.();
    wake = null;
  };

  const h = streamHost(host);
  const id = h.listen(method, payload, {
    onMessage: (m) => push({ kind: "value", value: m }),
    onError: (e) =>
      push({
        kind: "error",
        error:
          e instanceof WasmRpcError
            ? e
            : new WasmRpcError({
                code: (e as { code?: string })?.code ?? "unknown",
                method,
                message: String((e as { message?: string })?.message ?? e),
              }),
      }),
    onEnd: () => push({ kind: "end" }),
  });

  return {
    cancel: () => h.cancel(id),
    async *[Symbol.asyncIterator]() {
      for (;;) {
        while (queue.length === 0) {
          await new Promise<void>((r) => (wake = r));
        }
        const slot = queue.shift()!;
        if (slot.kind === "end") return;
        if (slot.kind === "error") throw slot.error;
        yield slot.value;
      }
    },
  };
}

/** Factory producing a strongly-typed server-streaming method. */
export function serverStream<Req extends DescMessage, Resp extends DescMessage>(
  transport: WasmRpcTransport,
  method: string,
  reqSchema: Req,
  respSchema: Resp,
): (init: MessageInitShape<Req>) => WasmRpcStream<MessageShape<Resp>> {
  return (init) => {
    const payload = toBinary(reqSchema, create(reqSchema, init));
    const raw = transport.listenRaw(method, payload);
    return {
      cancel: () => raw.cancel(),
      async *[Symbol.asyncIterator]() {
        for await (const frame of raw) {
          yield fromBinary(respSchema, frame);
        }
      },
    };
  };
}
