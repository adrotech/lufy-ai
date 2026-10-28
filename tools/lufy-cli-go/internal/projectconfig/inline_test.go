package projectconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineConfigRoundTripPreservesExecutionAndProjectValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project.yaml")
	body := "schema_version: 2\ntool: codex\nexecution:\n  mode: inline\n  delegation: off\ncustom_rule: keep\nvalidation:\n  allowed_commands:\n    implementer: [go test ./...]\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(cfg.Execution.Delegation) != "off" {
		t.Fatal("delegation lost")
	}
	encoded, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, want := range []string{"schema_version: 2", "mode: inline", "custom_rule: keep", "go test ./..."} {
		if !strings.Contains(text, want) {
			t.Fatalf("falta %q en %s", want, text)
		}
	}
	if strings.Contains(text, "methodology") {
		t.Fatalf("config conserva metodología: %s", text)
	}
}

func TestInlineConfigRejectsOldOrMixedModelsWithoutWriting(t *testing.T) {
	cases := []string{
		"schema_version: 2\nbase: &old {methodology_by_tier: null}\n<<: *old\ntool: codex\n",
		"schema_version: 2\ntool: codex\n---\nmethodology_by_tier: {}\n",
		"schema_version: 1\ntool: codex\n",
		"tool: codex\n",
		"schema_version: 2\ntool: codex\nmethodology_by_tier: {}\n",
		"schema_version: 2\ntool: codex\nmethodology_by_tier: null\n",
		"schema_version: 2\ntool: codex\nexecution: {mode: full-sdd, delegation: auto}\n",
		"schema_version: 2\ntool: codex\nexecution: {mode: inline, delegation: unknown}\n",
		"schema_version: 2\ntool: codex\nexecution: {mode: inline, delegation: auto, surprise: true}\n",
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "project.yaml")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatalf("se aceptó un modelo no soportado: %s", body)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != body {
				t.Fatalf("la lectura modificó el archivo: %v", err)
			}
		})
	}
}

func TestInlineDefaultsDoNotImposeTDDOrEvidenceCounts(t *testing.T) {
	cfg, err := Scan(t.TempDir(), fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TDD.Strict || cfg.TDD.TriangulateRequired {
		t.Fatal("universal TDD gate remains")
	}
	body, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "min_evidence_items") {
		t.Fatal("universal evidence count remains")
	}
	cfg.TDD.Strict = true
	body, err = Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "project.yaml")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Load(path)
	if err != nil || !reloaded.TDD.Strict {
		t.Fatal("explicit project preference was lost")
	}
	cfg.WorkflowLimits.Review.Extra = map[string]any{"min_evidence_items": 2}
	if _, err := Marshal(cfg); err == nil {
		t.Fatal("retired minimum was re-emitted")
	}
}
