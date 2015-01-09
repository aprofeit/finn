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

	player := &Player{
		ClientID:  conn.RemoteAddr().String(),
		PositionX: 10,
		PositionY: 10,
		AnchorX:   0.5,
		AnchorY:   0.5,
		Texture:   "south2.png",
		Direction: NoDirectionLabel,
	}

	h.World.Players = append(h.World.Players, player)

	go func() {
		for {
			world := <-h.WorldUpdates

			blob, err := json.Marshal(world)
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

type Player struct {
	ClientID  string  `json:"id"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	AnchorX   float64 `json:"anchor_x"`
	AnchorY   float64 `json:"anchor_y"`
	Texture   string  `json:"texture"`
	Direction string  `json:"direction"`
}

const NoDirectionLabel string = "none"
const WalkRate float64 = 3

type World struct {
	Players []*Player `json:"members"`
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func (p *Player) Update(elapsed time.Duration) {
	switch p.Direction {
	case "up":
		p.PositionY -= WalkRate
	case "down":
		p.PositionY += WalkRate
	case "left":
		p.PositionX -= WalkRate
	case "right":
		p.PositionX += WalkRate
	}
}

func main() {
	clientEvents := make(chan *ClientEvent)
	world := &World{}
	updates := make(chan *World)

	http.Handle("/", http.FileServer(http.Dir("public/")))
	websocketHandler := &WebSocketHandler{
		Upgrader:     websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		ClientEvents: clientEvents,
		World:        world,
		WorldUpdates: updates,
	}
	http.Handle("/websocket", websocketHandler)

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
								player.Direction = "left"
							case 38:
								player.Direction = "up"
							case 39:
								player.Direction = "right"
							case 40:
								player.Direction = "down"
							}
						} else if event.Event == "keyup" {
							player.Direction = NoDirectionLabel
						}
					}
				}
			}
		}
	}()

	log.Info("Listening on 3000")
	http.ListenAndServe(":3000", nil)
}
