package main

type Client struct {
	player     *Player
	remoteAddr string
}

func NewClient(player *Player, remoteAddr string) *Client {
	return &Client{
		player:     player,
		remoteAddr: remoteAddr,
	}
}

func (c *Client) Player() *Player {
	return c.player
}

func (c *Client) ID() string {
	return c.remoteAddr
}
