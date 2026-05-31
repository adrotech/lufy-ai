package domain

import "testing"

func TestValidateRoutingPolicyAllowsDefaults(t *testing.T) {
	if err := DefaultMethodologyByTier().ValidateRoutingPolicy(RoutingPolicyOptions{}); err != nil {
		t.Fatalf("defaults should validate: %v", err)
	}
}

func TestValidateRoutingPolicyBlocksT1None(t *testing.T) {
	cfg := DefaultMethodologyByTier()
	cfg[TierT1] = MethodologySelection{ID: MethodologyNone, Mode: MethodologyModeNone}

	if err := cfg.ValidateRoutingPolicy(RoutingPolicyOptions{}); err == nil {
		t.Fatalf("expected T1 none to be blocked")
	}
}

func TestValidateRoutingPolicyRequiresT2NoneRationale(t *testing.T) {
	cfg := DefaultMethodologyByTier()
	cfg[TierT2] = MethodologySelection{ID: MethodologyNone, Mode: MethodologyModeNone}

	if err := cfg.ValidateRoutingPolicy(RoutingPolicyOptions{}); err == nil {
		t.Fatalf("expected T2 none without rationale to be blocked")
	}
	if err := cfg.ValidateRoutingPolicy(RoutingPolicyOptions{AllowT2None: true, T2NoneReason: "bounded docs-only work"}); err != nil {
		t.Fatalf("expected T2 none with rationale to pass: %v", err)
	}
}
