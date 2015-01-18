package main

import (
	"math"
	"math/rand"
)

type Tile struct {
	Kind string
}

type World struct {
	Players  []*Player `json:"members"`
	TileGrid [][]*Tile
	Rooms    []*Rect
}

const WORLD_WIDTH int = 201
const WORLD_HEIGHT int = 61

func NewWorld() *World {
	cols := make([][]*Tile, WORLD_WIDTH)
	for x := range cols {
		cols[x] = make([]*Tile, WORLD_HEIGHT)

		for y := range cols[x] {
			cols[x][y] = &Tile{}
		}
	}

	return &World{
		TileGrid: cols,
		Rooms:    []*Rect{},
	}
}

func (w *World) AddPlayer(id string) {
	player := &Player{
		ClientID:  id,
		PositionX: 10,
		PositionY: 10,
		AnchorX:   0.5,
		AnchorY:   0.5,
		Texture:   "sprites/south2.png",
		Direction: "none",
	}

	w.Players = append(w.Players, player)
}

func (w *World) RemovePlayer(id string) {
	for i, player := range w.Players {
		if player.ClientID == id {
			w.Players = append(w.Players[:i], w.Players[i+1:]...)
			return
		}
	}
}

func (w *World) Generate() {
	w.fillAll("wall")
	w.addRooms()
}

const NUM_ROOM_ATTEMPTS int = 1000
const ROOM_EXTRA_SIZE int = 3

func (w *World) addRooms() {
	for i := 0; i < NUM_ROOM_ATTEMPTS; i++ {
		size := ((rand.Intn(3+ROOM_EXTRA_SIZE) + 1) * 2) + 1
		rectangularity := rand.Intn(1+size/2) * 2
		width := size
		height := size

		if rand.Intn(2) == 1 {
			width += rectangularity
		} else {
			height += rectangularity
		}

		x := rand.Intn((WORLD_WIDTH-width)/2)*2 + 1
		y := rand.Intn((WORLD_HEIGHT-height)/2)*2 + 1

		room := NewRect(x, y, width, height)
		overlaps := false
		for _, otherRoom := range w.Rooms {
			if room.DistanceTo(otherRoom) <= 0 {
				overlaps = true
				break
			}
		}

		if overlaps {
			continue
		}

		w.Rooms = append(w.Rooms, room)
		w.CarveRect(room, "floor")
	}
}

func (w *World) CarveRect(r *Rect, kind string) {
	for _, coord := range r.Coords() {
		w.TileGrid[coord.X][coord.Y].Kind = kind
	}
}

func (w *World) fillAll(kind string) {
	for x := range w.TileGrid {
		for _, tile := range w.TileGrid[x] {
			tile.Kind = "wall"
		}
	}
}

type Coordinate struct {
	X int
	Y int
}

type Edge struct {
	P1 *Coordinate
	P2 *Coordinate
}

func distanceBetween(x0, y0, x1, y1, x2, y2 int) float64 {
	return math.Abs(float64((y2-y1)*x0-(x2-x1)*y0+x2*y1-x2*x1)) / math.Sqrt(math.Pow(float64(y2-y1), 2)+math.Pow(float64(x2-x1), 2))
}

func (e *Edge) DistanceTo(other *Edge) float64 {
	shortestDistance := 1000000.0
	cases := []struct{ x0, y0, x1, y1, x2, y2 int }{
		{e.P1.X, e.P1.Y, other.P1.X, other.P1.Y, other.P2.X, other.P2.Y},
		{e.P2.X, e.P2.Y, other.P1.X, other.P1.Y, other.P2.X, other.P2.Y},
		{other.P1.X, other.P1.Y, e.P1.X, e.P1.Y, e.P2.X, e.P2.Y},
		{other.P2.X, other.P2.Y, e.P1.X, e.P1.Y, e.P2.X, e.P2.Y},
	}
	for _, c := range cases {
		if distance := distanceBetween(c.x0, c.y0, c.x1, c.y1, c.x2, c.y2); distance < shortestDistance {
			shortestDistance = distance
		}
	}

	return shortestDistance
}

func (r *Rect) DistanceTo(other *Rect) float64 {
	if r.Overlaps(other) {
		return 0
	}

	shortestDistance := 1000000.0
	for _, edge := range r.Edges() {
		for _, otherEdge := range other.Edges() {
			if distance := edge.DistanceTo(otherEdge); distance < shortestDistance {
				shortestDistance = distance
			}
		}
	}

	return shortestDistance
}

func (r *Rect) Overlaps(other *Rect) bool {
	return r.X1() <= other.X2() && r.X2() >= other.X1() && r.Y1() <= other.Y2() && r.Y2() >= other.Y1()
}

func (r *Rect) Edges() []*Edge {
	return []*Edge{
		&Edge{&Coordinate{r.X1(), r.Y1()}, &Coordinate{r.X2(), r.Y1()}},
		&Edge{&Coordinate{r.X1(), r.Y1()}, &Coordinate{r.X1(), r.Y2()}},
		&Edge{&Coordinate{r.X1(), r.Y2()}, &Coordinate{r.X2(), r.Y2()}},
		&Edge{&Coordinate{r.X2(), r.Y1()}, &Coordinate{r.X2(), r.Y2()}},
	}
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r *Rect) Coords() []*Coordinate {
	coords := []*Coordinate{}
	for x := r.X1(); x <= r.X2(); x++ {
		for y := r.Y1(); y <= r.Y2(); y++ {
			coords = append(coords, &Coordinate{x, y})
		}
	}
	return coords
}

func (r *Rect) X1() int {
	return r.X
}

func (r *Rect) X2() int {
	return r.X + r.Width - 1
}

func (r *Rect) Y1() int {
	return r.Y
}

func (r *Rect) Y2() int {
	return r.Y + r.Height - 1
}

func NewRect(x int, y int, width int, height int) *Rect {
	return &Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}
