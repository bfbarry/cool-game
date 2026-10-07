// HUD overlay: local player's health + points. Lives in plain DOM
// (overlaid via #ui-root), separate from the canvas world-render.
//
// TODO(owner): minimap, kill feed, respawn countdown, low-health vignette.
import type { PlayerState } from "../net/protocol";

export class HUD {
  private el: HTMLDivElement;

  constructor() {
    this.el = document.createElement("div");
    this.el.style.cssText =
      "position:absolute;top:12px;left:12px;background:rgba(0,0,0,0.5);" +
      "padding:8px 12px;border-radius:6px;font-size:14px;";
    document.getElementById("ui-root")!.appendChild(this.el);
  }

  update(local: PlayerState | undefined): void {
    if (!local) {
      this.el.textContent = "Connecting...";
      return;
    }
    this.el.textContent = `${local.name} — HP ${local.health}/${local.maxHealth} — ${local.points} pts — ${local.weaponId}`;
  }
}
