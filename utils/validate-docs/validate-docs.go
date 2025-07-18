package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProjectConfig represents auto-detected project configuration
type ProjectConfig struct {
	Name        string
	Type        string
	Directories []string
	Files       []string
	BuildTasks  []string
	HasMain     bool
	HasTests    bool
}

// ValidationResult holds the results of validation checks
type ValidationResult struct {
	Type    string // "error", "warning", "success"
	Message string
	Item    string
}

func main() {
	projectDir := flag.String("dir", ".", "Project directory to validate")
	flag.Parse()

	config, err := autoDetectProject(*projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Detected %s project: %s\n", config.Type, config.Name)

	results := validateProject(config, *projectDir)
	printResults(results)

	errors := countErrors(results)
	if errors > 0 {
		fmt.Printf("\nValidation failed with %d errors\n", errors)
		os.Exit(1)
	} else {
		fmt.Println("\nValidation passed!")
	}
}

// autoDetectProject detects project type and creates smart configuration
func autoDetectProject(projectDir string) (ProjectConfig, error) {
	var config ProjectConfig
	config.Name = filepath.Base(projectDir)

	// Detect project type
	if fileExists(filepath.Join(projectDir, "go.mod")) {
		config = detectGoProject(projectDir)
	} else if fileExists(filepath.Join(projectDir, "requirements.txt")) || fileExists(filepath.Join(projectDir, "pyproject.toml")) {
		config = detectPythonProject(projectDir)
	} else {
		return config, fmt.Errorf("unable to detect project type (no go.mod, requirements.txt, or pyproject.toml found)")
	}

	config.Name = filepath.Base(projectDir)
	return config, nil
}

// detectGoProject creates configuration for Go projects
func detectGoProject(projectDir string) ProjectConfig {
	config := ProjectConfig{Type: "go"}

	// Standard Go directories
	config.Directories = []string{"cmd", "internal", "pkg"}

	// Add optional directories if they exist
	optionalDirs := []string{"api", "web", "scripts", "docs", "build", "testdata"}
	for _, dir := range optionalDirs {
		if dirExists(filepath.Join(projectDir, dir)) {
			config.Directories = append(config.Directories, dir)
		}
	}

	// Required files
	config.Files = []string{"go.mod"}
	if fileExists(filepath.Join(projectDir, "go.sum")) {
		config.Files = append(config.Files, "go.sum")
	}

	// Detect build system
	if fileExists(filepath.Join(projectDir, "Taskfile.yml")) {
		config.BuildTasks = []string{"build", "test", "lint"}
	}

	config.HasMain = true
	config.HasTests = true

	return config
}

// detectPythonProject creates configuration for Python projects
func detectPythonProject(projectDir string) ProjectConfig {
	config := ProjectConfig{Type: "python"}

	// Standard Python directories
	config.Directories = []string{"src"}

	// Add optional directories if they exist
	optionalDirs := []string{"tests", "docs", "scripts", "data"}
	for _, dir := range optionalDirs {
		if dirExists(filepath.Join(projectDir, dir)) {
			config.Directories = append(config.Directories, dir)
		}
	}

	// Required dependency files
	if fileExists(filepath.Join(projectDir, "requirements.txt")) {
		config.Files = append(config.Files, "requirements.txt")
	}
	if fileExists(filepath.Join(projectDir, "pyproject.toml")) {
		config.Files = append(config.Files, "pyproject.toml")
	}

	// Detect build system
	if fileExists(filepath.Join(projectDir, "Taskfile.yml")) {
		config.BuildTasks = []string{"test", "lint"}
	}

	config.HasMain = true
	config.HasTests = true

	return config
}

// validateProject performs all validation checks
func validateProject(config ProjectConfig, projectDir string) []ValidationResult {
	var results []ValidationResult

	fmt.Printf("Validating %s project structure...\n", config.Name)

	// Validate directories
	results = append(results, validateDirectories(config.Directories, projectDir)...)

	// Validate files
	results = append(results, validateFiles(config.Files, projectDir)...)

	// Validate build system
	if len(config.BuildTasks) > 0 {
		results = append(results, validateBuildSystem(config.BuildTasks, projectDir)...)
	}

	// Validate functions using existing tools
	results = append(results, validateFunctions(config, projectDir)...)

	return results
}

// validateDirectories checks if directories exist
func validateDirectories(dirs []string, projectDir string) []ValidationResult {
	var results []ValidationResult

	for _, dir := range dirs {
		dirPath := filepath.Join(projectDir, dir)
		if dirExists(dirPath) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Directory exists: %s", dir), dir})
		} else {
			// Only error for standard directories, warn for optional ones
			msgType := "warning"
			if isStandardDirectory(dir) {
				msgType = "error"
			}
			results = append(results, ValidationResult{msgType, fmt.Sprintf("Directory missing: %s", dir), dir})
		}
	}

	return results
}

// validateFiles checks if required files exist
func validateFiles(files []string, projectDir string) []ValidationResult {
	var results []ValidationResult

	for _, file := range files {
		filePath := filepath.Join(projectDir, file)
		if fileExists(filePath) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("File exists: %s", file), file})
		} else {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Required file missing: %s", file), file})
		}
	}

	return results
}

// validateBuildSystem checks build system configuration
func validateBuildSystem(tasks []string, projectDir string) []ValidationResult {
	var results []ValidationResult

	if fileExists(filepath.Join(projectDir, "Taskfile.yml")) {
		results = append(results, validateTaskfile(tasks, projectDir)...)
	} else {
		results = append(results, ValidationResult{"warning", "No Taskfile.yml found (required for build system)", "build"})
	}

	return results
}

// validateTaskfile checks Taskfile.yml for expected tasks
func validateTaskfile(expectedTasks []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	taskfilePath := filepath.Join(projectDir, "Taskfile.yml")

	content, err := os.ReadFile(taskfilePath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading Taskfile.yml: %v", err), "Taskfile.yml"})
		return results
	}

	taskfileContent := string(content)
	for _, task := range expectedTasks {
		taskPattern := task + ":"
		if strings.Contains(taskfileContent, taskPattern) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Task found: %s", task), "Taskfile.yml"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Task missing: %s", task), "Taskfile.yml"})
		}
	}

	return results
}

// validateFunctions uses gofuncs.go or pyfuncs.py to validate function patterns
func validateFunctions(config ProjectConfig, projectDir string) []ValidationResult {
	var results []ValidationResult

	switch config.Type {
	case "go":
		results = append(results, validateGoFunctions(config, projectDir)...)
	case "python":
		results = append(results, validatePythonFunctions(config, projectDir)...)
	}

	return results
}

// validateGoFunctions uses gofuncs.go to validate Go function patterns
func validateGoFunctions(config ProjectConfig, projectDir string) []ValidationResult {
	var results []ValidationResult

	// Find gofuncs.go - try multiple locations
	var gofuncsPath string
	locations := []string{
		filepath.Join(projectDir, "utils", "gofuncs", "gofuncs.go"),
		filepath.Join(projectDir, "..", "gofuncs", "gofuncs.go"),
		filepath.Join(projectDir, "..", "..", "utils", "gofuncs", "gofuncs.go"),
	}

	for _, path := range locations {
		if fileExists(path) {
			gofuncsPath = path
			break
		}
	}

	if gofuncsPath == "" {
		results = append(results, ValidationResult{"warning", "gofuncs.go not found, skipping function validation", "functions"})
		return results
	}

	// Run gofuncs.go
	cmd := exec.Command("go", "run", gofuncsPath, "-dir", projectDir)
	output, err := cmd.Output()
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error running gofuncs: %v", err), "functions"})
		return results
	}

	functions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(functions) == 1 && functions[0] == "" {
		functions = []string{}
	}

	// Check for main functions if required
	if config.HasMain {
		hasMain := false
		for _, fn := range functions {
			if strings.Contains(fn, ":f:y:main:") {
				hasMain = true
				break
			}
		}
		if hasMain {
			results = append(results, ValidationResult{"success", "Found main function", "functions"})
		} else {
			results = append(results, ValidationResult{"warning", "No main function found", "functions"})
		}
	}

	// Check for test functions
	if config.HasTests {
		hasTests := false
		for _, fn := range functions {
			if strings.Contains(fn, ":f:y:Test") {
				hasTests = true
				break
			}
		}
		if hasTests {
			results = append(results, ValidationResult{"success", "Found test functions", "functions"})
		} else {
			results = append(results, ValidationResult{"warning", "No test functions found", "functions"})
		}
	}

	if len(functions) > 0 {
		results = append(results, ValidationResult{"success", fmt.Sprintf("Analyzed %d functions", len(functions)), "functions"})
	}

	return results
}

// validatePythonFunctions uses pyfuncs.py to validate Python function patterns
func validatePythonFunctions(config ProjectConfig, projectDir string) []ValidationResult {
	var results []ValidationResult

	// Find pyfuncs.py - try multiple locations
	var pyfuncsPath string
	locations := []string{
		filepath.Join(projectDir, "utils", "pyfuncs.py"),
		filepath.Join(projectDir, "..", "pyfuncs.py"),
		filepath.Join(projectDir, "..", "..", "utils", "pyfuncs.py"),
	}

	for _, path := range locations {
		if fileExists(path) {
			pyfuncsPath = path
			break
		}
	}

	if pyfuncsPath == "" {
		results = append(results, ValidationResult{"warning", "pyfuncs.py not found, skipping function validation", "functions"})
		return results
	}

	// Run pyfuncs.py
	cmd := exec.Command("python3", pyfuncsPath, "--dir", projectDir)
	output, err := cmd.Output()
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error running pyfuncs: %v", err), "functions"})
		return results
	}

	functions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(functions) == 1 && functions[0] == "" {
		functions = []string{}
	}

	// Check for main functions if required
	if config.HasMain {
		hasMain := false
		for _, fn := range functions {
			if strings.Contains(fn, ":f:y:main:") {
				hasMain = true
				break
			}
		}
		if hasMain {
			results = append(results, ValidationResult{"success", "Found main function", "functions"})
		} else {
			results = append(results, ValidationResult{"warning", "No main function found", "functions"})
		}
	}

	// Check for test functions
	if config.HasTests {
		hasTests := false
		for _, fn := range functions {
			if strings.Contains(fn, ":f:y:test_") {
				hasTests = true
				break
			}
		}
		if hasTests {
			results = append(results, ValidationResult{"success", "Found test functions", "functions"})
		} else {
			results = append(results, ValidationResult{"warning", "No test functions found", "functions"})
		}
	}

	if len(functions) > 0 {
		results = append(results, ValidationResult{"success", fmt.Sprintf("Analyzed %d functions", len(functions)), "functions"})
	}

	return results
}

// isStandardDirectory checks if a directory is considered standard/required
func isStandardDirectory(dir string) bool {
	standardDirs := []string{"cmd", "internal", "pkg", "src"}
	for _, std := range standardDirs {
		if dir == std {
			return true
		}
	}
	return false
}

// printResults displays validation results
func printResults(results []ValidationResult) {
	for _, result := range results {
		var icon string
		switch result.Type {
		case "error":
			icon = "❌"
		case "warning":
			icon = "⚠️"
		case "success":
			icon = "✅"
		}
		fmt.Printf("%s %s\n", icon, result.Message)
	}
}

// countErrors counts the number of error results
func countErrors(results []ValidationResult) int {
	count := 0
	for _, result := range results {
		if result.Type == "error" {
			count++
		}
	}
	return count
}

// Utility functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
