package main

type Client struct {
	player *Player
}

func NewClient(player *Player) *Client {
	return &Client{
		player: player,
	}
}

func (c *Client) Player() *Player {
	return c.player
}
