package registry

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestDefaultRegistryResolvesInitialAdapters(t *testing.T) {
	reg := Default()

	tool, err := reg.Tool(domain.ToolInitialDefault)
	if err != nil {
		t.Fatalf("tool lookup: %v", err)
	}
	if tool.ID() != domain.ToolInitialDefault {
		t.Fatalf("tool id = %s", tool.ID())
	}
	codexTool, err := reg.Tool(domain.ToolCodex)
	if err != nil {
		t.Fatalf("codex tool lookup: %v", err)
	}
	if codexTool.ID() != domain.ToolCodex || codexTool.Capabilities().DryRunOnly {
		t.Fatalf("codex tool = %s capabilities=%+v", codexTool.ID(), codexTool.Capabilities())
	}
	claudeTool, err := reg.Tool(domain.ToolClaudeCode)
	if err != nil {
		t.Fatalf("claude code tool lookup: %v", err)
	}
	if claudeTool.ID() != domain.ToolClaudeCode || !claudeTool.Capabilities().DryRunOnly {
		t.Fatalf("claude code tool = %s capabilities=%+v", claudeTool.ID(), claudeTool.Capabilities())
	}

	spec, err := reg.Methodology(domain.MethodologySpecWorkflow)
	if err != nil {
		t.Fatalf("methodology lookup: %v", err)
	}
	if spec.ID() != domain.MethodologySpecWorkflow {
		t.Fatalf("methodology id = %s", spec.ID())
	}

	lufy, err := reg.Methodology(domain.MethodologyLufyWorkflow)
	if err != nil {
		t.Fatalf("lufy-sdd lookup: %v", err)
	}
	if lufy.ID() != domain.MethodologyLufyWorkflow {
		t.Fatalf("lufy-sdd id = %s", lufy.ID())
	}
	if modes := lufy.SupportedModes(); len(modes) != 2 || modes[0] != domain.MethodologyModeFull || modes[1] != domain.MethodologyModeLite {
		t.Fatalf("lufy-sdd modes = %#v", modes)
	}

	none, err := reg.Methodology(domain.MethodologyNone)
	if err != nil {
		t.Fatalf("none lookup: %v", err)
	}
	if none.ID() != domain.MethodologyNone {
		t.Fatalf("none id = %s", none.ID())
	}
}

func TestDefaultRegistryListsWritableTools(t *testing.T) {
	reg := Default()

	got := reg.WritableToolIDs()
	want := []domain.ToolID{domain.ToolCodex, domain.ToolInitialDefault}
	if len(got) != len(want) {
		t.Fatalf("writable tools = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("writable tools = %#v", got)
		}
	}
}

func TestDefaultRegistryRejectsUnsupportedAdapters(t *testing.T) {
	reg := Default()

	if _, err := reg.Tool(domain.ToolID("other")); err == nil {
		t.Fatalf("expected unsupported tool error")
	}
	if _, err := reg.Methodology(domain.MethodologyID("spec-kit")); err == nil {
		t.Fatalf("expected unsupported methodology error")
	}
}
