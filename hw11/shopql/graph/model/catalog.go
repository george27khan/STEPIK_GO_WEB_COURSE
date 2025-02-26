package model

type Catalog struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParendID int    `json:"-"`
}

//func (c *Catalog) id() string {
//	return strconv.Itoa(c.ID)
//}
