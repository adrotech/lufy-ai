package projectconfig

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

type Scanner struct {
	Now func() time.Time
}

func Scan(root string, now time.Time) (ProjectConfig, error) {
	return Scanner{Now: func() time.Time { return now }}.Scan(root)
}

func (s Scanner) Scan(root string) (ProjectConfig, error) {
	now := s.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	stacks := scanRootStacks(root)
	for _, stack := range scanCommonSubdirs(root) {
		stacks = upsertStack(stacks, stack)
	}
	sort.SliceStable(stacks, func(i, j int) bool { return stacks[i].ID < stacks[j].ID })
	return ProjectConfig{
		SchemaVersion:     SchemaVersion,
		DetectedAt:        now().UTC(),
		Tool:              domain.ToolInitialDefault,
		MethodologyByTier: domain.DefaultMethodologyByTier(),
		ProjectProfile:    detectProjectProfile(root, stacks),
		Stacks:            stacks,
		CI:                scanCI(root),
		TDD:               defaultTDD(),
		Validation:        scanValidation(root, stacks),
		WorkflowLimits:    defaultWorkflowLimits(),
		ContextGraph:      DefaultContextGraphConfig(),
		Memory:            DefaultMemoryConfig(),
		ParallelExecution: DefaultParallelExecutionConfig(),
	}, nil
}

func scanCommonSubdirs(root string) []Stack {
	var stacks []Stack
	patterns := []string{"frontend", "backend", "web", "api", "apps/*", "packages/*"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(root, pattern))
		for _, match := range matches {
			if info, err := statPath(match); err != nil || !info.IsDir() {
				continue
			}
			cfg, _ := ScanShallow(match)
			stacks = append(stacks, cfg...)
		}
	}
	return stacks
}

func ScanShallow(root string) ([]Stack, error) {
	return scanRootStacks(root), nil
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

func pythonFrameworks(text string) []string {
	return containsAny(text, []string{"fastapi", "django", "flask"})
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
		parts = append(parts, stack.ID+" ("+state+")")
	}
	return strings.Join(parts, ", ")
}
