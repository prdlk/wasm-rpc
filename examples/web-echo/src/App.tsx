// web-echo: unary wasm-rpc calls from React through the generated
// TypeScript client. The "backend" is /server.wasm — a Go RPC server
// running inside the page.
import { useEffect, useState } from "react";
import {
  loadWasmRpc,
  WasmRpcError,
  type WasmRpcTransport,
} from "@prdlk/wasm-rpc-client";
import { EchoServiceClient } from "@prdlk/wasm-rpc-client/gen/echo/v1/echo_wasmrpc.pb.js";

type Entry = { kind: "ok" | "err"; text: string };

export default function App() {
  const [client, setClient] = useState<EchoServiceClient | null>(null);
  const [methods, setMethods] = useState<string[]>([]);
  const [message, setMessage] = useState("hello from the browser");
  const [repeat, setRepeat] = useState(1);
  const [log, setLog] = useState<Entry[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    let transport: WasmRpcTransport;
    loadWasmRpc("/server.wasm")
      .then((t) => {
        transport = t;
        setMethods(t.methods());
        setClient(new EchoServiceClient(t));
      })
      .catch((e) => setLoadError(String(e)));
  }, []);

  const push = (e: Entry) => setLog((l) => [e, ...l].slice(0, 20));

  const call = async (fn: () => Promise<{ message: string; unixMillis: bigint }>) => {
    try {
      const r = await fn();
      push({ kind: "ok", text: `${r.message}   (t=${r.unixMillis})` });
    } catch (e) {
      const msg =
        e instanceof WasmRpcError ? `${e.code}: ${e.message}` : String(e);
      push({ kind: "err", text: msg });
    }
  };

  if (loadError) return <pre style={{ color: "crimson" }}>{loadError}</pre>;
  if (!client) return <p>Instantiating Go Wasm module…</p>;

  return (
    <main style={{ fontFamily: "monospace", maxWidth: 640, margin: "2rem auto" }}>
      <h1>wasm-rpc · web-echo</h1>
      <p>
        Go server methods mounted in this page: <b>{methods.join(", ")}</b>
      </p>

      <label>
        message{" "}
        <input value={message} onChange={(e) => setMessage(e.target.value)} size={32} />
      </label>{" "}
      <label>
        repeat{" "}
        <input
          type="number"
          value={repeat}
          min={0}
          onChange={(e) => setRepeat(Number(e.target.value))}
          style={{ width: "5rem" }}
        />
      </label>

      <p>
        <button onClick={() => call(() => client.echo({ message, repeat }))}>
          Echo
        </button>{" "}
        <button onClick={() => call(() => client.reverse({ message }))}>
          Reverse
        </button>{" "}
        <button
          title="repeat > 1000 returns a typed invalid_argument error from Go"
          onClick={() => call(() => client.echo({ message, repeat: 5000 }))}
        >
          Trigger typed error
        </button>
      </p>

      <ul style={{ listStyle: "none", padding: 0 }}>
        {log.map((e, i) => (
          <li key={i} style={{ color: e.kind === "err" ? "crimson" : "inherit" }}>
            {e.kind === "err" ? "✖" : "✔"} {e.text}
          </li>
        ))}
      </ul>
    </main>
  );
}
