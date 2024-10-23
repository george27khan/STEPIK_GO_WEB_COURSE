package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func TagToMap(tag string) map[string]string {
	tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
	tag, _ = strings.CutSuffix(tag, "\"")
	optsMap := make(map[string]string)
	for _, opt := range strings.Split(tag, ",") {
		optParts := strings.Split(opt, "=")
		if len(optParts) > 1 {
			optsMap[optParts[0]] = optParts[1]
		} else {
			optsMap[optParts[0]] = ""
		}
	}
	fmt.Println(optsMap)
	return optsMap
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/profile":
		if false && r.Header.Get("X-Auth") != "100500" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"unauthorized\"}"))
			return
		}
		if r.Method != "" && "" != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("{\"error\": \"bad method\"}"))
			return
		}
		h.wrapperProfile(w, r)

	case "/user/create":
		if true && r.Header.Get("X-Auth") != "100500" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"unauthorized\"}"))
			return
		}
		if r.Method != "POST" && "POST" != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("{\"error\": \"bad method\"}"))
			return
		}
		h.wrapperCreate(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}
}

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// заполнение структуры params
	// валидирование параметров

	var (
		params     ProfileParams
		optsMap    map[string]string
		queryParam url.Values
		err        error
	)
	if r.Method == http.MethodPost {
		// Считывание тела запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			return
		}
		// Разбор строки с использованием url.ParseQuery
		queryParam, _ = url.ParseQuery(string(body))
	} else {
		queryParam = r.URL.Query()
	}

	optsMap = TagToMap(`apivalidator:"required"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Login = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Login = queryParam.Get(strings.ToLower("Login"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Login == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Login"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Login == "" {
			params.Login = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Login == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Login"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Login) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Login"), minVal)))
				return
			}
		}
	}

	fmt.Println("in params", params)

	res, err := h.Profile(ctx, params)
	// обработка ошибки
	if err != nil {
		switch err.(type) {
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	// обработка успешного ответа
	respMap := map[string]interface{}{"error": "", "response": res}
	resp, err := json.Marshal(&respMap)
	fmt.Println("resp", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(resp)
}

func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// заполнение структуры params
	// валидирование параметров

	var (
		params     CreateParams
		optsMap    map[string]string
		queryParam url.Values
		err        error
	)
	if r.Method == http.MethodPost {
		// Считывание тела запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			return
		}
		// Разбор строки с использованием url.ParseQuery
		queryParam, _ = url.ParseQuery(string(body))
	} else {
		queryParam = r.URL.Query()
	}

	optsMap = TagToMap(`apivalidator:"min=0,max=128"`)
	paramAge := ""
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		paramAge = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		paramAge = queryParam.Get(strings.ToLower("Age"))
		delete(optsMap, "paramname")
	}
	if params.Age, err = strconv.Atoi(paramAge); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", strings.ToLower("Age"))))
		return
	}

	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Age == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Age"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Age == 0 {
			params.Age, _ = strconv.Atoi(val)
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || params.Age < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", strings.ToLower("Age"), minVal)))
				return
			}
		}
		// если есть ограничения max
		if opt == "max" {
			if maxVal, err := strconv.Atoi(val); err != nil || params.Age > maxVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", strings.ToLower("Age"), maxVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"required,min=10"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Login = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Login = queryParam.Get(strings.ToLower("Login"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Login == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Login"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Login == "" {
			params.Login = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Login == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Login"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Login) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Login"), minVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"paramname=full_name"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Name = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Name = queryParam.Get(strings.ToLower("Name"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Name"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Name == "" {
			params.Name = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Name == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Name"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Name) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Name"), minVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"enum=user|moderator|admin,default=user"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Status = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Status = queryParam.Get(strings.ToLower("Status"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Status == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Status"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Status == "" {
			params.Status = val
		}
		fmt.Println("val ", val)
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Status == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Status"), strings.ReplaceAll(val, "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Status) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Status"), minVal)))
				return
			}
		}
	}

	fmt.Println("in params", params)

	res, err := h.Create(ctx, params)
	// обработка ошибки
	if err != nil {
		switch err.(type) {
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	// обработка успешного ответа
	respMap := map[string]interface{}{"error": "", "response": res}
	resp, err := json.Marshal(&respMap)
	fmt.Println("resp", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(resp)
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/create":
		if true && r.Header.Get("X-Auth") != "100500" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"unauthorized\"}"))
			return
		}
		if r.Method != "POST" && "POST" != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("{\"error\": \"bad method\"}"))
			return
		}
		h.wrapperCreate(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}
}

func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	// заполнение структуры params
	// валидирование параметров

	var (
		params     OtherCreateParams
		optsMap    map[string]string
		queryParam url.Values
		err        error
	)
	if r.Method == http.MethodPost {
		// Считывание тела запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			return
		}
		// Разбор строки с использованием url.ParseQuery
		queryParam, _ = url.ParseQuery(string(body))
	} else {
		queryParam = r.URL.Query()
	}

	optsMap = TagToMap(`apivalidator:"min=1,max=50"`)
	paramLevel := ""
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		paramLevel = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		paramLevel = queryParam.Get(strings.ToLower("Level"))
		delete(optsMap, "paramname")
	}
	if params.Level, err = strconv.Atoi(paramLevel); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", strings.ToLower("Level"))))
		return
	}

	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Level == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Level"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Level == 0 {
			params.Level, _ = strconv.Atoi(val)
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || params.Level < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", strings.ToLower("Level"), minVal)))
				return
			}
		}
		// если есть ограничения max
		if opt == "max" {
			if maxVal, err := strconv.Atoi(val); err != nil || params.Level > maxVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", strings.ToLower("Level"), maxVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"required,min=3"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Username = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Username = queryParam.Get(strings.ToLower("Username"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Username"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Username == "" {
			params.Username = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Username == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Username"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Username) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Username"), minVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"paramname=account_name"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Name = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Name = queryParam.Get(strings.ToLower("Name"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Name"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Name == "" {
			params.Name = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Name == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Name"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Name) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Name"), minVal)))
				return
			}
		}
	}

	optsMap = TagToMap(`apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`)
	// сперва проверка опции подмены имени атрибута
	if paramName, ok := optsMap["paramname"]; ok {
		params.Class = queryParam.Get(paramName)
		delete(optsMap, "paramname")
	} else {
		params.Class = queryParam.Get(strings.ToLower("Class"))
	}
	// далее проверяем остальные параметры
	for opt, val := range optsMap {
		if opt == "required" && params.Class == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("Class"))))
			return
		}
		// если есть дефолт значение и параметр принял дефолт значение
		if opt == "default" && params.Class == "" {
			params.Class = val
		}
		// если есть ограничения по значениям
		if opt == "enum" {
			ok := false
			for _, item := range strings.Split(val, "|") {
				if params.Class == item {
					ok = true
					break
				}
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("Class"), strings.ReplaceAll("", "|", ", "))))
				return
			}
		}
		// если есть ограничения min
		if opt == "min" {
			if minVal, err := strconv.Atoi(val); err != nil || len(params.Class) < minVal {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("Class"), minVal)))
				return
			}
		}
	}

	fmt.Println("in params", params)

	res, err := h.Create(ctx, params)
	// обработка ошибки
	if err != nil {
		switch err.(type) {
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}
	// обработка успешного ответа
	respMap := map[string]interface{}{"error": "", "response": res}
	resp, err := json.Marshal(&respMap)
	fmt.Println("resp", resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(resp)
}
