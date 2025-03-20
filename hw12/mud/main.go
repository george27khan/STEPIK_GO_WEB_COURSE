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

	time.Sleep(2 * time.Second)
}
