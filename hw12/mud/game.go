package main

import (
	"fmt"
	"strings"
)

// сюда писать код
// на сервер грузить только этот файл
const (
	lookAround = "осмотреться"
	goTo       = "идти"
	clothe     = "одеть"
	take       = "взять"
	use        = "применить"
	tell       = "сказать"
	tellPlayer = "сказать_игроку"
)

type World struct {
	StartLocation *Location
	Locations     map[string]*Location
	Commands      map[string]cmdFunc
	Players       map[string]*Player
}

type Location struct {
	Name          string
	LookAroundStr string
	GoStr         string
	Actions       []string
	PlaceItems    []*Place
	NextLocations []*Location
	Doors         []*Door
	Players       []*Player
}

type Door struct {
	Key          *Item
	OpenLocation *Location
	IsOpen       bool
}
type TargetItems struct {
}

type LocationObject struct {
	Name       string
	Properties map[string]interface{}
}

type LocationItems struct {
	Place string
	Items []*Item
}

type iCmdFunc interface {
	LookAround(command string, params []string)
	Go(command string, params []string)
	TakeItem(command string, params []string)
	ClotheItem(command string, params []string)
	Use(command string, params []string)
	TellPlayer(command string, params []string)
	Tell(command string, params []string)
}

type cmdFunc func(fnc iCmdFunc, command string, params []string)

var world World

func initGame() {
	var Room, Corridor, Kitchen, Outside Location

	corridorOutsideDoorKey := &Item{"ключи", take}

	Room = Location{
		Name:          "комната",
		LookAroundStr: "",
		GoStr:         "ты в своей комнате",
		NextLocations: []*Location{&Corridor},
		PlaceItems: []*Place{
			&Place{"на столе", []*Item{corridorOutsideDoorKey, &Item{"конспекты", take}}},
			&Place{"на стуле", []*Item{&Item{"рюкзак", clothe}}},
		},
		Players: make([]*Player, 0),
	}
	Kitchen = Location{
		Name:          "кухня",
		LookAroundStr: "ты находишься на кухне, на столе чай",
		GoStr:         "кухня, ничего интересного",
		Actions:       []string{"собрать рюкзак", "идти в универ"},
		NextLocations: []*Location{&Corridor},
		Players:       make([]*Player, 0),
	}
	Outside = Location{
		Name:          "улица",
		LookAroundStr: "",
		GoStr:         "на улице весна",
		NextLocations: []*Location{&Corridor},
		Players:       make([]*Player, 0),
	}
	Corridor = Location{
		Name:          "коридор",
		LookAroundStr: "",
		GoStr:         "ничего интересного",
		NextLocations: []*Location{&Kitchen, &Room, &Outside},
		Doors:         []*Door{&Door{corridorOutsideDoorKey, &Outside, false}},
		Players:       make([]*Player, 0),
	}
	Commands := map[string]cmdFunc{
		lookAround: iCmdFunc.LookAround,
		goTo:       iCmdFunc.Go,
		use:        iCmdFunc.Use,
		take:       iCmdFunc.TakeItem,
		clothe:     iCmdFunc.ClotheItem,
		tell:       iCmdFunc.Tell,
		tellPlayer: iCmdFunc.TellPlayer}
	world = World{
		StartLocation: &Kitchen,
		Locations:     map[string]*Location{"коридор": &Corridor, "комната": &Room, "кухня": &Kitchen, "улица": &Outside},
		Commands:      Commands,
		Players:       make(map[string]*Player),
	}
}

func (l *Location) getItem(itemName string) (*Item, error) {
	for i, place := range l.PlaceItems {
		for j, item := range place.Items {
			if item.Name == itemName {
				place.Items = append(place.Items[0:j], place.Items[j+1:]...)
				if len(place.Items) > 0 {
					l.PlaceItems[i].Items = place.Items
				} else {
					l.PlaceItems = append(l.PlaceItems[:i], l.PlaceItems[i+1:]...)
				}
				return item, nil
			}
		}
	}
	return nil, fmt.Errorf("нет такого")
}

// открывает дверь ключем, если он подходит
func (l *Location) OpenDoor(key *Item) string {
	for _, door := range l.Doors {
		if door.Key == key {
			door.IsOpen = true
			return "дверь открыта"
		}
	}
	return "дверь не открыта"
}

// удалить список нужных действий
func (l *Location) DelAction(actionName string) {
	for i, action := range l.Actions {
		if action == actionName {
			l.Actions = append(l.Actions[:i], l.Actions[i+1:]...)
		}
	}
}

func (l *Location) Describe(command string, player *Player) string {
	res := &strings.Builder{}
	if command == lookAround {
		res.WriteString(l.LookAroundStr)
	} else if command == goTo {
		res.WriteString(l.GoStr)
	}

	if command == "осмотреться" {
		cntActions := len(l.Actions)
		if cntActions > 0 {
			res.WriteString(", надо ")
			if cntActions > 1 {
				res.WriteString(strings.Join(l.Actions, " и "))
			} else {
				res.WriteString(l.Actions[0])
			}
		}
		cntPlace := len(l.PlaceItems)
		if l.PlaceItems != nil && len(l.PlaceItems) == 0 {
			res.WriteString(fmt.Sprintf("пустая %s", l.Name))
		} else {
			for j, place := range l.PlaceItems {
				res.WriteString(place.Name)
				cntItems := len(place.Items)
				if cntItems == 1 && len(l.PlaceItems) > 1 {
					res.WriteString(" - ")
				} else {
					res.WriteString(": ")
				}
				for i, item := range place.Items {
					res.WriteString(item.Name)
					if i < cntItems-1 {
						res.WriteString(", ")
					}
				}
				if j < cntPlace-1 {
					res.WriteString(", ")
				}
			}
		}
	}

	cntLocation := len(l.NextLocations)
	if cntLocation > 0 {
		res.WriteString(". можно пройти - ")
		for i, loc := range l.NextLocations {
			if l.Name == "улица" {
				res.WriteString("домой")
			} else {
				res.WriteString(loc.Name)
			}

			if i+1 < cntLocation {
				res.WriteString(", ")
			}
		}
	}

	if command == "осмотреться" && len(l.Players) > 1 {
		for _, p := range l.Players {
			if p.Name == player.Name {
				continue
			}
			res.WriteString(". Кроме вас тут ещё " + p.Name)
		}
	}
	return res.String()
}

// проверка закрыта ли локация дверью
func (l *Location) IsOpen(location string) bool {
	for _, door := range l.Doors {
		if door.OpenLocation.Name == location {
			if door.IsOpen {
				return true
			}
			return false
		}
	}
	return true
}

// проверка доступности ли локация дверью
func (l *Location) IsNextLocation(locationName string) bool {
	for _, location := range l.NextLocations {
		if location.Name == locationName {
			return true
		}
	}
	return false
}

// проверка доступности ли локация дверью
func (l *Location) DelPlayer(player *Player) {
	for i, p := range l.Players {
		if p == player {
			l.Players = append(l.Players[:i], l.Players[i+1:]...)
			return
		}
	}
}
