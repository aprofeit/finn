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
	}
	if p.MovingDown {
		p.Direction = "down"
	}
	if p.MovingLeft {
		p.Direction = "left"
	}
	if p.MovingRight {
		p.Direction = "right"
	}
	if !p.MovingRight && !p.MovingLeft && !p.MovingUp && !p.MovingDown {
		p.Direction = "none"
	}
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
