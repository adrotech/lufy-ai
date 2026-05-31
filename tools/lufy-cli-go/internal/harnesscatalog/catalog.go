package harnesscatalog

import (
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/registry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func Effective(base assets.Catalog, harness domain.HarnessConfig) (assets.Catalog, error) {
	return EffectiveWithRegistry(base, harness, registry.Default())
}

func EffectiveWithRegistry(base assets.Catalog, harness domain.HarnessConfig, reg adapterRegistry) (assets.Catalog, error) {
	normalized := harness.WithDefaults()
	if err := normalized.ValidateSupported(); err != nil {
		return assets.Catalog{}, err
	}

	toolAdapter, err := reg.Tool(normalized.Tool)
	if err != nil {
		return assets.Catalog{}, err
	}
	toolSpecs, err := toolAdapter.RenderSurface(ports.HarnessModel{
		Tool:              normalized.Tool,
		MethodologyByTier: normalized.MethodologyByTier,
	})
	if err != nil {
		return assets.Catalog{}, err
	}

	methodSpecs := map[domain.MethodologyID][]ports.AssetSpec{}
	selectedMethodologies := map[domain.MethodologyID]bool{}
	for tier, selection := range normalized.MethodologyByTier {
		selectedMethodologies[selection.ID] = true
		adapter, err := reg.Methodology(selection.ID)
		if err != nil {
			return assets.Catalog{}, err
		}
		rendered, err := adapter.RenderWorkflow(ports.WorkflowModel{Tier: tier, Selection: selection})
		if err != nil {
			return assets.Catalog{}, err
		}
		methodSpecs[selection.ID] = append(methodSpecs[selection.ID], rendered...)
	}

	out := make([]assets.Asset, 0, len(base.Assets))
	for _, asset := range base.Assets {
		if shouldInclude(asset, toolSpecs, methodSpecs, selectedMethodologies) {
			out = append(out, asset)
		}
	}
	return assets.Catalog{SourceRoot: base.SourceRoot, Assets: out}, nil
}

type adapterRegistry interface {
	Tool(domain.ToolID) (ports.ToolAdapter, error)
	Methodology(domain.MethodologyID) (ports.MethodologyAdapter, error)
}

func shouldInclude(asset assets.Asset, toolSpecs []ports.AssetSpec, methodSpecs map[domain.MethodologyID][]ports.AssetSpec, selectedMethodologies map[domain.MethodologyID]bool) bool {
	if isCoreAsset(asset) {
		return true
	}

	toolMatched := matchesSpecs(asset.TargetRel, toolSpecs)
	if asset.Methodology == "" || asset.Methodology == domain.MethodologyNone {
		return toolMatched
	}

	if !selectedMethodologies[asset.Methodology] {
		return false
	}
	if matchesSpecs(asset.TargetRel, methodSpecs[asset.Methodology]) {
		return true
	}
	return toolMatched && isMethodologyInstructionAsset(asset)
}

func isCoreAsset(asset assets.Asset) bool {
	return asset.Methodology == domain.MethodologyNone && (asset.Component == "harness-core" || asset.Component == "harness-reference")
}

func isMethodologyInstructionAsset(asset assets.Asset) bool {
	return asset.Component == "methodology-skill" || asset.Component == "methodology-command"
}

func matchesSpecs(targetRel string, specs []ports.AssetSpec) bool {
	target := filepath.ToSlash(targetRel)
	for _, spec := range specs {
		if spec.Scope == "preview" || spec.Policy == "dry-run" || spec.Policy == "merge-json" {
			continue
		}
		specTarget := filepath.ToSlash(spec.TargetRel)
		if target == specTarget || strings.HasPrefix(target, specTarget+"/") {
			return true
		}
	}
	return false
}
