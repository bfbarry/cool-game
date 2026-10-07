// Hostile NPC rendering. AI/behavior is entirely server-side
// (backend/internal/npc) — this just draws whatever state arrives.
//
// TODO(owner): distinct art per NPC `type`, aggro indicator from `aiState`.
import type { NPCState } from "../net/protocol";
import type { Renderer } from "../render/canvas";
import { drawHealthBar } from "../combat/healthbar";

export function drawNPC(r: Renderer, n: NPCState): void {
  r.circle(n.x, n.y, 18, "#8e24aa");
  drawHealthBar(r, n.x, n.y - 30, n.health, 40); // TODO(owner): pull real max health from npc.DefaultMaxHealth equivalent on client
}
