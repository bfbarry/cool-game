import { defineConfig } from "vite";

// Proxies /ws to the Go backend during local dev so the client can connect
// with a relative URL (see src/net/socket.ts). Run `go run ./cmd/server`
// in backend/ first.
export default defineConfig({
  server: {
    proxy: {
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
});
