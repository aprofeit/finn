package main

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	websocket.Upgrader
	clientEvents chan *ClientEvent
}

func (h WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}
		var event ClientEvent
		err = json.Unmarshal(p, &event)
		if err != nil {
			log.Error(err)
		}
		event.ClientID = conn.RemoteAddr().String()
		h.clientEvents <- &event
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

func main() {
	clientEvents := make(chan *ClientEvent)

	http.Handle("/", http.FileServer(http.Dir("public/")))
	websocketHandler := WebSocketHandler{
		Upgrader:     websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		clientEvents: clientEvents,
	}
	http.Handle("/websocket", websocketHandler)

	go func() {
		for {
			select {
			case event := <-clientEvents:
				log.Infof("received client event: %+v", event)
			}
		}
	}()

	http.ListenAndServe(":3000", nil)
}
