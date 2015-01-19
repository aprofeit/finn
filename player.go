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
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}

const WALK_RATE float64 = 0.16

func (p *Player) Update(elapsed time.Duration, world *World) {
	if p.MovingUp {
		p.Direction = "up"
		if tile := world.TileGrid[int(p.PositionX)][int(p.PositionY-WALK_RATE)]; tile.Kind != "wall" {
			p.PositionY -= WALK_RATE
		} else {
			p.PositionY = float64(int(p.PositionY))
		}
	}
	if p.MovingDown {
		p.Direction = "down"
		if tile := world.TileGrid[int(p.PositionX)][int(p.PositionY+WALK_RATE+p.Height)]; tile.Kind != "wall" {
			p.PositionY += WALK_RATE
		} else {
			p.PositionY = float64(int(p.PositionY+WALK_RATE+p.Height)) - p.Height
		}
	}
	if p.MovingLeft {
		p.Direction = "left"
		if tile := world.TileGrid[int(p.PositionX-WALK_RATE)][int(p.PositionY)]; tile.Kind != "wall" {
			p.PositionX -= WALK_RATE
		} else {
			p.PositionX = float64(int(p.PositionX))
		}
	}
	if p.MovingRight {
		p.Direction = "right"
		if tile := world.TileGrid[int(p.PositionX+WALK_RATE+p.Width)][int(p.PositionY)]; tile.Kind != "wall" {
			p.PositionX += WALK_RATE
		} else {
			p.PositionX = float64(int(p.PositionX+WALK_RATE+p.Width)) - p.Width
		}
	}
	if !p.MovingRight && !p.MovingLeft && !p.MovingUp && !p.MovingDown {
		p.Direction = "none"
	}
}
