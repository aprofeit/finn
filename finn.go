package main

import (
	"encoding/json"
	"net/http"

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

	player := &Player{
		ClientID:  conn.RemoteAddr().String(),
		PositionX: 10,
		PositionY: 10,
		AnchorX:   0.5,
		AnchorY:   0.5,
		Texture:   "south2.png",
		Direction: "",
	}

	h.World.Players = append(h.World.Players, player)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Infof("Client disconnected: ", err)
			return
		}
		event := &ClientEvent{ClientID: conn.RemoteAddr().String()}
		err = json.Unmarshal(p, event)
		if err != nil {
			log.Error(err)
		}
		h.ClientEvents <- event
		if err = conn.WriteMessage(messageType, p); err != nil {
			log.Error(err)
			return
		}
	}
}

type ClientEvent struct {
	ClientID string `json:"-"`
	Event    string `json:"event"`
	KeyCode  int    `json:"keycode"`
}

type Player struct {
	ClientID  string  `json:"id"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	AnchorX   float64 `json:"anchor_x"`
	AnchorY   float64 `json:"anchor_y"`
	Texture   string  `json:"texture"`
	Direction string  `json:"direction"`
}

type World struct {
	Players []*Player
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	clientEvents := make(chan *ClientEvent)
	world := &World{}

	http.Handle("/", http.FileServer(http.Dir("public/")))
	websocketHandler := &WebSocketHandler{
		Upgrader:     websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		ClientEvents: clientEvents,
		World:        world,
	}
	http.Handle("/websocket", websocketHandler)

	go func() {
		for {
			select {
			case event := <-clientEvents:
				for _, player := range world.Players {
					if player.ClientID == event.ClientID {
						if event.Event == "keydown" {
							switch event.KeyCode {
							case 37:
								player.Direction = "left"
							case 38:
								player.Direction = "up"
							case 39:
								player.Direction = "right"
							case 40:
								player.Direction = "down"
							}
						} else if event.Event == "keyup" {
							player.Direction = "none"
						}
					}
				}
			}
		}
	}()

	http.ListenAndServe(":3000", nil)
}
