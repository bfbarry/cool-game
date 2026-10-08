// HUD overlay: local player's health + points, plus the help hint/controls
// modal. Lives in plain DOM (overlaid via #ui-root), separate from the
// canvas world-render.
//
// TODO(owner): minimap, kill feed, respawn countdown, low-health vignette.
// Keep CONTROLS below in sync with actual keybindings as modules add them
// (see player/player.ts for WASD/mouse, store/store.ts for "B").
import type { PlayerState } from "../net/protocol";

const CONTROLS: Array<[string, string]> = [
  ["W A S D", "Move"],
  ["Mouse", "Aim"],
  ["Left Click", "Fire"],
  ["B", "Open / close store"],
  ["H", "Toggle this help menu"],
];

export class HUD {
  private el: HTMLDivElement;
  private helpHint: HTMLDivElement;
  private helpPanel: HTMLDivElement;
  private helpOpen = false;

  constructor() {
    const root = document.getElementById("ui-root")!;

    this.el = document.createElement("div");
    this.el.style.cssText =
      "position:absolute;top:12px;left:12px;background:rgba(0,0,0,0.5);" +
      "padding:8px 12px;border-radius:6px;font-size:14px;";
    root.appendChild(this.el);

    this.helpHint = document.createElement("div");
    this.helpHint.textContent = "Press H for help";
    this.helpHint.style.cssText =
      "position:absolute;top:12px;left:50%;transform:translateX(-50%);" +
      "background:rgba(0,0,0,0.5);padding:6px 12px;border-radius:6px;" +
      "font-size:13px;color:#ccc;";
    root.appendChild(this.helpHint);

    this.helpPanel = document.createElement("div");
    this.helpPanel.className = "interactive";
    this.helpPanel.style.cssText =
      "position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);" +
      "background:#222;color:#fff;padding:20px 24px;border-radius:8px;" +
      "min-width:220px;display:none;";
    this.helpPanel.innerHTML =
      `<h3 style="margin-top:0">Controls</h3>` +
      `<table style="border-spacing:0 6px">${CONTROLS.map(
        ([key, action]) =>
          `<tr><td style="padding-right:16px;font-family:monospace;color:#ffd700">${key}</td><td>${action}</td></tr>`,
      ).join("")}</table>` +
      `<p style="color:#999;font-size:12px;margin-bottom:0">Press H to close</p>`;
    root.appendChild(this.helpPanel);

    window.addEventListener("keydown", (e) => {
      if (e.key.toLowerCase() === "h") this.toggleHelp();
    });
  }

  private toggleHelp(): void {
    this.helpOpen = !this.helpOpen;
    this.helpPanel.style.display = this.helpOpen ? "block" : "none";
  }

  update(local: PlayerState | undefined): void {
    if (!local) {
      this.el.textContent = "Connecting...";
      return;
    }
    this.el.textContent = `${local.name} — HP ${local.health}/${local.maxHealth} — ${local.points} pts — ${local.weaponId}`;
  }
}
