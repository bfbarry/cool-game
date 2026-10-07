// Store / upgrade UI. Lets the player spend Points on weapons and
// vehicles. The server is the source of truth on what's affordable and
// whether the player is actually in a store zone (see
// backend/internal/economy/store.go, world.InStoreZone) — this UI is just
// a convenience front for `purchase` messages.
//
// TODO(owner): MapDef (net/protocol.ts) doesn't yet carry store zone
// geometry, so the UI currently opens on a keybind rather than real
// proximity detection. Once world/map.ts has zone coords from the server,
// open/close this based on actually standing in a zone (and send
// enter_store/exit_store accordingly instead of on every keypress).
import type { GameSocket } from "../net/socket";
import type { StoreCatalog, PurchaseResultData } from "../net/protocol";
import { MessageType } from "../net/protocol";

export class StoreUI {
  private panel: HTMLDivElement;
  private open = false;

  constructor(private socket: GameSocket, private catalog: StoreCatalog) {
    this.panel = document.createElement("div");
    this.panel.className = "interactive";
    this.panel.style.cssText =
      "position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);" +
      "background:#222;padding:20px;border-radius:8px;color:#fff;display:none;min-width:240px;";
    document.getElementById("ui-root")!.appendChild(this.panel);
    this.render();

    window.addEventListener("keydown", (e) => {
      if (e.key.toLowerCase() === "b") this.toggle();
    });

    socket.on(MessageType.PurchaseResult, (data) => this.onPurchaseResult(data as PurchaseResultData));
  }

  private toggle(): void {
    this.open = !this.open;
    this.panel.style.display = this.open ? "block" : "none";
    this.socket.send(this.open ? MessageType.EnterStore : MessageType.ExitStore, {});
  }

  private render(): void {
    this.panel.innerHTML = `<h3 style="margin-top:0">Store (B to close)</h3>`;
    for (const item of this.catalog.items) {
      const row = document.createElement("div");
      row.style.cssText = "display:flex;justify-content:space-between;gap:16px;padding:4px 0;";
      row.innerHTML = `<span>${item.name} (${item.kind})</span><span>${item.cost} pts</span>`;
      const btn = document.createElement("button");
      btn.textContent = "Buy";
      btn.onclick = () => this.socket.send(MessageType.Purchase, { itemId: item.id });
      row.appendChild(btn);
      this.panel.appendChild(row);
    }
  }

  private onPurchaseResult(result: PurchaseResultData): void {
    // TODO(owner): replace alert with a proper toast/notification in ui/hud.ts.
    if (!result.success) {
      console.warn(`purchase failed: ${result.reason}`);
    }
  }
}
