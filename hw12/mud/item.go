package main

type Item struct {
	Name   string
	Action string
}

type Place struct {
	Name  string
	Items []*Item
}
