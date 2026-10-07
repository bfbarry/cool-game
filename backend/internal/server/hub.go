// Package server owns the websocket transport layer: accepting
// connections, tracking connected clients, and moving bytes in/out. It
// knows nothing about game rules — incoming messages are handed to a
// GameHandler, and outgoing messages go out through the Sender interface
// Hub implements. This keeps server/ and game/ decoupled (wired together
// in cmd/server/main.go) so each can be worked on independently.
package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/nir-game/server/internal/protocol"
)

// GameHandler is implemented by game.World. The hub calls into it on
// connect/message/disconnect; it never reaches back into transport details.
type GameHandler interface {
	HandleConnect(playerID string)
	HandleMessage(playerID string, raw protocol.RawEnvelope)
	HandleDisconnect(playerID string)
}

// Sender lets game code push messages out without knowing about websockets.
type Sender interface {
	SendTo(playerID string, env protocol.Envelope)
	Broadcast(env protocol.Envelope)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // keyed by playerID

	register   chan *Client
	unregister chan *Client

	handler GameHandler
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// SetHandler wires the hub to game logic. Call once at startup before Run.
func (h *Hub) SetHandler(handler GameHandler) {
	h.handler = handler
}

// Run processes register/unregister events. Call in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.playerID] = c
			h.mu.Unlock()
			if h.handler != nil {
				h.handler.HandleConnect(c.playerID)
			}
		case c := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, c.playerID)
			h.mu.Unlock()
			close(c.send)
			if h.handler != nil {
				h.handler.HandleDisconnect(c.playerID)
			}
		}
	}
}

// --- Sender implementation ---

func (h *Hub) SendTo(playerID string, env protocol.Envelope) {
	b, err := json.Marshal(env)
	if err != nil {
		log.Printf("hub: marshal error for %s: %v", env.Type, err)
		return
	}
	h.mu.RLock()
	c, ok := h.clients[playerID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case c.send <- b:
	default:
		log.Printf("hub: send buffer full for player %s, dropping %s", playerID, env.Type)
	}
}

func (h *Hub) Broadcast(env protocol.Envelope) {
	b, err := json.Marshal(env)
	if err != nil {
		log.Printf("hub: marshal error for %s: %v", env.Type, err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.send <- b:
		default:
			log.Printf("hub: send buffer full for player %s, dropping %s", c.playerID, env.Type)
		}
	}
}
