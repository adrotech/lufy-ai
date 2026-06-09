package projectconfig

import (
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"gopkg.in/yaml.v3"
)

type RescanMerger struct{}

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

func MergeRescan(current, detected ProjectConfig) ProjectConfig {
	return BuildRescanPlan(current, detected).Merged
}

func BuildRescanPlan(current, detected ProjectConfig) RescanPlan {
	return RescanMerger{}.Build(current, detected)
}

func (m RescanMerger) Build(current, detected ProjectConfig) RescanPlan {
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
	if !isZeroMemoryConfig(current.Memory) {
		merged.Memory = mergeMemoryConfig(current.Memory, detected.Memory)
		if !reflect.DeepEqual(current.Memory, merged.Memory) {
			items = append(items, DriftItem{Category: "memory", Severity: "info", Path: "memory", Status: "applied", SuggestedAction: "Se completaron defaults de memoria Obsidian preservando overrides existentes."})
		}
	} else {
		items = append(items, DriftItem{Category: "memory", Severity: "info", Path: "memory", Status: "applied", SuggestedAction: "Se agregó memoria Obsidian portable como provider canónico del proyecto."})
	}
	if !isZeroParallelExecutionConfig(current.ParallelExecution) {
		merged.ParallelExecution = mergeParallelExecutionConfig(current.ParallelExecution, detected.ParallelExecution)
		if !reflect.DeepEqual(current.ParallelExecution, merged.ParallelExecution) {
			items = append(items, DriftItem{Category: "parallel_execution", Severity: "info", Path: "parallel_execution", Status: "applied", SuggestedAction: "Se completaron defaults de paralelismo gobernado preservando overrides existentes."})
		}
	} else {
		items = append(items, DriftItem{Category: "parallel_execution", Severity: "info", Path: "parallel_execution", Status: "applied", SuggestedAction: "Se agregó estrategia de paralelismo gobernada por review_slices independientes."})
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
	merged.ProjectProfile = mergeProjectProfile(current.ProjectProfile, detected.ProjectProfile)
	if !reflect.DeepEqual(current.ProjectProfile, merged.ProjectProfile) {
		items = append(items, DriftItem{Category: "project_profile", Severity: "info", Path: "project_profile.surfaces", Status: "applied", SuggestedAction: "Se preservaron superficies configuradas y se agregaron superficies nuevas detectadas."})
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

func mergeProjectProfile(current, detected ProjectProfile) ProjectProfile {
	if len(current.Surfaces) == 0 && len(current.Extra) == 0 {
		return detected
	}
	merged := current
	detectedByID := map[string]ProjectSurface{}
	for _, surface := range detected.Surfaces {
		detectedByID[surface.ID] = surface
	}
	byID := map[string]bool{}
	for i, surface := range merged.Surfaces {
		if isZeroArchitecture(surface.Architecture) {
			if fresh, ok := detectedByID[surface.ID]; ok && !isZeroArchitecture(fresh.Architecture) {
				surface.Architecture = fresh.Architecture
			}
		}
		merged.Surfaces[i] = ApplySurfaceDefaults(surface)
		byID[surface.ID] = true
	}
	for _, surface := range detected.Surfaces {
		if surface.ID == "" || byID[surface.ID] {
			continue
		}
		merged.Surfaces = append(merged.Surfaces, surface)
		byID[surface.ID] = true
	}
	sort.SliceStable(merged.Surfaces, func(i, j int) bool { return merged.Surfaces[i].ID < merged.Surfaces[j].ID })
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

func isZeroMemoryConfig(config MemoryConfig) bool {
	return reflect.DeepEqual(config, MemoryConfig{})
}

func mergeMemoryConfig(current, defaults MemoryConfig) MemoryConfig {
	merged := defaults
	if current.Provider != "" {
		merged.Provider = current.Provider
	}
	if current.Root != "" {
		merged.Root = current.Root
	}
	if current.GitPolicy != "" {
		merged.GitPolicy = current.GitPolicy
	}
	if current.SchemaVersion != 0 {
		merged.SchemaVersion = current.SchemaVersion
	}
	if current.Search != "" {
		merged.Search = current.Search
	}
	if current.BacklinksIndex != "" {
		merged.BacklinksIndex = current.BacklinksIndex
	}
	if len(current.Extra) > 0 {
		merged.Extra = current.Extra
	}
	return merged
}

func isZeroParallelExecutionConfig(config ParallelExecutionConfig) bool {
	return reflect.DeepEqual(config, ParallelExecutionConfig{})
}

func mergeParallelExecutionConfig(current, defaults ParallelExecutionConfig) ParallelExecutionConfig {
	merged := defaults
	merged.Enabled = current.Enabled
	if current.Strategy != "" {
		merged.Strategy = current.Strategy
	}
	if current.MaxParallelAgents != 0 {
		merged.MaxParallelAgents = current.MaxParallelAgents
	}
	if current.RequiresIndependentFiles {
		merged.RequiresIndependentFiles = true
	}
	if current.RequiresMergePlan {
		merged.RequiresMergePlan = true
	}
	if current.ValidationMode != "" {
		merged.ValidationMode = current.ValidationMode
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
