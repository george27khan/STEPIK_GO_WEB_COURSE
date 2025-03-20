package main

import (
	"fmt"
	"strings"
)

type Player struct {
	Name     string
	Actions  chan string
	Location *Location
}

func NewPlayer(name string) *Player {
	return &Player{name, make(chan string), nil}
}

func addPlayer(player *Player) {
	player.Location = world.StartLocation
	world.Players[player.Name] = player
}

func (p *Player) GetOutput() chan string {
	return p.Actions
}

func (p *Player) HandleInput(command string) {
	var (
		fnc cmdFunc
		ok  bool
	)
	params := strings.Split(command, " ")
	if fnc, ok = world.Commands[params[0]]; !ok {
		p.Actions <- "неизвестная команда"
	}
	fnc(p, params[0], params[1:])
}

func (p *Player) LookAround(command string, params []string) {
	p.Actions <- p.Location.Describe(command)
}

// {2, "идти коридор", "ничего интересного. можно пройти - кухня, комната, улица"},                                         // действие идти
func (p *Player) Go(command string, params []string) {
	var (
		location *Location
		ok       bool
	)
	if len(params) != 1 {
		fmt.Println("Ошибка Go")
		return
	}
	if location, ok = world.Locations[params[0]]; !ok {
		fmt.Println("Ошибочная локация ", params[0])
		return
	}
	p.Location = location
	p.Actions <- p.Location.Describe(command)
}

func (p *Player) Get(command string, params []string) {
	p.Actions <- "взять"
}
