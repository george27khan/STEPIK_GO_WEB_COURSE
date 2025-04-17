package main

import (
	"fmt"
	"strings"
)

type Player struct {
	Name     string
	Actions  chan string
	Location *Location
	Backpack []*Item
}

func NewPlayer(name string) *Player {
	return &Player{name, make(chan string), nil, nil}
}
func (p *Player) LookAround(command string, params []string) {
	p.Actions <- p.Location.Describe(command)
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
	if location, ok = world.Locations[params[0]]; !ok {
		fmt.Println("Ошибочная локация ", params[0])
		return
	}
	p.Location = location
	p.Actions <- p.Location.Describe(command)
}

func (p *Player) TakeItem(command string, params []string) {
	if p.Backpack == nil {
		p.Actions <- "некуда класть"
		return
	}
	item, err := p.Location.getItem(params[0])
	if err != nil {
		p.Actions <- err.Error()
		return
	}
	if item.Action != take {
		p.Actions <- fmt.Sprintf("%s недопустимое действие", params[0])
		return
	}
	p.Backpack = append(p.Backpack, item)
	p.Actions <- fmt.Sprintf("предмет добавлен в инвентарь: %s", item.Name)
}

func (p *Player) ClotheItem(command string, params []string) {
	item, err := p.Location.getItem(params[0])
	if err != nil {
		p.Actions <- err.Error()
		return
	}
	if item.Action != clothe {
		p.Actions <- fmt.Sprintf("%s недопустимое действие", params[1])
		return
	}
	if item.Name == "рюкзак" {
		p.Backpack = make([]*Item, 0)
	}
	p.Actions <- fmt.Sprintf("вы одели: %s", item.Name)
}

func (p *Player) getItem(itemName string) (*Item, error) {
	for _, item := range p.Backpack {
		if item.Name == itemName {
			return item, nil
		}
	}
	return nil, fmt.Errorf("итем отсутствует")
}

func (p *Player) Use(command string, params []string) {
	item, err := p.getItem(params[0])
	if err != nil {
		p.Actions <- err.Error()
		return
	}
	if item.Name == "ключ" {
		p.Actions <- p.Location.OpenDoor(item)
	}
	if item.Action != clothe {
		p.Actions <- fmt.Sprintf("%s недопустимое действие", params[1])
	}
	if item.Name == "рюкзак" {
		p.Backpack = make([]*Item, 0)
	}
	p.Actions <- fmt.Sprintf("вы одели: %s", item.Name)
}
