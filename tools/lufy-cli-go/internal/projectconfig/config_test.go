package projectconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScanDetectsGoStack(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\nrequire go.opentelemetry.io/otel v1.0.0\n")
	writeFile(t, root, ".golangci.yml", "linters: {}\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	stack := requireStack(t, cfg, "go")
	if !stack.Supported || stack.Version != "1.22" || stack.PackageManager != "go modules" {
		t.Fatalf("unexpected go stack: %#v", stack)
	}
	if stack.TestRunner.Command != "go test ./..." || stack.TestRunner.CoverageThreshold != 85 {
		t.Fatalf("unexpected go test config: %#v", stack.TestRunner)
	}
	if stack.Linter.Command != "golangci-lint run" || stack.Formatter.Command != "gofmt -w" {
		t.Fatalf("unexpected go tooling: %#v %#v", stack.Linter, stack.Formatter)
	}
	if !contains(stack.ObservabilityLibs, "go.opentelemetry.io") {
		t.Fatalf("missing observability lib: %#v", stack.ObservabilityLibs)
	}
}

func TestScanDetectsTypeScriptNextStack(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "package.json", `{"devDependencies":{"typescript":"5.4.0","vitest":"1.0.0","eslint":"8.0.0","prettier":"3.0.0"},"dependencies":{"react":"18.0.0","next":"14.0.0","pino":"8.0.0"}}`)
	writeFile(t, root, "tsconfig.json", "{}")
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	stack := requireStack(t, cfg, "typescript")
	for _, framework := range []string{"react", "next"} {
		if !contains(stack.Frameworks, framework) {
			t.Fatalf("missing framework %s in %#v", framework, stack.Frameworks)
		}
	}
	if stack.PackageManager != "pnpm" || stack.TestRunner.Command != "vitest run" || stack.StaticAnalysis.Command != "tsc --noEmit" {
		t.Fatalf("unexpected ts stack: %#v", stack)
	}
	if !contains(stack.ObservabilityLibs, "pino") {
		t.Fatalf("missing pino observability: %#v", stack.ObservabilityLibs)
	}
}

func TestScanDetectsJavaScriptToolingFromScripts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "package.json", `{"scripts":{"test":"vitest run","lint":"eslint .","format":"prettier --write ."}}`)

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	stack := requireStack(t, cfg, "javascript")
	if !stack.Supported || stack.TestRunner.Command != "vitest run" {
		t.Fatalf("unexpected javascript test config from scripts: %#v", stack)
	}
	if stack.Linter.Command != "eslint ." || stack.Formatter.Command != "prettier --write" {
		t.Fatalf("unexpected javascript tooling from scripts: %#v %#v", stack.Linter, stack.Formatter)
	}
	if stack.StaticAnalysis.Command != "" {
		t.Fatalf("javascript without tsconfig should not set static analysis: %#v", stack.StaticAnalysis)
	}
}

func TestScanDetectsPythonJVMRustAndCI(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pyproject.toml", "[project]\ndependencies = ['pytest', 'ruff', 'mypy', 'fastapi', 'structlog']\n")
	writeFile(t, root, "pom.xml", "<project><dependency>spring-boot</dependency><dependency>io.opentelemetry</dependency></project>")
	writeFile(t, root, "Cargo.toml", "[package]\nname='demo'\n")
	writeFile(t, root, ".github/workflows/ci.yml", "name: ci\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.CI.Detected || cfg.CI.Provider != "github-actions" || !contains(cfg.CI.Workflows, ".github/workflows/ci.yml") {
		t.Fatalf("unexpected ci config: %#v", cfg.CI)
	}
	python := requireStack(t, cfg, "python")
	if python.TestRunner.Command != "pytest" || python.Formatter.Command != "ruff format" || !contains(python.Frameworks, "fastapi") {
		t.Fatalf("unexpected python stack: %#v", python)
	}
	java := requireStack(t, cfg, "java")
	if java.PackageManager != "maven" || !contains(java.Frameworks, "spring-boot") {
		t.Fatalf("unexpected java stack: %#v", java)
	}
	rust := requireStack(t, cfg, "rust")
	if rust.Supported || rust.TestRunner.Command != "TODO" {
		t.Fatalf("unexpected rust placeholder: %#v", rust)
	}
}

func TestScanDetectsKotlinGradleKtsStack(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "build.gradle.kts", `plugins { kotlin("jvm") version "1.9.0" }`)

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	stack := requireStack(t, cfg, "kotlin")
	if !stack.Supported || stack.PackageManager != "gradle" {
		t.Fatalf("unexpected kotlin gradle stack: %#v", stack)
	}
	if stack.TestRunner.Command != "gradle test" || stack.Formatter.Command != "gradle spotlessApply" {
		t.Fatalf("unexpected kotlin gradle tooling: %#v %#v", stack.TestRunner, stack.Formatter)
	}
}

func TestScanDetectsAdditionalUnsupportedStacks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "composer.json", `{"require":{"php":"^8.2"}}`)
	writeFile(t, root, "Gemfile", "source 'https://rubygems.org'\n")
	writeFile(t, root, "mix.exs", "defmodule Demo.MixProject do\nend\n")
	writeFile(t, root, "Demo.csproj", `<Project Sdk="Microsoft.NET.Sdk"></Project>`)

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"dotnet", "elixir", "php", "ruby"} {
		stack := requireStack(t, cfg, id)
		if stack.Supported || stack.TestRunner.Command != "TODO" || stack.Notes == "" {
			t.Fatalf("unexpected unsupported placeholder for %s: %#v", id, stack)
		}
	}
}

func TestServiceRunBlocksForceAndRescan(t *testing.T) {
	var out strings.Builder
	codeRoot := t.TempDir()
	writeFile(t, codeRoot, "go.mod", "module example.com/app\n\ngo 1.22\n")
	svc := Service{Now: fixedTime}
	if err := svc.Run(Options{Target: codeRoot}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(codeRoot, ProjectConfigPath)); err != nil {
		t.Fatal(err)
	}
	if err := svc.Run(Options{Target: codeRoot}, &out); err == nil || !strings.Contains(err.Error(), "ya existe") {
		t.Fatalf("expected existing config error, got %v", err)
	}
	if err := svc.Run(Options{Target: codeRoot, Force: true}, &out); err != nil {
		t.Fatalf("force should replace config: %v", err)
	}
}

func TestRescanPreservesOverridesAndAddsStack(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	svc := Service{Now: fixedTime}
	if err := svc.Run(Options{Target: root}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ProjectConfigPath)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.LocBudget = 250
	cfg.DeliveryStrategy = "single-pr"
	cfg.Stacks[0].TestRunner.CoverageThreshold = 70
	cfg.Stacks[0].AntiPatterns = []string{"custom-go-smell"}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "package.json", `{"devDependencies":{"jest":"29.0.0"}}`)
	writeFile(t, root, "tsconfig.json", "{}")

	var out strings.Builder
	if err := svc.Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=stack") || !strings.Contains(out.String(), "status=applied") {
		t.Fatalf("rescan report missing stack drift: %s", out.String())
	}
	merged, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	goStack := requireStack(t, merged, "go")
	if goStack.TestRunner.CoverageThreshold != 70 || !contains(goStack.AntiPatterns, "custom-go-smell") {
		t.Fatalf("rescan did not preserve go overrides: %#v", goStack)
	}
	if merged.LocBudget != 250 || merged.DeliveryStrategy != "single-pr" {
		t.Fatalf("rescan did not preserve global overrides: %#v", merged)
	}
	_ = requireStack(t, merged, "typescript")
}

func TestRescanNoDriftDoesNotRewriteConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	svc := Service{Now: fixedTime}
	if err := svc.Run(Options{Target: root}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ProjectConfigPath)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	infoBefore, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	later := func() time.Time { return fixedTime().Add(2 * time.Hour) }
	var out strings.Builder
	if err := (Service{Now: later}).Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	infoAfter, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) || !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Fatalf("no-drift rescan rewrote config\nbefore=%s\nafter=%s", string(before), string(after))
	}
	if !strings.Contains(out.String(), "sin drift") || !strings.Contains(out.String(), "no fue reescrito") {
		t.Fatalf("no-drift report unexpected: %s", out.String())
	}
}

func TestRescanReportsToolingAndCIDrift(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "package.json", `{"scripts":{"test":"npm test"}}`)
	svc := Service{Now: fixedTime}
	if err := svc.Run(Options{Target: root}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "package.json", `{"devDependencies":{"vitest":"1.0.0"}}`)
	writeFile(t, root, ".github/workflows/ci.yml", "name: ci\n")
	var out strings.Builder
	if err := svc.Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=tooling") || !strings.Contains(out.String(), "category=ci") {
		t.Fatalf("rescan report missing tooling/ci drift: %s", out.String())
	}
	merged, err := Load(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	js := requireStack(t, merged, "javascript")
	if js.TestRunner.Command != "vitest run" {
		t.Fatalf("tooling drift was not applied: %#v", js.TestRunner)
	}
	if !merged.CI.Detected || merged.CI.Provider != "github-actions" {
		t.Fatalf("ci drift was not applied: %#v", merged.CI)
	}
}

func TestRescanMarksRemovedStackDeprecated(t *testing.T) {
	old := ProjectConfig{SchemaVersion: 1, Stacks: []Stack{{ID: "python", Supported: true, Frameworks: []string{}, Formatter: Formatter{FileExtensions: []string{".py"}}}}}
	fresh := ProjectConfig{SchemaVersion: 1, Stacks: []Stack{{ID: "go", Supported: true, Frameworks: []string{}, Formatter: Formatter{FileExtensions: []string{".go"}}}}}
	merged := MergeRescan(old, fresh)
	python := requireStack(t, merged, "python")
	if !python.Deprecated {
		t.Fatalf("expected removed python stack to be deprecated: %#v", python)
	}
}

func TestRescanPreservesUnknownFieldsAndOverrides(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
custom_top: keep-me
stacks:
  - id: go
    supported: true
    version: "1.21"
    frameworks: []
    test_runner:
      command: go test ./...
      coverage_command: go test -coverprofile=coverage.out ./...
      coverage_threshold: 70
    linter:
      command: go vet ./...
    formatter:
      command: gofmt -w
      file_extensions: [.go]
    static_analysis:
      command: go vet ./...
    anti_patterns: [custom-go-smell]
    observability_libs: []
    custom_stack: keep-stack
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
loc_budget: 250
delivery_strategy: single-pr
`)
	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"custom_top: keep-me", "custom_stack: keep-stack", "coverage_threshold: 70", "custom-go-smell"} {
		if !strings.Contains(text, want) {
			t.Fatalf("merged config missing %q:\n%s", want, text)
		}
	}
}

func TestRescanInvalidYAMLFailsWithoutMutation(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, "schema_version: [\n")
	path := filepath.Join(root, ProjectConfigPath)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "corrige o respalda") {
		t.Fatalf("expected actionable YAML error, got %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("invalid YAML rescan mutated file: %s", string(after))
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 20, 14, 0, 0, 0, time.UTC)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func requireStack(t *testing.T, cfg ProjectConfig, id string) Stack {
	t.Helper()
	for _, stack := range cfg.Stacks {
		if stack.ID == id {
			return stack
		}
	}
	t.Fatalf("stack %s not found in %#v", id, cfg.Stacks)
	return Stack{}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
