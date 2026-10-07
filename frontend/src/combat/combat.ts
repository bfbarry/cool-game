// Rendering for projectiles and combat feedback (hit flashes, death fx).
// Actual hit detection/damage is server-authoritative (see
// backend/internal/combat); this module only reacts to `damage`/`death`
// messages and draws what the server tells it happened.
//
// TODO(owner): hit flash / damage numbers / death animation / sound.
import type { ProjectileState, DamageData, DeathData } from "../net/protocol";
import type { Renderer } from "../render/canvas";

export function drawProjectiles(r: Renderer, projectiles: ProjectileState[]): void {
  for (const p of projectiles) {
    r.circle(p.x, p.y, 4, "#ffeb3b");
  }
}

/** Call when a `damage` message arrives. Currently a stub hook point. */
export function onDamage(_data: DamageData): void {
  // TODO(owner): flash the target, play a hit sound.
}

/** Call when a `death` message arrives. Currently a stub hook point. */
export function onDeath(_data: DeathData): void {
  // TODO(owner): death animation/fx, respawn countdown UI.
}
