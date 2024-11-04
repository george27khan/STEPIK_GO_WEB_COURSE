package main
import (
		"context"
		"fmt"
		"net/http"
		"strconv"
		"strings"
		"encoding/json"
		"io"
		"net/url"
		"regexp"
	)


		func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
				
				case "/user/profile":
					if false && r.Header.Get("X-Auth")!="100500" {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("{\"error\": \"unauthorized\"}"))
						return
					}
					if r.Method != ""&&""!="" {
						w.WriteHeader(http.StatusNotAcceptable)
						w.Write([]byte("{\"error\": \"bad method\"}"))
						return
					}
					h.wrapperProfile(w, r)
				
				case "/user/create":
					if true && r.Header.Get("X-Auth")!="100500" {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("{\"error\": \"unauthorized\"}"))
						return
					}
					if r.Method != "POST"&&"POST"!="" {
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
			params ProfileParams
			queryParam	url.Values
			err error
		)
		reEnum := regexp.MustCompile("enum=([\\w\\|]+)")
		reMin := regexp.MustCompile("min=(\\d+)")
		reMax := regexp.MustCompile("max=(\\d+)")
		reParamname := regexp.MustCompile("paramname=(\\w+)")
		reDefault := regexp.MustCompile("default=(\\w+)")

		checkParamInt := func(tag string, paramName string) (res int, errRes error) {
			//optEnum := reEnum.FindStringSubmatch(tag)[1]
			var optMin, optMax, optParamname, optDefault string
			if match:=reMin.FindStringSubmatch(tag); len(match)==2{
				optMin = match[1]
			}
			if match:=reMax.FindStringSubmatch(tag); len(match)==2{
				optMax = match[1]
			}
			if match:=reParamname.FindStringSubmatch(tag); len(match)==2{
				optParamname = match[1]
			}
			if match:=reDefault.FindStringSubmatch(tag); len(match)==2{
				optDefault = match[1]
			}
			paramName = strings.ToLower(paramName)
			param:=""

			if optParamname != ""{
				param = queryParam.Get(optParamname)
			} else {
				param = queryParam.Get(paramName)
			}
			// если есть дефолт значение и параметр не передали
			if optDefault != "" && param==""{
				param = optDefault
			}
			if param!=""{
				if res, err = strconv.Atoi(param); err!=nil{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			if strings.Contains(tag, "required") && res==0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
				errRes = fmt.Errorf("error")
				return
			}
			// если есть ограничения min
			if optMin != "" {
				if minVal, err := strconv.Atoi(optMin); err!=nil ||  res < minVal{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", paramName, minVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			// если есть ограничения max
			if optMax != "" {
				if maxVal, err := strconv.Atoi(optMax); err!=nil ||  res > maxVal {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", paramName, maxVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			return
		}

		checkParamString := func(tag string, paramName string) (res string, errRes error) {
			tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
			tag, _ = strings.CutSuffix(tag, "\"")
			for _, option := range strings.Split(tag, ",") {
				var optMin, optParamname, optEnum, optDefault string
				if match := reMin.FindStringSubmatch(option); len(match) == 2 {
					optMin = match[1]
				}
				if match := reParamname.FindStringSubmatch(option); len(match) == 2 {
					optParamname = match[1]
				}
				if match := reDefault.FindStringSubmatch(option); len(match) == 2 {
					optDefault = match[1]
				}
				if match := reEnum.FindStringSubmatch(option); len(match) == 2 {
					optEnum = match[1]
				}
				paramName = strings.ToLower(paramName)
	
				if optParamname != "" {
					res = queryParam.Get(optParamname)
				} else {
					res = queryParam.Get(paramName)
				}
				// если есть дефолт значение и параметр не передали
				if optDefault != "" && res == "" {
					res = optDefault
				}
				if strings.Contains(tag, "required") && res == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
				// если есть ограничения min
				if optMin != "" {
					if minVal, err := strconv.Atoi(optMin); err != nil || len(res) < minVal {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", paramName, minVal)))
						errRes = fmt.Errorf("error")
						return
					}
				}
				if optEnum != "" && res != "" {
					ok := false
					for _, item := range strings.Split(optEnum, "|") {
						if res == item {
							ok = true
							break
						}
					}
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", paramName, strings.ReplaceAll(optEnum, "|", ", "))))
						errRes = fmt.Errorf("error")
						return
					}
				}
			}
			return
		}
		checkParamInt=checkParamInt
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
		
		if params.Login, err = checkParamString(`apivalidator:"required"`, "Login"); err!=nil{
			return
		}
		
		
		
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
			respMap := map[string]interface{} {"error": "", "response":res}
			resp, err := json.Marshal(&respMap)
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
			params CreateParams
			queryParam	url.Values
			err error
		)
		reEnum := regexp.MustCompile("enum=([\\w\\|]+)")
		reMin := regexp.MustCompile("min=(\\d+)")
		reMax := regexp.MustCompile("max=(\\d+)")
		reParamname := regexp.MustCompile("paramname=(\\w+)")
		reDefault := regexp.MustCompile("default=(\\w+)")

		checkParamInt := func(tag string, paramName string) (res int, errRes error) {
			//optEnum := reEnum.FindStringSubmatch(tag)[1]
			var optMin, optMax, optParamname, optDefault string
			if match:=reMin.FindStringSubmatch(tag); len(match)==2{
				optMin = match[1]
			}
			if match:=reMax.FindStringSubmatch(tag); len(match)==2{
				optMax = match[1]
			}
			if match:=reParamname.FindStringSubmatch(tag); len(match)==2{
				optParamname = match[1]
			}
			if match:=reDefault.FindStringSubmatch(tag); len(match)==2{
				optDefault = match[1]
			}
			paramName = strings.ToLower(paramName)
			param:=""

			if optParamname != ""{
				param = queryParam.Get(optParamname)
			} else {
				param = queryParam.Get(paramName)
			}
			// если есть дефолт значение и параметр не передали
			if optDefault != "" && param==""{
				param = optDefault
			}
			if param!=""{
				if res, err = strconv.Atoi(param); err!=nil{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			if strings.Contains(tag, "required") && res==0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
				errRes = fmt.Errorf("error")
				return
			}
			// если есть ограничения min
			if optMin != "" {
				if minVal, err := strconv.Atoi(optMin); err!=nil ||  res < minVal{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", paramName, minVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			// если есть ограничения max
			if optMax != "" {
				if maxVal, err := strconv.Atoi(optMax); err!=nil ||  res > maxVal {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", paramName, maxVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			return
		}

		checkParamString := func(tag string, paramName string) (res string, errRes error) {
			tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
			tag, _ = strings.CutSuffix(tag, "\"")
			for _, option := range strings.Split(tag, ",") {
				var optMin, optParamname, optEnum, optDefault string
				if match := reMin.FindStringSubmatch(option); len(match) == 2 {
					optMin = match[1]
				}
				if match := reParamname.FindStringSubmatch(option); len(match) == 2 {
					optParamname = match[1]
				}
				if match := reDefault.FindStringSubmatch(option); len(match) == 2 {
					optDefault = match[1]
				}
				if match := reEnum.FindStringSubmatch(option); len(match) == 2 {
					optEnum = match[1]
				}
				paramName = strings.ToLower(paramName)
	
				if optParamname != "" {
					res = queryParam.Get(optParamname)
				} else {
					res = queryParam.Get(paramName)
				}
				// если есть дефолт значение и параметр не передали
				if optDefault != "" && res == "" {
					res = optDefault
				}
				if strings.Contains(tag, "required") && res == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
				// если есть ограничения min
				if optMin != "" {
					if minVal, err := strconv.Atoi(optMin); err != nil || len(res) < minVal {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", paramName, minVal)))
						errRes = fmt.Errorf("error")
						return
					}
				}
				if optEnum != "" && res != "" {
					ok := false
					for _, item := range strings.Split(optEnum, "|") {
						if res == item {
							ok = true
							break
						}
					}
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", paramName, strings.ReplaceAll(optEnum, "|", ", "))))
						errRes = fmt.Errorf("error")
						return
					}
				}
			}
			return
		}
		checkParamInt=checkParamInt
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
		
		if params.Login, err = checkParamString(`apivalidator:"required,min=10"`, "Login"); err!=nil{
			return
		}
		
		if params.Name, err = checkParamString(`apivalidator:"paramname=full_name"`, "Name"); err!=nil{
			return
		}
		
		if params.Status, err = checkParamString(`apivalidator:"enum=user|moderator|admin,default=user"`, "Status"); err!=nil{
			return
		}
		
		if params.Age, err = checkParamInt(`apivalidator:"min=0,max=128"`, "Age"); err!=nil{
			return
		}
		
		
		
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
			respMap := map[string]interface{} {"error": "", "response":res}
			resp, err := json.Marshal(&respMap)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(resp)
		}
		
		func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
				
				case "/user/create":
					if true && r.Header.Get("X-Auth")!="100500" {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("{\"error\": \"unauthorized\"}"))
						return
					}
					if r.Method != "POST"&&"POST"!="" {
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
			params OtherCreateParams
			queryParam	url.Values
			err error
		)
		reEnum := regexp.MustCompile("enum=([\\w\\|]+)")
		reMin := regexp.MustCompile("min=(\\d+)")
		reMax := regexp.MustCompile("max=(\\d+)")
		reParamname := regexp.MustCompile("paramname=(\\w+)")
		reDefault := regexp.MustCompile("default=(\\w+)")

		checkParamInt := func(tag string, paramName string) (res int, errRes error) {
			//optEnum := reEnum.FindStringSubmatch(tag)[1]
			var optMin, optMax, optParamname, optDefault string
			if match:=reMin.FindStringSubmatch(tag); len(match)==2{
				optMin = match[1]
			}
			if match:=reMax.FindStringSubmatch(tag); len(match)==2{
				optMax = match[1]
			}
			if match:=reParamname.FindStringSubmatch(tag); len(match)==2{
				optParamname = match[1]
			}
			if match:=reDefault.FindStringSubmatch(tag); len(match)==2{
				optDefault = match[1]
			}
			paramName = strings.ToLower(paramName)
			param:=""

			if optParamname != ""{
				param = queryParam.Get(optParamname)
			} else {
				param = queryParam.Get(paramName)
			}
			// если есть дефолт значение и параметр не передали
			if optDefault != "" && param==""{
				param = optDefault
			}
			if param!=""{
				if res, err = strconv.Atoi(param); err!=nil{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			if strings.Contains(tag, "required") && res==0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
				errRes = fmt.Errorf("error")
				return
			}
			// если есть ограничения min
			if optMin != "" {
				if minVal, err := strconv.Atoi(optMin); err!=nil ||  res < minVal{
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", paramName, minVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			// если есть ограничения max
			if optMax != "" {
				if maxVal, err := strconv.Atoi(optMax); err!=nil ||  res > maxVal {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", paramName, maxVal)))
					errRes = fmt.Errorf("error")
					return
				}
			}
			return
		}

		checkParamString := func(tag string, paramName string) (res string, errRes error) {
			tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
			tag, _ = strings.CutSuffix(tag, "\"")
			for _, option := range strings.Split(tag, ",") {
				var optMin, optParamname, optEnum, optDefault string
				if match := reMin.FindStringSubmatch(option); len(match) == 2 {
					optMin = match[1]
				}
				if match := reParamname.FindStringSubmatch(option); len(match) == 2 {
					optParamname = match[1]
				}
				if match := reDefault.FindStringSubmatch(option); len(match) == 2 {
					optDefault = match[1]
				}
				if match := reEnum.FindStringSubmatch(option); len(match) == 2 {
					optEnum = match[1]
				}
				paramName = strings.ToLower(paramName)
	
				if optParamname != "" {
					res = queryParam.Get(optParamname)
				} else {
					res = queryParam.Get(paramName)
				}
				// если есть дефолт значение и параметр не передали
				if optDefault != "" && res == "" {
					res = optDefault
				}
				if strings.Contains(tag, "required") && res == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", paramName)))
					errRes = fmt.Errorf("error")
					return
				}
				// если есть ограничения min
				if optMin != "" {
					if minVal, err := strconv.Atoi(optMin); err != nil || len(res) < minVal {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", paramName, minVal)))
						errRes = fmt.Errorf("error")
						return
					}
				}
				if optEnum != "" && res != "" {
					ok := false
					for _, item := range strings.Split(optEnum, "|") {
						if res == item {
							ok = true
							break
						}
					}
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", paramName, strings.ReplaceAll(optEnum, "|", ", "))))
						errRes = fmt.Errorf("error")
						return
					}
				}
			}
			return
		}
		checkParamInt=checkParamInt
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
		
		if params.Username, err = checkParamString(`apivalidator:"required,min=3"`, "Username"); err!=nil{
			return
		}
		
		if params.Name, err = checkParamString(`apivalidator:"paramname=account_name"`, "Name"); err!=nil{
			return
		}
		
		if params.Class, err = checkParamString(`apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`, "Class"); err!=nil{
			return
		}
		
		if params.Level, err = checkParamInt(`apivalidator:"min=1,max=50"`, "Level"); err!=nil{
			return
		}
		
		
		
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
			respMap := map[string]interface{} {"error": "", "response":res}
			resp, err := json.Marshal(&respMap)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(resp)
		}
		