// Map/world rendering: ground, boundaries, and (once the protocol carries
// it) parkour platforms + store zone markers.
//
// TODO(owner): MapDef (net/protocol.ts) currently only has width/height.
// Extend it (and backend/internal/protocol + world.Default()) to include
// platforms/storeZones/checkpoints, then replace the placeholder grid
// below with real level geometry.
import type { MapDef } from "../net/protocol";
import type { Renderer } from "../render/canvas";

export function drawMap(r: Renderer, map: MapDef): void {
  // World boundary.
  r.rect(0, 0, map.width, 8, "#444");
  r.rect(0, map.height - 8, map.width, 8, "#444");
  r.rect(0, 0, 8, map.height, "#444");
  r.rect(map.width - 8, 0, 8, map.height, "#444");

  // Light grid so motion is readable even before real level art exists.
  const gridSize = 200;
  for (let x = 0; x < map.width; x += gridSize) {
    r.line(x, 0, x, map.height, "#26263a");
  }
  for (let y = 0; y < map.height; y += gridSize) {
    r.line(0, y, map.width, y, "#26263a");
  }
}
