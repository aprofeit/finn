package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	websocket.Upgrader
	ClientEvents chan *ClientEvent
	World        *World
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
	}
	log.Infof("Client connected %v", conn.RemoteAddr().String())

	playerID := conn.RemoteAddr().String()
	openTile := h.World.getSpawn()
	player := NewPlayer(playerID, float64(openTile.X), float64(openTile.Y))
	updater := h.World.AddPlayer(player)

	go func() {
		for {
			world := <-updater.c

			var current *Player
			world.Lock()
			for _, player := range world.Players {
				if player.ClientID == playerID {
					current = player
				}
			}
			world.Unlock()
			if current == nil {
				return
			}
			blob, err := world.MarshalMembers(current)
			if err != nil {
				log.Errorf("Marshaling world update %v", world)
			}

			if err = conn.WriteMessage(websocket.TextMessage, blob); err != nil {
				log.Errorf("Writing update to client: %v", err)
				return
			}
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Infof("Client disconnected %v", conn.RemoteAddr().String())
			h.World.RemovePlayer(playerID)
			return
		}
		event := &ClientEvent{ClientID: conn.RemoteAddr().String()}
		err = json.Unmarshal(p, event)
		if err != nil {
			log.Errorf("Unmarshaling client event json: %v", err)
		}
		h.ClientEvents <- event
	}
}

type ClientEvent struct {
	ClientID string `json:"-"`
	Event    string `json:"event"`
	KeyCode  int    `json:"keycode"`
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	clientEvents := make(chan *ClientEvent)
	world := NewWorld()
	world.Generate()
	world.Print(os.Stdout)

	http.Handle("/", http.FileServer(http.Dir("public/")))
	websocketHandler := &WebSocketHandler{
		Upgrader:     websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		ClientEvents: clientEvents,
		World:        world,
	}
	http.Handle("/websocket", websocketHandler)
	http.HandleFunc("/tiles", func(w http.ResponseWriter, r *http.Request) {
		blob, err := world.MarshalTiles()
		if err != nil {
			log.Error("marshaling tiles: %v", err)
		}
		w.Write(blob)
	})
	go world.Update()

	go func() {
		for {
			select {
			case event := <-clientEvents:
				for _, player := range world.Players {
					if player.ClientID == event.ClientID {
						if event.Event == "keydown" {
							switch event.KeyCode {
							case 37:
								player.MovingLeft = true
							case 38:
								player.MovingUp = true
							case 39:
								player.MovingRight = true
							case 40:
								player.MovingDown = true
							case 32:
								player.StartShot(world)
							default:
								log.Debugf("Unknown key pressed: %v", event.KeyCode)
							}
						}
						if event.Event == "keyup" {
							switch event.KeyCode {
							case 37:
								player.MovingLeft = false
							case 38:
								player.MovingUp = false
							case 39:
								player.MovingRight = false
							case 40:
								player.MovingDown = false
							case 32:
								player.EndShot()
							}
						}
					}
				}
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Infof("Listening on %v", port)
	http.ListenAndServe(":"+port, nil)
}
