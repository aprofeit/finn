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
	world         *World  `json:"-"`
	Score         int     `json:"score"`
}

type FloatCoordinate struct {
	X, Y float64
}

const WALK_RATE float64 = 0.15
const BULLET_SPEED float64 = 0.25

func NewPlayer(id string, x, y float64) *Player {
	return &Player{
		Z:             1,
		ClientID:      id,
		PositionX:     x,
		PositionY:     y,
		AnchorX:       0.5,
		AnchorY:       0.25,
		Texture:       "sprites/south2.png",
		Direction:     "none",
		Width:         0.4,
		Height:        0.3,
		lastDirection: "down",
	}
}

func (p *Player) Die(w *World) {
	p.Score = 0
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

func (p *Player) corners() []*FloatCoordinate {
	return []*FloatCoordinate{
		&FloatCoordinate{p.PositionX, p.PositionY},
		&FloatCoordinate{p.PositionX, p.PositionY + p.Height},
		&FloatCoordinate{p.PositionX + p.Width, p.PositionY},
		&FloatCoordinate{p.PositionX + p.Width, p.PositionY + p.Height},
	}
}

func (p *Player) collidesAt(x, y float64) bool {
	collides := false
	for _, coord := range p.corners() {
		tile := p.world.TileGrid[int(coord.X+x)][int(coord.Y+y)]
		if tile.Kind == "wall" {
			collides = true
		}
	}

	return collides
}

func (p *Player) Update(elapsed time.Duration) {
	if p.Dead {
		return
	}
	if p.MovingUp {
		p.Direction = "up"
		if p.collidesAt(0, -WALK_RATE) {
			p.PositionY = float64(int(p.PositionY)) + 0.001
		} else {
			p.PositionY -= WALK_RATE
		}
	}
	if p.MovingDown {
		p.Direction = "down"
		if p.collidesAt(0, WALK_RATE) {
			p.PositionY = float64(int(p.PositionY+WALK_RATE+p.Height)) - p.Height - 0.001
		} else {
			p.PositionY += WALK_RATE
		}
	}
	if p.MovingLeft {
		p.Direction = "left"
		if p.collidesAt(-WALK_RATE, 0) {
			p.PositionX = float64(int(p.PositionX)) + 0.001
		} else {
			p.PositionX -= WALK_RATE
		}
	}
	if p.MovingRight {
		p.Direction = "right"
		if p.collidesAt(WALK_RATE, 0) {
			p.PositionX = float64(int(p.PositionX+WALK_RATE+p.Width)) - p.Width - 0.001
		} else {
			p.PositionX += WALK_RATE
		}
	}
	if !p.MovingRight && !p.MovingLeft && !p.MovingUp && !p.MovingDown {
		if p.Direction != "none" {
			p.lastDirection = p.Direction
		}
		p.Direction = "none"
	}
}
