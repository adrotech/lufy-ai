package domain

import "testing"

func TestDefaultMethodologyByTier(t *testing.T) {
	defaults := DefaultMethodologyByTier()

	cases := map[Tier]MethodologySelection{
		TierT1: {ID: MethodologySpecWorkflow, Mode: MethodologyModeFull, Required: true},
		TierT2: {ID: MethodologySpecWorkflow, Mode: MethodologyModeLite, Required: true},
		TierT3: {ID: MethodologyNone, Mode: MethodologyModeNone, Required: false},
	}

	for tier, want := range cases {
		got, ok := defaults[tier]
		if !ok {
			t.Fatalf("missing default for %s", tier)
		}
		if got != want {
			t.Fatalf("default for %s = %+v, want %+v", tier, got, want)
		}
	}
	if err := defaults.ValidateSupported(); err != nil {
		t.Fatalf("defaults should validate: %v", err)
	}
}

func TestHarnessConfigDefaults(t *testing.T) {
	cfg := HarnessConfig{}.WithDefaults()

	if cfg.Tool != ToolInitialDefault {
		t.Fatalf("tool default = %s", cfg.Tool)
	}
	selection, err := cfg.MethodologyByTier.SelectionFor(TierT2)
	if err != nil {
		t.Fatalf("selection for T2: %v", err)
	}
	if selection.ID != MethodologySpecWorkflow || selection.Mode != MethodologyModeLite || !selection.Required {
		t.Fatalf("unexpected T2 selection: %+v", selection)
	}
	if err := cfg.ValidateSupported(); err != nil {
		t.Fatalf("default harness config should validate: %v", err)
	}
}

func TestMethodologyByTierWithDefaultsPreservesCompatibleOverrides(t *testing.T) {
	cfg := MethodologyByTier{
		TierT3: {ID: MethodologySpecWorkflow, Mode: MethodologyModeLite, Required: false},
	}.WithDefaults()

	if cfg[TierT1].ID != MethodologySpecWorkflow || cfg[TierT2].Mode != MethodologyModeLite {
		t.Fatalf("missing default tiers after merge: %+v", cfg)
	}
	if cfg[TierT3].ID != MethodologySpecWorkflow || cfg[TierT3].Mode != MethodologyModeLite {
		t.Fatalf("override was not preserved: %+v", cfg[TierT3])
	}
}

func TestMethodologyByTierRejectsUnsupportedValues(t *testing.T) {
	cases := []MethodologyByTier{
		{Tier("T4"): {ID: MethodologySpecWorkflow, Mode: MethodologyModeFull, Required: true}},
		{TierT1: {ID: MethodologyID("other"), Mode: MethodologyModeFull, Required: true}},
		{TierT1: {ID: MethodologySpecWorkflow, Mode: MethodologyMode("expanded"), Required: true}},
		{TierT3: {ID: MethodologyNone, Mode: MethodologyModeLite, Required: false}},
	}

	for _, tc := range cases {
		if err := tc.ValidateSupported(); err == nil {
			t.Fatalf("expected validation error for %+v", tc)
		}
	}
}

func TestHarnessConfigAcceptsKnownDryRunTool(t *testing.T) {
	for _, tool := range []ToolID{ToolCodex, ToolClaudeCode} {
		cfg := HarnessConfig{Tool: tool}
		if err := cfg.ValidateSupported(); err != nil {
			t.Fatalf("expected known dry-run tool %s to validate: %v", tool, err)
		}
	}
}

func TestHarnessConfigRejectsUnsupportedTool(t *testing.T) {
	cfg := HarnessConfig{Tool: ToolID("other")}

	if err := cfg.ValidateSupported(); err == nil {
		t.Fatalf("expected unsupported tool to fail")
	}
}
