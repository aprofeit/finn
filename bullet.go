package main

import "time"

type Bullet struct {
	PositionX, PositionY, VelocityX, VelocityY float64
}

func NewBullet(x, y, xVel, yVel float64) *Bullet {
	return &Bullet{x, y, xVel, yVel}
}

func (b *Bullet) Update(elapsed time.Duration, world *World) {
	if tile := world.TileGrid[int(b.PositionX+b.VelocityX)][int(b.PositionY+b.VelocityY)]; tile.Kind != "wall" {
		b.PositionX += b.VelocityX
		b.PositionY += b.VelocityY
	} else {
		world.RemoveProjectile(b)
	}
}
