package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
)

type persons struct {
	XMLName xml.Name `xml:"root"`
	Usr     []Usr    `xml:"row"`
}

type Usr struct {
	ID            string `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      string `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           string `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
	name          string
}

func usrSort(usr []Usr, field string, order_by int, limit int, offset int) {
	if order_by != OrderByAsIs {
		sortFunc := func(i, j int) bool {
			var namei, namej bytes.Buffer
			//Id, Age, Name
			if field == "Name" || field == "" {
				namei.WriteString(usr[i].FirstName)
				namei.WriteString(usr[i].LastName)
				namej.WriteString(usr[j].FirstName)
				namej.WriteString(usr[j].LastName)
				if order_by == OrderByAsc {
					return namei.String() < namej.String()
				} else if order_by == OrderByDesc {
					return namei.String() > namej.String()
				}
			} else if field == "Age" {
				if order_by == OrderByAsc {
					return usr[i].Age < usr[j].Age
				} else if order_by == OrderByDesc {
					return usr[i].Age > usr[j].Age
				}
			} else if field == "Id" {
				if order_by == OrderByAsc {
					return usr[i].ID < usr[j].ID
				} else if order_by == OrderByDesc {
					return usr[i].ID > usr[j].ID
				}
			}
			return false
		}
		sort.SliceStable(usr, sortFunc)
	}
	usr = usr[offset:limit]

}

// код писать тут
// query - что искать. Ищем по полям записи Name и About просто подстроку, без регулярок. Name - это first_name + last_name из xml (вам надо руками пройтись в цикле по записям и сделать такой, автоматом нельзя). Если поле пустое - то возвращаем все записи (поиск пустой подстроки всегда возвращает true), т.е. делаем только логику сортировки
// order_field - по какому полю сортировать. Работает по полям Id, Age, Name, если пустой - то сортируем по Name, если что-то другое - SearchServer ругается ошибкой.
// order_by - направление сортировки (как есть, по убыванию, по возрастанию), в client.go есть соответствующие константы
// limit - сколько записей вернуть
// offset - начиня с какой записи вернуть (сколько пропустить с начала) - нужно для огранизации постраничной навигации
func main() {
	var (
		pers                    persons
		query, orderField       string
		limit, offset, order_by int
		buffer                  bytes.Buffer
		usrRes                  []Usr
	)
	query = "est a"
	orderField = "Name"
	order_by = OrderByDesc
	limit = 10
	offset = 0
	if orderField != "" && orderField != "Id" && orderField != "Age" && orderField != "Name" {
		fmt.Println("Error")
		return
	}

	fileData, err := os.ReadFile("dataset.xml")
	if err != nil {
		panic("Error file open. " + err.Error())
	}
	if err := xml.Unmarshal(fileData, &pers); err != nil {
		fmt.Println(err.Error())
	}
	usrRes = make([]Usr, 0, len(pers.Usr))

	//фильтрация
	if query != "" {
		for _, user := range pers.Usr {
			buffer.WriteString(user.FirstName)
			buffer.WriteString(user.LastName)
			if strings.Contains(buffer.String(), query) || strings.Contains(user.About, query) {
				usrRes = append(usrRes, user)
			}
		}
	}
	usrSort(usrRes, orderField, order_by, limit, offset)
	txt, _ := xml.MarshalIndent(usrRes, "", "	")
	fmt.Println(string(txt))

}
