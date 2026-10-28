package registry

import (
	"fmt"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/tool/codex"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Registry struct {
	tools map[domain.ToolID]ports.ToolAdapter
}

func Default() Registry {
	return Registry{tools: map[domain.ToolID]ports.ToolAdapter{domain.ToolCodex: codex.New()}}
}
func (r Registry) Tool(id domain.ToolID) (ports.ToolAdapter, error) {
	adapter, ok := r.tools[id]
	if !ok {
		return nil, fmt.Errorf("tool adapter no soportado %q; disponible: codex", id)
	}
	return adapter, nil
}
func (r Registry) ToolIDs() []domain.ToolID {
	if _, ok := r.tools[domain.ToolCodex]; ok {
		return []domain.ToolID{domain.ToolCodex}
	}
	return nil
}
func (r Registry) WritableToolIDs() []domain.ToolID {
	if a, ok := r.tools[domain.ToolCodex]; ok && !a.Capabilities().DryRunOnly {
		return []domain.ToolID{domain.ToolCodex}
	}
	return nil
}
