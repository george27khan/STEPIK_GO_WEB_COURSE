package main

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
	"strings"
	"text/template"
)

type funcOpts struct {
	Url    string
	Auth   bool
	Method string
}

type ParamsServeHTTP struct {
	RecieverType string
	FuncInfoArr  []FuncInfo
}

type FuncInfo struct {
	Path         string
	FuncName     string
	FuncInParam  [][]string
	FuncOutParam [][]string
}

type structInfo struct {
	Name       string
	Attributes []structAttrInfo
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
			case "{{.Path}}":
				h.wrapper{{.FuncName}}(w, r)
			{{end}}
			default:
				w.WriteHeader(http.StatusNotFound)
		}
		`))
	//шаблон тела функции ServeHTTP
	wrapFncTmpl = template.Must(template.New("wrapFncTmpl").Parse(`
		func (h {{.RecieverType}}) wrapper{{.FuncName}}() {
		// заполнение структуры params
		//		// валидирование параметров
			res, err := h.{{.FuncName}}(ctx, params)
		//		// прочие обработки
		`))

	//шаблон тела функции ServeHTTP
	createStructTmpl = template.Must(template.New("createStructTmpl").Parse(`
		var params {{.Name}}
		{{range .Attributes}}
            optsMap := tagToMap({{.Tag}})
			if {{.Type}} == "int" {
				// сперва проверка опции подмены имени атрибута
				if paramName, ok := optsMap["paramname"]; ok {
					params.{{.Name}}, _ = strconv.Atoi(r.Header.Get(paramName))
					delete(optsMap, "paramname") 
				} else {
					params.{{.Name}}, _ = strconv.Atoi(r.Header.Get(strings.ToLower({{.Name}})))
				}
				// далее проверяем остальные параметры
				for opt, val := range optsMap{
					if opt == "required" && val==0 {
						fmt.Println("Error")
					}
					// если есть дефолт значение и параметр принял дефолт значение
					if opt == "default" && val==0 {
						params.{{.Name}} = val
					}
					// если есть ограничения min
					if opt == "min"{
						if minVal, err := strconv.Atoi(val); err!=nil ||  params.{{.Name}} < minVal{
							fmt.Println("Error")
						}
					}
					// если есть ограничения max
					if opt == "max"{
						if maxVal, err := strconv.Atoi(val); err!=nil ||  params.{{.Name}} > maxVal {
							fmt.Println("Error")
						}
					}
				}
			} else if {{.Type}} == "string" {
				// сперва проверка опции подмены имени атрибута
				if paramName, ok := optsMap["paramname"]; ok {
					params.{{.Name}}, _ = r.Header.Get(paramName)
					delete(optsMap, "paramname") 
				} else {
					params.{{.Name}}, _ = r.Header.Get(strings.ToLower({{.Name}}))
				}
				// далее проверяем остальные параметры
				for opt, val := range optsMap{
					if opt == "required" && val==0 {
						fmt.Println("Error")
					}
					// если есть дефолт значение и параметр принял дефолт значение
					if opt == "default" && val=="" {
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
							if !ok {
								fmt.Println("Error")
							}
						}
					}
					// если есть ограничения min
					if opt == "min"{
						if minVal, err := strconv.Atoi(val); err!=nil || len(params.{{.Name}}) < minVal{
							fmt.Println("Error")
						}
					}
				}
			}
		{{end}}
		`))
)

func tagToMap(tag string) map[string]string {
	tag, _ = strings.CutPrefix(tag, "apivalidator:\"")
	tag, _ = strings.CutSuffix(tag, "\"")
	optsMap := make(map[string]string)
	for _, opt := range strings.Split(tag, ",") {
		optParts := strings.Split(opt, "=")
		optsMap[optParts[0]] = optParts[1]
	}
	fmt.Println(optsMap)
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
		opts     funcOpts
		APIparts map[string][]FuncInfo
	)
	ctx := context.Background()

	fset := token.NewFileSet() //Отслеживание позиций в исходном коде
	node, err := parser.ParseFile(fset, "api.go", nil, parser.ParseComments)
	if err != nil {
		slog.Log(ctx, slog.LevelError, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, `package `+node.Name.Name) //название пакета
	fmt.Fprintln(os.Stdout)                            // empty line

	structParts := make(map[string][]structAttrInfo)
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
					if _, ok := structParts[typeName]; ok {
						structParts[typeName] = append(structParts[typeName], structAttrInfo{attr.Names[0].String(), expr.Name, tag})
					} else {
						structParts[typeName] = []structAttrInfo{structAttrInfo{attr.Names[0].String(), expr.Name, tag}}
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
	//	createStructTmpl.Execute(os.Stdout, structInfo{name, info})
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
		fmt.Println(recvType)
		//массив хранения входящих параметров [название, тип]
		funcInParam := make([][]string, 0)
		//funcOut   := make([][]string,0)
		// разбор параметров функции
		fmt.Println("Входящие параметры:")
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
		fmt.Println(funcOutParam)

		if _, ok := APIparts[recvType]; ok {
			APIparts[recvType] = append(APIparts[recvType], FuncInfo{opts.Url, fnc.Name.String(), funcInParam, funcOutParam})
		} else {
			APIparts[recvType] = []FuncInfo{FuncInfo{opts.Url, fnc.Name.String(), funcInParam, funcOutParam}}
		}
		//fmt.Println(fnc.Name, fnc.Doc.Text(), fnc.Recv.List[0].Names[0], fnc.Recv.List[0].)
		//fmt.Printf("%+v\n", fnc.Recv.List[0].Type.(*ast.StarExpr).X)

		//формируем функцию ServeHTTP

		//fmt.Println("res: ", opts)
		//FUNC_LOOP:
		//	for _, block := range fnc.{
		//
		//	}
	}
	fmt.Println(APIparts)
	//for recvType, info := range APIparts {
	//	ServeHTTPtmpl.Execute(os.Stdout, ParamsServeHTTP{recvType, info})
	//}

	for recvType, info := range APIparts {
		ServeHTTPtmpl.Execute(os.Stdout, ParamsServeHTTP{recvType, info})
	}
}
