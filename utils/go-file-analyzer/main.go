package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fileStat struct {
	Path          string
	Lines         int
	Types         []string
	Funcs         []funcStat
	Buckets       buckets
	MethodCounts  map[string]int
	TopLevelFuncs int
	Notes         []string
}

type funcStat struct {
	Name  string
	Lines int
	Stmts int
}

type buckets struct {
	Over200    int
	Between120 int
	Between80  int
	Under80    int
}

func main() {
	rootDir := flag.String("dir", ".", "Directory to scan for Go files")
	topN := flag.Int("n", 20, "Number of files to display")
	topFuncs := flag.Int("topfuncs", 5, "Number of functions to list per file")
	flag.Parse()

	files, err := collectGoFiles(*rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No Go files found")
		return
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Lines == files[j].Lines {
			return files[i].Path < files[j].Path
		}
		return files[i].Lines > files[j].Lines
	})

	limit := *topN
	if limit <= 0 || limit > len(files) {
		limit = len(files)
	}

	for i := 0; i < limit; i++ {
		printFileReport(files[i], *topFuncs)
	}
}

func collectGoFiles(root string) ([]fileStat, error) {
	var results []fileStat
	gitIgnoreDirs := loadGitIgnoreDirs(root)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			relDir, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if shouldSkipDir(d.Name(), relDir, gitIgnoreDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		lines, err := countLines(path)
		if err != nil {
			return err
		}

		funcs, types, buckets, methodCounts, topLevelFuncs, notes, err := analyzeFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		results = append(results, fileStat{
			Path:          rel,
			Lines:         lines,
			Types:         types,
			Funcs:         funcs,
			Buckets:       buckets,
			MethodCounts:  methodCounts,
			TopLevelFuncs: topLevelFuncs,
			Notes:         notes,
		})

		return nil
	})

	return results, err
}

func shouldSkipDir(name, relPath string, gitIgnoreDirs []string) bool {
	switch name {
	case ".git", "vendor", "node_modules", "dist", "build", "coverage":
		return true
	default:
		for _, dir := range gitIgnoreDirs {
			if relPath == dir || strings.HasPrefix(relPath, dir+string(os.PathSeparator)) {
				return true
			}
		}
		return false
	}
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	count := 0
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func analyzeFile(path string) ([]funcStat, []string, buckets, map[string]int, int, []string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, buckets{}, nil, 0, nil, err
	}

	var funcs []funcStat
	typeSet := map[string]struct{}{}
	methodCounts := map[string]int{}
	topLevelFuncs := 0
	buckets := buckets{}
	imports := collectImports(file)

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				typeSet[ts.Name.Name] = struct{}{}
			}
		case *ast.FuncDecl:
			if d.Body == nil {
				continue
			}
			start := fset.Position(d.Pos()).Line
			end := fset.Position(d.End()).Line
			lines := end - start + 1
			stmts := len(d.Body.List)

			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recv := types.ExprString(d.Recv.List[0].Type)
				name = fmt.Sprintf("(%s) %s", recv, name)
				methodCounts[recv]++
			} else {
				topLevelFuncs++
			}

			funcs = append(funcs, funcStat{Name: name, Lines: lines, Stmts: stmts})
			addToBuckets(&buckets, lines)
		}
	}

	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	sort.Slice(funcs, func(i, j int) bool {
		if funcs[i].Lines == funcs[j].Lines {
			return funcs[i].Name < funcs[j].Name
		}
		return funcs[i].Lines > funcs[j].Lines
	})

	notes := buildNotes(types, buckets, imports)

	return funcs, types, buckets, methodCounts, topLevelFuncs, notes, nil
}

func addToBuckets(b *buckets, lines int) {
	switch {
	case lines > 200:
		b.Over200++
	case lines >= 120:
		b.Between120++
	case lines >= 80:
		b.Between80++
	default:
		b.Under80++
	}
}

func loadGitIgnoreDirs(root string) []string {
	path := filepath.Join(root, ".gitignore")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	var dirs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasSuffix(line, "/") {
			continue
		}

		trimmed := strings.TrimPrefix(line, "/")
		trimmed = strings.TrimSuffix(trimmed, "/")
		clean := filepath.Clean(trimmed)
		if clean != "." && clean != "" {
			dirs = append(dirs, clean)
		}
	}

	return dirs
}

func printFileReport(f fileStat, topFuncs int) {
	fmt.Printf("%8d %s\n", f.Lines, f.Path)

	fmt.Printf("    types: %d", len(f.Types))
	if len(f.Types) > 0 {
		fmt.Printf(" (%s)", strings.Join(f.Types, ", "))
	}
	fmt.Println()

	if len(f.MethodCounts) == 0 && f.TopLevelFuncs == 0 {
		fmt.Println("    methods: none")
	} else {
		var parts []string
		for _, recv := range sortedKeys(f.MethodCounts) {
			parts = append(parts, fmt.Sprintf("%s=%d", recv, f.MethodCounts[recv]))
		}
		parts = append(parts, fmt.Sprintf("top-level=%d", f.TopLevelFuncs))
		fmt.Printf("    methods: %s\n", strings.Join(parts, "; "))
	}

	fmt.Printf("    buckets: >200=%d, 120-200=%d, 80-119=%d, <80=%d\n", f.Buckets.Over200, f.Buckets.Between120, f.Buckets.Between80, f.Buckets.Under80)

	if len(f.Funcs) > 0 {
		limit := topFuncs
		if limit <= 0 || limit > len(f.Funcs) {
			limit = len(f.Funcs)
		}
		fmt.Printf("    funcs (top %d by lines):\n", limit)
		for i := 0; i < limit; i++ {
			fn := f.Funcs[i]
			fmt.Printf("      %4d lines | stmts=%d | %s\n", fn.Lines, fn.Stmts, fn.Name)
		}
	}

	for _, note := range f.Notes {
		fmt.Printf("    note: %s\n", note)
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func collectImports(file *ast.File) []string {
	imports := make([]string, 0, len(file.Imports))
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		imports = append(imports, path)
	}
	return imports
}

func buildNotes(types []string, buckets buckets, imports []string) []string {
	var notes []string

	if len(types) > 3 {
		notes = append(notes, fmt.Sprintf(">3 types in file (types=%d)", len(types)))
	}

	var longFuncNotes []string
	if buckets.Over200 > 0 {
		longFuncNotes = append(longFuncNotes, fmt.Sprintf("%d funcs >200 lines", buckets.Over200))
	}
	if buckets.Between120 > 0 {
		longFuncNotes = append(longFuncNotes, fmt.Sprintf("%d funcs 120-200 lines", buckets.Between120))
	}
	if len(longFuncNotes) > 0 {
		notes = append(notes, strings.Join(longFuncNotes, "; "))
	}

	if hasMixedConcerns(imports) {
		notes = append(notes, "mixed concerns? imports net/http and database/package")
	}

	return notes
}

func hasMixedConcerns(imports []string) bool {
	importSet := map[string]struct{}{}
	for _, imp := range imports {
		importSet[imp] = struct{}{}
	}

	http := false
	db := false
	if _, ok := importSet["net/http"]; ok {
		http = true
	}
	for _, candidate := range []string{"database/sql", "gorm.io/gorm"} {
		if _, ok := importSet[candidate]; ok {
			db = true
			break
		}
	}

	return http && db
}
