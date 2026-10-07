# Nir Game

Browser multiplayer 2D shooter — Go backend, TypeScript frontend.

**Start here: [`CONTEXT.md`](./CONTEXT.md)** — architecture, wire protocol, data models, and
the module ownership table that splits the codebase so multiple people can build different
parts (networking, player, combat, economy/store, vehicles, NPCs, world/parkour, rendering) at
the same time.

## Quick start

```bash
# backend
cd backend && go run ./cmd/server      # :8080, websocket at /ws

# frontend (separate terminal)
cd frontend && npm install && npm run dev
```
