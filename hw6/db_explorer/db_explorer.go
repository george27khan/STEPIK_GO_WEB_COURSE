package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

// CaseResponse
type respMap map[string]interface{}

type colInfo struct {
	Field        sql.NullString
	TypeName     sql.NullString
	Collation    sql.NullString
	Null         sql.NullString
	Key          sql.NullString
	DefaultValue sql.NullString
	Extra        sql.NullString
	Privileges   sql.NullString
	Comment      sql.NullString
}

type DBexplorer struct {
	TabsInfo map[string][]colInfo
	DB       *sql.DB
}

// считывание списка таблиц отдельной функцией, для устранения проблемы открытия второго соединения
func initTables(db *sql.DB) (tables []string) {
	var tableName string
	rows, err := db.Query("SHOW TABLES")
	defer rows.Close()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			slog.Error(err.Error())
		}
		tables = append(tables, tableName)
	}
	return
}

// Создаем "конструктор" для Person с зачитыванием информации о таблицах
func NewExplorer(db *sql.DB) *DBexplorer {
	var col colInfo
	colInfoMap := make(map[string][]colInfo)
	for _, tableName := range initTables(db) {
		cols, err := db.Query(fmt.Sprintf("SHOW FULL COLUMNS FROM %s", tableName))
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		for cols.Next() {
			if err := cols.Scan(&col.Field, &col.TypeName, &col.Collation, &col.Null, &col.Key, &col.DefaultValue, &col.Extra, &col.Privileges, &col.Comment); err != nil {
				slog.Error(err.Error())
				return nil
			}
			colInfoMap[tableName] = append(colInfoMap[tableName], col)
		}
		cols.Close()
	}
	return &DBexplorer{
		TabsInfo: colInfoMap,
		DB:       db,
	}
}

// точка входа
func NewDbExplorer(db *sql.DB) (*http.ServeMux, error) {
	explorer := NewExplorer(db) // при инициализации регистрируем все таблицы в сущность
	mux := http.NewServeMux()
	mux.HandleFunc("/", explorer.rootHandler)

	//зарегистрируем роуты для всех таблиц
	for tableName, _ := range explorer.TabsInfo {
		mux.HandleFunc(fmt.Sprintf("/%s", tableName), explorer.tableHandler)
		mux.HandleFunc(fmt.Sprintf("/%s/", tableName), explorer.tableRowHandler)
	}
	return mux, nil
}

func (exp *DBexplorer) tableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		exp.getRows(w, r)
	} else if r.Method == http.MethodPut {

	}
}

func (exp *DBexplorer) getRows(w http.ResponseWriter, r *http.Request) {
	var (
		limit, offset       int
		limitStr, offsetStr string
		tableName           string
		colNameSlice        []string
		colsInfo            []colInfo
		ok                  bool
	)
	tableName, _ = strings.CutPrefix(r.URL.Path, "/")

	if colsInfo, ok = exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	}

	if limitStr = r.URL.Query().Get("limit"); limitStr == "" {
		limit = 5
	} else {
		if val, err := strconv.Atoi(limitStr); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			return
		} else {
			limit = val
		}
	}
	if offsetStr = r.URL.Query().Get("offset"); offsetStr == "" {
		offset = 0
	} else {
		if val, err := strconv.Atoi(offsetStr); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		} else {
			offset = val
		}
	}

	query := fmt.Sprintf("select * from %s limit ? offset ?", tableName)
	rows, err := exp.DB.Query(query, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	defer rows.Close()
	attrs := make([]interface{}, len(colsInfo))
	attrsPntr := make([]interface{}, len(colsInfo))
	for i, _ := range attrsPntr {
		attrsPntr[i] = &attrs[i]
	}

	if colNameSlice, err = rows.Columns(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	result := make([]map[string]interface{}, 0, limit-offset)
	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(attrsPntr...); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			return
		}
		colAttr := make(map[string]interface{})
		for j, name := range colNameSlice {
			switch val := attrs[j].(type) {
			case []byte:
				colAttr[name] = string(val)
			default:
				colAttr[name] = attrs[j]
			}
		}
		result = append(result, colAttr)
	}
	js, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"response\": {\"records\": %s}}", string(js))))
}

func (exp *DBexplorer) tableRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		exp.getRow(w, r)
	} else if r.Method == http.MethodPut {
		exp.putRow(w, r)
	}
}

func (exp *DBexplorer) putRow(w http.ResponseWriter, r *http.Request) {
	var (
		tableName string
		//colNameSlice []string
		colsInfo []colInfo
		ok       bool
		incomRow map[string]interface{}
	)
	tableName = strings.Split(r.URL.Path, "/")[1]
	if colsInfo, ok = exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	incomRow = make(map[string]interface{})
	if err := json.Unmarshal(body, &incomRow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	exp.DB.Exec("insert into %s values(?)", tableName)
	insertCols := strings.Builder{}
	insertVals := strings.Builder{}
	for _, attr := range colsInfo {
		//проверяем наличие атрибута
		if val, ok := incomRow[attr.Field.String]; !ok {
			//если атрибута нету и он в базе определен как not null, то завершаем работу
			if attr.Null.String == "NO" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("{\"error\": \"empty param value\"}"))
				return
			}
		} else {
			insertCols.WriteString(attr.Field.String)
			insertCols.WriteString(",")
			insertVals.WriteString("?,")
		}
	}
	fmt.Println("body", colsInfo)
	fmt.Println("body", incomRow)

}

func (exp *DBexplorer) getRow(w http.ResponseWriter, r *http.Request) {
	var (
		colNameSlice []string
		colsInfo     []colInfo
		id           int
		tableName    string
		err          error
		ok           bool
	)
	pathParts := strings.Split(r.URL.Path, "/")
	tableName = pathParts[1]
	if id, err = strconv.Atoi(pathParts[2]); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	}
	if colsInfo, ok = exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	}
	rows, err := exp.DB.Query(fmt.Sprintf("select * from %s where id = ?", tableName), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err.Error()))
		return
	}
	defer rows.Close()

	if !rows.Next() {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	} else {
		attrs := make([]interface{}, len(colsInfo))
		attrsPntr := make([]interface{}, len(colsInfo))
		for i, _ := range attrsPntr {
			attrsPntr[i] = &attrs[i]
		}
		if colNameSlice, err = rows.Columns(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err.Error()))
			return
		}
		result := make(map[string]interface{}, 1)
		if err = rows.Scan(attrsPntr...); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err.Error()))
			return
		}
		for j, name := range colNameSlice {
			switch val := attrs[j].(type) {
			case []byte:
				result[name] = string(val)
			default:
				result[name] = attrs[j]
			}
		}
		js, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"response\": {\"record\": %s}}", string(js))))
	}
}

func (exp *DBexplorer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if tableName, _ := strings.CutPrefix(r.URL.Path, "/"); tableName != "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("{\"error\": \"unknown table\"}"))
			return
		}

		tables := make([]string, 0, len(exp.TabsInfo))
		for tabName, _ := range exp.TabsInfo {
			tables = append(tables, tabName)
		}
		sort.Strings(tables)
		if resp, err := json.Marshal(respMap{"response": respMap{"tables": tables}}); err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return
		}
	}
}
