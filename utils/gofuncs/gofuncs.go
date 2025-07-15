package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FunctionInfo struct {
	File     string
	Line     int
	Type     string
	Exported bool
	Name     string
	Receiver string
	Params   []string
	Returns  []string
}

func main() {
	dirFlag := flag.String("dir", ".", "Directory to scan for Go files")
	flag.Parse()

	functions, err := extractFunctions(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	sort.Slice(functions, func(i, j int) bool {
		if functions[i].File != functions[j].File {
			return functions[i].File < functions[j].File
		}
		return functions[i].Line < functions[j].Line
	})

	for _, fn := range functions {
		fmt.Println(formatFunction(fn))
	}
}

func extractFunctions(dir string) ([]FunctionInfo, error) {
	var functions []FunctionInfo
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		node, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name == nil {
					return true
				}

				pos := fset.Position(x.Pos())
				fn := FunctionInfo{
					File:     relPath,
					Line:     pos.Line,
					Name:     x.Name.Name,
					Exported: ast.IsExported(x.Name.Name),
				}

				if x.Recv != nil && len(x.Recv.List) > 0 {
					fn.Type = "m"
					fn.Receiver = typeToString(x.Recv.List[0].Type)
				} else {
					fn.Type = "f"
				}

				if x.Type.Params != nil {
					for _, param := range x.Type.Params.List {
						paramType := typeToString(param.Type)
						if len(param.Names) == 0 {
							fn.Params = append(fn.Params, paramType)
						} else {
							for range param.Names {
								fn.Params = append(fn.Params, paramType)
							}
						}
					}
				}

				if x.Type.Results != nil {
					for _, result := range x.Type.Results.List {
						resultType := typeToString(result.Type)
						if len(result.Names) == 0 {
							fn.Returns = append(fn.Returns, resultType)
						} else {
							for range result.Names {
								fn.Returns = append(fn.Returns, resultType)
							}
						}
					}
				}

				functions = append(functions, fn)
			}
			return true
		})

		return nil
	})

	return functions, err
}

func typeToString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + typeToString(x.X)
	case *ast.ArrayType:
		if x.Len == nil {
			return "[]" + typeToString(x.Elt)
		}
		return "[" + typeToString(x.Len) + "]" + typeToString(x.Elt)
	case *ast.MapType:
		return "map[" + typeToString(x.Key) + "]" + typeToString(x.Value)
	case *ast.ChanType:
		switch x.Dir {
		case ast.SEND:
			return "chan<- " + typeToString(x.Value)
		case ast.RECV:
			return "<-chan " + typeToString(x.Value)
		default:
			return "chan " + typeToString(x.Value)
		}
	case *ast.FuncType:
		return "func" + buildFuncSignature(x)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.SelectorExpr:
		return typeToString(x.X) + "." + x.Sel.Name
	case *ast.Ellipsis:
		return "..." + typeToString(x.Elt)
	default:
		return "unknown"
	}
}

func buildFuncSignature(ft *ast.FuncType) string {
	var params, results []string

	if ft.Params != nil {
		for _, param := range ft.Params.List {
			paramType := typeToString(param.Type)
			if len(param.Names) == 0 {
				params = append(params, paramType)
			} else {
				for range param.Names {
					params = append(params, paramType)
				}
			}
		}
	}

	if ft.Results != nil {
		for _, result := range ft.Results.List {
			resultType := typeToString(result.Type)
			if len(result.Names) == 0 {
				results = append(results, resultType)
			} else {
				for range result.Names {
					results = append(results, resultType)
				}
			}
		}
	}

	sig := "(" + strings.Join(params, ",") + ")"
	if len(results) > 0 {
		if len(results) == 1 {
			sig += results[0]
		} else {
			sig += "(" + strings.Join(results, ",") + ")"
		}
	}

	return sig
}

func formatFunction(fn FunctionInfo) string {
	exported := "n"
	if fn.Exported {
		exported = "y"
	}

	signature := "(" + strings.Join(fn.Params, ",") + ")"
	if len(fn.Returns) > 0 {
		if len(fn.Returns) == 1 {
			signature += fn.Returns[0]
		} else {
			signature += "(" + strings.Join(fn.Returns, ",") + ")"
		}
	}

	if fn.Type == "m" {
		return fmt.Sprintf("%s:%d:%s:%s:%s:%s:%s", fn.File, fn.Line, fn.Type, exported, fn.Name, fn.Receiver, signature)
	}
	return fmt.Sprintf("%s:%d:%s:%s:%s:%s", fn.File, fn.Line, fn.Type, exported, fn.Name, signature)
}
