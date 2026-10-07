// Local input capture + rendering for players (local and remote). Server
// is authoritative; this module reads keyboard/mouse, reports intent, and
// draws whatever PlayerState the server last sent (see game/world.ts).
//
// TODO(owner): local-player movement prediction (apply input immediately
// client-side, reconcile against server snapshots) for responsiveness.
import type { PlayerState } from "../net/protocol";
import type { Renderer } from "../render/canvas";
import { drawHealthBar } from "../combat/healthbar";

export class InputState {
  dx = 0;
  dy = 0;
  aimAngle = 0;
  firing = false;
  private seq = 0;

  nextSeq(): number {
    return ++this.seq;
  }
}

/** Wires keyboard + mouse to an InputState. Call once; read `state` each frame. */
export function bindInput(canvas: HTMLCanvasElement, state: InputState): void {
  const keys = new Set<string>();
  window.addEventListener("keydown", (e) => keys.add(e.key.toLowerCase()));
  window.addEventListener("keyup", (e) => keys.delete(e.key.toLowerCase()));

  const updateAxes = () => {
    state.dx = (keys.has("d") ? 1 : 0) - (keys.has("a") ? 1 : 0);
    state.dy = (keys.has("s") ? 1 : 0) - (keys.has("w") ? 1 : 0);
  };
  window.setInterval(updateAxes, 16);

  canvas.addEventListener("mousemove", (e) => {
    const rect = canvas.getBoundingClientRect();
    const cx = canvas.width / 2;
    const cy = canvas.height / 2;
    // Local player is always camera-centered (see Renderer.centerOn), so
    // aim angle is just screen-center -> mouse.
    state.aimAngle = Math.atan2(e.clientY - rect.top - cy, e.clientX - rect.left - cx);
  });
  canvas.addEventListener("mousedown", () => (state.firing = true));
  canvas.addEventListener("mouseup", () => (state.firing = false));
}

/** Draws one player entity (local or remote) from the latest server state. */
export function drawPlayer(r: Renderer, p: PlayerState, isLocal: boolean): void {
  r.circle(p.x, p.y, 24, p.cosmetics.colorPrimary || "#3b82f6");
  // Facing/aim indicator.
  const tipX = p.x + Math.cos(p.aimAngle) * 32;
  const tipY = p.y + Math.sin(p.aimAngle) * 32;
  r.line(p.x, p.y, tipX, tipY, p.cosmetics.colorSecondary || "#1e3a8a", 3);

  r.text(p.x, p.y - 36, p.name, isLocal ? "#ffd700" : "#fff");
  drawHealthBar(r, p.x, p.y - 44, p.health, p.maxHealth);

  // TODO(player owner): draw hat/cosmetic (p.cosmetics.hatId) once there's
  // sprite art; currently cosmetics only affect color.
}
