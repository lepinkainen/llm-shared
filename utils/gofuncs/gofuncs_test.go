package main

import (
	"go/parser"
	"os"
	"path/filepath"
	"testing"
)

func TestTypeToString(t *testing.T) {
	tests := map[string]string{
		"*int":                "*int",
		"[]string":            "[]string",
		"map[string]int":      "map[string]int",
		"chan<- bool":         "chan<- bool",
		"<-chan int":          "<-chan int",
		"struct{}":            "struct{}",
		"pkg.Type":            "pkg.Type",
		"func(a int) error":   "func(int)error",
		"func() (int, error)": "func()(int,error)",
	}

	for expr, want := range tests {
		t.Run(expr, func(t *testing.T) {
			node, err := parser.ParseExpr(expr)
			if err != nil {
				t.Fatalf("ParseExpr failed: %v", err)
			}
			got := typeToString(node)
			if got != want {
				t.Errorf("typeToString(%q) = %q, want %q", expr, got, want)
			}
		})
	}
}

func TestExtractFunctions(t *testing.T) {
	dir := t.TempDir()
	source := `package sample

func helper(a string) error { return nil }

type thing struct{}

func (t *thing) Do(x int) (string, error) { return "", nil }
`
	err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte(source), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	funcs, err := extractFunctions(dir)
	if err != nil {
		t.Fatalf("extractFunctions failed: %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("extractFunctions returned %d functions, want 2", len(funcs))
	}

	helper := funcs[0]
	if helper.File != "sample.go" {
		t.Errorf("helper.File = %q, want %q", helper.File, "sample.go")
	}
	if helper.Line != 3 {
		t.Errorf("helper.Line = %d, want 3", helper.Line)
	}
	if helper.Name != "helper" {
		t.Errorf("helper.Name = %q, want %q", helper.Name, "helper")
	}
	if len(helper.Params) != 1 || helper.Params[0] != "string" {
		t.Errorf("helper.Params = %v, want [string]", helper.Params)
	}
	if len(helper.Returns) != 1 || helper.Returns[0] != "error" {
		t.Errorf("helper.Returns = %v, want [error]", helper.Returns)
	}

	method := funcs[1]
	if method.Name != "Do" {
		t.Errorf("method.Name = %q, want %q", method.Name, "Do")
	}
	if method.Type != "m" {
		t.Errorf("method.Type = %q, want %q", method.Type, "m")
	}
	if method.Receiver != "*thing" {
		t.Errorf("method.Receiver = %q, want %q", method.Receiver, "*thing")
	}
	if len(method.Params) != 1 || method.Params[0] != "int" {
		t.Errorf("method.Params = %v, want [int]", method.Params)
	}
	if len(method.Returns) != 2 || method.Returns[0] != "string" || method.Returns[1] != "error" {
		t.Errorf("method.Returns = %v, want [string error]", method.Returns)
	}
	if !method.Exported {
		t.Errorf("method.Exported = false, want true")
	}
}

func TestFormatFunction(t *testing.T) {
	fn := FunctionInfo{
		File:     "sample.go",
		Line:     10,
		Type:     "f",
		Exported: true,
		Name:     "Main",
		Params:   []string{"int"},
		Returns:  []string{"string"},
	}

	got := formatFunction(fn)
	want := "sample.go:10:f:y:Main:(int)string"
	if got != want {
		t.Errorf("formatFunction(fn) = %q, want %q", got, want)
	}

	method := FunctionInfo{
		File:     "sample.go",
		Line:     5,
		Type:     "m",
		Exported: false,
		Name:     "Do",
		Receiver: "*Thing",
		Params:   []string{"string", "int"},
		Returns:  []string{"error"},
	}
	got = formatFunction(method)
	want = "sample.go:5:m:n:Do:*Thing:(string,int)error"
	if got != want {
		t.Errorf("formatFunction(method) = %q, want %q", got, want)
	}
}
