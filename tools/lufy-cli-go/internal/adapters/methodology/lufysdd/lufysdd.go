package lufysdd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	workflow "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufysdd"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (Adapter) ID() domain.MethodologyID {
	return domain.MethodologyLufyWorkflow
}

func (Adapter) SupportedModes() []domain.MethodologyMode {
	return []domain.MethodologyMode{domain.MethodologyModeFull, domain.MethodologyModeLite}
}

func (Adapter) RenderWorkflow(model ports.WorkflowModel) ([]ports.AssetSpec, error) {
	if model.Selection.ID != domain.MethodologyLufyWorkflow {
		return nil, nil
	}

	switch model.Selection.Mode {
	case domain.MethodologyModeFull:
		return append(baseAssets(), ports.AssetSpec{
			ID:        "methodology.lufy-sdd.specs",
			TargetRel: lufypaths.LufySDD + "/specs",
			Policy:    "managed",
			Scope:     "project",
		}), nil
	case domain.MethodologyModeLite:
		return baseAssets(), nil
	default:
		return nil, fmt.Errorf("lufy-sdd no soporta mode %s", model.Selection.Mode)
	}
}

func (Adapter) VerifyWorkflow(target ports.Target, tier domain.Tier) ([]ports.Check, error) {
	root, err := platform.SafeJoin(target.Root, lufypaths.LufySDD)
	if err != nil {
		return nil, err
	}
	required := []string{"README.md", "config.yaml", "actions", "templates", "changes", "decisions", "verification", "archive"}
	if tier == domain.TierT1 {
		required = append(required, "specs")
	}
	checks := []ports.Check{}
	for _, rel := range required {
		path, joinErr := platform.SafeJoin(root, rel)
		if joinErr != nil {
			return nil, joinErr
		}
		info, statErr := os.Lstat(path)
		checkPath := filepath.ToSlash(filepath.Join(lufypaths.LufySDD, rel))
		if statErr != nil {
			checks = append(checks, ports.Check{Level: "fail", Path: checkPath, Message: "asset Lufy SDD faltante"})
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			checks = append(checks, ports.Check{Level: "fail", Path: checkPath, Message: "symlink no permitido en workflow Lufy SDD"})
			continue
		}
		wantDir := rel != "README.md" && rel != "config.yaml"
		if wantDir != info.IsDir() {
			checks = append(checks, ports.Check{Level: "fail", Path: checkPath, Message: "tipo de asset Lufy SDD inválido"})
			continue
		}
		if rel == "config.yaml" {
			body, readErr := os.ReadFile(path)
			if readErr != nil || !strings.Contains(string(body), "version: 1") || !strings.Contains(string(body), "workflow: lufy-sdd") {
				checks = append(checks, ports.Check{Level: "fail", Path: checkPath, Message: "config Lufy SDD inválida o incompatible"})
			}
		}
	}
	changesRoot, joinErr := platform.SafeJoin(root, "changes")
	if joinErr == nil {
		entries, readErr := os.ReadDir(changesRoot)
		if readErr == nil {
			sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				report, validateErr := workflow.NewService().Check(target.Root, entry.Name(), tier == domain.TierT1)
				if validateErr != nil {
					checks = append(checks, ports.Check{Level: "fail", Path: filepath.ToSlash(filepath.Join(lufypaths.LufySDD, "changes", entry.Name())), Message: validateErr.Error()})
					continue
				}
				for _, diagnostic := range report.Diagnostics {
					level := "warn"
					if diagnostic.Level == workflow.LevelError {
						level = "fail"
					}
					checks = append(checks, ports.Check{Level: level, Path: diagnostic.Path, Message: diagnostic.Code + ": " + diagnostic.Message})
				}
			}
		}
	}
	if len(checks) == 0 {
		checks = append(checks, ports.Check{Level: "info", Path: filepath.ToSlash(lufypaths.LufySDD), Message: "workflow Lufy SDD listo"})
	}
	return checks, nil
}

func baseAssets() []ports.AssetSpec {
	return []ports.AssetSpec{
		{ID: "methodology.lufy-sdd.readme", TargetRel: lufypaths.LufySDD + "/README.md", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.config", TargetRel: lufypaths.LufySDD + "/config.yaml", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.actions", TargetRel: lufypaths.LufySDD + "/actions", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.templates", TargetRel: lufypaths.LufySDD + "/templates", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.changes", TargetRel: lufypaths.LufySDD + "/changes", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.decisions", TargetRel: lufypaths.LufySDD + "/decisions", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.verification", TargetRel: lufypaths.LufySDD + "/verification", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.archive", TargetRel: lufypaths.LufySDD + "/archive", Policy: "managed", Scope: "project"},
	}
}
