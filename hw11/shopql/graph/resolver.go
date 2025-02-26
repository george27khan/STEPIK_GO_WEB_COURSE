package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"shopql/graph/model"
	"strconv"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Data LocalDB
}

type Data struct {
	Catalog CatalogJS  `json:"catalog"`
	Seller  []SellerJS `json:"sellers"`
}

type ItemJS struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	InStock  int    `json:"in_stock"`
	SellerId string `json:"seller_id"`
}
type CatalogJS struct {
	Id     int         `json:"id"`
	Name   string      `json:"name"`
	Childs []CatalogJS `json:"childs"`
	Items  []ItemJS    `json:"items"`
}

type SellerJS struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Deals string `json:"deals"`
}

type LocalDB struct {
	Item    map[string]model.Item
	Catalog map[int]model.Catalog
	Seller  map[string]model.Seller
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

	db := LocalDB{make(map[string]model.Item),
		make(map[int]model.Catalog),
		make(map[string]model.Seller)}
	for _, seller := range tmp.Seller {
		s := model.Seller{
			ID:    strconv.Itoa(seller.ID),
			Name:  seller.Name,
			Deals: seller.Deals,
		}
		db.Seller[s.ID] = s
	}
	loadCatalog(&db, tmp.Catalog, 0)
	return &Resolver{db}
}

func loadCatalog(db *LocalDB, catalog CatalogJS, parentID int) {
	db.Catalog[catalog.Id] = model.Catalog{
		ID:       catalog.Id,
		Name:     catalog.Name,
		ParendID: parentID,
	}
	for _, cat := range catalog.Childs {
		loadCatalog(db, cat, catalog.Id)
	}
	for _, item := range catalog.Items {
		db.Item[strconv.Itoa(item.Id)] = model.Item{
			ID:        strconv.Itoa(item.Id),
			Name:      item.Name,
			InStock:   item.InStock,
			SellerID:  item.SellerId,
			CatalogID: strconv.Itoa(catalog.Id),
		}
	}
}
