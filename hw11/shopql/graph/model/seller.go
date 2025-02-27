package model

type Seller struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Deals string `json:"deals"`
}

//func (c *Catalog) id() string {
//	return strconv.Itoa(c.ID)
//}
