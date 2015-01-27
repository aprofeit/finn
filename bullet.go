package main

import "time"

type Bullet struct {
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Texture   string  `json:"texture"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	player    *Player
}

func NewBullet(x, y, xVel, yVel float64, p *Player) *Bullet {
	return &Bullet{x, y, xVel, yVel, "sprites/bullet.png", 0.1, 0.1, p}
}

func (b *Bullet) Update(elapsed time.Duration, world *World) {
	if tile := world.TileGrid[int(b.PositionX+b.VelocityX)][int(b.PositionY+b.VelocityY)]; tile.Kind != "wall" {
		b.PositionX += b.VelocityX
		b.PositionY += b.VelocityY

		for _, player := range world.Players {
			if player == b.player {
				continue
			}
			if (player.PositionX < b.PositionX && player.PositionX+player.Width > b.PositionX && player.PositionY < b.PositionY && player.PositionY+player.Height > b.PositionY) || (player.PositionX < b.PositionX+b.Width && player.PositionX+player.Width > b.PositionX+b.Width && player.PositionY < b.PositionY+b.Height && player.PositionY+player.Height > b.PositionY+b.Height) {
				player.Die(world)
				b.player.Score += 1
				world.RemoveProjectile(b)
			}
		}
	} else {
		world.RemoveProjectile(b)
	}
}
