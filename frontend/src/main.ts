// Entry point: customization screen -> connect -> join -> start loop.
// Each piece this wires together lives in its own module per CONTEXT.md §6
// so contributors can work on one without touching this file much.
import { GameSocket } from "./net/socket";
import { ClientWorld } from "./game/world";
import { GameLoop } from "./game/loop";
import { Renderer } from "./render/canvas";
import { InputState, bindInput } from "./player/player";
import { showCustomizationScreen } from "./player/customization";
import { StoreUI } from "./store/store";
import { MessageType, type JoinData, type WelcomeData } from "./net/protocol";

function main(): void {
  const canvas = document.getElementById("game-canvas") as HTMLCanvasElement;
  const renderer = new Renderer(canvas);
  const input = new InputState();
  bindInput(canvas, input);

  showCustomizationScreen((join: JoinData) => {
    const socket = new GameSocket();
    const world = new ClientWorld(socket);

    socket.on(MessageType.Welcome, (data) => {
      new StoreUI(socket, (data as WelcomeData).storeCatalog);
    });

    socket.on("_open", () => socket.send(MessageType.Join, join));

    socket.connect();

    const loop = new GameLoop(socket, world, renderer, input);
    loop.start();
  });
}

main();
