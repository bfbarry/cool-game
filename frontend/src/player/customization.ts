// Pre-game character customization screen. Shown once before connecting;
// result is sent as the `join` message (see net/protocol.ts JoinData).
//
// TODO(owner): replace the plain-DOM form with real art (skin/hat preview
// canvas), keep HatID options in sync with backend/internal/player/customization.go's
// ValidHatIDs.
import type { JoinData } from "../net/protocol";

const HAT_OPTIONS = ["none", "cap", "helmet", "beanie"];

export function showCustomizationScreen(onReady: (join: JoinData) => void): void {
  const root = document.getElementById("ui-root")!;
  const panel = document.createElement("div");
  panel.className = "interactive";
  panel.style.cssText =
    "position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);" +
    "background:#222;padding:24px;border-radius:8px;display:flex;flex-direction:column;gap:12px;min-width:260px;";

  panel.innerHTML = `
    <h2 style="margin:0">Customize your player</h2>
    <label>Name <input id="cfg-name" maxlength="16" value="Player" /></label>
    <label>Primary color <input id="cfg-color1" type="color" value="#3b82f6" /></label>
    <label>Secondary color <input id="cfg-color2" type="color" value="#1e3a8a" /></label>
    <label>Hat
      <select id="cfg-hat">
        ${HAT_OPTIONS.map((h) => `<option value="${h}">${h}</option>`).join("")}
      </select>
    </label>
    <button id="cfg-start">Enter game</button>
  `;
  root.appendChild(panel);

  panel.querySelector<HTMLButtonElement>("#cfg-start")!.onclick = () => {
    const join: JoinData = {
      name: panel.querySelector<HTMLInputElement>("#cfg-name")!.value || "Player",
      colorPrimary: panel.querySelector<HTMLInputElement>("#cfg-color1")!.value,
      colorSecondary: panel.querySelector<HTMLInputElement>("#cfg-color2")!.value,
      hatId: panel.querySelector<HTMLSelectElement>("#cfg-hat")!.value,
    };
    root.removeChild(panel);
    onReady(join);
  };
}
