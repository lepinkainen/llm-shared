package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoDetectProjectGo(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0644))

	config, err := autoDetectProject(dir)
	require.NoError(t, err)
	require.Equal(t, "go", config.Type)
	require.Contains(t, config.Directories, "cmd")
	require.Contains(t, config.Files, "go.mod")
	require.True(t, config.HasMain)
	require.True(t, config.HasTests)
}

func TestDetectPythonProject(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0644))

	config := detectPythonProject(dir)
	require.Equal(t, "python", config.Type)
	require.Contains(t, config.Directories, "src")
	require.Contains(t, config.Files, "requirements.txt")
}

func TestValidateDirectoriesAndFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "cmd"), 0755))

	results := validateDirectories([]string{"cmd", "docs"}, dir)
	require.Len(t, results, 2)
	require.Equal(t, "success", results[0].Type)
	require.Equal(t, "warning", results[1].Type)

	results = validateFiles([]string{"go.mod"}, dir)
	require.Len(t, results, 1)
	require.Equal(t, "error", results[0].Type)
}

func TestValidateBuildSystemWithTaskfile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte("tasks:\n  build:\n    cmds: []\n  test:\n    cmds: []\n"), 0644))

	results := validateBuildSystem([]string{"build", "lint"}, dir)
	require.Len(t, results, 2)
	require.Equal(t, "success", results[0].Type)
	require.Equal(t, "warning", results[1].Type)
}

func TestValidateGoFunctionsUsesStubGoBinary(t *testing.T) {
	dir := t.TempDir()
	gofuncsPath := filepath.Join(dir, "utils", "gofuncs")
	require.NoError(t, os.MkdirAll(gofuncsPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(gofuncsPath, "gofuncs.go"), []byte("package main\nfunc main() {}\n"), 0644))

	// Stub go binary so we don't actually run the tool
	fakeBin := filepath.Join(dir, "bin")
	require.NoError(t, os.MkdirAll(fakeBin, 0755))
	fakeGo := filepath.Join(fakeBin, "go")
	script := "#!/bin/sh\necho \"file.go:1:f:y:main:()\"\necho \"file.go:2:f:y:TestThing:()\"\n"
	require.NoError(t, os.WriteFile(fakeGo, []byte(script), 0755))
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	results := validateGoFunctions(ProjectConfig{
		Type:     "go",
		HasMain:  true,
		HasTests: true,
	}, dir)

	require.Condition(t, func() bool {
		return containsResult(results, "success", "Found main function") &&
			containsResult(results, "success", "Found test functions") &&
			containsResult(results, "success", "Analyzed 2 functions")
	}, "unexpected results: %+v", results)
}

func TestCountErrors(t *testing.T) {
	results := []ValidationResult{
		{Type: "success"}, {Type: "error"}, {Type: "error"}, {Type: "warning"},
	}
	require.Equal(t, 2, countErrors(results))
}

func containsResult(results []ValidationResult, typ, messageContains string) bool {
	for _, r := range results {
		if r.Type == typ && r.Message == messageContains {
			return true
		}
	}
	return false
}
