package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationConfig represents the structure of the validation configuration
type ValidationConfig struct {
	Project struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	} `yaml:"project"`
	Validation struct {
		Directories struct {
			Required []string `yaml:"required"`
			Optional []string `yaml:"optional"`
		} `yaml:"directories"`
		Files struct {
			Required []string `yaml:"required"`
			Patterns []string `yaml:"patterns"`
		} `yaml:"files"`
		Dependencies struct {
			GoMod     []string `yaml:"go_mod,omitempty"`
			PackageJs []string `yaml:"package_json,omitempty"`
			PipReqs   []string `yaml:"requirements_txt,omitempty"`
			CargoToml []string `yaml:"cargo_toml,omitempty"`
		} `yaml:"dependencies"`
		BuildSystem struct {
			Taskfile []string `yaml:"taskfile,omitempty"`
			Make     []string `yaml:"make,omitempty"`
			Npm      []string `yaml:"npm,omitempty"`
			Cargo    []string `yaml:"cargo,omitempty"`
		} `yaml:"build_system"`
		Interfaces []struct {
			File            string   `yaml:"file"`
			Interface       string   `yaml:"interface"`
			Implementations []string `yaml:"implementations"`
		} `yaml:"interfaces,omitempty"`
	} `yaml:"validation"`
}

// ValidationResult holds the results of validation checks
type ValidationResult struct {
	Type    string // "error", "warning", "success"
	Message string
	File    string
}

// Colors for console output
const (
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorYellow = "\033[1;33m"
	ColorNC     = "\033[0m" // No Color
)

func main() {
	configFlag := flag.String("config", "", "Path to validation config file (.doc-validator.yml)")
	projectDir := flag.String("dir", ".", "Project directory to validate")
	autoDetect := flag.Bool("auto-detect", false, "Auto-detect project type and use default validation")
	initFlag := flag.String("init", "", "Initialize config for project type (go, node, python, rust)")
	flag.Parse()

	if *initFlag != "" {
		if err := initializeConfig(*initFlag, *projectDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Initialized .doc-validator.yml for %s project\n", *initFlag)
		return
	}

	var config ValidationConfig
	var err error

	if *autoDetect || *configFlag == "" {
		config, err = autoDetectConfig(*projectDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error auto-detecting project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔍 Auto-detected %s project: %s\n", config.Project.Type, config.Project.Name)
	} else {
		config, err = loadConfig(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	results := validateProject(config, *projectDir)

	errors, warnings, successes := categorizeResults(results)

	printResults(results)
	printSummary(len(errors), len(warnings), len(successes))

	if len(errors) > 0 {
		os.Exit(1)
	}
}

// autoDetectConfig detects project type and generates appropriate validation config
func autoDetectConfig(projectDir string) (ValidationConfig, error) {
	var config ValidationConfig

	// Detect project type
	if fileExists(filepath.Join(projectDir, "go.mod")) {
		config = generateGoConfig(projectDir)
	} else if fileExists(filepath.Join(projectDir, "package.json")) {
		config = generateNodeConfig(projectDir)
	} else if fileExists(filepath.Join(projectDir, "requirements.txt")) || fileExists(filepath.Join(projectDir, "pyproject.toml")) {
		config = generatePythonConfig(projectDir)
	} else if fileExists(filepath.Join(projectDir, "Cargo.toml")) {
		config = generateRustConfig(projectDir)
	} else {
		return config, fmt.Errorf("unable to detect project type")
	}

	// Set project name from directory
	config.Project.Name = filepath.Base(projectDir)

	return config, nil
}

// generateGoConfig creates a default config for Go projects
func generateGoConfig(projectDir string) ValidationConfig {
	config := ValidationConfig{}
	config.Project.Type = "go"

	// Standard Go project structure
	config.Validation.Directories.Required = []string{"cmd", "internal", "pkg"}
	config.Validation.Directories.Optional = []string{"scripts", "testdata", "build", "docs"}

	config.Validation.Files.Required = []string{"go.mod"}
	config.Validation.Files.Patterns = []string{"*.go", "cmd/**/*.go", "internal/**/*.go", "pkg/**/*.go"}

	// Common Go dependencies
	config.Validation.Dependencies.GoMod = []string{
		"github.com/spf13/viper",
		"github.com/alecthomas/kong",
	}

	// Check for build systems
	if fileExists(filepath.Join(projectDir, "Taskfile.yml")) {
		config.Validation.BuildSystem.Taskfile = []string{"build", "test", "lint"}
	}
	if fileExists(filepath.Join(projectDir, "Makefile")) {
		config.Validation.BuildSystem.Make = []string{"build", "test", "clean"}
	}

	return config
}

// generateNodeConfig creates a default config for Node.js projects
func generateNodeConfig(projectDir string) ValidationConfig {
	config := ValidationConfig{}
	config.Project.Type = "node"

	config.Validation.Directories.Required = []string{"src"}
	config.Validation.Directories.Optional = []string{"dist", "build", "docs", "tests", "__tests__"}

	config.Validation.Files.Required = []string{"package.json"}
	config.Validation.Files.Patterns = []string{"src/**/*.js", "src/**/*.ts", "src/**/*.jsx", "src/**/*.tsx"}

	config.Validation.BuildSystem.Npm = []string{"build", "test", "lint"}

	return config
}

// generatePythonConfig creates a default config for Python projects
func generatePythonConfig(projectDir string) ValidationConfig {
	config := ValidationConfig{}
	config.Project.Type = "python"

	config.Validation.Directories.Required = []string{"src"}
	config.Validation.Directories.Optional = []string{"tests", "docs", "scripts"}

	config.Validation.Files.Required = []string{"requirements.txt"}
	config.Validation.Files.Patterns = []string{"src/**/*.py", "*.py"}

	return config
}

// generateRustConfig creates a default config for Rust projects
func generateRustConfig(projectDir string) ValidationConfig {
	config := ValidationConfig{}
	config.Project.Type = "rust"

	config.Validation.Directories.Required = []string{"src"}
	config.Validation.Directories.Optional = []string{"tests", "examples", "benches"}

	config.Validation.Files.Required = []string{"Cargo.toml"}
	config.Validation.Files.Patterns = []string{"src/**/*.rs", "*.rs"}

	config.Validation.BuildSystem.Cargo = []string{"build", "test", "clippy"}

	return config
}

// loadConfig loads validation config from YAML file
func loadConfig(configPath string) (ValidationConfig, error) {
	var config ValidationConfig

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	return config, err
}

// validateProject performs all validation checks
func validateProject(config ValidationConfig, projectDir string) []ValidationResult {
	var results []ValidationResult

	fmt.Printf("🔍 Validating %s Documentation...\n\n", config.Project.Name)

	// Validate directories
	fmt.Println("📁 Validating directory structure...")
	results = append(results, validateDirectories(config.Validation.Directories, projectDir)...)

	// Validate files
	fmt.Println("\n📄 Validating file references...")
	results = append(results, validateFiles(config.Validation.Files, projectDir)...)

	// Validate dependencies
	if config.Project.Type != "" {
		fmt.Println("\n🔧 Validating dependencies...")
		results = append(results, validateDependencies(config.Validation.Dependencies, projectDir, config.Project.Type)...)
	}

	// Validate build system
	fmt.Println("\n📋 Validating build system...")
	results = append(results, validateBuildSystem(config.Validation.BuildSystem, projectDir)...)

	// Validate interfaces (Go-specific)
	if len(config.Validation.Interfaces) > 0 {
		fmt.Println("\n🏗️  Validating interface implementations...")
		results = append(results, validateInterfaces(config.Validation.Interfaces, projectDir)...)
	}

	return results
}

// validateDirectories checks if documented directories exist
func validateDirectories(dirs struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}, projectDir string) []ValidationResult {
	var results []ValidationResult

	for _, dir := range dirs.Required {
		dirPath := filepath.Join(projectDir, dir)
		if dirExists(dirPath) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Required directory exists: %s", dir), dir})
		} else {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Required directory missing: %s", dir), dir})
		}
	}

	for _, dir := range dirs.Optional {
		dirPath := filepath.Join(projectDir, dir)
		if dirExists(dirPath) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Optional directory exists: %s", dir), dir})
		}
	}

	return results
}

// validateFiles checks if documented files exist
func validateFiles(files struct {
	Required []string `yaml:"required"`
	Patterns []string `yaml:"patterns"`
}, projectDir string) []ValidationResult {
	var results []ValidationResult

	for _, file := range files.Required {
		filePath := filepath.Join(projectDir, file)
		if fileExists(filePath) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Required file exists: %s", file), file})
		} else {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Required file missing: %s", file), file})
		}
	}

	for _, pattern := range files.Patterns {
		matches, err := filepath.Glob(filepath.Join(projectDir, pattern))
		if err != nil {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Error checking pattern %s: %v", pattern, err), pattern})
			continue
		}

		if len(matches) > 0 {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Pattern matched %d files: %s", len(matches), pattern), pattern})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Pattern matched no files: %s", pattern), pattern})
		}
	}

	return results
}

// validateDependencies checks project dependencies
func validateDependencies(deps struct {
	GoMod     []string `yaml:"go_mod,omitempty"`
	PackageJs []string `yaml:"package_json,omitempty"`
	PipReqs   []string `yaml:"requirements_txt,omitempty"`
	CargoToml []string `yaml:"cargo_toml,omitempty"`
}, projectDir string, projectType string) []ValidationResult {
	var results []ValidationResult

	switch projectType {
	case "go":
		return validateGoMod(deps.GoMod, projectDir)
	case "node":
		return validatePackageJson(deps.PackageJs, projectDir)
	case "python":
		return validateRequirementsTxt(deps.PipReqs, projectDir)
	case "rust":
		return validateCargoToml(deps.CargoToml, projectDir)
	}

	return results
}

// validateGoMod checks Go module dependencies
func validateGoMod(expectedDeps []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	goModPath := filepath.Join(projectDir, "go.mod")

	if !fileExists(goModPath) {
		results = append(results, ValidationResult{"error", "go.mod file not found", "go.mod"})
		return results
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading go.mod: %v", err), "go.mod"})
		return results
	}

	goModContent := string(content)
	for _, dep := range expectedDeps {
		if strings.Contains(goModContent, dep) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Dependency found in go.mod: %s", dep), "go.mod"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected dependency not found in go.mod: %s", dep), "go.mod"})
		}
	}

	return results
}

// validatePackageJson checks Node.js dependencies
func validatePackageJson(expectedDeps []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	packagePath := filepath.Join(projectDir, "package.json")

	if !fileExists(packagePath) {
		results = append(results, ValidationResult{"error", "package.json file not found", "package.json"})
		return results
	}

	content, err := os.ReadFile(packagePath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading package.json: %v", err), "package.json"})
		return results
	}

	packageContent := string(content)
	for _, dep := range expectedDeps {
		if strings.Contains(packageContent, dep) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Dependency found in package.json: %s", dep), "package.json"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected dependency not found in package.json: %s", dep), "package.json"})
		}
	}

	return results
}

// validateRequirementsTxt checks Python dependencies
func validateRequirementsTxt(expectedDeps []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	reqPath := filepath.Join(projectDir, "requirements.txt")

	if !fileExists(reqPath) {
		results = append(results, ValidationResult{"error", "requirements.txt file not found", "requirements.txt"})
		return results
	}

	content, err := os.ReadFile(reqPath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading requirements.txt: %v", err), "requirements.txt"})
		return results
	}

	reqContent := string(content)
	for _, dep := range expectedDeps {
		if strings.Contains(reqContent, dep) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Dependency found in requirements.txt: %s", dep), "requirements.txt"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected dependency not found in requirements.txt: %s", dep), "requirements.txt"})
		}
	}

	return results
}

// validateCargoToml checks Rust dependencies
func validateCargoToml(expectedDeps []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	cargoPath := filepath.Join(projectDir, "Cargo.toml")

	if !fileExists(cargoPath) {
		results = append(results, ValidationResult{"error", "Cargo.toml file not found", "Cargo.toml"})
		return results
	}

	content, err := os.ReadFile(cargoPath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading Cargo.toml: %v", err), "Cargo.toml"})
		return results
	}

	cargoContent := string(content)
	for _, dep := range expectedDeps {
		if strings.Contains(cargoContent, dep) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Dependency found in Cargo.toml: %s", dep), "Cargo.toml"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected dependency not found in Cargo.toml: %s", dep), "Cargo.toml"})
		}
	}

	return results
}

// validateBuildSystem checks build system configuration
func validateBuildSystem(buildSystem struct {
	Taskfile []string `yaml:"taskfile,omitempty"`
	Make     []string `yaml:"make,omitempty"`
	Npm      []string `yaml:"npm,omitempty"`
	Cargo    []string `yaml:"cargo,omitempty"`
}, projectDir string) []ValidationResult {
	var results []ValidationResult

	if len(buildSystem.Taskfile) > 0 {
		results = append(results, validateTaskfile(buildSystem.Taskfile, projectDir)...)
	}

	if len(buildSystem.Make) > 0 {
		results = append(results, validateMakefile(buildSystem.Make, projectDir)...)
	}

	if len(buildSystem.Npm) > 0 {
		results = append(results, validateNpmScripts(buildSystem.Npm, projectDir)...)
	}

	if len(buildSystem.Cargo) > 0 {
		results = append(results, validateCargoCommands(buildSystem.Cargo, projectDir)...)
	}

	return results
}

// validateTaskfile checks Taskfile.yml for expected tasks
func validateTaskfile(expectedTasks []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	taskfilePath := filepath.Join(projectDir, "Taskfile.yml")

	if !fileExists(taskfilePath) {
		results = append(results, ValidationResult{"error", "Taskfile.yml not found", "Taskfile.yml"})
		return results
	}

	content, err := os.ReadFile(taskfilePath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading Taskfile.yml: %v", err), "Taskfile.yml"})
		return results
	}

	taskfileContent := string(content)
	for _, task := range expectedTasks {
		taskPattern := task + ":"
		if strings.Contains(taskfileContent, taskPattern) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Task found in Taskfile.yml: %s", task), "Taskfile.yml"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected task not found in Taskfile.yml: %s", task), "Taskfile.yml"})
		}
	}

	return results
}

// validateMakefile checks Makefile for expected targets
func validateMakefile(expectedTargets []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	makefilePath := filepath.Join(projectDir, "Makefile")

	if !fileExists(makefilePath) {
		results = append(results, ValidationResult{"error", "Makefile not found", "Makefile"})
		return results
	}

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading Makefile: %v", err), "Makefile"})
		return results
	}

	makefileContent := string(content)
	for _, target := range expectedTargets {
		targetPattern := target + ":"
		if strings.Contains(makefileContent, targetPattern) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Target found in Makefile: %s", target), "Makefile"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected target not found in Makefile: %s", target), "Makefile"})
		}
	}

	return results
}

// validateNpmScripts checks package.json for expected scripts
func validateNpmScripts(expectedScripts []string, projectDir string) []ValidationResult {
	var results []ValidationResult
	packagePath := filepath.Join(projectDir, "package.json")

	if !fileExists(packagePath) {
		results = append(results, ValidationResult{"error", "package.json not found", "package.json"})
		return results
	}

	content, err := os.ReadFile(packagePath)
	if err != nil {
		results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading package.json: %v", err), "package.json"})
		return results
	}

	packageContent := string(content)
	for _, script := range expectedScripts {
		scriptPattern := fmt.Sprintf("\"%s\":", script)
		if strings.Contains(packageContent, scriptPattern) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Script found in package.json: %s", script), "package.json"})
		} else {
			results = append(results, ValidationResult{"warning", fmt.Sprintf("Expected script not found in package.json: %s", script), "package.json"})
		}
	}

	return results
}

// validateCargoCommands checks Cargo.toml and common Cargo commands
func validateCargoCommands(expectedCommands []string, projectDir string) []ValidationResult {
	var results []ValidationResult

	// For Cargo, we primarily check that Cargo.toml exists and is valid
	cargoPath := filepath.Join(projectDir, "Cargo.toml")
	if !fileExists(cargoPath) {
		results = append(results, ValidationResult{"error", "Cargo.toml not found", "Cargo.toml"})
		return results
	}

	for _, cmd := range expectedCommands {
		results = append(results, ValidationResult{"success", fmt.Sprintf("Cargo command expected: %s", cmd), "Cargo.toml"})
	}

	return results
}

// validateInterfaces checks Go interface implementations
func validateInterfaces(interfaces []struct {
	File            string   `yaml:"file"`
	Interface       string   `yaml:"interface"`
	Implementations []string `yaml:"implementations"`
}, projectDir string) []ValidationResult {
	var results []ValidationResult

	for _, iface := range interfaces {
		interfaceFile := filepath.Join(projectDir, iface.File)

		if !fileExists(interfaceFile) {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Interface file not found: %s", iface.File), iface.File})
			continue
		}

		content, err := os.ReadFile(interfaceFile)
		if err != nil {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading interface file %s: %v", iface.File, err), iface.File})
			continue
		}

		if strings.Contains(string(content), iface.Interface) {
			results = append(results, ValidationResult{"success", fmt.Sprintf("Interface found: %s in %s", iface.Interface, iface.File), iface.File})
		} else {
			results = append(results, ValidationResult{"error", fmt.Sprintf("Interface not found: %s in %s", iface.Interface, iface.File), iface.File})
			continue
		}

		// Check implementations
		for _, implFile := range iface.Implementations {
			implPath := filepath.Join(projectDir, implFile)
			if !fileExists(implPath) {
				results = append(results, ValidationResult{"error", fmt.Sprintf("Implementation file not found: %s", implFile), implFile})
				continue
			}

			implContent, err := os.ReadFile(implPath)
			if err != nil {
				results = append(results, ValidationResult{"error", fmt.Sprintf("Error reading implementation file %s: %v", implFile, err), implFile})
				continue
			}

			if strings.Contains(string(implContent), iface.Interface) {
				results = append(results, ValidationResult{"success", fmt.Sprintf("Implementation references interface %s: %s", iface.Interface, implFile), implFile})
			} else {
				results = append(results, ValidationResult{"warning", fmt.Sprintf("Implementation may not reference interface %s: %s", iface.Interface, implFile), implFile})
			}
		}
	}

	return results
}

// categorizeResults separates results by type
func categorizeResults(results []ValidationResult) ([]ValidationResult, []ValidationResult, []ValidationResult) {
	var errors, warnings, successes []ValidationResult

	for _, result := range results {
		switch result.Type {
		case "error":
			errors = append(errors, result)
		case "warning":
			warnings = append(warnings, result)
		case "success":
			successes = append(successes, result)
		}
	}

	return errors, warnings, successes
}

// printResults displays validation results with colors
func printResults(results []ValidationResult) {
	// Sort results by type (errors first, then warnings, then successes)
	sort.Slice(results, func(i, j int) bool {
		typeOrder := map[string]int{"error": 0, "warning": 1, "success": 2}
		return typeOrder[results[i].Type] < typeOrder[results[j].Type]
	})

	for _, result := range results {
		var color, icon string
		switch result.Type {
		case "error":
			color = ColorRed
			icon = "❌"
		case "warning":
			color = ColorYellow
			icon = "⚠️ "
		case "success":
			color = ColorGreen
			icon = "✅"
		}

		fmt.Printf("%s%s %s%s\n", color, icon, result.Message, ColorNC)
	}
}

// printSummary displays final validation summary
func printSummary(errors, warnings, successes int) {
	fmt.Println("\n📊 Validation Summary")
	fmt.Println("====================")

	if errors == 0 && warnings == 0 {
		fmt.Printf("%s🎉 All documentation validation checks passed!%s\n", ColorGreen, ColorNC)
	} else if errors == 0 {
		fmt.Printf("%s⚠️  Documentation validation completed with %d warnings%s\n", ColorYellow, warnings, ColorNC)
	} else {
		fmt.Printf("%s❌ Documentation validation failed with %d errors and %d warnings%s\n", ColorRed, errors, warnings, ColorNC)
		fmt.Println("")
		fmt.Println("Please update the documentation to reflect the current codebase structure.")
	}

	fmt.Printf("\nResults: %d successes, %d warnings, %d errors\n", successes, warnings, errors)
}

// initializeConfig creates a new validation config file
func initializeConfig(projectType, projectDir string) error {
	var config ValidationConfig

	switch projectType {
	case "go":
		config = generateGoConfig(projectDir)
	case "node":
		config = generateNodeConfig(projectDir)
	case "python":
		config = generatePythonConfig(projectDir)
	case "rust":
		config = generateRustConfig(projectDir)
	default:
		return fmt.Errorf("unsupported project type: %s", projectType)
	}

	config.Project.Name = filepath.Base(projectDir)

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	configPath := filepath.Join(projectDir, ".doc-validator.yml")
	return os.WriteFile(configPath, data, 0644)
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
