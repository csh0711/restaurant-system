package model

import "github.com/google/uuid"

type OrderID uuid.UUID

func (id OrderID) String() string {
	return uuid.UUID(id).String()
}

type TableID string

type Status string

type MenuItem string

const (
	CaesarSalad      MenuItem = "Caesar Salad"
	MargheritaPizza  MenuItem = "Margherita Pizza"
	PastaCarbonara   MenuItem = "Pasta Carbonara"
	BeefBurger       MenuItem = "Beef Burger"
	ChocolateFondant MenuItem = "Chocolate Fondant"
	Kaesespaetzle    MenuItem = "Käsespätzle"
)

var allItems = map[MenuItem]struct{}{
	CaesarSalad:      {},
	MargheritaPizza:  {},
	PastaCarbonara:   {},
	BeefBurger:       {},
	ChocolateFondant: {},
	Kaesespaetzle:    {},
}

func IsValid(i MenuItem) bool {
	_, ok := allItems[i]
	return ok
}

func List() []MenuItem {
	items := make([]MenuItem, 0, len(allItems))
	for k := range allItems {
		items = append(items, k)
	}
	return items
}
