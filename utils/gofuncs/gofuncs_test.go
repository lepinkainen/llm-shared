package main

import (
	"go/parser"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			require.Equal(t, want, typeToString(node))
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
	require.NoError(t, err)

	funcs, err := extractFunctions(dir)
	require.NoError(t, err)
	require.Len(t, funcs, 2)

	helper := funcs[0]
	require.Equal(t, "sample.go", helper.File)
	require.Equal(t, 3, helper.Line)
	require.Equal(t, "helper", helper.Name)
	require.Equal(t, []string{"string"}, helper.Params)
	require.Equal(t, []string{"error"}, helper.Returns)

	method := funcs[1]
	require.Equal(t, "Do", method.Name)
	require.Equal(t, "m", method.Type)
	require.Equal(t, "*thing", method.Receiver)
	require.Equal(t, []string{"int"}, method.Params)
	require.Equal(t, []string{"string", "error"}, method.Returns)
	require.True(t, method.Exported)
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

	require.Equal(t, "sample.go:10:f:y:Main:(int)string", formatFunction(fn))

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
	require.Equal(t, "sample.go:5:m:n:Do:*Thing:(string,int)error", formatFunction(method))
}
