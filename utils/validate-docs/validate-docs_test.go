package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAutoDetectProjectGo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	config, err := autoDetectProject(dir)
	if err != nil {
		t.Fatalf("autoDetectProject failed: %v", err)
	}
	if config.Type != "go" {
		t.Errorf("config.Type = %q, want %q", config.Type, "go")
	}
	if !contains(config.Directories, "cmd") {
		t.Errorf("config.Directories does not contain 'cmd': %v", config.Directories)
	}
	if !contains(config.Files, "go.mod") {
		t.Errorf("config.Files does not contain 'go.mod': %v", config.Files)
	}
	if !config.HasMain {
		t.Errorf("config.HasMain = false, want true")
	}
	if !config.HasTests {
		t.Errorf("config.HasTests = false, want true")
	}
}

func TestDetectPythonProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	config := detectPythonProject(dir)
	if config.Type != "python" {
		t.Errorf("config.Type = %q, want %q", config.Type, "python")
	}
	if !contains(config.Directories, "src") {
		t.Errorf("config.Directories does not contain 'src': %v", config.Directories)
	}
	if !contains(config.Files, "requirements.txt") {
		t.Errorf("config.Files does not contain 'requirements.txt': %v", config.Files)
	}
}

func TestValidateDirectoriesAndFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "cmd"), 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	results := validateDirectories([]string{"cmd", "docs"}, dir)
	if len(results) != 2 {
		t.Fatalf("validateDirectories returned %d results, want 2", len(results))
	}
	if results[0].Type != "success" {
		t.Errorf("results[0].Type = %q, want %q", results[0].Type, "success")
	}
	if results[1].Type != "warning" {
		t.Errorf("results[1].Type = %q, want %q", results[1].Type, "warning")
	}

	results = validateFiles([]string{"go.mod"}, dir)
	if len(results) != 1 {
		t.Fatalf("validateFiles returned %d results, want 1", len(results))
	}
	if results[0].Type != "error" {
		t.Errorf("results[0].Type = %q, want %q", results[0].Type, "error")
	}
}

func TestValidateBuildSystemWithTaskfile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte("tasks:\n  build:\n    cmds: []\n  test:\n    cmds: []\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	results := validateBuildSystem([]string{"build", "lint"}, dir)
	if len(results) != 2 {
		t.Fatalf("validateBuildSystem returned %d results, want 2", len(results))
	}
	if results[0].Type != "success" {
		t.Errorf("results[0].Type = %q, want %q", results[0].Type, "success")
	}
	if results[1].Type != "warning" {
		t.Errorf("results[1].Type = %q, want %q", results[1].Type, "warning")
	}
}

func TestValidateGoFunctionsUsesStubGoBinary(t *testing.T) {
	dir := t.TempDir()
	gofuncsPath := filepath.Join(dir, "utils", "gofuncs")
	if err := os.MkdirAll(gofuncsPath, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gofuncsPath, "gofuncs.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Stub go binary so we don't actually run the tool
	fakeBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(fakeBin, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	fakeGo := filepath.Join(fakeBin, "go")
	script := "#!/bin/sh\necho \"file.go:1:f:y:main:()\"\necho \"file.go:2:f:y:TestThing:()\"\n"
	if err := os.WriteFile(fakeGo, []byte(script), 0755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	results := validateGoFunctions(ProjectConfig{
		Type:     "go",
		HasMain:  true,
		HasTests: true,
	}, dir)

	hasMainFunc := containsResult(results, "success", "Found main function")
	hasTestFunc := containsResult(results, "success", "Found test functions")
	hasAnalyzed := containsResult(results, "success", "Analyzed 2 functions")

	if !hasMainFunc || !hasTestFunc || !hasAnalyzed {
		t.Errorf("validateGoFunctions returned unexpected results: %+v", results)
	}
}

func TestCountErrors(t *testing.T) {
	results := []ValidationResult{
		{Type: "success"}, {Type: "error"}, {Type: "error"}, {Type: "warning"},
	}
	got := countErrors(results)
	want := 2
	if got != want {
		t.Errorf("countErrors(results) = %d, want %d", got, want)
	}
}

func containsResult(results []ValidationResult, typ, messageContains string) bool {
	for _, r := range results {
		if r.Type == typ && r.Message == messageContains {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
