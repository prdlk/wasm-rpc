// web-streaming: server-streaming wasm-rpc from React. The Go module
// inside the page emits a paced tick stream; the UI consumes it as an
// AsyncIterable via the generated client and can cancel mid-flight.
import { useEffect, useRef, useState } from "react";
import {
  loadWasmRpc,
  WasmRpcError,
  type WasmRpcStream,
} from "@prdlk/wasm-rpc-client";
import { TickerServiceClient } from "@prdlk/wasm-rpc-client/gen/ticker/v1/ticker_wasmrpc.pb.js";

type Tick = { topic: string; seq: number; unixMillis: bigint };

export default function App() {
  const [client, setClient] = useState<TickerServiceClient | null>(null);
  const [topic, setTopic] = useState("prices");
  const [count, setCount] = useState(20);
  const [intervalMs, setIntervalMs] = useState(250);
  const [ticks, setTicks] = useState<Tick[]>([]);
  const [status, setStatus] = useState<"idle" | "streaming" | "done" | "cancelled" | "error">("idle");
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const streamRef = useRef<WasmRpcStream<Tick> | null>(null);

  useEffect(() => {
    loadWasmRpc("/server.wasm")
      .then((t) => setClient(new TickerServiceClient(t)))
      .catch((e) => setLoadError(String(e)));
  }, []);

  const start = async (topicOverride?: string) => {
    if (!client || status === "streaming") return;
    setTicks([]);
    setError(null);
    setStatus("streaming");
    const stream = client.subscribe({ topic: topicOverride ?? topic, count, intervalMs });
    streamRef.current = stream;
    try {
      for await (const tick of stream) {
        setTicks((t) => [tick, ...t]);
      }
      setStatus((s) => (s === "streaming" ? "done" : s));
    } catch (e) {
      setError(e instanceof WasmRpcError ? `${e.code}: ${e.message}` : String(e));
      setStatus("error");
    } finally {
      streamRef.current = null;
    }
  };

  const cancel = () => {
    streamRef.current?.cancel();
    setStatus("cancelled");
  };

  if (loadError) return <pre style={{ color: "crimson" }}>{loadError}</pre>;
  if (!client) return <p>Instantiating Go Wasm module…</p>;

  return (
    <main style={{ fontFamily: "monospace", maxWidth: 640, margin: "2rem auto" }}>
      <h1>wasm-rpc · web-streaming</h1>
      <p>
        <code>ticker.v1.TickerService/Subscribe</code> — a Go
        server-stream running inside this page, consumed as an{" "}
        <code>AsyncIterable</code>.
      </p>

      <label>
        topic <input value={topic} onChange={(e) => setTopic(e.target.value)} />
      </label>{" "}
      <label>
        count{" "}
        <input
          type="number"
          value={count}
          min={1}
          onChange={(e) => setCount(Number(e.target.value))}
          style={{ width: "5rem" }}
        />
      </label>{" "}
      <label>
        interval ms{" "}
        <input
          type="number"
          value={intervalMs}
          min={0}
          onChange={(e) => setIntervalMs(Number(e.target.value))}
          style={{ width: "5rem" }}
        />
      </label>

      <p>
        <button onClick={() => void start()} disabled={status === "streaming"}>
          Subscribe
        </button>{" "}
        <button onClick={cancel} disabled={status !== "streaming"}>
          Cancel
        </button>{" "}
        <button
          title="topic 'explode' panics the Go handler; the panic is recovered and surfaces as a typed internal error"
          onClick={() => {
            setTopic("explode");
            void start("explode");
          }}
          disabled={status === "streaming"}
        >
          Trigger handler panic
        </button>{" "}
        status: <b>{status}</b>
      </p>

      {error && <p style={{ color: "crimson" }}>✖ {error}</p>}

      <ul style={{ listStyle: "none", padding: 0 }}>
        {ticks.map((t) => (
          <li key={`${t.topic}-${t.seq}`}>
            #{t.seq} <b>{t.topic}</b> @ {new Date(Number(t.unixMillis)).toLocaleTimeString()}
          </li>
        ))}
      </ul>
    </main>
  );
}
