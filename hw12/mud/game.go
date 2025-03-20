package main

import (
	"strings"
)

// сюда писать код
// на сервер грузить только этот файл

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
	NextLocations []*Location
}

type iCmdFunc interface {
	LookAround(command string, params []string)
	Go(command string, params []string)
	Get(command string, params []string)
}
type cmdFunc func(fnc iCmdFunc, command string, params []string)

const (
	lookAround = "осмотреться"
	goTo       = "идти"
)

var world World

func initGame() {
	var Room, Corridor, Kitchen, Outside Location
	//{4, "осмотреться", "на столе: ключи, конспекты, на стуле - рюкзак. можно пройти - коридор"},
	Room = Location{
		Name:          "комната",
		LookAroundStr: "",
		GoStr:         "ты в своей комнате",
		NextLocations: []*Location{&Corridor},
	}
	Kitchen = Location{
		Name:          "кухня",
		LookAroundStr: "ты находишься на кухне, на столе чай",
		Actions:       []string{"собрать рюкзак", "идти в универ"},
		NextLocations: []*Location{&Corridor},
	}
	Outside = Location{
		Name:          "улица",
		LookAroundStr: "",
		NextLocations: []*Location{&Corridor},
	}
	Corridor = Location{
		Name:          "коридор",
		LookAroundStr: "",
		GoStr:         "ничего интересного",
		NextLocations: []*Location{&Kitchen, &Room, &Outside},
	}
	Commands := map[string]cmdFunc{
		lookAround:       iCmdFunc.LookAround,
		goTo:             iCmdFunc.Go,
		"применить":      iCmdFunc.LookAround,
		"взять":          iCmdFunc.Get,
		"одеть":          iCmdFunc.LookAround,
		"сказать":        iCmdFunc.LookAround,
		"сказать_игроку": iCmdFunc.LookAround}
	world = World{
		StartLocation: &Kitchen,
		Locations:     map[string]*Location{"коридор": &Corridor, "комната": &Room, "кухня": &Kitchen, "улица": &Outside},
		Commands:      Commands,
		Players:       make(map[string]*Player),
	}
}

func (l *Location) Describe(command string) string {
	res := &strings.Builder{}
	if command == lookAround {
		res.WriteString(l.LookAroundStr)
	} else if command == goTo {
		res.WriteString(l.GoStr)
	}

	cntActions := len(l.Actions)
	if cntActions > 0 {
		res.WriteString(", надо ")
		if cntActions > 1 {
			res.WriteString(strings.Join(l.Actions, " и "))
		} else {
			res.WriteString(l.Actions[0])
		}
	}

	cntLocation := len(l.NextLocations)
	if cntLocation > 0 {
		res.WriteString(". можно пройти - ")
		for i, loc := range l.NextLocations {
			res.WriteString(loc.Name)
			if i+1 < cntLocation {
				res.WriteString(", ")
			}
		}
	}
	return res.String()

}
