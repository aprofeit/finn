package main

type Inventory struct {
	Items []*Item
}

func (i *Inventory) Add(item *Item) {
	i.Items = append(i.Items, item)
}

func (i *Inventory) Remove(id string) {
}

type Item struct {
	Position *Position
}

type Position struct {
	X float64
	Y float64
}
