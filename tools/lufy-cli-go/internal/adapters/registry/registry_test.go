package registry

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"testing"
)

func TestOnlyCodexIsRegistered(t *testing.T) {
	r := Default()
	if ids := r.ToolIDs(); len(ids) != 1 || ids[0] != domain.ToolCodex {
		t.Fatalf("tools: %v", ids)
	}
	if ids := r.WritableToolIDs(); len(ids) != 1 || ids[0] != domain.ToolCodex {
		t.Fatalf("writable tools: %v", ids)
	}
	if _, err := r.Tool(domain.ToolCodex); err != nil {
		t.Fatal(err)
	}
	for _, id := range []domain.ToolID{domain.ToolID("opencode"), domain.ToolID("claude-code"), "unknown"} {
		if _, err := r.Tool(id); err == nil {
			t.Fatalf("registered %s", id)
		}
	}
	if len((Registry{}).ToolIDs()) != 0 || len((Registry{}).WritableToolIDs()) != 0 {
		t.Fatal("empty registry has tools")
	}
}
