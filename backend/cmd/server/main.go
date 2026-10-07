// Command server is the entrypoint: wires together the websocket hub
// (internal/server) and the authoritative game world (internal/game), then
// serves HTTP. See CONTEXT.md for the overall architecture.
package main

import (
	"log"
	"net/http"

	"github.com/nir-game/server/internal/game"
	"github.com/nir-game/server/internal/server"
)

func main() {
	hub := server.NewHub()
	world := game.NewWorld(hub)
	hub.SetHandler(world)

	go hub.Run()
	go world.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWS(hub, w, r)
	})

	addr := ":8080"
	log.Printf("nir-game server listening on %s (ws endpoint: /ws)", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
