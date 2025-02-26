package graph

import (
	"encoding/json"
	"fmt"
	"os"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Data LocalDB
}

type Data struct {
	Catalog CatalogJS `json:"catalog"`
	Seller  []Seller  `json:"sellers"`
}
type Seller struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Deals int    `json:"deals"`
}

type ItemJS struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	InStock  int    `json:"in_stock"`
	SellerId int    `json:"seller_id"`
}
type CatalogJS struct {
	Id     int         `json:"id"`
	Name   string      `json:"name"`
	Childs []CatalogJS `json:"childs"`
	Items  []ItemJS    `json:"items"`
}
type Item struct {
	Id        int
	Name      string
	InStock   int
	SellerID  int
	CatalogID int
}
type Catalog struct {
	Id       int
	Name     string
	ParendID int
}

type LocalDB struct {
	Item    []Item
	Catalog []Catalog
	Seller  []Seller
}

func NewResolver() *Resolver {
	var tmp Data
	data, err := os.ReadFile("testdata.json")
	if err != nil {
		fmt.Println("Ошибка отрытия файла", err.Error())
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		fmt.Println("Ошибка разбора файла", err.Error())
	}
	//fmt.Println(tmp.Catalog)

	db := LocalDB{}
	for _, val := range tmp.Seller {
		db.Seller = append(db.Seller, val)
	}
	loadCatalog(&db, tmp.Catalog, 0)
	return &Resolver{db}
}

func loadCatalog(db *LocalDB, catalog CatalogJS, parentID int) {
	db.Catalog = append(db.Catalog, Catalog{
		Id:       catalog.Id,
		Name:     catalog.Name,
		ParendID: parentID,
	})
	for _, cat := range catalog.Childs {
		loadCatalog(db, cat, catalog.Id)
	}
	for _, item := range catalog.Items {
		db.Item = append(db.Item, Item{
			Id:        item.Id,
			Name:      item.Name,
			InStock:   item.InStock,
			SellerID:  item.SellerId,
			CatalogID: catalog.Id,
		})
	}
}
