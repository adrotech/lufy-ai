package projectconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScanEmitsDisabledAdaptiveRoutingDefaults(t *testing.T) {
	t.Parallel()
	cfg, err := Scan(t.TempDir(), time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	assertAdaptiveRouting(t, cfg.AdaptiveRouting, false, "shadow", "deterministic-v1", 900, 32, 128, 5)
	data, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"adaptive_routing:", "enabled: false", "mode: shadow", "policy_version: deterministic-v1", "lease_ttl_seconds: 900"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("config no contiene %q:\n%s", want, data)
		}
	}
}

func TestLoadPartialAdaptiveRoutingCompletesDefaultsAndPreservesExtras(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "project.yaml")
	body := []byte("schema_version: 1\nadaptive_routing:\n  enabled: true\n  max_candidates: 7\n  future_policy: keep\n")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	assertAdaptiveRouting(t, cfg.AdaptiveRouting, true, "shadow", "deterministic-v1", 900, 7, 128, 5)
	if cfg.AdaptiveRouting.Extra["future_policy"] != "keep" {
		t.Fatalf("extras no preservados: %#v", cfg.AdaptiveRouting.Extra)
	}
}

func TestRescanPreservesAdaptiveRoutingOverridesWithoutActivatingMissingConfig(t *testing.T) {
	t.Parallel()
	detected, err := Scan(t.TempDir(), time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	current := ProjectConfig{AdaptiveRouting: AdaptiveRoutingConfig{
		Enabled: false, Mode: "advisory", MaxCandidates: 9, Extra: map[string]any{"future": "keep"},
	}}
	merged := MergeRescan(current, detected)
	assertAdaptiveRouting(t, merged.AdaptiveRouting, false, "advisory", "deterministic-v1", 900, 9, 128, 5)
	if merged.AdaptiveRouting.Extra["future"] != "keep" {
		t.Fatalf("extras no preservados: %#v", merged.AdaptiveRouting.Extra)
	}

	mergedMissing := MergeRescan(ProjectConfig{}, detected)
	if mergedMissing.AdaptiveRouting.Enabled {
		t.Fatal("rescan no debe activar adaptive routing")
	}
}

func TestLoadRejectsInvalidAdaptiveRouting(t *testing.T) {
	t.Parallel()
	tests := []string{
		"mode: autonomous",
		"lease_ttl_seconds: -1",
		"max_candidates: 33",
		"max_waiting_items: 2048",
		"starvation_after_cycles: -1",
	}
	for _, field := range tests {
		field := field
		t.Run(strings.ReplaceAll(field, " ", "_"), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "project.yaml")
			body := []byte("schema_version: 1\nadaptive_routing:\n  enabled: true\n  " + field + "\n")
			if err := os.WriteFile(path, body, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "adaptive_routing") {
				t.Fatalf("error=%v, se esperaba path canónico", err)
			}
		})
	}
}

func assertAdaptiveRouting(t *testing.T, got AdaptiveRoutingConfig, enabled bool, mode, policy string, ttl, candidates, waiting, starvation int) {
	t.Helper()
	if got.Enabled != enabled || got.Mode != mode || got.PolicyVersion != policy || got.LeaseTTLSeconds != ttl ||
		got.MaxCandidates != candidates || got.MaxWaitingItems != waiting || got.StarvationAfterCycles != starvation {
		t.Fatalf("adaptive routing = %#v", got)
	}
}
