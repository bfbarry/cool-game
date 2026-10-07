# Nir Game — Project Context

Browser-based multiplayer 2D top-down shooter. Players customize a character, fight each
other with hitscan/projectile weapons and health bars, earn **Points** from PvP kills, killing
hostile NPCs, and parkour challenges, then spend Points in in-world **Stores** on weapon
upgrades and — at higher cost — vehicles (e.g. helicopters).

This doc is the shared source of truth. It defines the architecture, the wire protocol between
client and server, and the module boundaries so several people can build different pieces at
once without stepping on each other. **Read the "Module Ownership" table first** — pick a row,
work inside those files, respect the protocol contract in `backend/internal/protocol` /
`frontend/src/net/protocol.ts` (keep both in sync when you change a message shape).

---

## 1. Tech Stack

| Layer | Choice | Notes |
|---|---|---|
| Backend | Go | Single authoritative game server. No external game-engine dependency; plain `net/http` + a websocket lib. |
| Realtime transport | **WebSockets** (primary) | Client input up, authoritative state snapshots down. JSON messages (see §3). |
| Realtime transport | **WebRTC DataChannel** (optional, later) | Not needed for v1. Candidate for peer voice chat, or for a future low-latency unreliable channel for position updates once the game has scale. Do not block v1 on this. |
| Frontend | TypeScript + Canvas 2D | Build tool: Vite. No framework needed for the game canvas; a light UI layer (store, HUD) can be plain DOM + TS. |
| Serialization | JSON over WebSocket | Simple, debuggable, good enough for a 2D shooter's tick rate. Revisit (e.g. MessagePack) only if bandwidth becomes a real problem. |

### Why WebSockets, not WebRTC, for game state

The server must stay **authoritative** (positions, health, points, purchases) so players can't
cheat by forging client messages — that requires a central hub the server controls, which is
exactly what a WebSocket hub gives you for free. WebRTC data channels are peer-oriented and add
NAT-traversal/signaling complexity that buys nothing here. Keep WebRTC in the toolbox for voice
chat or future optimization, not as the game-state transport.

---

## 2. High-Level Architecture

```
                     ┌─────────────────────────────────────────────┐
                     │                 Go Server                    │
                     │                                               │
  Browser ──WS──▶    │  server/   Hub + per-connection Client pumps  │
  (TS client)        │     │                                         │
                      │     ▼                                        │
                     │  game/     Fixed-tick World loop (20 Hz)      │
                     │     │  owns authoritative state, calls into:  │
                     │     ├──▶ player/    player state, movement    │
                     │     ├──▶ combat/    shooting, damage, health  │
                     │     ├──▶ economy/   points ledger, store      │
                     │     ├──▶ vehicles/  helicopters etc.          │
                     │     ├──▶ npc/       hostile NPC AI             │
                     │     └──▶ world/     map, parkour geometry      │
                     │     │                                         │
                     │     ▼                                        │
                     │  protocol/  message structs (shared contract) │
                     └─────────────────────────────────────────────┘
                                       │ JSON snapshots each tick
                                       ▼
                     ┌─────────────────────────────────────────────┐
                     │              TypeScript Client                │
                     │  net/      WebSocket client, protocol mirror  │
                     │  game/     client World model + render loop   │
                     │  render/   Canvas 2D renderer                 │
                     │  player/   local input, customization UI      │
                     │  combat/   shooting fx, health bar rendering  │
                     │  store/    upgrade/purchase UI                │
                     │  vehicles/ vehicle rendering + controls       │
                     │  npc/      NPC rendering                      │
                     │  world/    map/parkour rendering              │
                     │  ui/       HUD (points, health, prompts)      │
                     └─────────────────────────────────────────────┘
```

The server is the only source of truth. The client renders the latest snapshot and predicts
local movement for responsiveness; it never decides who died or how many points it has.

---

## 3. Wire Protocol

All messages are JSON with an envelope:

```json
{ "type": "string", "data": { ... } }
```

Defined once per language and kept in lockstep:
- Go: `backend/internal/protocol/messages.go`
- TS: `frontend/src/net/protocol.ts`

### Client → Server

| type | data | purpose |
|---|---|---|
| `join` | `{ name, colorPrimary, colorSecondary, hatId }` | Sent once after connecting; requests spawn with chosen cosmetics. |
| `input` | `{ seq, dx, dy, aimAngle, firing, tick }` | Per-frame movement/aim/fire intent. Server is authoritative over the result. |
| `enter_store` / `exit_store` | `{}` | Client notifies UI intent; server still gates purchases by proximity. |
| `purchase` | `{ itemId }` | Buy a weapon/vehicle/cosmetic with Points. Server validates funds + location. |
| `enter_vehicle` / `exit_vehicle` | `{ vehicleId }` | Request to mount/dismount a vehicle. |

### Server → Client

| type | data | purpose |
|---|---|---|
| `welcome` | `{ playerId, map, storeCatalog }` | First message after `join`; static/startup data. |
| `state` | `{ tick, players[], npcs[], vehicles[], projectiles[] }` | Authoritative snapshot, broadcast every tick. |
| `player_joined` / `player_left` | `{ playerId }` | Roster changes. |
| `damage` | `{ targetId, sourceId, amount, newHealth }` | Lets client play hit fx / update health bar immediately. |
| `death` | `{ victimId, killerId, pointsAwarded, cause }` | `cause`: `"player" \| "npc_kill" \| "fall" \| ...` |
| `points_update` | `{ playerId, points }` | Pushed on any point-earning event. |
| `purchase_result` | `{ success, itemId, points, reason }` | Response to `purchase`. |

See `backend/internal/protocol/messages.go` for the exact Go structs — treat it as canonical;
this table is the human-readable summary.

---

## 4. Core Data Models

```go
Player struct {
    ID         string
    Name       string
    X, Y       float64
    VX, VY     float64
    AimAngle   float64
    Health     int
    MaxHealth  int
    Points     int
    WeaponID   string
    Cosmetics  Cosmetics
    State      string // "on_foot" | "in_vehicle"
    VehicleID  string // set when State == in_vehicle
}

Weapon struct {
    ID         string
    Name       string
    Damage     int
    FireRateMs int
    BulletSpeed float64
    Range      float64
    Cost       int
    Tier       int
}

Vehicle struct {
    ID       string
    Type     string // "helicopter" ...
    X, Y     float64
    Angle    float64
    Health   int
    PilotID  string
    Cost     int
}

NPC struct {
    ID       string
    Type     string
    X, Y     float64
    Health   int
    Damage   int
    AIState  string
}

StoreItem struct {
    ID   string
    Name string
    Kind string // "weapon" | "vehicle" | "cosmetic"
    Cost int
}
```

Mirror these as TS interfaces in `frontend/src/net/protocol.ts`. Keep field names identical
(camelCase on the wire; Go structs use JSON tags to match).

---

## 5. Points Economy (tune freely, starting values)

| Event | Points |
|---|---|
| Kill another player | 100 |
| Kill a hostile NPC | 25 |
| Clear a parkour checkpoint | 15 |
| Finish a parkour course | 75 bonus |

Weapons cost a few hundred points (tiered upgrades: pistol → SMG → rifle → sniper). Vehicles
(helicopter, etc.) are a large sink — thousands of points — so they stay a late-game goal.
Exact numbers live in `backend/internal/economy/store.go` as the single source of truth; don't
hardcode prices anywhere else.

---

## 6. Module Ownership

Pick a row. Work mainly inside those files. If you need to change the protocol, update both
the Go and TS protocol files in the same PR and call it out.

| Area | Backend files | Frontend files | Depends on |
|---|---|---|---|
| **Networking / connection hub** | `backend/internal/server/*`, `backend/cmd/server/main.go` | `frontend/src/net/*` | protocol |
| **Protocol contract** | `backend/internal/protocol/messages.go` | `frontend/src/net/protocol.ts` | — (everyone depends on this; coordinate changes) |
| **Player movement & customization** | `backend/internal/player/*` | `frontend/src/player/*` | protocol, render |
| **Combat (shooting, health, damage)** | `backend/internal/combat/*` | `frontend/src/combat/*` | player |
| **Points & Store economy** | `backend/internal/economy/*` | `frontend/src/store/*`, `frontend/src/ui/*` | player, combat |
| **Vehicles (helicopters etc.)** | `backend/internal/vehicles/*` | `frontend/src/vehicles/*` | player, economy |
| **Hostile NPCs** | `backend/internal/npc/*` | `frontend/src/npc/*` | combat |
| **World / map / parkour** | `backend/internal/world/*` | `frontend/src/world/*` | player |
| **Rendering / HUD** | — | `frontend/src/render/*`, `frontend/src/ui/*` | everything (reads client World model) |
| **Game loop orchestration** | `backend/internal/game/*` | `frontend/src/game/*` | all gameplay modules |

Every package above currently contains stub types/functions with `TODO(owner):` comments
describing what to implement — start there.

---

## 7. Tick Rate & Loop

- Server runs a fixed 20 Hz tick (50ms) in `backend/internal/game/loop.go`: read buffered
  inputs → update player/NPC/vehicle state → resolve combat → broadcast a `state` snapshot.
- Client runs a `requestAnimationFrame` loop in `frontend/src/game/loop.ts`: sends `input` at
  render rate (throttled to ~20-30/s), interpolates between the last two `state` snapshots for
  smooth rendering of remote entities, and predicts the local player's own movement.

---

## 8. Directory Layout

```
nir-game/
├── CONTEXT.md
├── backend/
│   ├── go.mod
│   ├── cmd/server/main.go
│   └── internal/
│       ├── protocol/   # wire message structs
│       ├── server/     # websocket hub + client pumps
│       ├── game/        # fixed-tick world loop
│       ├── player/      # player state + customization
│       ├── combat/      # weapons, damage, health
│       ├── economy/     # points, store catalog, purchases
│       ├── vehicles/    # vehicles (helicopter, ...)
│       ├── npc/         # hostile NPC state + AI
│       └── world/       # map + parkour geometry
└── frontend/
    ├── package.json
    ├── tsconfig.json
    ├── vite.config.ts
    ├── index.html
    └── src/
        ├── main.ts
        ├── net/         # websocket client + protocol mirror
        ├── game/        # client world model + loop
        ├── render/       # canvas renderer
        ├── player/       # local input + customization UI
        ├── combat/       # shooting fx + health bars
        ├── store/        # upgrade/purchase UI
        ├── vehicles/     # vehicle rendering + controls
        ├── npc/          # NPC rendering
        ├── world/        # map/parkour rendering
        └── ui/           # HUD
```

---

## 9. Running It

```bash
# backend
cd backend
go mod tidy
go run ./cmd/server          # starts on :8080, ws endpoint at /ws

# frontend (separate terminal)
cd frontend
npm install
npm run dev                  # Vite dev server, proxies /ws to :8080
```

---

## 10. Contribution Workflow

1. Claim a row in §6 (comment in the PR / issue tracker, or just say so in chat).
2. Don't change files outside your area without a heads-up — especially `protocol/`.
3. Keep server authoritative: never trust a client-sent position, health, or point value;
   always compute/validate on the backend.
4. When you touch the protocol, update **both** `messages.go` and `protocol.ts`, and bump the
   relevant table in §3 of this doc in the same change.
5. Prices, damage numbers, tick rate — treat the constants called out in this doc as the
   starting values; tune in code, but keep this doc's tables updated if you change them
   meaningfully.

---

## 11. Roadmap (suggested milestones)

1. **M1 — Connect & move**: join flow, customization, server-authoritative movement, broadcast
   loop, canvas rendering of self + remote players.
2. **M2 — Combat**: shooting, projectiles/hitscan, health bars, death/respawn, points on kill.
3. **M3 — Economy**: store proximity UI, purchase flow, weapon tiers.
4. **M4 — World**: parkour platforming geometry + checkpoints, hostile NPCs with basic AI.
5. **M5 — Vehicles**: helicopters — purchase, mount/dismount, flight controls, vehicle combat.
6. **M6 — Polish**: interpolation/prediction smoothing, spectating/respawn UX, optional voice
   chat over WebRTC.
