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
	Path     string
	FuncName string
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
			if {{.Type}} == "int" {
				params.{{.Name}}, _ = strconv.Atoi(r.Header.Get(strings.ToLower({{.Name}})))
			} else if {{.Type}} == "string" {
				params.{{.Name}} = r.Header.Get(strings.ToLower({{.Name}}))
			}
		{{end}}

		`))
)

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

	fmt.Println(strings.Split("enum=warrior|sorcerer|rouge,default=warrior", ","))
	return
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
	for name, info := range structParts {
		createStructTmpl.Execute(os.Stdout, structInfo{name, info})
	}
	return

	APIparts = make(map[string][]FuncInfo)
	for _, part := range node.Decls {
		fnc, ok := part.(*ast.FuncDecl) // Проверяем, является ли узел объявлением функции
		if !ok {
			//fmt.Printf("SKIP %#T is not *ast.FuncDecl\n", fnc)
			continue
		}

		comment := fnc.Doc.Text()
		if !strings.HasPrefix(comment, "apigen:api") {
			continue
		}
		strFuncOpts, _ := strings.CutPrefix(comment, "apigen:api ")
		if err := json.Unmarshal([]byte(strFuncOpts), &opts); err != nil {
			slog.Log(ctx, slog.LevelError, err.Error())
			return
		}
		//получатель метода
		recvType := ""
		if fnc.Recv != nil {
			// Получаем тип получателя
			recvType = ""
			switch expr := fnc.Recv.List[0].Type.(type) {
			case *ast.Ident:
				recvType = expr.Name
			case *ast.StarExpr:
				// Указатель на тип
				if ident, ok := expr.X.(*ast.Ident); ok {
					recvType = "*" + ident.Name
				}
			}
		}
		if _, ok := APIparts[recvType]; ok {
			APIparts[recvType] = append(APIparts[recvType], FuncInfo{opts.Url, fnc.Name.String()})
		} else {
			APIparts[recvType] = []FuncInfo{FuncInfo{opts.Url, fnc.Name.String()}}
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
	for recvType, info := range APIparts {
		ServeHTTPtmpl.Execute(os.Stdout, ParamsServeHTTP{recvType, info})
	}
}
