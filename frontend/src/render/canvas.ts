// Thin wrapper around a 2D canvas context: resize handling + a handful of
// drawing primitives other render modules (player/combat/npc/vehicles/world)
// build on. Keep camera/viewport logic here so everyone draws in the same
// coordinate space.
export class Renderer {
  readonly canvas: HTMLCanvasElement;
  readonly ctx: CanvasRenderingContext2D;

  // Camera follows the local player; world-space coordinates get translated
  // by this offset before drawing. See CONTEXT.md for world size.
  camera = { x: 0, y: 0 };

  constructor(canvas: HTMLCanvasElement) {
    this.canvas = canvas;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("render: 2d context unavailable");
    this.ctx = ctx;
    window.addEventListener("resize", () => this.resize());
    this.resize();
  }

  private resize(): void {
    this.canvas.width = window.innerWidth;
    this.canvas.height = window.innerHeight;
  }

  clear(color = "#1a1a2e"): void {
    this.ctx.fillStyle = color;
    this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
  }

  /** Centers the camera on a world-space point (typically the local player). */
  centerOn(x: number, y: number): void {
    this.camera.x = x - this.canvas.width / 2;
    this.camera.y = y - this.canvas.height / 2;
  }

  toScreen(x: number, y: number): [number, number] {
    return [x - this.camera.x, y - this.camera.y];
  }

  circle(x: number, y: number, radius: number, color: string): void {
    const [sx, sy] = this.toScreen(x, y);
    this.ctx.beginPath();
    this.ctx.arc(sx, sy, radius, 0, Math.PI * 2);
    this.ctx.fillStyle = color;
    this.ctx.fill();
  }

  rect(x: number, y: number, w: number, h: number, color: string): void {
    const [sx, sy] = this.toScreen(x, y);
    this.ctx.fillStyle = color;
    this.ctx.fillRect(sx, sy, w, h);
  }

  line(x1: number, y1: number, x2: number, y2: number, color: string, width = 1): void {
    const [sx1, sy1] = this.toScreen(x1, y1);
    const [sx2, sy2] = this.toScreen(x2, y2);
    this.ctx.strokeStyle = color;
    this.ctx.lineWidth = width;
    this.ctx.beginPath();
    this.ctx.moveTo(sx1, sy1);
    this.ctx.lineTo(sx2, sy2);
    this.ctx.stroke();
  }

  text(x: number, y: number, text: string, color = "#fff", font = "12px sans-serif"): void {
    const [sx, sy] = this.toScreen(x, y);
    this.ctx.fillStyle = color;
    this.ctx.font = font;
    this.ctx.textAlign = "center";
    this.ctx.fillText(text, sx, sy);
  }
}
