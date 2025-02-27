package model

type Item struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	InStock   int    `json:"inStock"`
	SellerID  int    `json:"sellerId"`
	CatalogID int    `json:"catalogId"`
}

//func (c *Catalog) id() string {
//	return strconv.Itoa(c.ID)
//}
