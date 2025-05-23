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
	p.Actions <- p.Location.Describe(command, p)
}

func addPlayer(player *Player) {
	player.Location = world.StartLocation
	world.StartLocation.Players = append(world.StartLocation.Players, player)
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
		return
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
	locationName := params[0]
	//fmt.Println("p.Location.Name ", p.Location.Name, locationName)
	if !p.Location.IsNextLocation(locationName) {
		p.Actions <- fmt.Sprintf("нет пути в %s", locationName)
		return
	}

	if !p.Location.IsOpen(locationName) {
		p.Actions <- "дверь закрыта"
		return
	}
	if location, ok = world.Locations[locationName]; !ok {
		fmt.Println("Ошибочная локация ", locationName)
		return
	}
	if location, ok = world.Locations[params[0]]; !ok {
		fmt.Println("Ошибочная локация ", locationName)
		return
	}
	p.Location.DelPlayer(p)
	p.Location = location

	p.Actions <- p.Location.Describe(command, p)
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
		world.Locations["кухня"].DelAction("собрать рюкзак")
	}
	p.Actions <- fmt.Sprintf("вы одели: %s", item.Name)
}

func (p *Player) getItem(itemName string) (*Item, error) {
	for _, item := range p.Backpack {
		if item.Name == itemName {
			return item, nil
		}
	}
	return nil, fmt.Errorf("нет предмета в инвентаре - %s", itemName)
}

func (p *Player) Use(command string, params []string) {
	item, err := p.getItem(params[0])
	target := params[1]
	if err != nil {
		p.Actions <- err.Error()
		return
	}
	if item.Name == "ключи" && target == "дверь" {
		p.Actions <- p.Location.OpenDoor(item)
		return
	}
	if item.Name == "ключи" && target != "дверь" {
		p.Actions <- "не к чему применить"
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

func (p *Player) TellPlayer(command string, params []string) {
	playerName := params[0]
	message := ""
	if len(params[1:]) == 0 {
		message = fmt.Sprintf("%s выразительно молчит, смотря на вас", p.Name)
	} else {
		text := strings.Join(params[1:], " ")
		message = fmt.Sprintf("%s говорит вам: %s", p.Name, text)
	}

	for _, targetPlayer := range p.Location.Players {
		if targetPlayer.Name == playerName {
			targetPlayer.Actions <- message
			return
		}
	}
	p.Actions <- "тут нет такого игрока"

}

func (p *Player) Tell(command string, params []string) {
	text := strings.Join(params, " ")
	message := fmt.Sprintf("%s говорит: %s", p.Name, text)
	for _, targetPlayer := range p.Location.Players {
		targetPlayer.Actions <- message
	}
}
