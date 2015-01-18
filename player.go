package main

import "time"

type Player struct {
	ClientID    string  `json:"id"`
	PositionX   float64 `json:"position_x"`
	PositionY   float64 `json:"position_y"`
	AnchorX     float64 `json:"anchor_x"`
	AnchorY     float64 `json:"anchor_y"`
	Texture     string  `json:"texture"`
	Direction   string  `json:"direction"`
	MovingUp    bool
	MovingDown  bool
	MovingLeft  bool
	MovingRight bool
}

func (p *Player) Update(elapsed time.Duration) {
	if p.MovingUp {
		p.Direction = "up"
		p.PositionY -= WalkRate
	}
	if p.MovingDown {
		p.Direction = "down"
		p.PositionY += WalkRate
	}
	if p.MovingLeft {
		p.Direction = "left"
		p.PositionX -= WalkRate
	}
	if p.MovingRight {
		p.Direction = "right"
		p.PositionX += WalkRate
	}
	if !p.MovingRight && !p.MovingLeft && !p.MovingUp && !p.MovingDown {
		p.Direction = "none"
	}
}
