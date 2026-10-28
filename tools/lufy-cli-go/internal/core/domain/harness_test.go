package domain

import "testing"

func TestInlineDefaults(t *testing.T) {
	c := DefaultHarnessConfig()
	if c.Tool != ToolCodex || c.Execution.Mode != ExecutionInline || c.Execution.Delegation != DelegationAuto {
		t.Fatalf("defaults: %#v", c)
	}
	if c.ToolIsExplicit() || c.ExecutionIsExplicit() {
		t.Fatal("defaults became explicit")
	}
	if err := c.ValidateSupported(); err != nil {
		t.Fatal(err)
	}
	if !(HarnessConfig{Tool: ToolCodex}).ToolIsExplicit() || !(HarnessConfig{Execution: ExecutionConfig{Delegation: DelegationOff}}).ExecutionIsExplicit() {
		t.Fatal("explicit selection lost")
	}
}
func TestRejectUnsupportedExecution(t *testing.T) {
	for _, c := range []HarnessConfig{
		{Tool: ToolID("opencode")}, {Tool: ToolID("claude-code")}, {Execution: ExecutionConfig{Mode: "sdd"}}, {Execution: ExecutionConfig{Delegation: "always"}},
	} {
		if c.ValidateSupported() == nil {
			t.Fatalf("accepted %#v", c)
		}
	}
	for _, d := range []DelegationPolicy{DelegationAuto, DelegationOnRequest, DelegationOff} {
		if err := (ExecutionConfig{Mode: ExecutionInline, Delegation: d}).Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOnlyCodexIdentifierIsSupported(t *testing.T) {
	if !ToolCodex.Valid() || ToolID("opencode").Valid() || ToolID("claude-code").Valid() {
		t.Fatal("identificador de host retirado aceptado")
	}
}
