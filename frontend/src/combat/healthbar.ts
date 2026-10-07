import type { Renderer } from "../render/canvas";

/** Shared health bar drawing used for players and NPCs. */
export function drawHealthBar(r: Renderer, x: number, y: number, health: number, maxHealth: number): void {
  const width = 48;
  const height = 6;
  const pct = maxHealth > 0 ? Math.max(0, health / maxHealth) : 0;

  r.rect(x - width / 2, y, width, height, "#000");
  const color = pct > 0.5 ? "#4caf50" : pct > 0.2 ? "#ffb300" : "#e53935";
  r.rect(x - width / 2, y, width * pct, height, color);
}
