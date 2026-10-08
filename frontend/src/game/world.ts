// Client-side mirror of authoritative server state. Owns nothing by
// itself — just listens to socket events and keeps the latest snapshot
// around for render/loop code to read. No game logic/decisions belong
// here; the server decided all of this already.
//
// TODO(owner): snapshot interpolation — currently the render loop reads
// `latest` directly, which will look jittery at a 20Hz tick rate. Keep the
// last two snapshots and lerp between them based on elapsed time.
import type { GameSocket } from "../net/socket";
import {
  MessageType,
  type StateData,
  type WelcomeData,
  type PlayerJoinedData,
  type PlayerLeftData,
  type DamageData,
  type DeathData,
  type PointsUpdateData,
} from "../net/protocol";
import { onDamage, onDeath } from "../combat/combat";

export class ClientWorld {
  localPlayerId = "";
  welcome: WelcomeData | null = null;
  latest: StateData = { tick: 0, players: [], npcs: [], vehicles: [], projectiles: [] };

  constructor(private socket: GameSocket) {
    socket.on(MessageType.Welcome, (data) => this.onWelcome(data as WelcomeData));
    socket.on(MessageType.State, (data) => this.onState(data as StateData));
    socket.on(MessageType.PlayerJoined, (data) => this.onPlayerJoined(data as PlayerJoinedData));
    socket.on(MessageType.PlayerLeft, (data) => this.onPlayerLeft(data as PlayerLeftData));
    socket.on(MessageType.Damage, (data) => onDamage(data as DamageData));
    socket.on(MessageType.Death, (data) => onDeath(data as DeathData));
    socket.on(MessageType.PointsUpdate, (data) => this.onPointsUpdate(data as PointsUpdateData));
  }

  // Defense-in-depth: the server always sends [] for empty lists (never
  // null), but normalize here too so one malformed snapshot can't throw
  // inside the render loop's `for...of` and kill requestAnimationFrame.
  private onState(data: StateData): void {
    this.latest = {
      tick: data.tick,
      players: data.players ?? [],
      npcs: data.npcs ?? [],
      vehicles: data.vehicles ?? [],
      projectiles: data.projectiles ?? [],
    };
  }

  private onWelcome(data: WelcomeData): void {
    this.localPlayerId = data.playerId;
    this.welcome = data;
  }

  private onPlayerJoined(_data: PlayerJoinedData): void {
    // TODO(owner): toast/notification, sound.
  }

  private onPlayerLeft(data: PlayerLeftData): void {
    this.latest.players = this.latest.players.filter((p) => p.id !== data.playerId);
  }

  private onPointsUpdate(data: PointsUpdateData): void {
    const p = this.latest.players.find((pl) => pl.id === data.playerId);
    if (p) p.points = data.points;
  }

  localPlayer() {
    return this.latest.players.find((p) => p.id === this.localPlayerId);
  }
}
