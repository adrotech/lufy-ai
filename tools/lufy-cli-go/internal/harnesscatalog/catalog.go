package harnesscatalog

import (
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/registry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type adapterRegistry interface {
	Tool(domain.ToolID) (ports.ToolAdapter, error)
}

func Effective(base assets.Catalog, cfg domain.HarnessConfig) (assets.Catalog, error) {
	return EffectiveWithRegistry(base, cfg, registry.Default())
}

func EffectiveWithRegistry(base assets.Catalog, cfg domain.HarnessConfig, reg adapterRegistry) (assets.Catalog, error) {
	cfg = cfg.WithDefaults()
	if err := cfg.ValidateSupported(); err != nil {
		return assets.Catalog{}, err
	}
	adapter, err := reg.Tool(cfg.Tool)
	if err != nil {
		return assets.Catalog{}, err
	}
	specs, err := adapter.RenderSurface(ports.HarnessModel{Tool: cfg.Tool, Execution: cfg.Execution})
	if err != nil {
		return assets.Catalog{}, err
	}
	for _, path := range []string{"lufy-ia.harness.md", ".lufy/README.md", ".lufy/contracts/README.md", ".lufy/contracts/work.md", ".lufy/contracts/evidence.md", ".lufy/contracts/delivery.md"} {
		specs = append(specs, ports.AssetSpec{TargetRel: path, Policy: "managed", Scope: "project"})
	}
	out := assets.Catalog{SourceRoot: base.SourceRoot}
	for _, asset := range base.Assets {
		if matchesSpecs(asset, specs) {
			out.Assets = append(out.Assets, asset)
		}
	}
	return out, nil
}

func matchesSpecs(asset assets.Asset, specs []ports.AssetSpec) bool {
	target := filepath.ToSlash(asset.TargetRel)
	for _, spec := range specs {
		if spec.Scope == "preview" || spec.Policy == "dry-run" || spec.Policy == "merge-json" {
			continue
		}
		path := filepath.ToSlash(spec.TargetRel)
		if target == path || strings.HasPrefix(target, path+"/") || (asset.Kind == assets.KindDir && strings.HasPrefix(path, target+"/")) {
			return true
		}
	}
	return false
}
