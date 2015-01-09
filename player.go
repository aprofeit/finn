package main

import "time"

type Player struct {
	ClientID  string  `json:"id"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	AnchorX   float64 `json:"anchor_x"`
	AnchorY   float64 `json:"anchor_y"`
	Texture   string  `json:"texture"`
	Direction string  `json:"direction"`
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
