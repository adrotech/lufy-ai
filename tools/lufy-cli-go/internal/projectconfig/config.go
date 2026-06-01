package projectconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"gopkg.in/yaml.v3"
)

const (
	ProjectConfigPath = ".lufy/project.yaml"
	SchemaVersion     = 1
)

type Options struct {
	Target string
	Force  bool
	Rescan bool
}

type Service struct {
	Now func() time.Time
}

func NewService() Service {
	return Service{Now: func() time.Time { return time.Now().UTC() }}
}

type ProjectConfig struct {
	SchemaVersion     int                      `yaml:"schema_version"`
	DetectedAt        time.Time                `yaml:"detected_at"`
	Tool              domain.ToolID            `yaml:"tool"`
	MethodologyByTier domain.MethodologyByTier `yaml:"methodology_by_tier"`
	Stacks            []Stack                  `yaml:"stacks"`
	CI                CIConfig                 `yaml:"ci"`
	TDD               TDDConfig                `yaml:"tdd"`
	Validation        ValidationConfig         `yaml:"validation"`
	WorkflowLimits    WorkflowLimits           `yaml:"workflow_limits"`
	Extra             map[string]any           `yaml:",inline,omitempty"`
}

type Stack struct {
	ID                string         `yaml:"id"`
	Supported         bool           `yaml:"supported"`
	Deprecated        bool           `yaml:"deprecated,omitempty"`
	Version           string         `yaml:"version,omitempty"`
	PackageManager    string         `yaml:"package_manager,omitempty"`
	Frameworks        []string       `yaml:"frameworks"`
	TestRunner        CommandConfig  `yaml:"test_runner"`
	Linter            CommandConfig  `yaml:"linter"`
	Formatter         Formatter      `yaml:"formatter"`
	StaticAnalysis    CommandConfig  `yaml:"static_analysis"`
	AntiPatterns      []string       `yaml:"anti_patterns"`
	ObservabilityLibs []string       `yaml:"observability_libs"`
	Notes             string         `yaml:"notes,omitempty"`
	Extra             map[string]any `yaml:",inline,omitempty"`
}

type RescanPlan struct {
	Merged     ProjectConfig
	Items      []DriftItem
	HasChanges bool
}

type DriftItem struct {
	Category        string
	Severity        string
	Path            string
	StackID         string
	Status          string
	SuggestedAction string
}

type CommandConfig struct {
	Command           string `yaml:"command,omitempty"`
	CoverageCommand   string `yaml:"coverage_command,omitempty"`
	CoverageThreshold int    `yaml:"coverage_threshold,omitempty"`
	AutoFix           string `yaml:"auto_fix,omitempty"`
}

type Formatter struct {
	Command        string   `yaml:"command,omitempty"`
	FileExtensions []string `yaml:"file_extensions"`
}

type CIConfig struct {
	Detected  bool     `yaml:"detected"`
	Provider  string   `yaml:"provider,omitempty"`
	Workflows []string `yaml:"workflows"`
}

type TDDConfig struct {
	Strict              bool     `yaml:"strict"`
	TriangulateRequired bool     `yaml:"triangulate_required"`
	EdgeCaseCategories  []string `yaml:"edge_case_categories"`
}

type ValidationConfig struct {
	AllowedCommands ValidationAllowedCommands `yaml:"allowed_commands"`
}

type ValidationAllowedCommands struct {
	Implementer []string `yaml:"implementer"`
}

type WorkflowLimits struct {
	Sizing                  WorkflowSizing  `yaml:"sizing"`
	Routing                 WorkflowRouting `yaml:"routing"`
	ProposalSlicingStrategy string          `yaml:"proposal_slicing_strategy"`
	DeliveryBatchStrategy   string          `yaml:"delivery_batch_strategy"`
	StopRules               []string        `yaml:"stop_rules"`
	Preflight               []string        `yaml:"preflight"`
	Extra                   map[string]any  `yaml:",inline,omitempty"`
}

type WorkflowSizing struct {
	LOCBudget int            `yaml:"loc_budget"`
	Extra     map[string]any `yaml:",inline,omitempty"`
}

type WorkflowRouting struct {
	Strategy string         `yaml:"strategy"`
	Extra    map[string]any `yaml:",inline,omitempty"`
}

func (s Service) Run(opts Options, out io.Writer) error {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return fmt.Errorf("resolver target: %w", err)
	}
	configPath := filepath.Join(target, ProjectConfigPath)
	_, statErr := os.Stat(configPath)
	exists := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	if exists && !opts.Force && !opts.Rescan {
		return fmt.Errorf("%s ya existe; usa --rescan para preservar overrides o --force para reemplazarlo", ProjectConfigPath)
	}

	detected, err := Scan(target, s.Now())
	if err != nil {
		return err
	}
	finalConfig := detected
	var report *RescanPlan
	if opts.Rescan && exists {
		current, err := Load(configPath)
		if err != nil {
			return fmt.Errorf("leer config existente para rescan: corrige o respalda %s antes de reintentar: %w", ProjectConfigPath, err)
		}
		plan := BuildRescanPlan(current, detected)
		report = &plan
		finalConfig = plan.Merged
		if !plan.HasChanges {
			printRescanReport(out, plan)
			return nil
		}
	}
	content, err := Marshal(finalConfig)
	if err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(configPath, content, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "Generado %s\n", configPath)
	fmt.Fprintf(out, "Stacks detectados: %s\n", stackSummary(finalConfig.Stacks))
	if report != nil {
		printRescanReport(out, *report)
	}
	return nil
}

func (s Service) Ensure(target string) (bool, error) {
	target, err := platform.ResolveTargetPath(target)
	if err != nil {
		return false, fmt.Errorf("resolver target: %w", err)
	}
	configPath := filepath.Join(target, ProjectConfigPath)
	if _, err := os.Stat(configPath); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	detected, err := Scan(target, s.Now())
	if err != nil {
		return false, err
	}
	content, err := Marshal(detected)
	if err != nil {
		return false, err
	}
	if err := platform.WriteFileAtomic(configPath, content, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func Load(path string) (ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return ProjectConfig{}, err
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return ProjectConfig{}, fmt.Errorf("formato inválido: se esperaba un mapa YAML")
	}
	var cfg ProjectConfig
	if err := node.Content[0].Decode(&cfg); err != nil {
		return ProjectConfig{}, err
	}
	if err := validateHarnessConfig(cfg); err != nil {
		return ProjectConfig{}, err
	}
	return cfg, nil
}

func Marshal(cfg ProjectConfig) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return append([]byte("# Generated by lufy-ai init. Edit this file to tune project rules.\n"), data...), nil
}

func Scan(root string, now time.Time) (ProjectConfig, error) {
	stacks := []Stack{}
	if exists(root, "go.mod") {
		stacks = append(stacks, scanGo(root))
	}
	if exists(root, "package.json") {
		stacks = append(stacks, scanJS(root))
	}
	if existsAny(root, "pyproject.toml", "requirements.txt", "setup.py") {
		stacks = append(stacks, scanPython(root))
	}
	if existsAny(root, "pom.xml", "build.gradle", "build.gradle.kts") {
		stacks = append(stacks, scanJVM(root))
	}
	if exists(root, "Cargo.toml") {
		stacks = append(stacks, unsupportedStack("rust", "cargo", ".rs", "Soporte oficial pendiente. Completar manualmente o esperar release."))
	}
	stacks = append(stacks, scanUnsupportedStacks(root)...)
	for _, stack := range scanCommonSubdirs(root) {
		stacks = upsertStack(stacks, stack)
	}
	sort.SliceStable(stacks, func(i, j int) bool { return stacks[i].ID < stacks[j].ID })
	return ProjectConfig{
		SchemaVersion:     SchemaVersion,
		DetectedAt:        now.UTC(),
		Tool:              domain.ToolInitialDefault,
		MethodologyByTier: domain.DefaultMethodologyByTier(),
		Stacks:            stacks,
		CI:                scanCI(root),
		TDD:               defaultTDD(),
		Validation:        scanValidation(root, stacks),
		WorkflowLimits:    defaultWorkflowLimits(),
	}, nil
}

func MergeRescan(current, detected ProjectConfig) ProjectConfig {
	return BuildRescanPlan(current, detected).Merged
}

func BuildRescanPlan(current, detected ProjectConfig) RescanPlan {
	merged := detected
	merged.Extra = sanitizedTopLevelExtra(current.Extra)
	items := legacyWorkflowFieldItems(current.Extra)
	if item, changed := mergeHarnessConfig(&merged, current); changed {
		items = append(items, item)
	}
	if !isZeroWorkflowLimits(current.WorkflowLimits) {
		merged.WorkflowLimits = mergeWorkflowLimits(current.WorkflowLimits, detected.WorkflowLimits)
		if !reflect.DeepEqual(current.WorkflowLimits, merged.WorkflowLimits) {
			items = append(items, DriftItem{Category: "workflow_limits", Severity: "info", Path: "workflow_limits", Status: "applied", SuggestedAction: "Se completaron defaults canónicos faltantes dentro de workflow_limits preservando overrides existentes."})
		}
	} else {
		items = append(items, DriftItem{Category: "workflow_limits", Severity: "info", Path: "workflow_limits", Status: "applied", SuggestedAction: "Se agregó workflow_limits como fuente canónica única para sizing, slicing, delivery, stop rules y preflight."})
	}
	if len(current.TDD.EdgeCaseCategories) > 0 || current.TDD.Strict || current.TDD.TriangulateRequired {
		merged.TDD = current.TDD
	}
	if !isZeroValidationConfig(current.Validation) {
		merged.Validation = mergeValidationConfig(current.Validation, detected.Validation)
		if !reflect.DeepEqual(current.Validation, merged.Validation) {
			items = append(items, DriftItem{Category: "validation", Severity: "info", Path: "validation.allowed_commands", Status: "applied", SuggestedAction: "Se preservaron comandos permitidos manuales y se agregaron comandos detectados faltantes."})
		}
	} else if !isZeroValidationConfig(detected.Validation) {
		items = append(items, DriftItem{Category: "validation", Severity: "info", Path: "validation.allowed_commands", Status: "applied", SuggestedAction: "Se agregó allowlist de validación para roles a partir del toolchain detectado."})
	}
	if !reflect.DeepEqual(current.CI, detected.CI) {
		items = append(items, DriftItem{Category: "ci", Severity: "info", Path: "ci", Status: "applied", SuggestedAction: "Revisa workflows detectados y ajusta manualmente si necesitas una política distinta."})
	}
	currentByID := map[string]Stack{}
	for _, stack := range current.Stacks {
		currentByID[stack.ID] = stack
	}
	for i := range merged.Stacks {
		if old, ok := currentByID[merged.Stacks[i].ID]; ok {
			fresh := merged.Stacks[i]
			merged.Stacks[i] = preserveUserManaged(old, fresh)
			if stackGeneratedDrift(old, fresh) {
				items = append(items, DriftItem{Category: "tooling", Severity: "info", Path: fmt.Sprintf("stacks.%s", fresh.ID), StackID: fresh.ID, Status: "applied", SuggestedAction: "Revisa los comandos detectados y re-aplica overrides manuales solo si corresponden."})
			}
			delete(currentByID, merged.Stacks[i].ID)
		} else {
			items = append(items, DriftItem{Category: "stack", Severity: "info", Path: fmt.Sprintf("stacks.%s", merged.Stacks[i].ID), StackID: merged.Stacks[i].ID, Status: "applied", SuggestedAction: "Revisa la configuración generada para el nuevo stack y completa placeholders si existen."})
		}
	}
	for _, old := range currentByID {
		status := "detected"
		if !old.Deprecated {
			old.Deprecated = true
			status = "applied"
		}
		items = append(items, DriftItem{Category: "stale", Severity: "warning", Path: fmt.Sprintf("stacks.%s", old.ID), StackID: old.ID, Status: status, SuggestedAction: "Verifica si el stack fue removido; el rescan no borra entradas ni archivos automáticamente."})
		merged.Stacks = append(merged.Stacks, old)
	}
	sort.SliceStable(merged.Stacks, func(i, j int) bool { return merged.Stacks[i].ID < merged.Stacks[j].ID })
	hasChanges := len(items) > 0 && hasProjectMutation(current, merged)
	if !hasChanges {
		merged.DetectedAt = current.DetectedAt
	}
	return RescanPlan{Merged: merged, Items: items, HasChanges: hasChanges}
}

func validateHarnessConfig(cfg ProjectConfig) error {
	return domain.HarnessConfig{
		Tool:              cfg.Tool,
		MethodologyByTier: cfg.MethodologyByTier,
	}.ValidateSupported()
}

func mergeHarnessConfig(merged *ProjectConfig, current ProjectConfig) (DriftItem, bool) {
	if merged.Tool == "" {
		merged.Tool = domain.ToolInitialDefault
	}
	if len(merged.MethodologyByTier) == 0 {
		merged.MethodologyByTier = domain.DefaultMethodologyByTier()
	}
	before := domain.HarnessConfig{
		Tool:              current.Tool,
		MethodologyByTier: current.MethodologyByTier,
	}.WithDefaults()
	if current.Tool != "" {
		merged.Tool = current.Tool
	}
	if len(current.MethodologyByTier) > 0 {
		merged.MethodologyByTier = current.MethodologyByTier.WithDefaults()
	}
	after := domain.HarnessConfig{
		Tool:              merged.Tool,
		MethodologyByTier: merged.MethodologyByTier,
	}.WithDefaults()
	if current.Tool != "" && len(current.MethodologyByTier) >= len(domain.DefaultMethodologyByTier()) && reflect.DeepEqual(before, after) {
		return DriftItem{}, false
	}
	return DriftItem{
		Category:        "harness",
		Severity:        "info",
		Path:            "tool,methodology_by_tier",
		Status:          "applied",
		SuggestedAction: "Se completaron defaults de tool y methodology_by_tier preservando overrides compatibles.",
	}, true
}

func sanitizedTopLevelExtra(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return extra
	}
	cleaned := map[string]any{}
	for key, value := range extra {
		if key == "loc_budget" || key == "delivery_strategy" {
			continue
		}
		cleaned[key] = value
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func legacyWorkflowFieldItems(extra map[string]any) []DriftItem {
	items := []DriftItem{}
	for _, key := range []string{"loc_budget", "delivery_strategy"} {
		if _, ok := extra[key]; ok {
			items = append(items, DriftItem{Category: "workflow_limits", Severity: "warning", Path: key, Status: "unsupported", SuggestedAction: "Mueve este override legacy a workflow_limits; los campos top-level loc_budget y delivery_strategy ya no son fuentes canónicas."})
		}
	}
	return items
}

func isZeroWorkflowLimits(limits WorkflowLimits) bool {
	return reflect.DeepEqual(limits, WorkflowLimits{})
}

func mergeWorkflowLimits(current, defaults WorkflowLimits) WorkflowLimits {
	merged := defaults
	if current.Sizing.LOCBudget != 0 || len(current.Sizing.Extra) > 0 {
		merged.Sizing = mergeWorkflowSizing(current.Sizing, defaults.Sizing)
	}
	if current.Routing.Strategy != "" || len(current.Routing.Extra) > 0 {
		merged.Routing = mergeWorkflowRouting(current.Routing, defaults.Routing)
	}
	if current.ProposalSlicingStrategy != "" {
		merged.ProposalSlicingStrategy = current.ProposalSlicingStrategy
	}
	if current.DeliveryBatchStrategy != "" {
		merged.DeliveryBatchStrategy = current.DeliveryBatchStrategy
	}
	if len(current.StopRules) > 0 {
		merged.StopRules = current.StopRules
	}
	if len(current.Preflight) > 0 {
		merged.Preflight = current.Preflight
	}
	if len(current.Extra) > 0 {
		merged.Extra = current.Extra
	}
	return merged
}

func isZeroValidationConfig(config ValidationConfig) bool {
	return len(config.AllowedCommands.Implementer) == 0
}

func mergeValidationConfig(current, detected ValidationConfig) ValidationConfig {
	merged := current
	merged.AllowedCommands.Implementer = unique(append(append([]string{}, current.AllowedCommands.Implementer...), detected.AllowedCommands.Implementer...))
	return merged
}

func mergeWorkflowSizing(current, defaults WorkflowSizing) WorkflowSizing {
	merged := defaults
	if current.LOCBudget != 0 {
		merged.LOCBudget = current.LOCBudget
	}
	if len(current.Extra) > 0 {
		merged.Extra = current.Extra
	}
	return merged
}

func mergeWorkflowRouting(current, defaults WorkflowRouting) WorkflowRouting {
	merged := defaults
	if current.Strategy != "" {
		merged.Strategy = current.Strategy
	}
	if len(current.Extra) > 0 {
		merged.Extra = current.Extra
	}
	return merged
}

func hasProjectMutation(current, merged ProjectConfig) bool {
	currentComparable := current
	mergedComparable := merged
	mergedComparable.DetectedAt = currentComparable.DetectedAt
	currentBytes, currentErr := yaml.Marshal(currentComparable)
	mergedBytes, mergedErr := yaml.Marshal(mergedComparable)
	if currentErr == nil && mergedErr == nil {
		return string(currentBytes) != string(mergedBytes)
	}
	return !reflect.DeepEqual(currentComparable, mergedComparable)
}

func preserveUserManaged(old, fresh Stack) Stack {
	if old.TestRunner.CoverageThreshold != 0 && old.TestRunner.CoverageThreshold != fresh.TestRunner.CoverageThreshold {
		fresh.TestRunner.CoverageThreshold = old.TestRunner.CoverageThreshold
	}
	if len(old.AntiPatterns) > 0 && !equalStrings(old.AntiPatterns, fresh.AntiPatterns) {
		fresh.AntiPatterns = old.AntiPatterns
	}
	if old.Notes != "" {
		fresh.Notes = old.Notes
	}
	if len(old.Extra) > 0 {
		fresh.Extra = old.Extra
	}
	return fresh
}

func stackGeneratedDrift(old, fresh Stack) bool {
	oldGenerated := old
	freshGenerated := fresh
	oldGenerated.TestRunner.CoverageThreshold = freshGenerated.TestRunner.CoverageThreshold
	oldGenerated.AntiPatterns = freshGenerated.AntiPatterns
	oldGenerated.Notes = freshGenerated.Notes
	oldGenerated.Extra = freshGenerated.Extra
	oldBytes, oldErr := yaml.Marshal(oldGenerated)
	freshBytes, freshErr := yaml.Marshal(freshGenerated)
	if oldErr == nil && freshErr == nil {
		return string(oldBytes) != string(freshBytes)
	}
	return !reflect.DeepEqual(oldGenerated, freshGenerated)
}

func printRescanReport(out io.Writer, plan RescanPlan) {
	if len(plan.Items) == 0 {
		fmt.Fprintf(out, "Rescan completado: sin drift detectable; %s no fue reescrito.\n", ProjectConfigPath)
		return
	}
	fmt.Fprintln(out, "Rescan drift report:")
	for _, item := range plan.Items {
		stack := item.StackID
		if stack == "" {
			stack = "-"
		}
		fmt.Fprintf(out, "- category=%s severity=%s path=%s stack=%s status=%s action=%s\n", item.Category, item.Severity, item.Path, stack, item.Status, item.SuggestedAction)
	}
}

func scanGo(root string) Stack {
	version := ""
	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "go ") {
				version = strings.TrimSpace(strings.TrimPrefix(line, "go "))
				break
			}
		}
	}
	return Stack{ID: "go", Supported: true, Version: version, PackageManager: "go modules", Frameworks: []string{}, TestRunner: CommandConfig{Command: "go test ./...", CoverageCommand: "go test -coverprofile=coverage.out ./...", CoverageThreshold: 85}, Linter: CommandConfig{Command: goLinter(root), AutoFix: goLinterFix(root)}, Formatter: Formatter{Command: "gofmt -w", FileExtensions: []string{".go"}}, StaticAnalysis: CommandConfig{Command: "go vet ./..."}, AntiPatterns: []string{"mock.Anything", "gomonkey", "fmt.Println", "panic("}, ObservabilityLibs: detectInFiles(root, []string{"go.mod"}, []string{"go.opentelemetry.io", "datadog/dd-trace-go", "meli/fury_go-toolkit-otel"})}
}

func scanJS(root string) Stack {
	pkg := readPackageJSON(root)
	id := "javascript"
	if exists(root, "tsconfig.json") {
		id = "typescript"
	}
	frameworks := []string{}
	if hasDep(pkg, "react") {
		frameworks = append(frameworks, "react")
	}
	if hasDep(pkg, "next") {
		frameworks = append(frameworks, "react", "next")
	}
	if hasDepPrefix(pkg, "@remix-run/") {
		frameworks = append(frameworks, "react", "remix")
	}
	for _, fw := range []string{"vue", "svelte"} {
		if hasDep(pkg, fw) {
			frameworks = append(frameworks, fw)
		}
	}
	frameworks = unique(frameworks)
	test := "npm test"
	coverage := "npm test -- --coverage"
	for _, candidate := range []string{"vitest", "jest", "mocha"} {
		if hasDep(pkg, candidate) {
			test = candidate + " run"
			if candidate == "jest" || candidate == "mocha" {
				test = candidate
			}
			coverage = test + " --coverage"
			break
		}
	}
	linter := "TODO: eslint ."
	if hasDep(pkg, "eslint") || existsGlob(root, ".eslintrc*") {
		linter = "eslint ."
	}
	formatter := "TODO: prettier --write"
	if hasDep(pkg, "prettier") || existsGlob(root, ".prettierrc*") {
		formatter = "prettier --write"
	}
	static := ""
	if id == "typescript" {
		static = "tsc --noEmit"
	}
	return Stack{ID: id, Supported: true, PackageManager: jsPackageManager(root), Frameworks: frameworks, TestRunner: CommandConfig{Command: test, CoverageCommand: coverage, CoverageThreshold: 80}, Linter: CommandConfig{Command: linter, AutoFix: linter + " --fix"}, Formatter: Formatter{Command: formatter, FileExtensions: []string{".ts", ".tsx", ".js", ".jsx", ".json", ".md", ".css"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"console.log", "// @ts-ignore", "as any", "useEffect(() => {}, [])"}, ObservabilityLibs: detectPackageLibs(pkg, []string{"@opentelemetry/api", "pino", "winston"})}
}

func scanPython(root string) Stack {
	text := readKnown(root, "pyproject.toml", "requirements.txt", "setup.py")
	test := "python -m unittest"
	if strings.Contains(text, "pytest") {
		test = "pytest"
	}
	linter := "TODO: ruff check"
	fix := "TODO: ruff check --fix"
	formatter := "TODO: ruff format"
	if strings.Contains(text, "ruff") {
		linter, fix, formatter = "ruff check", "ruff check --fix", "ruff format"
	} else if strings.Contains(text, "flake8") {
		linter = "flake8"
	} else if strings.Contains(text, "pylint") {
		linter = "pylint"
	}
	if strings.Contains(text, "black") && !strings.Contains(text, "ruff") {
		formatter = "black ."
	}
	static := "TODO: mypy ."
	if strings.Contains(text, "mypy") {
		static = "mypy ."
	}
	pm := "pip"
	if exists(root, "uv.lock") {
		pm = "uv"
	} else if exists(root, "poetry.lock") {
		pm = "poetry"
	}
	return Stack{ID: "python", Supported: true, PackageManager: pm, Frameworks: pythonFrameworks(text), TestRunner: CommandConfig{Command: test, CoverageCommand: "pytest --cov", CoverageThreshold: 80}, Linter: CommandConfig{Command: linter, AutoFix: fix}, Formatter: Formatter{Command: formatter, FileExtensions: []string{".py"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"print(", "except:", "MagicMock()", "unittest.skip"}, ObservabilityLibs: containsAny(text, []string{"opentelemetry", "structlog"})}
}

func scanJVM(root string) Stack {
	id := "java"
	pm := "maven"
	test := "mvn test"
	coverage := "mvn test jacoco:report"
	lint := "mvn checkstyle:check"
	format := "mvn spotless:apply"
	static := "mvn spotbugs:check"
	if existsAny(root, "build.gradle", "build.gradle.kts") {
		pm, test, coverage, lint, format, static = "gradle", "gradle test", "gradle test jacocoTestReport", "gradle check", "gradle spotlessApply", "gradle spotbugsMain"
		if exists(root, "build.gradle.kts") {
			id = "kotlin"
		}
	}
	text := readKnown(root, "pom.xml", "build.gradle", "build.gradle.kts")
	frameworks := []string{}
	if strings.Contains(text, "spring-boot") || strings.Contains(text, "springframework.boot") {
		frameworks = append(frameworks, "spring-boot")
	}
	return Stack{ID: id, Supported: true, PackageManager: pm, Frameworks: frameworks, TestRunner: CommandConfig{Command: test, CoverageCommand: coverage, CoverageThreshold: 80}, Linter: CommandConfig{Command: lint}, Formatter: Formatter{Command: format, FileExtensions: []string{".java", ".kt"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"System.out.println", "e.printStackTrace()", "@SuppressWarnings(\"unchecked\")"}, ObservabilityLibs: containsAny(text, []string{"io.opentelemetry", "io.micrometer", "logback", "log4j"})}
}

func unsupportedStack(id, pm, ext, notes string) Stack {
	return Stack{ID: id, Supported: false, PackageManager: pm, Frameworks: []string{}, TestRunner: CommandConfig{Command: "TODO", CoverageCommand: "TODO", CoverageThreshold: 80}, Linter: CommandConfig{Command: "TODO"}, Formatter: Formatter{Command: "TODO", FileExtensions: []string{ext}}, StaticAnalysis: CommandConfig{}, AntiPatterns: []string{}, ObservabilityLibs: []string{}, Notes: notes}
}

func scanCI(root string) CIConfig {
	ci := CIConfig{Detected: false, Workflows: []string{}}
	if entries, err := os.ReadDir(filepath.Join(root, ".github/workflows")); err == nil {
		ci.Detected, ci.Provider = true, "github-actions"
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml")) {
				ci.Workflows = append(ci.Workflows, filepath.ToSlash(filepath.Join(".github/workflows", entry.Name())))
			}
		}
	}
	if exists(root, ".gitlab-ci.yml") {
		ci.Detected, ci.Provider = true, "gitlab-ci"
		ci.Workflows = append(ci.Workflows, ".gitlab-ci.yml")
	}
	if exists(root, "Jenkinsfile") {
		ci.Detected, ci.Provider = true, "jenkins"
		ci.Workflows = append(ci.Workflows, "Jenkinsfile")
	}
	if exists(root, ".circleci/config.yml") {
		ci.Detected, ci.Provider = true, "circleci"
		ci.Workflows = append(ci.Workflows, ".circleci/config.yml")
	}
	sort.Strings(ci.Workflows)
	return ci
}

func defaultTDD() TDDConfig {
	return TDDConfig{Strict: true, TriangulateRequired: true, EdgeCaseCategories: []string{"boundary", "error_path", "concurrency", "data_shape", "time_sensitive"}}
}

func scanValidation(root string, stacks []Stack) ValidationConfig {
	allowed := []string{}
	for _, stack := range stacks {
		if stack.PackageManager != "pnpm" || (stack.ID != "typescript" && stack.ID != "javascript") {
			continue
		}
		scripts := readPackageScripts(root)
		for _, script := range []string{"typecheck", "lint", "test", "build"} {
			if _, ok := scripts[script]; ok {
				allowed = append(allowed, "pnpm "+script+"*")
			}
		}
	}
	return ValidationConfig{AllowedCommands: ValidationAllowedCommands{Implementer: unique(allowed)}}
}

func defaultWorkflowLimits() WorkflowLimits {
	return WorkflowLimits{Sizing: WorkflowSizing{LOCBudget: 400}, Routing: WorkflowRouting{Strategy: "proportional-sdd"}, ProposalSlicingStrategy: "review-slices-on-multi-risk", DeliveryBatchStrategy: "ask-on-risk", StopRules: []string{"pause_on_scope_growth", "escalate_on_security_or_delivery_risk", "stop_before_unauthorized_git_or_gh"}, Preflight: []string{"read_project_config", "confirm_applicable_toolchain", "plan_grouped_validation"}}
}

func scanCommonSubdirs(root string) []Stack {
	var stacks []Stack
	patterns := []string{"frontend", "backend", "web", "api", "apps/*", "packages/*"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(root, pattern))
		for _, match := range matches {
			if info, err := os.Stat(match); err != nil || !info.IsDir() {
				continue
			}
			cfg, _ := ScanShallow(match)
			stacks = append(stacks, cfg...)
		}
	}
	return stacks
}

func ScanShallow(root string) ([]Stack, error) {
	var stacks []Stack
	if exists(root, "go.mod") {
		stacks = append(stacks, scanGo(root))
	}
	if exists(root, "package.json") {
		stacks = append(stacks, scanJS(root))
	}
	if existsAny(root, "pyproject.toml", "requirements.txt", "setup.py") {
		stacks = append(stacks, scanPython(root))
	}
	if existsAny(root, "pom.xml", "build.gradle", "build.gradle.kts") {
		stacks = append(stacks, scanJVM(root))
	}
	if exists(root, "Cargo.toml") {
		stacks = append(stacks, unsupportedStack("rust", "cargo", ".rs", "Soporte oficial pendiente. Completar manualmente o esperar release."))
	}
	stacks = append(stacks, scanUnsupportedStacks(root)...)
	return stacks, nil
}

func scanUnsupportedStacks(root string) []Stack {
	var stacks []Stack
	notes := "Soporte oficial pendiente. Completar manualmente o esperar release."
	if exists(root, "composer.json") {
		stacks = append(stacks, unsupportedStack("php", "composer", ".php", notes))
	}
	if exists(root, "Gemfile") {
		stacks = append(stacks, unsupportedStack("ruby", "bundler", ".rb", notes))
	}
	if exists(root, "mix.exs") {
		stacks = append(stacks, unsupportedStack("elixir", "mix", ".exs", notes))
	}
	if existsGlob(root, "*.csproj") || existsGlob(root, "*.sln") {
		stacks = append(stacks, unsupportedStack("dotnet", "dotnet", ".cs", notes))
	}
	return stacks
}

func readPackageJSON(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return map[string]string{}
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]string{}
	}
	deps := map[string]string{}
	for _, key := range []string{"dependencies", "devDependencies"} {
		if section, ok := raw[key].(map[string]any); ok {
			for name := range section {
				deps[name] = key
			}
		}
	}
	if section, ok := raw["scripts"].(map[string]any); ok {
		for _, value := range section {
			command, ok := value.(string)
			if !ok {
				continue
			}
			for _, tool := range []string{"vitest", "jest", "mocha", "eslint", "prettier"} {
				if strings.Contains(command, tool) {
					deps[tool] = "scripts"
				}
			}
		}
	}
	return deps
}

func readPackageScripts(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return map[string]string{}
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]string{}
	}
	scripts := map[string]string{}
	if section, ok := raw["scripts"].(map[string]any); ok {
		for name, value := range section {
			command, ok := value.(string)
			if ok {
				scripts[name] = command
			}
		}
	}
	return scripts
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

func existsAny(root string, rels ...string) bool {
	for _, rel := range rels {
		if exists(root, rel) {
			return true
		}
	}
	return false
}

func existsGlob(root, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(root, pattern))
	return len(matches) > 0
}

func hasDep(deps map[string]string, name string) bool { _, ok := deps[name]; return ok }

func hasDepPrefix(deps map[string]string, prefix string) bool {
	for name := range deps {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func jsPackageManager(root string) string {
	switch {
	case exists(root, "pnpm-lock.yaml"):
		return "pnpm"
	case exists(root, "yarn.lock"):
		return "yarn"
	case exists(root, "package-lock.json"):
		return "npm"
	default:
		return "npm"
	}
}

func goLinter(root string) string {
	if existsAny(root, ".golangci.yml", ".golangci.yaml") {
		return "golangci-lint run"
	}
	return "go vet ./..."
}

func goLinterFix(root string) string {
	if existsAny(root, ".golangci.yml", ".golangci.yaml") {
		return "golangci-lint run --fix"
	}
	return ""
}

func readKnown(root string, names ...string) string {
	var buf strings.Builder
	for _, name := range names {
		if data, err := os.ReadFile(filepath.Join(root, name)); err == nil {
			buf.Write(data)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func containsAny(text string, candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		if strings.Contains(text, candidate) {
			out = append(out, candidate)
		}
	}
	return out
}

func detectInFiles(root string, files, candidates []string) []string {
	return containsAny(readKnown(root, files...), candidates)
}

func detectPackageLibs(deps map[string]string, candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		if hasDep(deps, candidate) {
			out = append(out, candidate)
		}
	}
	return out
}

func pythonFrameworks(text string) []string {
	return containsAny(text, []string{"fastapi", "django", "flask"})
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func upsertStack(stacks []Stack, stack Stack) []Stack {
	for i := range stacks {
		if stacks[i].ID == stack.ID {
			stacks[i].Frameworks = unique(append(stacks[i].Frameworks, stack.Frameworks...))
			stacks[i].ObservabilityLibs = unique(append(stacks[i].ObservabilityLibs, stack.ObservabilityLibs...))
			return stacks
		}
	}
	return append(stacks, stack)
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stackSummary(stacks []Stack) string {
	if len(stacks) == 0 {
		return "ninguno"
	}
	parts := make([]string, 0, len(stacks))
	for _, stack := range stacks {
		state := "supported"
		if !stack.Supported {
			state = "unsupported"
		}
		if stack.Deprecated {
			state = "deprecated"
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", stack.ID, state))
	}
	return strings.Join(parts, ", ")
}
