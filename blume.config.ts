import { defineConfig } from "blume";

export default defineConfig({
  title: "wasm-rpc",
  description:
    "Run a Go Protobuf RPC server inside WebAssembly — in the browser via syscall/js or in any WASI host via a go:wasmexport ABI — with generated clients for TypeScript, Dart, Kotlin, Swift, and Go.",

  github: {
    owner: "prdlk",
    repo: "wasm-rpc",
    branch: "main",
  },

  content: {
    root: "docs",
  },

  // The repository's `examples/` folder holds Go and Vite apps, not
  // previewable <Component> sources. Point Blume elsewhere so it never
  // tries to render them as islands.
  examples: "docs/_examples",

  theme: {
    accent: "violet",
    radius: "md",
    mode: "system",
  },

  navigation: {
    sidebar: {
      display: "group",
    },
  },

  markdown: {
    code: {
      icons: true,
    },
  },

  // GitHub Pages project site: https://prdlk.github.io/wasm-rpc
  deployment: {
    output: "static",
    site: "https://prdlk.github.io",
    base: "/wasm-rpc",
  },
});
