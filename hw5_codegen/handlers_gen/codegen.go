package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

type tpl struct {
	URL             string
	Auth            bool
	Method          string
	StructName      string
	FuncName        string
	SecondParamType string
	FirstResultType string
	APIValidator    string
}

type httpTpl struct {
	StructName string
	Cases      string
}

var (
	serverHTTPTpl = template.Must(template.New("serverHTTPTpl").Parse(`
func (h *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { {{.Cases}}
    default:
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "{\"error\": \"unknown method\"}")
    }
}
`))

	handlerTpl = template.Must(template.New("handlerTpl").Parse(`
func (h *{{.StructName}}) handler{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
{{.APIValidator}}
}
`))
)

func getAPIValidator(funcName string, paramType string, resultType string, method string, auth bool, node *ast.File) (result string) {
	result = result + `	field := ""
`
	for _, s := range node.Decls {
		g, ok := s.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", s)
			continue
		}
		if g.Tok.String() != "type" {
			fmt.Println("SKIP token Tok is not type \"type\"")
			continue
		}
		if g.Specs[0].(*ast.TypeSpec).Name.Name == paramType {
			if auth {
				result = result + `	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "{\"error\": \"unauthorized\"}")
		return
	}
`
			}
			if method != "" {
				result = result + `	if r.Method != "` + method + `" {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, "{\"error\": \"bad method\"}")
		return
	}			
`
			}
			result = result + `	s := ` + paramType + "{}\n"
			for _, field := range g.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
				paramsAPIValidation := make(map[string]string)
				paramsAPIValidation["field_name"] = field.Names[0].Name
				paramsAPIValidation["field_type"] = field.Type.(*ast.Ident).Name
				field.Tag.Value = strings.Replace(field.Tag.Value, "\"`", "", 1)
				strs := strings.Split(strings.Replace(field.Tag.Value, "`apivalidator:\"", "", 1), ",")
				for _, str := range strs {
					param := strings.Split(str, "=")
					if len(param) == 2 {
						paramsAPIValidation[param[0]] = param[1]
					} else {
						paramsAPIValidation[param[0]] = ""
					}
				}
				fieldName := strings.ToLower(paramsAPIValidation["field_name"])
				if val, ok := paramsAPIValidation["paramname"]; ok {
					fieldName = val
				}
				result = result + `	field = r.FormValue("` + fieldName + `")
`
				if val, ok := paramsAPIValidation["default"]; ok {
					result = result + `	if field == "" {
		field = "` + val + `"
	}
`
				}
				if paramsAPIValidation["field_type"] == "int" {
					result = result + `	intField, err := strconv.Atoi(field)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` must be int\"}")
		return		
	}
	s.` + paramsAPIValidation["field_name"] + ` = intField
`
				} else {
					result = result + `	s.` + paramsAPIValidation["field_name"] + ` = field
`
				}
				if _, ok := paramsAPIValidation["required"]; ok {
					result = result + `	if s.` + paramsAPIValidation["field_name"] + ` == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` must me not empty\"}")
		return
	}
`
				}
				//fmt.Printf("\n %#v", paramsAPIValidation)
				if val, ok := paramsAPIValidation["min"]; ok && paramsAPIValidation["field_type"] == "string" {
					result = result + `	if len(s.` + paramsAPIValidation["field_name"] + `) < ` + val + ` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` len must be >= ` + val + `\"}")
		return
	}
`
				}
				if val, ok := paramsAPIValidation["min"]; ok && paramsAPIValidation["field_type"] == "int" {
					result = result + `	if s.` + paramsAPIValidation["field_name"] + ` < ` + val + ` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` must be >= ` + val + `\"}")
		return
	}
`
				}
				if val, ok := paramsAPIValidation["max"]; ok && paramsAPIValidation["field_type"] == "int" {
					result = result + `	if s.` + paramsAPIValidation["field_name"] + ` > ` + val + ` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` must be <= ` + val + `\"}")
		return
	}
`
				}
				if val, ok := paramsAPIValidation["enum"]; ok {
					enum := strings.Split(val, "|")
					enumCondition := ""
					strError := ""
					for k, v := range enum {
						if k == 0 {
							enumCondition = enumCondition + `s.` + paramsAPIValidation["field_name"] + ` != "` + v + `"`
							strError = strError + v
						} else {
							enumCondition = enumCondition + ` && s.` + paramsAPIValidation["field_name"] + ` != "` + v + `"`
							strError = strError + `, ` + v
						}
					}
					result = result + `	if ` + enumCondition + ` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"` + strings.ToLower(paramsAPIValidation["field_name"]) + ` must be one of [` + strError + `]\"}")
		return
	}
`
				}
			}
		}
	}
	result = result + `	var ctx context.Context
	var` + resultType + `, err := h.` + funcName + `(ctx, s)
	if err != nil {
		switch err.(type) {
		case ApiError:
			//fmt.Printf("\nAAAAAAA %T\n", v)
			err := err.(ApiError)
			w.WriteHeader(err.HTTPStatus)
			io.WriteString(w, "{\"error\": \""+err.Error()+"\"}")
			return
		default:
			//fmt.Printf("\nBBBBBBBBB %T\n", v)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "{\"error\": \"bad user\"}")
			return
		}
	}
	answer, err := json.Marshal(map[string]interface{}{"error": "", "response": var` + resultType + `})
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(answer)`
	return
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	out, _ := os.Create(os.Args[2])
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"	
)`)

	httpTpls := make(map[string]*httpTpl)

	for _, f := range node.Decls {
		g, ok := f.(*ast.FuncDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.FuncDecl\n", f)
			continue
		}
		if g.Doc == nil {
			fmt.Println("SKIP.", g.Name.Name, "method dosn't have comment")
			continue
		}

		if !strings.HasPrefix(g.Doc.List[0].Text, "// apigen:api") {
			fmt.Printf("SKIP struct %#v doesnt have apigen mark\n", g.Name.Name)
			continue
		}

		structName := g.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		funcName := g.Name.Name
		resultType := g.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		paramType := g.Type.Params.List[1].Type.(*ast.Ident).Name
		t := &tpl{
			"",
			false,
			"",
			structName,
			funcName,
			paramType,
			resultType,
			"",
		}
		str := strings.Replace(g.Doc.List[0].Text, "// apigen:api ", "", 1)
		json.Unmarshal([]byte(str), t)
		t.APIValidator = getAPIValidator(funcName, paramType, resultType, t.Method, t.Auth, node)
		//fmt.Printf("%#v", t)
		if httpTpls[structName] == nil {
			httpTpls[structName] = &httpTpl{}
		}
		httpTpls[structName].StructName = structName
		httpTpls[structName].Cases = httpTpls[structName].Cases + "\n\tcase \"" + t.URL + "\":\n" +
			"\t\th.handler" + t.FuncName + "(w, r)"
		handlerTpl.Execute(out, t)
	}

	for _, v := range httpTpls {
		serverHTTPTpl.Execute(out, v)
		//fmt.Printf("struct:\n\t%#v\n\n", v)
	}
}
