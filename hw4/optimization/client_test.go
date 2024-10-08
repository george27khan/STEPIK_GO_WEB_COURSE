package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const secretKey = "secretKey"

type persons struct {
	XMLName xml.Name `xml:"root"`
	Usr     []Usr    `xml:"row"`
}

type Usr struct {
	ID            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      string `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
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

func usrSort(usr []User, field string, order_by int) []User {
	if order_by != OrderByAsIs {
		sortFunc := func(i, j int) bool {
			if field == "Name" || field == "" {
				if order_by == OrderByAsc {
					return usr[i].Name < usr[j].Name
				} else if order_by == OrderByDesc {
					return usr[i].Name > usr[j].Name
				}
			} else if field == "Age" {
				if order_by == OrderByAsc {
					return usr[i].Age < usr[j].Age
				} else if order_by == OrderByDesc {
					return usr[i].Age > usr[j].Age
				}
			} else if field == "Id" {
				if order_by == OrderByAsc {
					return usr[i].Id < usr[j].Id
				} else if order_by == OrderByDesc {
					return usr[i].Id > usr[j].Id
				}
			}
			return false
		}
		sort.SliceStable(usr, sortFunc)
	}
	return usr
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	var (
		pers                                               persons
		query, orderField, limitStr, offsetStr, orderByStr string
		limit, offset, orderBy                             int
		buffer                                             bytes.Buffer
		usrRes                                             []User
	)
	if key, ok := r.Header["Accesstoken"]; !ok || key[0] != secretKey {
		w.WriteHeader(http.StatusUnauthorized)
	}
	query = r.URL.Query().Get("query")
	orderField = r.URL.Query().Get("order_field")
	orderByStr = r.URL.Query().Get("order_by")
	limitStr = r.URL.Query().Get("limit")
	offsetStr = r.URL.Query().Get("offset")

	if orderField != "" && orderField != "Id" && orderField != "Age" && orderField != "Name" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"Error\": \"ErrorBadOrderField\"}")
		return
	}

	if val, err := strconv.Atoi(orderByStr); err != nil || val != OrderByAsc && val != OrderByDesc && val != OrderByAsIs {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: incorrect order_by='%s'", orderByStr)
		return
	} else {
		orderBy = val
	}
	if val, err := strconv.Atoi(limitStr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: incorrect limit='%s'", limitStr)
		return
	} else {
		limit = val
	}
	if val, err := strconv.Atoi(offsetStr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: incorrect offset='%s'", offsetStr)
		return
	} else {
		offset = val
	}

	fileData, err := os.ReadFile("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error file open. (%s)", err.Error())
		return
	}
	if err := xml.Unmarshal(fileData, &pers); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error Unmarshal. (%s)", err.Error())
		return
	}
	usrRes = make([]User, 0, len(pers.Usr))

	//фильтрация
	if query != "" {
		for _, user := range pers.Usr {
			buffer.WriteString(user.FirstName)
			buffer.WriteString(user.LastName)
			if strings.Contains(buffer.String(), query) || strings.Contains(user.About, query) {
				usrRes = append(usrRes, User{user.ID, buffer.String(), user.Age, user.About, user.Gender})
			}
			buffer.Reset()
		}
	}

	usrResSort := usrSort(usrRes, orderField, orderBy)

	//ограничиваем краевые случаи длинны
	if limit > len(usrResSort) {
		limit = len(usrResSort)
	}
	if offset > len(usrResSort) {
		offset = len(usrResSort)
	}
	usrResSort = usrResSort[offset:limit]
	fmt.Println("res", len(usrResSort))

	res, _ := json.Marshal(&usrResSort)

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

type TestCaseErr struct {
	id            int
	client        SearchClient
	server        *httptest.Server
	requestParams SearchRequest
	result        string
}

type TestCaseUser struct {
	id            int
	client        SearchClient
	requestParams SearchRequest
	result        []User
}

type TestCaseNum struct {
	id            int
	client        SearchClient
	requestParams SearchRequest
	result        int
}

type TestCaseBool struct {
	id            int
	client        SearchClient
	requestParams SearchRequest
	result        bool
}

type TestCaseServerAnswer struct {
	id            int
	client        SearchClient
	requestParams SearchRequest
	result        interface{}
}

func mock500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
func mock400EmptyErrResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
func mock400UnknownErrResp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	resp, _ := json.Marshal(SearchErrorResponse{"MOCK_ERROR"})
	w.Write(resp)
}

func mockBodyErrResp(w http.ResponseWriter, r *http.Request) {
}

func mockTimeoutErrResp(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	ts500 := httptest.NewServer(http.HandlerFunc(mock500))
	tsEmptyErr := httptest.NewServer(http.HandlerFunc(mock400EmptyErrResp))
	tsUnknownErr := httptest.NewServer(http.HandlerFunc(mock400UnknownErrResp))
	tsRespBodyErr := httptest.NewServer(http.HandlerFunc(mockBodyErrResp))
	tsRespTimeoutErr := httptest.NewServer(http.HandlerFunc(mockTimeoutErrResp))
	caseClientErr := []TestCaseErr{
		TestCaseErr{1, SearchClient{"", ""}, ts, SearchRequest{-1, 0, "", "", 0}, "limit must be > 0"},
		TestCaseErr{2, SearchClient{"", ""}, ts, SearchRequest{0, -1, "", "", 0}, "offset must be > 0"},
		TestCaseErr{3, SearchClient{"badToken", ts.URL}, ts, SearchRequest{}, "Bad AccessToken"},
		TestCaseErr{4, SearchClient{secretKey, ts.URL}, ts, SearchRequest{0, 0, "", "GUID", 0}, "OrderFeld GUID invalid"},

		TestCaseErr{5, SearchClient{"", tsRespBodyErr.URL}, tsRespBodyErr, SearchRequest{10, 0, "", "GUID", 0}, "cant unpack result json: unexpected end of JSON input"},
		TestCaseErr{6, SearchClient{"", ts500.URL}, ts500, SearchRequest{10, 0, "", "GUID", 0}, "SearchServer fatal error"},
		TestCaseErr{7, SearchClient{"", tsEmptyErr.URL}, tsEmptyErr, SearchRequest{10, 0, "", "GUID", 0}, "cant unpack error json: unexpected end of JSON input"},
		TestCaseErr{8, SearchClient{"", tsUnknownErr.URL}, tsUnknownErr, SearchRequest{10, 0, "", "GUID", 0}, "unknown bad request error: MOCK_ERROR"},
		TestCaseErr{9, SearchClient{"", tsRespTimeoutErr.URL}, tsRespTimeoutErr, SearchRequest{10, 0, "", "GUID", 0}, "timeout for limit=11&offset=0&order_by=0&order_field=GUID&query="},
		TestCaseErr{10, SearchClient{"", ""}, ts, SearchRequest{0, 0, "", "", 0}, "unknown error Get \"?limit=1&offset=0&order_by=0&order_field=&query=\": unsupported protocol scheme \"\""},
	}
	for _, c := range caseClientErr {
		_, err := c.client.FindUsers(c.requestParams)
		if err.Error() != c.result {
			t.Errorf("Test [%d] expected (%s), got (%s)", c.id, c.result, err.Error())
		}
	}

	caseServerAnswer := []TestCaseServerAnswer{
		TestCaseServerAnswer{1, SearchClient{secretKey, ts.URL}, SearchRequest{25, 0, "Brooks", "Name", 0}, []User{User{2, "BrooksAguilar", 25, "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n", "male"}}},
		TestCaseServerAnswer{2, SearchClient{secretKey, ts.URL}, SearchRequest{25, 0, "excepteur", "", 0}, 13},
		TestCaseServerAnswer{3, SearchClient{secretKey, ts.URL}, SearchRequest{2, 0, "excepteur", "", 0}, 2},
	}
	for _, c := range caseServerAnswer {
		resp, err := c.client.FindUsers(c.requestParams)
		if err != nil {
			t.Errorf("Test [%d] expected (%+v), got error(%s)", c.id, c.result, err.Error())
		} else if c.id == 1 {
			if !reflect.DeepEqual(resp.Users, c.result) {
				t.Errorf("Test [%d] expected (%+v), got (%+v)", c.id, c.result, resp.Users)
			}
		} else if !reflect.DeepEqual(len(resp.Users), c.result) {
			t.Errorf("Test [%d] expected (%+v), got (%+v)", c.id, c.result, len(resp.Users))
			fmt.Println("len(resp.Users)", len(resp.Users))
		}
	}

	caseMaxLimit := TestCaseNum{1, SearchClient{secretKey, ts.URL}, SearchRequest{35, 0, "nisi", "Name", 0}, 25}
	resp, err := caseMaxLimit.client.FindUsers(caseMaxLimit.requestParams)
	if err != nil {
		t.Errorf("Test [%d] expected (%+v), got error(%s)", caseMaxLimit.id, caseMaxLimit.result, err.Error())
	} else if len(resp.Users) == caseMaxLimit.result {
		t.Errorf("Test [%d] expected (%+v), got (%+v)", caseMaxLimit.id, caseMaxLimit.result, len(resp.Users))
	}

	caseNextPage := TestCaseBool{1, SearchClient{secretKey, ts.URL}, SearchRequest{35, 0, "do", "Name", 0}, true}
	resp, err = caseNextPage.client.FindUsers(caseNextPage.requestParams)
	//fmt.Println("len ", len(resp.Users), resp.NextPage, caseNextPage.result)
	if err != nil {
		t.Errorf("Test [%d] expected (%v), got error(%s)", caseNextPage.id, caseNextPage.result, err.Error())
	} else if resp.NextPage != caseNextPage.result {
		t.Errorf("Test [%d] expected (%v), got (%v)", caseNextPage.id, caseNextPage.result, resp.NextPage)
	}

}
