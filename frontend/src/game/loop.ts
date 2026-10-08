// Client render/input loop: requestAnimationFrame-driven, sends input at a
// throttled rate, renders the latest ClientWorld snapshot.
//
// TODO(owner): local movement prediction (see player/player.ts TODO) —
// right now the local player only moves once the server echoes back a new
// position in a `state` snapshot, which will feel laggy over real network
// latency.
import type { GameSocket } from "../net/socket";
import type { ClientWorld } from "./world";
import type { Renderer } from "../render/canvas";
import type { InputState } from "../player/player";
import { drawPlayer } from "../player/player";
import { drawNPC } from "../npc/npc";
import { drawVehicle } from "../vehicles/vehicle";
import { drawProjectiles } from "../combat/combat";
import { drawMap } from "../world/map";
import { HUD } from "../ui/hud";
import { MessageType } from "../net/protocol";

const INPUT_SEND_HZ = 30;

export class GameLoop {
  private hud = new HUD();
  private lastInputSent = 0;

  constructor(
    private socket: GameSocket,
    private world: ClientWorld,
    private renderer: Renderer,
    private input: InputState,
  ) {}

  start(): void {
    requestAnimationFrame(this.frame);
  }

  private frame = (now: number): void => {
    try {
      this.maybeSendInput(now);
      this.render();
    } catch (err) {
      // A single bad frame (e.g. an unexpected snapshot shape) should never
      // silently kill the loop — log and keep going rather than leaving the
      // player stuck on a frozen screen.
      console.error("game loop: frame error", err);
    }
    requestAnimationFrame(this.frame);
  };

  private maybeSendInput(now: number): void {
    const minInterval = 1000 / INPUT_SEND_HZ;
    if (now - this.lastInputSent < minInterval) return;
    this.lastInputSent = now;
    this.socket.send(MessageType.Input, {
      seq: this.input.nextSeq(),
      dx: this.input.dx,
      dy: this.input.dy,
      aimAngle: this.input.aimAngle,
      firing: this.input.firing,
      tick: this.world.latest.tick,
    });
  }

  private render(): void {
    const local = this.world.localPlayer();
    if (local) this.renderer.centerOn(local.x, local.y);

    this.renderer.clear();
    if (this.world.welcome) drawMap(this.renderer, this.world.welcome.map);

    for (const v of this.world.latest.vehicles) drawVehicle(this.renderer, v);
    for (const n of this.world.latest.npcs) drawNPC(this.renderer, n);
    for (const p of this.world.latest.players) drawPlayer(this.renderer, p, p.id === this.world.localPlayerId);
    drawProjectiles(this.renderer, this.world.latest.projectiles);

    this.hud.update(local);
  }
}
