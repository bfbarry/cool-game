package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nir-game/server/internal/protocol"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// TODO(server owner): restrict CheckOrigin before deploying beyond local dev.
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	sendBuffer = 64
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	playerID string
}

// ServeWS upgrades an HTTP request to a websocket connection, registers a
// new Client with the hub, and starts its pumps. Mount at /ws (see
// cmd/server/main.go).
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("server: upgrade failed: %v", err)
		return
	}
	c := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBuffer),
		playerID: newPlayerID(),
	}
	hub.register <- c

	go c.writePump()
	go c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(8192)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("server: read error for %s: %v", c.playerID, err)
			}
			return
		}
		var raw protocol.RawEnvelope
		if err := json.Unmarshal(msg, &raw); err != nil {
			log.Printf("server: bad envelope from %s: %v", c.playerID, err)
			continue
		}
		if c.hub.handler != nil {
			c.hub.handler.HandleMessage(c.playerID, raw)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
