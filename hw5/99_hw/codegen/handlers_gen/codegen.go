package main

// go build .\handlers_gen\codegen.go создание exe файла
// .\codegen.exe api.go api_handlers.go генерация файла из апи через генератор
// код писать тут

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	_ "reflect"
	"strings"
	"text/template"
)

type ParamsServeHTTP struct {
	RecieverType string
	FuncInfoArr  []FuncInfo
}
type ParamsWrapFunc struct {
	RecieverType string
	FuncName     string
	CreateParams string
}
type FuncAPIOpts struct {
	URL    string
	Auth   bool
	Method string
}
type FuncInfo struct {
	FuncAPIOpts
	FuncName     string
	FuncInParam  [][]string
	FuncOutParam [][]string
}

type structInfo struct {
	Name          string
	IntAttributes []structAttrInfo
	StrAttributes []structAttrInfo
}
type structAttrInfo struct {
	Name string
	Type string
	Tag  string
}

var (
	// шаблон сигнатуры функции ServeHTTP
	ServeHTTPtmpl = template.Must(template.New("ServeHTTPtmpl").Parse(`
		func (h {{.RecieverType}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
				{{range .FuncInfoArr}}
				case "{{.URL}}":
					if {{.Auth}} && r.Header.Get("X-Auth")!="100500" {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("{\"error\": \"unauthorized\"}"))
						return
					}
					if r.Method != "{{.Method}}"&&"{{.Method}}"!="" {
						w.WriteHeader(http.StatusNotAcceptable)
						w.Write([]byte("{\"error\": \"bad method\"}"))
						return
					}
					h.wrapper{{.FuncName}}(w, r)
				{{end}}
				default:
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("{\"error\": \"unknown method\"}"))
			}
		}
		`))
	//шаблон тела функции ServeHTTP
	wrapFncTmpl = template.Must(template.New("wrapFncTmpl").Parse(`
		func (h {{.RecieverType}}) wrapper{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		// заполнение структуры params
		// валидирование параметров
			{{.CreateParams}}
			fmt.Println("in params", params)

			res, err := h.{{.FuncName}}(ctx, params)
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
			fmt.Println("resp", resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(resp)
		}
		`))

	//шаблон тела функции ServeHTTP
	createStructTmpl = template.Must(template.New("createStructTmpl").Parse(`
		var (
			params {{.Name}}
			optsMap  map[string]string
			queryParam	url.Values
			err error
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
		{{range .IntAttributes}}
            optsMap = TagToMap({{.Tag}})
			param{{.Name}}:=""
			// сперва проверка опции подмены имени атрибута
			if paramName, ok := optsMap["paramname"]; ok {
				param{{.Name}} = queryParam.Get(paramName)
				delete(optsMap, "paramname") 
			} else {
				param{{.Name}} =  queryParam.Get(strings.ToLower("{{.Name}}"))
				delete(optsMap, "paramname") 
			}
			if params.{{.Name}}, err = strconv.Atoi(param{{.Name}}); err!=nil{
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be int\"}", strings.ToLower("{{.Name}}"))))
				return
			}

			// далее проверяем остальные параметры
			for opt, val := range optsMap{
				if opt == "required" && params.{{.Name}}==0 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("{{.Name}}"))))
					return
				}
				// если есть дефолт значение и параметр принял дефолт значение
				if opt == "default" && params.{{.Name}}==0 {
					params.{{.Name}},_ = strconv.Atoi(val)
				}
				// если есть ограничения min
				if opt == "min"{
					if minVal, err := strconv.Atoi(val); err!=nil ||  params.{{.Name}} < minVal{
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be >= %v\"}", strings.ToLower("{{.Name}}"),minVal)))
						return
					}
				}
				// если есть ограничения max
				if opt == "max"{
					if maxVal, err := strconv.Atoi(val); err!=nil ||  params.{{.Name}} > maxVal {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be <= %v\"}", strings.ToLower("{{.Name}}"),maxVal)))
						return
					}
				}
			}
		{{end}}
		{{range .StrAttributes}}
			optsMap = TagToMap({{.Tag}})
			// сперва проверка опции подмены имени атрибута
			if paramName, ok := optsMap["paramname"]; ok {
				params.{{.Name}} = queryParam.Get(paramName)
				delete(optsMap, "paramname") 
			} else {
				params.{{.Name}} = queryParam.Get(strings.ToLower("{{.Name}}"))
			}
			// далее проверяем остальные параметры
			for opt, val := range optsMap{
				if opt == "required" && params.{{.Name}}=="" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must me not empty\"}", strings.ToLower("{{.Name}}"))))
					return
				}
				// если есть дефолт значение и параметр принял дефолт значение
				if opt == "default" && params.{{.Name}}=="" {
					params.{{.Name}} = val
				}
				// если есть ограничения по значениям
				if opt == "enum"{
					ok := false
					for _, item := range strings.Split(val, "|") {
						if params.{{.Name}} == item {
							ok = true
							break
						}
					}
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s must be one of [%s]\"}", strings.ToLower("{{.Name}}"), strings.ReplaceAll(val,"|",", "))))	
						return
					}
				}
				// если есть ограничения min
				if opt == "min"{
					if minVal, err := strconv.Atoi(val); err!=nil || len(params.{{.Name}}) < minVal{
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("{\"error\": \"%s len must be >= %v\"}", strings.ToLower("{{.Name}}"),minVal)))
						return
					}
				}
			}
		{{end}}
		`))
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
	return optsMap
}

func checkParamTag(param interface{}, tag string) {
	//required - поле не должно быть пустым (не должно иметь значение по-умолчанию)
	//paramname - если указано - то брать из параметра с этим именем, иначе lowercase от имени
	//enum - "одно из"
	//default - если указано и приходит пустое значение (значение по-умолчанию) - устанавливать то что написано указано в default
	//min - >= X для типа int, для строк len(str) >=
	//max - <= X для типа int

	var (
		attrType, valStr string
		valInt           int
	)
	switch tp := param.(type) {
	case int:
		attrType = "int"
		valInt = param.(int)
		tp = tp
	case string:
		attrType = "string"
		valStr = param.(string)
	}
	for _, opt := range strings.Split(tag, ",") {
		if opt == "required" {
			if attrType == "int" && valInt == 0 || attrType == "string" && valStr == "" {
				fmt.Println("Error")
			}
		}
		if opt == "required" {
			if attrType == "int" && valInt == 0 || attrType == "string" && valStr == "" {
				fmt.Println("Error")
			}
		}
	}
}

//func ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	r.Header.Get()
//	strconv.Atoi()
//}

//	func (h *SomeStructName ) wrapperDoSomeJob() {
//		// заполнение структуры params
//		// валидирование параметров
//		res, err := h.DoSomeJob(ctx, params)
//		// прочие обработки
//	}
func main() {
	var (
		opts     FuncAPIOpts
		APIparts map[string][]FuncInfo
	)
	//fmt.Println(fmt.Sprintf("\"error\": \"%s must me not empty\"}", strings.ToLower("Age")))
	//return
	//t := TagToMap(`apivalidator:"required"`)
	//fmt.Println(t["required"])
	//return
	ctx := context.Background()

	fset := token.NewFileSet() //Отслеживание позиций в исходном коде
	node, err := parser.ParseFile(fset, "api.go", nil, parser.ParseComments)
	//node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		slog.Log(ctx, slog.LevelError, err.Error())
		return
	}
	outFile, _ := os.Create(os.Args[2])
	//outFile, _ := os.Create("111.go")
	out := outFile
	//out := os.Stdout

	fmt.Fprintln(out, `package `+node.Name.Name) //название пакета
	fmt.Fprintln(out, `import (
		"context"
		"fmt"
		"net/http"
		"strconv"
		"strings"
		"encoding/json"
		"io"
		"net/url"
	)`) //название пакета
	fmt.Fprintln(out) // empty line

	fmt.Fprintln(out, `func TagToMap(tag string) map[string]string {
		tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
		tag, _ = strings.CutSuffix(tag, "\"")
		optsMap := make(map[string]string)
		for _, opt := range strings.Split(tag, ",") {
			optParts := strings.Split(opt, "=")
			if len(optParts)>1{
				optsMap[optParts[0]] = optParts[1]
			} else {
				optsMap[optParts[0]]=""
			}
		}
		fmt.Println(optsMap)
		return optsMap
	}`) //название пакета

	structPartsInt := make(map[string][]structAttrInfo)
	structPartsStr := make(map[string][]structAttrInfo)
	//цикл для сбора информации по структурам
	for _, part := range node.Decls {
		g, ok := part.(*ast.GenDecl) // Проверяем, является ли узел для типов
		if !ok {
			//fmt.Printf("SKIP %#T is not *ast.FuncDecl\n", fnc)
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				//fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
				continue
			}
			typeName := currType.Name.String()
			for _, attr := range currType.Type.(*ast.StructType).Fields.List {
				switch expr := attr.Type.(type) {
				case *ast.Ident:
					//fmt.Printf("type: %T data: %+v\n", expr., expr.Name)
					tag := ""
					if attr.Tag != nil {
						tag = attr.Tag.Value
					}
					if expr.Name == "int" {
						if _, ok := structPartsInt[typeName]; ok {
							structPartsInt[typeName] = append(structPartsInt[typeName], structAttrInfo{attr.Names[0].String(), expr.Name, tag})
						} else {
							structPartsInt[typeName] = []structAttrInfo{structAttrInfo{attr.Names[0].String(), expr.Name, tag}}
						}
					}
					if expr.Name == "string" {
						if _, ok := structPartsStr[typeName]; ok {
							structPartsStr[typeName] = append(structPartsStr[typeName], structAttrInfo{attr.Names[0].String(), expr.Name, tag})
						} else {
							structPartsStr[typeName] = []structAttrInfo{structAttrInfo{attr.Names[0].String(), expr.Name, tag}}
						}
					}

				}
			}
			//fmt.Println(structParts)
			//fmt.Println(len(currType.Type.(*ast.StructType).Fields.List))
			//spec.(*ast.TypeSpec) + currType.Type.(*ast.StructType) -
		}
	}
	//fmt.Println(structParts)
	//for name, info := range structParts {
	//	createStructTmpl.Execute(out, structInfo{name, info})
	//}

	APIparts = make(map[string][]FuncInfo)
	//цикл для сбора информации по функциям
	for _, part := range node.Decls {
		fnc, ok := part.(*ast.FuncDecl) // Проверяем, является ли узел объявлением функции
		if !ok {
			//fmt.Printf("SKIP %#T is not *ast.FuncDecl\n", fnc)
			continue
		}

		comment := fnc.Doc.Text()
		//если у функции нет указания на генерацию апи, то пропускаем
		if !strings.HasPrefix(comment, "apigen:api") {
			continue
		}

		//разбираем указания генерации
		strFuncOpts, _ := strings.CutPrefix(comment, "apigen:api ")
		if err := json.Unmarshal([]byte(strFuncOpts), &opts); err != nil {
			slog.Log(ctx, slog.LevelError, err.Error())
			return
		}

		//анализ получателя метода
		recvType := ""
		if fnc.Recv != nil {
			// Получаем тип получателя
			recvType = ""
			switch expr := fnc.Recv.List[0].Type.(type) {
			case *ast.Ident:
				// простой тип
				recvType = expr.Name
			case *ast.StarExpr:
				// Указатель на тип
				if ident, ok := expr.X.(*ast.Ident); ok {
					recvType = "*" + ident.Name
				}
			}
		}
		//fmt.Println(recvType)
		//массив хранения входящих параметров [название, тип]
		funcInParam := make([][]string, 0)
		//funcOut   := make([][]string,0)
		// разбор параметров функции
		for _, param := range fnc.Type.Params.List {
			switch expr := param.Type.(type) {
			case *ast.Ident:
				// простой тип
				funcInParam = append(funcInParam, []string{param.Names[0].Name, expr.Name})
			case *ast.StarExpr:
				// Указатель на тип
				funcInParam = append(funcInParam, []string{param.Names[0].Name, "*" + expr.X.(*ast.Ident).Name})
			case *ast.SelectorExpr:
				// для контекста, и скорее всего для интерфейсынх типов, не проверялдо конца
				funcInParam = append(funcInParam, []string{param.Names[0].Name, expr.X.(*ast.Ident).Name + "." + expr.Sel.Name})
			}
			//for _, name := range param.Names {
			//	fmt.Printf("  %s %s\n", name.Name, param.Type.(*ast.SelectorExpr).Sel)
			//}
		}
		//массив хранения исходящих параметров [название(если есть), тип]
		funcOutParam := make([][]string, 0)
		for _, param := range fnc.Type.Results.List {
			resultRow := []string{"", ""}
			if len(param.Names) > 0 {
				resultRow[0] = param.Names[0].Name
			}
			switch expr := param.Type.(type) {
			case *ast.Ident:
				// простой тип
				resultRow[1] = expr.Name
				funcOutParam = append(funcOutParam, resultRow)
			case *ast.StarExpr:
				// Указатель на тип
				resultRow[1] = "*" + expr.X.(*ast.Ident).Name
				funcOutParam = append(funcOutParam, resultRow)
			case *ast.SelectorExpr:
				// для контекста, и скорее всего для интерфейсынх типов, не проверялдо конца
				resultRow[1] = expr.X.(*ast.Ident).Name + "." + expr.Sel.Name
				funcOutParam = append(funcOutParam, resultRow)
			}
		}
		//fmt.Println(funcOutParam)

		if _, ok := APIparts[recvType]; ok {
			APIparts[recvType] = append(APIparts[recvType], FuncInfo{opts, fnc.Name.String(), funcInParam, funcOutParam})
		} else {
			APIparts[recvType] = []FuncInfo{FuncInfo{opts, fnc.Name.String(), funcInParam, funcOutParam}}
		}
		//fmt.Println(fnc.Name, fnc.Doc.Text(), fnc.Recv.List[0].Names[0], fnc.Recv.List[0].)
		//fmt.Printf("%+v\n", fnc.Recv.List[0].Type.(*ast.StarExpr).X)
	}
	fmt.Println(APIparts)

	for recvType, info := range APIparts {
		//генерация функции ServeHTTP для структуры
		ServeHTTPtmpl.Execute(out, ParamsServeHTTP{recvType, info})
		//генерация оберток для методов структуры
		for _, funcInfo := range info {
			var checkParamText strings.Builder
			//fmt.Println(funcInfo.FuncInParam[1][1])
			structName := strings.TrimLeft(funcInfo.FuncInParam[1][1], "*")

			createStructTmpl.Execute(&checkParamText, structInfo{structName, structPartsInt[structName], structPartsStr[structName]})
			//fmt.Println(checkParamText.String())
			wrapFncTmpl.Execute(out, ParamsWrapFunc{recvType, funcInfo.FuncName, checkParamText.String()})
		}

	}
}
