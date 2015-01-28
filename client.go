package main

type Client struct {
	player     *Player
	remoteAddr string
	Score      int
	HighScore  int
}

func NewClient(player *Player, remoteAddr string) *Client {
	client := &Client{
		player:     player,
		remoteAddr: remoteAddr,
		Score:      0,
	}
	player.client = client
	return client
}

func (c *Client) Player() *Player {
	return c.player
}

func (c *Client) ID() string {
	return c.remoteAddr
}
