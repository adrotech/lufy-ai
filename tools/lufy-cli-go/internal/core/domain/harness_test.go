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
