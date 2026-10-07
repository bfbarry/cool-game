// Mirrors backend/internal/protocol/messages.go field-for-field. If you
// change a shape here, change it there too (and update CONTEXT.md §3) in
// the same change.

export interface Envelope<T = unknown> {
  type: string;
  data: T;
}

// ---- Client -> Server payloads ----

export interface JoinData {
  name: string;
  colorPrimary: string;
  colorSecondary: string;
  hatId: string;
}

export interface InputData {
  seq: number;
  dx: number; // movement intent, -1..1
  dy: number;
  aimAngle: number; // radians
  firing: boolean;
  tick: number;
}

export type EnterStoreData = Record<string, never>;
export type ExitStoreData = Record<string, never>;

export interface PurchaseData {
  itemId: string;
}

export interface EnterVehicleData {
  vehicleId: string;
}

export interface ExitVehicleData {
  vehicleId: string;
}

// ---- Server -> Client payloads ----

export interface Cosmetics {
  colorPrimary: string;
  colorSecondary: string;
  hatId: string;
}

export type PlayerLocomotionState = "on_foot" | "in_vehicle";

export interface PlayerState {
  id: string;
  name: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
  aimAngle: number;
  health: number;
  maxHealth: number;
  points: number;
  weaponId: string;
  cosmetics: Cosmetics;
  state: PlayerLocomotionState;
  vehicleId: string;
}

export interface NPCState {
  id: string;
  type: string;
  x: number;
  y: number;
  health: number;
  aiState: string;
}

export interface VehicleState {
  id: string;
  type: string;
  x: number;
  y: number;
  angle: number;
  health: number;
  pilotId: string;
}

export interface ProjectileState {
  id: string;
  x: number;
  y: number;
  angle: number;
  speed: number;
  owner: string;
}

export interface MapDef {
  width: number;
  height: number;
  // TODO(world owner): platforms, parkour checkpoints, store zones, spawn points.
}

export type StoreItemKind = "weapon" | "vehicle" | "cosmetic";

export interface StoreItem {
  id: string;
  name: string;
  kind: StoreItemKind;
  cost: number;
}

export interface StoreCatalog {
  items: StoreItem[];
}

export interface WelcomeData {
  playerId: string;
  map: MapDef;
  storeCatalog: StoreCatalog;
}

export interface StateData {
  tick: number;
  players: PlayerState[];
  npcs: NPCState[];
  vehicles: VehicleState[];
  projectiles: ProjectileState[];
}

export interface PlayerJoinedData {
  playerId: string;
}

export interface PlayerLeftData {
  playerId: string;
}

export interface DamageData {
  targetId: string;
  sourceId: string;
  amount: number;
  newHealth: number;
}

export type DeathCause = "player" | "npc_kill" | "fall" | string;

export interface DeathData {
  victimId: string;
  killerId: string;
  pointsAwarded: number;
  cause: DeathCause;
}

export interface PointsUpdateData {
  playerId: string;
  points: number;
}

export interface PurchaseResultData {
  success: boolean;
  itemId: string;
  points: number;
  reason?: string;
}

// Message type string constants — keep in sync with protocol.go's Type* consts.
export const MessageType = {
  Join: "join",
  Input: "input",
  EnterStore: "enter_store",
  ExitStore: "exit_store",
  Purchase: "purchase",
  EnterVehicle: "enter_vehicle",
  ExitVehicle: "exit_vehicle",

  Welcome: "welcome",
  State: "state",
  PlayerJoined: "player_joined",
  PlayerLeft: "player_left",
  Damage: "damage",
  Death: "death",
  PointsUpdate: "points_update",
  PurchaseResult: "purchase_result",
} as const;
