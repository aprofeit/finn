package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	websocket.Upgrader
	ClientEvents chan *ClientEvent
	World        *World
	WorldUpdates chan *World
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
	}
	log.Infof("Client connected %v", conn.RemoteAddr().String())

	playerID := conn.RemoteAddr().String()
	h.World.AddPlayer(playerID)

	go func() {
		for {
			world := <-h.WorldUpdates

			blob, err := world.MarshalMembers()
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

const WalkRate float64 = 5

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	clientEvents := make(chan *ClientEvent)
	world := NewWorld()
	world.Generate()
	updates := make(chan *World)

	http.Handle("/", http.FileServer(http.Dir("public/")))
	websocketHandler := &WebSocketHandler{
		Upgrader:     websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		ClientEvents: clientEvents,
		World:        world,
		WorldUpdates: updates,
	}
	http.Handle("/websocket", websocketHandler)
	http.HandleFunc("/tiles", func(w http.ResponseWriter, r *http.Request) {
		blob, err := world.MarshalTiles()
		if err != nil {
			log.Error("marshaling tiles: %v", err)
		}
		w.Write(blob)
	})
	http.HandleFunc("/world", func(w http.ResponseWriter, r *http.Request) {
		for y := 0; y < WORLD_HEIGHT; y++ {
			for x := 0; x < WORLD_WIDTH; x++ {
				switch world.TileGrid[x][y].Kind {
				case "wall":
					w.Write([]byte("X"))
				case "floor":
					w.Write([]byte(" "))
				}
			}
			w.Write([]byte("\n"))
		}
	})

	go func() {
		for now := range time.Tick(time.Second / 30) {
			last := time.Now()
			for _, player := range world.Players {
				player.Update(time.Since(last))
				last = now

				updates <- world
			}
		}
	}()

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
							}
						}
					}
				}
			}
		}
	}()

	log.Info("Listening on 3000")
	http.ListenAndServe(":3000", nil)
}
