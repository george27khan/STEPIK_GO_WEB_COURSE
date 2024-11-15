package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type colDescr struct {
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

type tableInfo struct {
	Cols  []colDescr
	PKCol string // работает только для single pk
}

type DBexplorer struct {
	TabsInfo map[string]*tableInfo
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
	var col colDescr
	colInfoMap := make(map[string]*tableInfo)
	for _, tableName := range initTables(db) {
		cols, err := db.Query(fmt.Sprintf("SHOW FULL COLUMNS FROM %s", tableName))
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		colInfoMap[tableName] = &tableInfo{}
		colInfoMap[tableName].Cols = make([]colDescr, 0)
		for cols.Next() {
			if err := cols.Scan(&col.Field, &col.TypeName, &col.Collation, &col.Null, &col.Key, &col.DefaultValue, &col.Extra, &col.Privileges, &col.Comment); err != nil {
				slog.Error(err.Error())
				return nil
			}
			colInfoMap[tableName].Cols = append(colInfoMap[tableName].Cols, col)
			if col.Key.String == "PRI" {
				colInfoMap[tableName].PKCol = col.Field.String // запоминает поле PK
			}
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
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (exp *DBexplorer) getRows(w http.ResponseWriter, r *http.Request) {
	const (
		limitDef  = 5
		offsetDef = 0
	)
	var (
		limit, offset       int
		limitStr, offsetStr string
		tableName           string
		colNameSlice        []string
		colsDescr           []colDescr
		ok                  bool
		info                *tableInfo
	)
	tableName, _ = strings.CutPrefix(r.URL.Path, "/")

	if info, ok = exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	} else {
		colsDescr = info.Cols
	}

	if limitStr = r.URL.Query().Get("limit"); limitStr == "" {
		limit = limitDef
	} else {
		// по тестам если пришло не число, заменяем дефолт значением
		if val, err := strconv.Atoi(limitStr); err != nil {
			limit = limitDef
		} else {
			limit = val
		}
	}
	if offsetStr = r.URL.Query().Get("offset"); offsetStr == "" {
		offset = offsetDef
	} else {
		// по тестам если пришло не число, заменяем дефолт значением
		if val, err := strconv.Atoi(offsetStr); err != nil {
			offset = offsetDef
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
	attrs := make([]interface{}, len(colsDescr))
	attrsPntr := make([]interface{}, len(colsDescr))
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
	} else if r.Method == http.MethodPost {
		exp.postRow(w, r)
	} else if r.Method == http.MethodDelete {
		exp.deleteRow(w, r)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (exp *DBexplorer) putRow(w http.ResponseWriter, r *http.Request) {
	var (
		tableName, colNamePK string
		colsDescr            []colDescr
		incomRow             map[string]interface{}
		attrsInsert          []interface{}
	)
	tableName = strings.Split(r.URL.Path, "/")[1]
	if info, ok := exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	} else {
		colsDescr = info.Cols
		colNamePK = info.PKCol
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
	insertCols := strings.Builder{}
	insertVals := strings.Builder{}
	attrsInsert = make([]interface{}, 0, len(incomRow))
	for _, attr := range colsDescr {
		//проверяем наличие атрибута
		if val, ok := incomRow[attr.Field.String]; !ok {
			//если атрибута нету и он в базе определен как not null и не определено дефолт значение, то подменяем пустой строкой
			if attr.Null.String == "NO" && attr.DefaultValue.String == "" {
				insertCols.WriteString(attr.Field.String)
				insertCols.WriteString(",")
				insertVals.WriteString("?,")
				attrsInsert = append(attrsInsert, "")
			}
		} else if attr.Extra.String != "auto_increment" {
			// добавляем атрибут в запрос только если не автоинкремент
			insertCols.WriteString(attr.Field.String)
			insertCols.WriteString(",")
			insertVals.WriteString("?,")
			attrsInsert = append(attrsInsert, val)
		}
	}
	result, err := exp.DB.Exec(fmt.Sprintf("insert into %s (%s) values(%s)",
		tableName,
		strings.TrimRight(insertCols.String(), ","),
		strings.TrimRight(insertVals.String(), ",")),
		attrsInsert...,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\": \"insert error\"}"))
		return
	}
	if id, err := result.LastInsertId(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\": \"get last id error\"}"))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"response\": {\"%s\": %v}}", colNamePK, id)))
		return
	}
}

func (exp *DBexplorer) postRow(w http.ResponseWriter, r *http.Request) {
	var (
		tableName, colNamePK string
		colsDescr            []colDescr
		id                   int
		err                  error
	)
	pathParts := strings.Split(r.URL.Path, "/")
	tableName = pathParts[1]
	//получаем id записи из запроса
	if id, err = strconv.Atoi(pathParts[2]); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	}
	if info, ok := exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	} else {
		colsDescr = info.Cols
		colNamePK = info.PKCol
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//распакрвываем тело в мап интерфейсов
	incomUpdAttr := make(map[string]interface{})
	if err := json.Unmarshal(body, &incomUpdAttr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	updateCols := strings.Builder{}
	valsUpdate := make([]interface{}, 0, len(incomUpdAttr))
	//формируем куски запроса
	for attr, val := range incomUpdAttr {
		//проверка пришедших атрибутов и их значений
		for _, info := range colsDescr {
			if info.Field.String == attr {
				// поиск среди полей на обновление PRIMARY KEY
				if info.Key.String == "PRI" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"field %s have invalid type\"}", attr)))
					return
				}
				//проверка обновления на null, если в базе поле как not null
				if info.Null.String == "NO" && val == nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"field %s have invalid type\"}", attr)))
					return
				}
				//проверка на типы пришедшего значения для поля, не все типы учтены...
				switch t := val.(type) {
				case float64:
					if !strings.Contains(info.TypeName.String, "int") &&
						!strings.Contains(info.TypeName.String, "float") &&
						!strings.Contains(info.TypeName.String, "double") {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"field %s have invalid type\"}", attr)))
						return
					}
				case string:
					if info.TypeName.String != "text" && !strings.HasPrefix(info.TypeName.String, "varchar") {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"field %s have invalid type\"}", attr)))
						return
					}
				default:
					fmt.Println("default", t, val)
				}
			}
		}
		updateCols.WriteString(attr)
		updateCols.WriteString(" = ?,")
		valsUpdate = append(valsUpdate, val)
	}
	valsUpdate = append(valsUpdate, id)
	result, err := exp.DB.Exec(fmt.Sprintf("update %s set %s where %s = ?",
		tableName,
		strings.TrimRight(updateCols.String(), ","),
		colNamePK),
		valsUpdate...)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}

	if updatedRows, err := result.RowsAffected(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"response\": {\"updated\": %v}}", updatedRows)))
		return
	}
}

func parseURL(url string, tabInfo map[string]*tableInfo) (tableName string, id int, err error) {
	pathParts := strings.Split(url, "/")
	tableName = pathParts[1]
	//получаем id записи из запроса
	if id, err = strconv.Atoi(pathParts[2]); err != nil {
		return "", 0, err
	}
	if _, ok := tabInfo[tableName]; !ok {
		return "", 0, errors.New("table not found")
	}
	return tableName, id, nil
}

func (exp *DBexplorer) deleteRow(w http.ResponseWriter, r *http.Request) {
	var (
		tableName string
		id        int
		err       error
	)
	//разбираем URL
	tableName, id, err = parseURL(r.URL.Path, exp.TabsInfo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
	}
	result, err := exp.DB.Exec(fmt.Sprintf("delete from %s where id = ?", tableName), id)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}

	if deletedRows, err := result.RowsAffected(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"response\": {\"deleted\": %v}}", deletedRows)))
		return
	}
}

func (exp *DBexplorer) getRow(w http.ResponseWriter, r *http.Request) {
	var (
		colNameSlice         []string
		colsDescr            []colDescr
		id                   int
		tableName, colNamePK string
		err                  error
		ok                   bool
		info                 *tableInfo
	)
	pathParts := strings.Split(r.URL.Path, "/")
	tableName = pathParts[1]
	//получаем id записи из запроса
	if id, err = strconv.Atoi(pathParts[2]); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	}

	if info, ok = exp.TabsInfo[tableName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown table\"}"))
		return
	} else {
		colsDescr = info.Cols
		colNamePK = info.PKCol
		fmt.Println(tableName, colNamePK)
	}

	rows, err := exp.DB.Query(fmt.Sprintf("select * from %s where %s = ?", tableName, colNamePK), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	defer rows.Close()
	if !rows.Next() { // проверяем нашли ли записи, если нет, то ошибка
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	} else {
		attrs := make([]interface{}, len(colsDescr))
		attrsPntr := make([]interface{}, len(colsDescr))
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
