package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	players := map[string]*Player{
		"Tristan": NewPlayer("Tristan"),
	}

	mu := &sync.Mutex{}

	initGame()
	fmt.Println("players[\"Tristan\"]", players["Tristan"].Location)
	addPlayer(players["Tristan"])
	go func() {
		output := players["Tristan"].GetOutput()
		for msg := range output {
			mu.Lock()
			fmt.Println(msg)
			mu.Unlock()
		}
	}()
	players["Tristan"].HandleInput("осмотреться")
	players["Tristan"].HandleInput("идти коридор")
	players["Tristan"].HandleInput("идти комната")
	players["Tristan"].HandleInput("осмотреться")
	players["Tristan"].HandleInput("одеть рюкзак")
	players["Tristan"].HandleInput("взять ключи")
	players["Tristan"].HandleInput("взять ключи")
	players["Tristan"].HandleInput("взять конспекты")
	players["Tristan"].HandleInput("идти коридор")
	players["Tristan"].HandleInput("идти улица")
	players["Tristan"].HandleInput("применить ключи дверь")
	players["Tristan"].HandleInput("применить телефон шкаф")
	players["Tristan"].HandleInput("применить ключи шкаф")
	players["Tristan"].HandleInput("идти улица")
	//players["Tristan"].HandleInput("идти улица")
	//{21, "идти улица", "дверь закрыта"},                                  //условие не удовлетворено
	//{22, "применить ключи дверь", "дверь открыта"},                       //состояние изменилось
	//{23, "применить телефон шкаф", "нет предмета в инвентаре - телефон"}, // нет предмета
	//{24, "применить ключи шкаф", "не к чему применить"},                  // предмет есть, но применить его к этому нельзя
	//{25, "идти улица", "на улице весна. можно пройти - домой"},
	time.Sleep(2 * time.Second)
}
