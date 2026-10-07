// Thin WebSocket client: connects, sends typed envelopes, and dispatches
// incoming envelopes to per-type listeners. Game-specific reactions to
// messages (update local world model, etc.) belong in game/world.ts, not
// here — this file should stay transport-only.
import type { Envelope } from "./protocol";

type Listener = (data: unknown) => void;

export class GameSocket {
  private ws: WebSocket | null = null;
  private listeners = new Map<string, Set<Listener>>();
  private url: string;

  constructor(url = GameSocket.defaultUrl()) {
    this.url = url;
  }

  static defaultUrl(): string {
    const proto = window.location.protocol === "https:" ? "wss" : "ws";
    return `${proto}://${window.location.host}/ws`;
  }

  connect(): void {
    this.ws = new WebSocket(this.url);
    this.ws.addEventListener("open", () => this.emit("_open", undefined));
    this.ws.addEventListener("close", () => this.emit("_close", undefined));
    this.ws.addEventListener("error", (e) => this.emit("_error", e));
    this.ws.addEventListener("message", (event) => {
      let envelope: Envelope;
      try {
        envelope = JSON.parse(event.data as string);
      } catch (err) {
        console.error("socket: bad message", err);
        return;
      }
      this.emit(envelope.type, envelope.data);
    });
  }

  /** Register a handler for a given server->client message type (see protocol.MessageType). */
  on(type: string, listener: Listener): void {
    if (!this.listeners.has(type)) this.listeners.set(type, new Set());
    this.listeners.get(type)!.add(listener);
  }

  off(type: string, listener: Listener): void {
    this.listeners.get(type)?.delete(listener);
  }

  private emit(type: string, data: unknown): void {
    this.listeners.get(type)?.forEach((l) => l(data));
  }

  /** Send a client->server message (see protocol.MessageType for valid types). */
  send(type: string, data: unknown): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn(`socket: dropped "${type}", not connected`);
      return;
    }
    this.ws.send(JSON.stringify({ type, data } satisfies Envelope));
  }
}
