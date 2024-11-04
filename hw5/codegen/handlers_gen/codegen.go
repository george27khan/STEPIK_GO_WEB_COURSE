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
		{{range .Attributes}}
		if params.{{.Name}}, err = checkParam{{.Type}}({{.Tag}}, "{{.Name}}"); err!=nil{
			return
		}
		{{end}}
		
		`))
)

// Функция для перевода первых букв каждого слова в верхний регистр
func toTitleCase(s string) string {
	// Разбиваем строку на слова
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			// Преобразуем первую букву в верхний регистр и добавляем оставшуюся часть слова
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	// Соединяем слова обратно в строку
	return strings.Join(words, " ")
}

func main() {
	var (
		opts     FuncAPIOpts
		APIparts map[string][]FuncInfo
	)

	ctx := context.Background()

	fset := token.NewFileSet() //Отслеживание позиций в исходном коде
	//разбор входящих аргументов
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		slog.Log(ctx, slog.LevelError, err.Error())
		return
	}
	outFile, _ := os.Create(os.Args[2])
	out := outFile //os.Stdout

	//пакет
	fmt.Fprintln(out, `package `+node.Name.Name)
	//импорты
	fmt.Fprintln(out, `import (
		"context"
		"fmt"
		"net/http"
		"strconv"
		"strings"
		"encoding/json"
		"io"
		"net/url"
		"regexp"
	)`)

	fmt.Fprintln(out) // empty line

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
						structParts[typeName] = append(structParts[typeName], structAttrInfo{attr.Names[0].String(), toTitleCase(expr.Name), tag})
					} else {
						structParts[typeName] = []structAttrInfo{structAttrInfo{attr.Names[0].String(), toTitleCase(expr.Name), tag}}
					}
				}
			}
		}
	}

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
		// разбор параметров функции
		for _, param := range fnc.Type.Params.List {
			switch expr := param.Type.(type) {
			case *ast.Ident: // простой тип
				funcInParam = append(funcInParam, []string{param.Names[0].Name, expr.Name})
			case *ast.StarExpr: // Указатель на тип
				funcInParam = append(funcInParam, []string{param.Names[0].Name, "*" + expr.X.(*ast.Ident).Name})
			case *ast.SelectorExpr: // для контекста, и скорее всего для интерфейсынх типов, не проверялдо конца
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

		if _, ok := APIparts[recvType]; ok {
			APIparts[recvType] = append(APIparts[recvType], FuncInfo{opts, fnc.Name.String(), funcInParam, funcOutParam})
		} else {
			APIparts[recvType] = []FuncInfo{FuncInfo{opts, fnc.Name.String(), funcInParam, funcOutParam}}
		}
		//fmt.Println(fnc.Name, fnc.Doc.Text(), fnc.Recv.List[0].Names[0], fnc.Recv.List[0].)
		//fmt.Printf("%+v\n", fnc.Recv.List[0].Type.(*ast.StarExpr).X)
	}

	//собираем код из шаблонов
	for recvType, info := range APIparts {
		//генерация функции ServeHTTP для основной структуры
		ServeHTTPtmpl.Execute(out, ParamsServeHTTP{recvType, info})

		//генерация оберток для методов структуры
		for _, funcInfo := range info {
			var checkParamText strings.Builder
			structName := strings.TrimLeft(funcInfo.FuncInParam[1][1], "*")

			createStructTmpl.Execute(&checkParamText, structInfo{structName, structParts[structName]})
			wrapFncTmpl.Execute(out, ParamsWrapFunc{recvType, funcInfo.FuncName, checkParamText.String()})
		}

	}
}
