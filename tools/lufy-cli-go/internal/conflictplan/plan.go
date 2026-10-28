package conflictplan

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target  string
	JSON    bool
	Scope   assets.Scope
	Harness domain.HarnessConfig
}

type Report struct {
	TargetRoot  string   `json:"targetRoot"`
	OK          bool     `json:"ok"`
	Summary     Summary  `json:"summary"`
	Groups      []Group  `json:"groups"`
	Items       []Item   `json:"items"`
	NextActions []string `json:"nextActions"`
}

type Summary struct {
	Conflicts          int `json:"conflicts"`
	Groups             int `json:"groups"`
	HighRisk           int `json:"highRisk"`
	MediumRisk         int `json:"mediumRisk"`
	ParallelCandidates int `json:"parallelCandidates"`
}

type Group struct {
	Category      string   `json:"category"`
	Risk          string   `json:"risk"`
	Count         int      `json:"count"`
	Paths         []string `json:"paths"`
	ParallelGroup string   `json:"parallelGroup"`
}

type Item struct {
	Path             string   `json:"path"`
	Category         string   `json:"category"`
	Status           string   `json:"status"`
	Policy           string   `json:"policy,omitempty"`
	Risk             string   `json:"risk"`
	Recommendation   string   `json:"recommendation"`
	Reason           string   `json:"reason"`
	AvailableActions []string `json:"availableActions"`
	ParallelGroup    string   `json:"parallelGroup"`
	CurrentHash      string   `json:"currentHash,omitempty"`
	SourceHash       string   `json:"sourceHash,omitempty"`
}

type Service struct {
	installer installer.Service
}

func NewService() Service {
	return Service{installer: installer.NewService()}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	report, err := s.Build(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "%s\n", body)
		return err
	}
	printReport(report, stdout)
	return nil
}

func (s Service) Build(opts Options) (Report, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Report{}, err
	}
	plan, err := s.installer.BuildPlan(installer.Options{Target: target, Scope: opts.Scope, Harness: opts.Harness})
	if err != nil {
		return Report{}, err
	}
	report := Report{TargetRoot: target, OK: true}
	assetByPath := map[string]assets.Asset{}
	for _, asset := range plan.Catalog.Assets {
		assetByPath[filepath.ToSlash(asset.TargetRel)] = asset
	}
	for _, conflict := range plan.Conflicts {
		item := itemFromConflict(conflict, assetByPath)
		report.Items = append(report.Items, item)
	}
	sort.Slice(report.Items, func(i, j int) bool {
		if report.Items[i].Category == report.Items[j].Category {
			return report.Items[i].Path < report.Items[j].Path
		}
		return report.Items[i].Category < report.Items[j].Category
	})
	report.Groups = groupsForItems(report.Items)
	report.Summary = summaryFor(report)
	report.OK = len(report.Items) == 0
	report.NextActions = nextActions(report)
	return report, nil
}

func itemFromConflict(conflict installer.Conflict, assetByPath map[string]assets.Asset) Item {
	path := filepath.ToSlash(conflict.Path)
	asset := assetByPath[path]
	category := categoryFor(path)
	risk := conflict.Risk
	if risk == "" {
		risk = riskFor(path, conflict.Reason)
	}
	recommendation := recommendationFor(path, conflict.Reason, conflict.Policy)
	return Item{
		Path:             path,
		Category:         category,
		Status:           statusFor(conflict.Reason),
		Policy:           string(conflict.Policy),
		Risk:             risk,
		Recommendation:   recommendation,
		Reason:           reasonWithComponent(conflict.Reason, asset.Component),
		AvailableActions: availableActionsFor(recommendation),
		ParallelGroup:    parallelGroupFor(category),
		CurrentHash:      conflict.CurrentHash,
		SourceHash:       conflict.SourceHash,
	}
}

func categoryFor(path string) string {
	switch {
	case strings.HasPrefix(path, ".agents/skills/"):
		return ".agents/skills"
	case strings.HasPrefix(path, ".codex/"):
		return ".codex"
	case strings.HasPrefix(path, ".lufy/"):
		return ".lufy"
	case path == "lufy-ia.harness.md" || path == "AGENTS.md":
		return "root/config"
	default:
		return "other"
	}
}

func statusFor(reason string) string {
	if strings.Contains(reason, "no gestionado") {
		return "unmanaged_conflict"
	}
	if strings.Contains(reason, "drift") {
		return "managed_drift"
	}
	if strings.Contains(reason, "symlink") || strings.Contains(reason, "archivo regular") {
		return "unsafe_target"
	}
	return "blocked_conflict"
}

func riskFor(path, reason string) string {
	if strings.Contains(reason, "symlink") || strings.Contains(reason, "archivo regular") {
		return "high"
	}
	if categoryFor(path) == "root/config" {
		return "high"
	}
	return "medium"
}

func recommendationFor(path, reason string, policy assets.Policy) string {
	if strings.Contains(reason, "symlink") || strings.Contains(reason, "archivo regular") {
		return "block"
	}
	if policy == assets.PolicyMergeBlock {
		return "merge"
	}
	switch categoryFor(path) {
	case "root/config":
		return "block"
	default:
		return "merge"
	}
}

func availableActionsFor(recommendation string) []string {
	actions := []string{"keep-local", "accept-managed", "merge", "backup-and-replace", "block"}
	if recommendation == "block" {
		return []string{"keep-local", "block"}
	}
	return actions
}

func parallelGroupFor(category string) string {
	return strings.NewReplacer("/", "-", ".", "", " ", "-").Replace(category)
}

func reasonWithComponent(reason, component string) string {
	if component == "" {
		return reason
	}
	return reason + "; component=" + component
}

func groupsForItems(items []Item) []Group {
	byCategory := map[string]*Group{}
	for _, item := range items {
		group := byCategory[item.Category]
		if group == nil {
			group = &Group{Category: item.Category, Risk: item.Risk, ParallelGroup: item.ParallelGroup}
			byCategory[item.Category] = group
		}
		group.Count++
		group.Paths = append(group.Paths, item.Path)
		if riskRank(item.Risk) > riskRank(group.Risk) {
			group.Risk = item.Risk
		}
	}
	groups := make([]Group, 0, len(byCategory))
	for _, group := range byCategory {
		sort.Strings(group.Paths)
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Category < groups[j].Category })
	return groups
}

func riskRank(risk string) int {
	switch risk {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func summaryFor(report Report) Summary {
	summary := Summary{Conflicts: len(report.Items), Groups: len(report.Groups), ParallelCandidates: len(report.Groups)}
	for _, item := range report.Items {
		switch item.Risk {
		case "high":
			summary.HighRisk++
		case "medium":
			summary.MediumRisk++
		}
	}
	return summary
}

func nextActions(report Report) []string {
	actions := []string{}
	if len(report.Items) > 0 {
		actions = append(actions, "Revisar grupos por categoria; no sobrescribir conflictos no gestionados sin decision humana.")
		actions = append(actions, "Usar lufy-ai install --target <dir> --dry-run para ver acciones completas de instalacion.")
	}
	if len(actions) == 0 {
		actions = append(actions, "No hay conflictos de instalacion detectados; puedes continuar con setup/install.")
	}
	return actions
}

func printReport(report Report, stdout io.Writer) {
	fmt.Fprintf(stdout, "Plan de conflictos para %s\n", report.TargetRoot)
	fmt.Fprintf(stdout, "Conflictos: %d grupos=%d\n", report.Summary.Conflicts, report.Summary.Groups)
	if len(report.Items) == 0 {
		fmt.Fprintln(stdout, "No hay conflictos de instalacion detectados")
	}
	for _, group := range report.Groups {
		fmt.Fprintf(stdout, "\n[%s] %d conflicto(s) riesgo=%s parallel_group=%s\n", group.Category, group.Count, group.Risk, group.ParallelGroup)
		for _, item := range report.Items {
			if item.Category != group.Category {
				continue
			}
			fmt.Fprintf(stdout, "- %s status=%s recomendacion=%s riesgo=%s\n", item.Path, item.Status, item.Recommendation, item.Risk)
			fmt.Fprintf(stdout, "  razon: %s\n", item.Reason)
			fmt.Fprintf(stdout, "  acciones: %s\n", strings.Join(item.AvailableActions, ", "))
		}
	}
	fmt.Fprintln(stdout, "\nSiguientes acciones:")
	for _, action := range report.NextActions {
		fmt.Fprintf(stdout, "- %s\n", action)
	}
}
