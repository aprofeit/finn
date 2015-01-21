package main

import "time"

type Bullet struct {
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Texture   string  `json:"texture"`
}

func NewBullet(x, y, xVel, yVel float64) *Bullet {
	return &Bullet{x, y, xVel, yVel, "sprites/bullet.png"}
}

func (b *Bullet) Update(elapsed time.Duration, world *World) {
	if tile := world.TileGrid[int(b.PositionX+b.VelocityX)][int(b.PositionY+b.VelocityY)]; tile.Kind != "wall" {
		b.PositionX += b.VelocityX
		b.PositionY += b.VelocityY
	} else {
		world.RemoveProjectile(b)
	}
}
