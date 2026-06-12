package registry

import (
	"fmt"
	"sort"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/methodology/lufysdd"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/methodology/none"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/methodology/openspec"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/tool/claudecode"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/tool/codex"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/tool/opencode"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Registry struct {
	tools       map[domain.ToolID]ports.ToolAdapter
	methodology map[domain.MethodologyID]ports.MethodologyAdapter
}

func Default() Registry {
	return Registry{
		tools: map[domain.ToolID]ports.ToolAdapter{
			domain.ToolInitialDefault: opencode.New(),
			domain.ToolCodex:          codex.New(),
			domain.ToolClaudeCode:     claudecode.New(),
		},
		methodology: map[domain.MethodologyID]ports.MethodologyAdapter{
			domain.MethodologySpecWorkflow: openspec.New(),
			domain.MethodologyLufyWorkflow: lufysdd.New(),
			domain.MethodologyNone:         none.New(),
		},
	}
}

func (r Registry) Tool(id domain.ToolID) (ports.ToolAdapter, error) {
	adapter, ok := r.tools[id]
	if !ok {
		return nil, fmt.Errorf("tool adapter no soportado %q; disponibles: %v", id, r.ToolIDs())
	}
	return adapter, nil
}

func (r Registry) Methodology(id domain.MethodologyID) (ports.MethodologyAdapter, error) {
	adapter, ok := r.methodology[id]
	if !ok {
		return nil, fmt.Errorf("methodology adapter no soportado %q; disponibles: %v", id, r.MethodologyIDs())
	}
	return adapter, nil
}

func (r Registry) ToolIDs() []domain.ToolID {
	ids := make([]domain.ToolID, 0, len(r.tools))
	for id := range r.tools {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r Registry) WritableToolIDs() []domain.ToolID {
	ids := make([]domain.ToolID, 0, len(r.tools))
	for id, adapter := range r.tools {
		if !adapter.Capabilities().DryRunOnly {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r Registry) MethodologyIDs() []domain.MethodologyID {
	ids := make([]domain.MethodologyID, 0, len(r.methodology))
	for id := range r.methodology {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
