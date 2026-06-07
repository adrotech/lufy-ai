package projectconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
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
	writeFile(t, root, "package.json", `{"scripts":{"typecheck":"tsc --noEmit","lint":"eslint .","test":"vitest run","build":"tsc --noEmit && next build"},"devDependencies":{"typescript":"5.4.0","vitest":"1.0.0","eslint":"8.0.0","prettier":"3.0.0"},"dependencies":{"react":"18.0.0","next":"14.0.0","pino":"8.0.0"}}`)
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
	for _, command := range []string{"pnpm typecheck*", "pnpm lint*", "pnpm test*", "pnpm build*"} {
		if !contains(cfg.Validation.AllowedCommands.Implementer, command) {
			t.Fatalf("missing implementer validation command %q: %#v", command, cfg.Validation.AllowedCommands.Implementer)
		}
	}
	surface := requireSurface(t, cfg, "web-app")
	if surface.Type != "frontend" ||
		surface.Architecture.Preferred != "feature_driven" ||
		!contains(surface.AgentLens.PrimaryConcerns, "accessibility") ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_driven_structure") ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_colocation") ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_public_barrels_index_ts") ||
		!contains(surface.AgentLens.PrimaryConcerns, "pages_as_routing_only") ||
		!contains(surface.AgentLens.ValidationExpectations, "browser_check_when_ui_changes") ||
		!contains(surface.AgentLens.ValidationExpectations, "feature_boundary_review") {
		t.Fatalf("unexpected frontend surface: %#v", surface)
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

func TestScanDetectsBackendCLIInfraAndFullstackSurfaces(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, "api/openapi.yaml", "openapi: 3.0.0\n")
	writeFile(t, root, "internal/controllers/users.go", "package controllers\n")
	writeFile(t, root, "internal/services/users.go", "package services\n")
	writeFile(t, root, "internal/repositories/users.go", "package repositories\n")
	writeFile(t, root, "package.json", `{"dependencies":{"react":"18.0.0","next":"14.0.0"},"devDependencies":{"typescript":"5.4.0"}}`)
	writeFile(t, root, "tsconfig.json", "{}")
	writeFile(t, root, "main.tf", "terraform {}\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if surface := requireSurface(t, cfg, "api"); surface.Type != "backend" ||
		surface.Architecture.Preferred != "controller_service_repository" ||
		!contains(surface.Architecture.Detected, "controller_service_repository") ||
		!contains(surface.AgentLens.PrimaryConcerns, "api_contracts") {
		t.Fatalf("unexpected backend surface: %#v", surface)
	}
	if surface := requireSurface(t, cfg, "web-app"); surface.Type != "frontend" {
		t.Fatalf("unexpected frontend surface: %#v", surface)
	}
	if surface := requireSurface(t, cfg, "infra"); surface.Type != "infra" {
		t.Fatalf("unexpected infra surface: %#v", surface)
	}
	if surface := requireSurface(t, cfg, "fullstack-flow"); surface.Type != "fullstack" ||
		!contains(surface.Connects, "api") ||
		!contains(surface.Connects, "web-app") ||
		surface.Architecture.Preferred != "controller_service_repository" ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_driven_frontend_structure") ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_colocation") ||
		!contains(surface.AgentLens.PrimaryConcerns, "feature_public_barrels_index_ts") ||
		!contains(surface.AgentLens.ValidationExpectations, "feature_boundary_review") {
		t.Fatalf("unexpected fullstack surface: %#v", surface)
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
	data, err := os.ReadFile(filepath.Join(codeRoot, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	assertWorkflowLimitsOnly(t, string(data))
}

func TestServiceRunPromptErrorDoesNotWriteProjectConfig(t *testing.T) {
	var out strings.Builder
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	cancelled := errors.New("project_profile cancelado")

	err := (Service{Now: fixedTime}).Run(Options{
		Target: root,
		ProfilePrompt: func(ProjectConfig) (ProjectProfile, error) {
			return ProjectProfile{}, cancelled
		},
	}, &out)
	if !errors.Is(err, cancelled) {
		t.Fatalf("expected prompt cancellation, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ProjectConfigPath)); !os.IsNotExist(statErr) {
		t.Fatalf("project config should not be written after prompt error, stat=%v", statErr)
	}
}

func TestServiceEnsureCreatesProjectConfigWhenMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")

	created, err := (Service{Now: fixedTime}).Ensure(root)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected Ensure to create missing project config")
	}
	data, err := os.ReadFile(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"schema_version: 1", "detected_at: 2026-05-20T14:00:00Z", "id: go", "workflow_limits:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated config missing %q:\n%s", want, text)
		}
	}

	created, err = (Service{Now: fixedTime}).Ensure(root)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected Ensure to preserve existing project config")
	}
}

func TestServiceEnsurePreservesExistingInvalidProjectConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, "schema_version: [\n")

	created, err := (Service{Now: fixedTime}).Ensure(root)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected Ensure to leave existing project config untouched")
	}
	data, err := os.ReadFile(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "schema_version: [\n" {
		t.Fatalf("existing project config was overwritten: %q", data)
	}
}

func TestScanGeneratesCanonicalWorkflowLimits(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkflowLimits.Sizing.LOCBudget != 400 {
		t.Fatalf("unexpected workflow sizing defaults: %#v", cfg.WorkflowLimits.Sizing)
	}
	if cfg.WorkflowLimits.Routing.Strategy == "" || cfg.WorkflowLimits.ProposalSlicingStrategy == "" || cfg.WorkflowLimits.DeliveryBatchStrategy == "" {
		t.Fatalf("missing workflow limits defaults: %#v", cfg.WorkflowLimits)
	}
	if len(cfg.WorkflowLimits.StopRules) == 0 || len(cfg.WorkflowLimits.Preflight) == 0 {
		t.Fatalf("missing workflow gates: %#v", cfg.WorkflowLimits)
	}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assertWorkflowLimitsOnly(t, string(data))
}

func TestScanGeneratesHarnessDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")

	cfg, err := Scan(root, fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tool != domain.ToolInitialDefault {
		t.Fatalf("unexpected tool default: %s", cfg.Tool)
	}
	if cfg.MethodologyByTier[domain.TierT1].ID != domain.MethodologySpecWorkflow || cfg.MethodologyByTier[domain.TierT1].Mode != domain.MethodologyModeFull {
		t.Fatalf("unexpected T1 methodology default: %#v", cfg.MethodologyByTier[domain.TierT1])
	}
	if cfg.MethodologyByTier[domain.TierT3].ID != domain.MethodologyNone || cfg.MethodologyByTier[domain.TierT3].Mode != domain.MethodologyModeNone {
		t.Fatalf("unexpected T3 methodology default: %#v", cfg.MethodologyByTier[domain.TierT3])
	}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"tool: opencode", "methodology_by_tier:", "T1:", "id: openspec", "mode: full", "T3:", "id: none"} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated config missing harness default %q:\n%s", want, text)
		}
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
	cfg.WorkflowLimits.Sizing.LOCBudget = 250
	cfg.WorkflowLimits.DeliveryBatchStrategy = "single-pr"
	cfg.WorkflowLimits.ProposalSlicingStrategy = "custom-review-slices"
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
	if merged.WorkflowLimits.Sizing.LOCBudget != 250 || merged.WorkflowLimits.DeliveryBatchStrategy != "single-pr" || merged.WorkflowLimits.ProposalSlicingStrategy != "custom-review-slices" {
		t.Fatalf("rescan did not preserve workflow limits overrides: %#v", merged.WorkflowLimits)
	}
	_ = requireStack(t, merged, "typescript")
}

func TestRescanMergesPartialMethodologyByTierWithDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
tool: opencode
methodology_by_tier:
  T3:
    id: openspec
    mode: lite
    required: false
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
workflow_limits:
  sizing:
    loc_budget: 250
`)

	var out strings.Builder
	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=harness") || !strings.Contains(out.String(), "status=applied") {
		t.Fatalf("rescan did not report harness default completion: %s", out.String())
	}
	merged, err := Load(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	if merged.MethodologyByTier[domain.TierT1].ID != domain.MethodologySpecWorkflow || merged.MethodologyByTier[domain.TierT2].Mode != domain.MethodologyModeLite {
		t.Fatalf("rescan did not fill missing methodology defaults: %#v", merged.MethodologyByTier)
	}
	if merged.MethodologyByTier[domain.TierT3].ID != domain.MethodologySpecWorkflow || merged.MethodologyByTier[domain.TierT3].Mode != domain.MethodologyModeLite {
		t.Fatalf("rescan did not preserve T3 override: %#v", merged.MethodologyByTier)
	}
}

func TestRescanMergesPartialWorkflowLimitsWithDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
workflow_limits:
  sizing:
    loc_budget: 250
`)

	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	merged, err := Load(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	if merged.WorkflowLimits.Sizing.LOCBudget != 250 {
		t.Fatalf("rescan did not preserve partial sizing override: %#v", merged.WorkflowLimits)
	}
	if merged.WorkflowLimits.Routing.Strategy == "" || merged.WorkflowLimits.ProposalSlicingStrategy == "" || merged.WorkflowLimits.DeliveryBatchStrategy == "" {
		t.Fatalf("rescan did not fill missing workflow defaults: %#v", merged.WorkflowLimits)
	}
	if len(merged.WorkflowLimits.StopRules) == 0 || len(merged.WorkflowLimits.Preflight) == 0 {
		t.Fatalf("rescan did not fill missing workflow gates: %#v", merged.WorkflowLimits)
	}
}

func TestRescanPreservesProjectProfileAndAddsDetectedSurfaces(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, "cmd/app/main.go", "package main\n")
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
tool: opencode
methodology_by_tier:
  T1:
    id: openspec
    mode: full
    required: true
  T2:
    id: openspec
    mode: lite
    required: true
  T3:
    id: none
    mode: none
    required: false
project_profile:
  surfaces:
    - id: custom-product
      type: backend
      roots: [service]
      stacks: [go]
      frameworks: []
      agent_lens:
        primary_concerns: [custom-domain]
        validation_expectations: [custom-validation]
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
workflow_limits:
  sizing:
    loc_budget: 400
  routing:
    strategy: proportional-sdd
  proposal_slicing_strategy: review-slices-on-multi-risk
  delivery_batch_strategy: ask-on-risk
  stop_rules: [pause_on_scope_growth]
  preflight: [read_project_config]
`)

	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	custom := requireSurface(t, cfg, "custom-product")
	if custom.Type != "backend" || !contains(custom.AgentLens.PrimaryConcerns, "custom-domain") {
		t.Fatalf("manual surface not preserved: %#v", custom)
	}
	if custom.Architecture.Preferred != "controller_service_repository" || !custom.Architecture.ReviewRequired {
		t.Fatalf("manual backend surface did not receive architecture defaults: %#v", custom.Architecture)
	}
	if surface := requireSurface(t, cfg, "cli"); surface.Type != "cli" {
		t.Fatalf("detected CLI surface not added: %#v", cfg.ProjectProfile.Surfaces)
	}
}

func TestLoadRejectsUnsupportedHarnessConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
tool: other
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
workflow_limits: {}
`)

	if _, err := Load(filepath.Join(root, ProjectConfigPath)); err == nil {
		t.Fatalf("expected unsupported harness config to fail")
	}
}

func TestRescanPersistsOnlyWorkflowLimitDefaultCompletion(t *testing.T) {
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
	cfg.WorkflowLimits = WorkflowLimits{Sizing: WorkflowSizing{LOCBudget: 250}}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := svc.Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=workflow_limits") || !strings.Contains(out.String(), "status=applied") {
		t.Fatalf("rescan did not report workflow default completion: %s", out.String())
	}
	merged, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if merged.WorkflowLimits.Sizing.LOCBudget != 250 || merged.WorkflowLimits.Routing.Strategy == "" || len(merged.WorkflowLimits.Preflight) == 0 {
		t.Fatalf("rescan did not persist completed workflow defaults: %#v", merged.WorkflowLimits)
	}
}

func TestRescanAddsMissingWorkflowLimitsWithoutOtherDrift(t *testing.T) {
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
	cfg.WorkflowLimits = WorkflowLimits{}
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := svc.Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=workflow_limits") || !strings.Contains(out.String(), "status=applied") {
		t.Fatalf("rescan did not report missing workflow_limits completion: %s", out.String())
	}
	merged, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if isZeroWorkflowLimits(merged.WorkflowLimits) || merged.WorkflowLimits.Routing.Strategy == "" || len(merged.WorkflowLimits.StopRules) == 0 {
		t.Fatalf("rescan did not persist missing workflow defaults: %#v", merged.WorkflowLimits)
	}
}

func TestWorkflowLimitsNestedExtrasDoNotDropDefaults(t *testing.T) {
	detected := ProjectConfig{WorkflowLimits: defaultWorkflowLimits()}
	current := detected
	current.WorkflowLimits = WorkflowLimits{
		Sizing:  WorkflowSizing{Extra: map[string]any{"custom_sizing": "keep"}},
		Routing: WorkflowRouting{Extra: map[string]any{"custom_routing": "keep"}},
	}

	plan := BuildRescanPlan(current, detected)
	if !plan.HasChanges {
		t.Fatalf("expected workflow default completion to require write: %#v", plan)
	}
	if plan.Merged.WorkflowLimits.Sizing.LOCBudget != 400 || plan.Merged.WorkflowLimits.Routing.Strategy == "" {
		t.Fatalf("nested extras dropped defaults: %#v", plan.Merged.WorkflowLimits)
	}
	if plan.Merged.WorkflowLimits.Sizing.Extra["custom_sizing"] != "keep" || plan.Merged.WorkflowLimits.Routing.Extra["custom_routing"] != "keep" {
		t.Fatalf("nested extras were not preserved: %#v", plan.Merged.WorkflowLimits)
	}
}

func TestRescanReportsAndRemovesLegacyWorkflowFields(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, ProjectConfigPath, `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
loc_budget: 250
delivery_strategy: single-pr
workflow_limits:
  sizing:
    loc_budget: 300
  routing:
    strategy: custom
  proposal_slicing_strategy: custom-slices
  delivery_batch_strategy: custom-batches
  stop_rules: [custom-stop]
  preflight: [custom-preflight]
`)

	var out strings.Builder
	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "category=workflow_limits") || !strings.Contains(out.String(), "status=unsupported") {
		t.Fatalf("rescan report missing legacy workflow warning: %s", out.String())
	}
	mergedData, err := os.ReadFile(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	text := string(mergedData)
	assertWorkflowLimitsOnly(t, text)
	for _, want := range []string{"loc_budget: 300", "strategy: custom", "proposal_slicing_strategy: custom-slices", "delivery_batch_strategy: custom-batches", "custom-stop", "custom-preflight"} {
		if !strings.Contains(text, want) {
			t.Fatalf("merged config missing workflow override %q:\n%s", want, text)
		}
	}
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
workflow_limits:
  sizing:
    loc_budget: 250
  routing:
    strategy: custom
  proposal_slicing_strategy: custom-review-slices
  delivery_batch_strategy: single-pr
  stop_rules: [custom-stop]
  preflight: [custom-preflight]
`)
	if err := (Service{Now: fixedTime}).Run(Options{Target: root, Rescan: true}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ProjectConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"custom_top: keep-me", "custom_stack: keep-stack", "coverage_threshold: 70", "custom-go-smell", "proposal_slicing_strategy: custom-review-slices", "delivery_batch_strategy: single-pr"} {
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

func TestScanShallowDetectsSupportedAndUnsupportedStacks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/app\n\ngo 1.22\n")
	writeFile(t, root, "package.json", `{"scripts":{"test":"vitest","lint":"eslint ."},"dependencies":{"next":"latest","react":"latest","@opentelemetry/api":"latest"},"devDependencies":{"prettier":"latest"}}`)
	writeFile(t, root, "tsconfig.json", `{}`)
	writeFile(t, root, "pnpm-lock.yaml", "lock\n")
	writeFile(t, root, "pyproject.toml", "[project]\ndependencies=['fastapi','pytest','ruff']\n")
	writeFile(t, root, "pom.xml", "<project></project>")
	writeFile(t, root, "Cargo.toml", "[package]\nname='rusty'\n")
	writeFile(t, root, "composer.json", `{}`)
	writeFile(t, root, "Gemfile", "source 'https://rubygems.org'\n")
	writeFile(t, root, "mix.exs", "defmodule App.MixProject do end\n")
	writeFile(t, root, "app.csproj", "<Project />")

	stacks, err := ScanShallow(root)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]Stack{}
	for _, stack := range stacks {
		ids[stack.ID] = stack
	}
	for _, id := range []string{"go", "typescript", "python", "java", "rust", "php", "ruby", "elixir", "dotnet"} {
		if _, ok := ids[id]; !ok {
			t.Fatalf("missing stack %s in %#v", id, stacks)
		}
	}
	if ids["typescript"].PackageManager != "pnpm" || !contains(ids["typescript"].Frameworks, "next") || !contains(ids["typescript"].ObservabilityLibs, "@opentelemetry/api") {
		t.Fatalf("typescript stack unexpected: %#v", ids["typescript"])
	}
	if ids["rust"].Supported || ids["php"].Supported {
		t.Fatalf("unsupported stacks should be marked unsupported: rust=%#v php=%#v", ids["rust"], ids["php"])
	}
}

func TestProjectConfigHelpers(t *testing.T) {
	stacks := upsertStack([]Stack{{ID: "typescript", Frameworks: []string{"react"}, ObservabilityLibs: []string{"pino"}}}, Stack{ID: "typescript", Frameworks: []string{"next", "react"}, ObservabilityLibs: []string{"pino", "winston"}})
	if len(stacks) != 1 || !contains(stacks[0].Frameworks, "next") || !contains(stacks[0].ObservabilityLibs, "winston") {
		t.Fatalf("upsertStack did not merge unique fields: %#v", stacks)
	}
	stacks = upsertStack(stacks, Stack{ID: "go"})
	if len(stacks) != 2 {
		t.Fatalf("upsertStack did not append new stack: %#v", stacks)
	}
	if equalStrings([]string{"a"}, []string{"a", "b"}) || equalStrings([]string{"a"}, []string{"b"}) || !equalStrings([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatalf("equalStrings returned unexpected result")
	}
	if stackSummary(nil) != "ninguno" || !strings.Contains(stackSummary([]Stack{{ID: "old", Deprecated: true}}), "deprecated") {
		t.Fatalf("stackSummary unexpected")
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

func requireSurface(t *testing.T, cfg ProjectConfig, id string) ProjectSurface {
	t.Helper()
	for _, surface := range cfg.ProjectProfile.Surfaces {
		if surface.ID == id {
			return surface
		}
	}
	t.Fatalf("surface %s not found in %#v", id, cfg.ProjectProfile.Surfaces)
	return ProjectSurface{}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertWorkflowLimitsOnly(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "workflow_limits:") {
		t.Fatalf("missing workflow_limits block:\n%s", text)
	}
	for _, legacy := range []string{"\nloc_budget:", "\ndelivery_strategy:"} {
		if strings.Contains(text, legacy) {
			t.Fatalf("found legacy top-level workflow field %q:\n%s", legacy, text)
		}
	}
}
