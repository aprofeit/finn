package main

import "time"

type Player struct {
	Z             int     `json:"z"`
	ClientID      string  `json:"id"`
	PositionX     float64 `json:"position_x"`
	PositionY     float64 `json:"position_y"`
	AnchorX       float64 `json:"anchor_x"`
	AnchorY       float64 `json:"anchor_y"`
	Texture       string  `json:"texture"`
	Direction     string  `json:"direction"`
	MovingUp      bool    `json:"-"`
	MovingDown    bool    `json:"-"`
	MovingLeft    bool    `json:"-"`
	MovingRight   bool    `json:"-"`
	Width         float64 `json:"width"`
	Height        float64 `json:"height"`
	hasShot       bool    `json:"-"`
	lastDirection string  `json:"-"`
	Dead          bool    `json:"dead"`
}

const WALK_RATE float64 = 0.15
const BULLET_SPEED float64 = 0.25

func (p *Player) Die(w *World) {
	p.Dead = true
}

func (p *Player) Facing() string {
	if p.Direction != "none" {
		return p.Direction
	} else {
		return p.lastDirection
	}
}

func (p *Player) StartShot(w *World) {
	if p.hasShot || p.Dead {
		return
	} else {
		p.hasShot = true
		var bullet *Bullet
		switch p.Facing() {
		case "up":
			bullet = NewBullet(p.PositionX+p.Width/2-0.05, p.PositionY, 0, -BULLET_SPEED, p)
		case "down":
			bullet = NewBullet(p.PositionX+p.Width/2-0.05, p.PositionY, 0, BULLET_SPEED, p)
		case "left":
			bullet = NewBullet(p.PositionX, p.PositionY+p.Height/2-0.05, -BULLET_SPEED, 0, p)
		case "right":
			bullet = NewBullet(p.PositionX, p.PositionY+p.Height/2-0.05, BULLET_SPEED, 0, p)
		}
		w.AddProjectile(bullet)
	}
}

func (p *Player) EndShot() {
	p.hasShot = false
}

func (p *Player) Update(elapsed time.Duration, world *World) {
	if p.Dead {
		return
	}
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
		if p.Direction != "none" {
			p.lastDirection = p.Direction
		}
		p.Direction = "none"
	}
}
