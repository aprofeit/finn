package main

import (
	"bufio"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Tile struct {
	Kind   string `json:"kind"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Z      int    `json:"z"`
	region int    `json:"-"`
}

type playerUpdater struct {
	c  chan *World
	id string
}

type World struct {
	Players        []*Player
	TileGrid       [][]*Tile
	Rooms          []*Rect
	currentRegion  int
	projectiles    []*Bullet
	playerUpdaters []*playerUpdater
	connectors     map[*Tile]map[int]bool
	sync.Mutex
}

func (w *World) floorTiles() []*Tile {
	floorTiles := []*Tile{}
	for _, tile := range w.Tiles() {
		if tile.Kind == "floor" {
			floorTiles = append(floorTiles, tile)
		}
	}
	return floorTiles
}

func (w *World) getSpawn() *Tile {
	tiles := w.floorTiles()

	return tiles[rand.Intn(len(tiles))]
}

func (w *World) Update() {
	for now := range time.Tick(time.Second / 30) {
		last := time.Now()
		w.Lock()
		for _, player := range w.Players {
			player.Update(time.Since(last))
		}
		for _, bullet := range w.projectiles {
			bullet.Update(time.Since(last), w)
		}
		w.Unlock()
		last = now
		for _, updater := range w.playerUpdaters {
			updater.c <- w
		}
	}
}

func (w *World) AddProjectile(b *Bullet) {
	w.projectiles = append(w.projectiles, b)
}

func (w *World) RemoveProjectile(b *Bullet) {
	for i, bullet := range w.projectiles {
		if bullet == b {
			w.projectiles = append(w.projectiles[:i], w.projectiles[i+1:]...)
		}
		break
	}
}

func (w *World) Tiles() []*Tile {
	tiles := []*Tile{}
	for x := range w.TileGrid {
		for _, tile := range w.TileGrid[x] {
			tiles = append(tiles, tile)
		}
	}
	return tiles
}

func (w *World) MarshalMembers(current *Player) ([]byte, error) {
	otherPlayers := []*Player{}
	for _, player := range w.Players {
		if player.ClientID != current.ClientID {
			otherPlayers = append(otherPlayers, player)
		}
	}
	return json.Marshal(map[string]interface{}{
		"members":     otherPlayers,
		"projectiles": w.projectiles,
		"current":     current,
	})
}

func (w *World) MarshalTiles() ([]byte, error) {
	return json.Marshal(w.Tiles())
}

const WORLD_WIDTH int = 27
const WORLD_HEIGHT int = 19

func NewWorld() *World {
	cols := make([][]*Tile, WORLD_WIDTH)
	for x := range cols {
		cols[x] = make([]*Tile, WORLD_HEIGHT)

		for y := range cols[x] {
			cols[x][y] = &Tile{X: x, Y: y, Z: 0}
		}
	}

	return &World{
		TileGrid: cols,
		Rooms:    []*Rect{},
	}
}

func (w *World) AddPlayer(player *Player) *playerUpdater {
	w.Lock()
	defer w.Unlock()
	player.world = w
	w.Players = append(w.Players, player)
	updater := &playerUpdater{c: make(chan *World), id: player.ClientID}
	w.playerUpdaters = append(w.playerUpdaters, updater)
	return updater
}

func (w *World) RemovePlayer(id string) {
	for i, player := range w.Players {
		if player.ClientID == id {
			w.Players = append(w.Players[:i], w.Players[i+1:]...)

			for j, updater := range w.playerUpdaters {
				if updater.id == id {
					w.playerUpdaters = append(w.playerUpdaters[:j], w.playerUpdaters[j+1:]...)
				}
			}
			return
		}
	}
}

func (w *World) Generate() {
	w.fillAll("wall")
	w.addRooms()

	for x := 1; x < WORLD_WIDTH; x += 2 {
		for y := 1; y < WORLD_HEIGHT; y += 2 {
			tile := w.TileGrid[x][y]
			if tile.Kind != "wall" {
				continue
			}
			w.GrowMaze(x, y)
		}
	}

	w.connectRegions()
	w.removeDeadEnds()
}

func (w *World) Print(writer io.Writer) {
	bufWriter := bufio.NewWriter(writer)
	bufWriter.Write([]byte("\033[2J"))
	for y := 0; y < WORLD_HEIGHT; y++ {
		for x := 0; x < WORLD_WIDTH; x++ {
			switch tile := w.TileGrid[x][y]; tile.Kind {
			case "wall":
				bufWriter.Write([]byte("▓▓"))
			case "floor":
				if tile.region == 0 {
					bufWriter.Write([]byte("░░"))
				} else {
					bufWriter.Write([]byte("  "))
				}
			}
		}
		bufWriter.Write([]byte("\n"))
	}
	bufWriter.Flush()
}

func (w *World) connectRegions() {
	w.connectors = make(map[*Tile]map[int]bool)
	for x := range w.TileGrid {
		for y, tile := range w.TileGrid[x] {
			if x == 0 || x == len(w.TileGrid)-1 || y == 0 || y == len(w.TileGrid[x])-1 {
				continue
			}

			if tile.Kind != "wall" {
				continue
			}

			regions := make(map[int]bool)
			for _, adjs := range []struct{ x, y int }{
				{x, y + 1},
				{x, y - 1},
				{x + 1, y},
				{x - 1, y},
			} {
				adjTile := w.TileGrid[adjs.x][adjs.y]
				if adjTile.Kind != "floor" {
					continue
				}
				regions[adjTile.region] = true
			}

			if len(regions) < 2 {
				continue
			}

			w.connectors[tile] = regions
		}
	}

	connectors := []*Tile{}
	for key, _ := range w.connectors {
		connectors = append(connectors, key)
	}

	merged := make(map[int]int)
	openRegions := make(map[int]bool)
	for i := 1; i <= w.currentRegion; i++ {
		merged[i] = i
		openRegions[i] = true
	}

	for len(openRegions) > 1 {
		connector := connectors[rand.Intn(len(connectors))]
		w.addJunction(connector)

		regions := make(map[int]bool)
		for region := range w.connectors[connector] {
			regions[merged[region]] = true
		}
		var dest int
		var sources []int
		i := 0
		for region := range regions {
			if i == 0 {
				dest = region
			} else {
				sources = append(sources, region)
			}
			i++
		}

		for i := range merged {
			for _, source := range sources {
				if source == merged[i] {
					merged[i] = dest
					for _, tile := range w.Tiles() {
						if tile.region == source {
							tile.region = dest
						}
					}
				}
			}
		}

		for _, source := range sources {
			delete(openRegions, source)
		}

		for i, c := range connectors {
			if c.X == connector.X && c.Y == connector.Y+1 {
				connectors = append(connectors[:i], connectors[i+1:]...)
				w.removeConnector(c)
				continue
			} else if c.X == connector.X && c.Y == connector.Y-1 {
				connectors = append(connectors[:i], connectors[i+1:]...)
				w.removeConnector(c)
				continue
			} else if c.X+1 == connector.X && c.Y == connector.Y {
				connectors = append(connectors[:i], connectors[i+1:]...)
				w.removeConnector(c)
				continue
			} else if c.X-1 == connector.X && c.Y == connector.Y {
				connectors = append(connectors[:i], connectors[i+1:]...)
				w.removeConnector(c)
				continue
			}

			regions := map[int]bool{}
			for region := range w.connectors[c] {
				regions[merged[region]] = true
			}
			if len(regions) > 1 {
				continue
			}

			if i < len(connectors) {
				head := connectors[:i]
				tail := connectors[i+1:]
				connectors = append(head, tail...)
				w.removeConnector(c)
			}
		}
	}
}

func (w *World) removeConnector(t *Tile) {
	delete(w.connectors, t)
}

func (w *World) addJunction(tile *Tile) {
	tile.Kind = "floor"
}

func (w *World) removeDeadEnds() {
	done := false
	for !done {
		done = true

		for _, tile := range w.Tiles() {
			if tile.X == 0 || tile.X == len(w.TileGrid)-1 || tile.Y == 0 || tile.Y == len(w.TileGrid[tile.X])-1 {
				continue
			}

			if tile.Kind == "wall" {
				continue
			}

			exits := 0
			for _, adjs := range []struct{ x, y int }{
				{tile.X, tile.Y + 1},
				{tile.X, tile.Y - 1},
				{tile.X + 1, tile.Y},
				{tile.X - 1, tile.Y},
			} {
				adjTile := w.TileGrid[adjs.x][adjs.y]
				if adjTile.Kind != "wall" {
					exits++
				}
			}

			if exits != 1 {
				continue
			}
			done = false
			tile.Kind = "wall"
		}
	}
}

const WINDING_PERCENT int = 2

func (w *World) GrowMaze(x, y int) {
	pos := &Coordinate{x, y}
	cells := []*Coordinate{}
	w.startRegion()
	w.Carve(x, y)
	cells = append(cells, pos)
	var lastDir string

	for len(cells) > 0 {
		cell := cells[len(cells)-1]

		unmadeCells := []string{}
		for _, direction := range []string{
			"north",
			"south",
			"east",
			"west",
		} {
			if w.canCarve(cell.X, cell.Y, direction) {
				unmadeCells = append(unmadeCells, direction)
			}
		}

		if len(unmadeCells) > 0 {
			dir := ""
			unmadeCellsContainsLast := false
			for _, direction := range unmadeCells {
				if direction == lastDir {
					unmadeCellsContainsLast = true
				}
			}
			if unmadeCellsContainsLast && rand.Intn(100) > WINDING_PERCENT {
				dir = lastDir
			} else {
				dir = unmadeCells[rand.Intn(len(unmadeCells))]
			}

			switch dir {
			case "north":
				w.Carve(cell.X, cell.Y+1)
				w.Carve(cell.X, cell.Y+2)
				cells = append(cells, &Coordinate{cell.X, cell.Y + 2})
			case "south":
				w.Carve(cell.X, cell.Y-1)
				w.Carve(cell.X, cell.Y-2)
				cells = append(cells, &Coordinate{cell.X, cell.Y - 2})
			case "east":
				w.Carve(cell.X+1, cell.Y)
				w.Carve(cell.X+2, cell.Y)
				cells = append(cells, &Coordinate{cell.X + 2, cell.Y})
			case "west":
				w.Carve(cell.X-1, cell.Y)
				w.Carve(cell.X-2, cell.Y)
				cells = append(cells, &Coordinate{cell.X - 2, cell.Y})
			}

			lastDir = dir
		} else {
			cells = cells[:len(cells)-1]
			lastDir = ""
		}
	}
}

func (w *World) canCarve(x int, y int, direction string) bool {
	switch direction {
	case "north":
		if y+3 > WORLD_HEIGHT {
			return false
		}

		if w.TileGrid[x][y+2].Kind == "wall" {
			return true
		}
	case "south":
		if y-3 < 0 {
			return false
		}

		if w.TileGrid[x][y-2].Kind == "wall" {
			return true
		}
	case "east":
		if x+3 > WORLD_WIDTH {
			return false
		}

		if w.TileGrid[x+2][y].Kind == "wall" {
			return true
		}
	case "west":
		if x-3 < 0 {
			return false
		}

		if w.TileGrid[x-2][y].Kind == "wall" {
			return true
		}
	}

	return false
}

func (w *World) Carve(x, y int) {
	tile := w.TileGrid[x][y]
	tile.Kind = "floor"
	tile.region = w.currentRegion
}

const NUM_ROOM_ATTEMPTS int = 1000
const ROOM_EXTRA_SIZE int = 1

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
		w.startRegion()
		w.CarveRect(room, "floor")
	}
}

func (w *World) startRegion() {
	w.currentRegion += 1
}

func (w *World) CarveRect(r *Rect, kind string) {
	for _, coord := range r.Coords() {
		tile := w.TileGrid[coord.X][coord.Y]
		tile.Kind = kind
		tile.region = w.currentRegion
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
