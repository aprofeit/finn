package main

type World struct {
	Players []*Player `json:"members"`
}

func (w *World) AddPlayer(id string) {
	player := &Player{
		ClientID:  id,
		PositionX: 10,
		PositionY: 10,
		AnchorX:   0.5,
		AnchorY:   0.5,
		Texture:   "sprites/south2.png",
		Direction: NoDirectionLabel,
	}

	w.Players = append(w.Players, player)
}

func (w *World) RemovePlayer(id string) {
	for i, player := range w.Players {
		if player.ClientID == id {
			w.Players = append(w.Players[:i], w.Players[i+1:]...)
			return
		}
	}
}
