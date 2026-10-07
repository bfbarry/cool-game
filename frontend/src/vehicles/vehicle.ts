// Vehicle rendering + mount/dismount interaction (helicopters etc. bought
// in the store — see backend/internal/vehicles).
//
// TODO(owner): real sprite/shape per vehicle type, flight controls while
// piloting (currently reuses the same WASD input — server-side vehicle
// physics isn't wired to `input` yet, see game/loop.go TODO in backend).
import type { VehicleState } from "../net/protocol";
import type { Renderer } from "../render/canvas";
import { drawHealthBar } from "../combat/healthbar";
import type { GameSocket } from "../net/socket";
import { MessageType } from "../net/protocol";

export function drawVehicle(r: Renderer, v: VehicleState): void {
  r.rect(v.x - 20, v.y - 12, 40, 24, "#607d8b");
  drawHealthBar(r, v.x, v.y - 24, v.health, 300); // TODO(owner): pull real max health from vehicles.Catalog equivalent on client
}

/** Request to mount the nearest vehicle (bind to a key, e.g. "F"). */
export function requestMount(socket: GameSocket, vehicleId: string): void {
  socket.send(MessageType.EnterVehicle, { vehicleId });
}

export function requestDismount(socket: GameSocket, vehicleId: string): void {
  socket.send(MessageType.ExitVehicle, { vehicleId });
}
