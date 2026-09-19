package codexsurface

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAcceptsLifecycleConfigAndConservativeRules(t *testing.T) {
	target := t.TempDir()
	writeSurfaceFile(t, filepath.Join(target, ".codex", "config.toml"), "[features]\nmulti_agent = true\nhooks = true\n")
	writeSurfaceFile(t, filepath.Join(target, ".codex", "hooks.json"), `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"lufy-ai lifecycle codex","commandWindows":"lufy-ai lifecycle codex"}]}],"SubagentStop":[{"hooks":[{"type":"command","command":"lufy-ai lifecycle codex","commandWindows":"lufy-ai lifecycle codex"}]}],"Stop":[{"hooks":[{"type":"command","command":"lufy-ai lifecycle codex","commandWindows":"lufy-ai lifecycle codex"}]}],"SessionEnd":[{"hooks":[{"type":"command","command":"lufy-ai lifecycle codex","commandWindows":"lufy-ai lifecycle codex"}]}]}}`)
	writeSurfaceFile(t, filepath.Join(target, ".codex", "rules", "lufy.rules"), `prefix_rule(
pattern = ["git", ["commit", "push", "merge", "tag"]],
decision = "prompt",
match = ["git commit -m x"],
not_match = ["git status"],
)
prefix_rule(
pattern = ["gh", "pr", ["create", "merge", "ready", "review"]],
decision = "prompt",
match = ["gh pr create"],
not_match = ["gh pr view"],
)
prefix_rule(
pattern = ["git", "reset", "--hard"],
decision = "forbidden",
match = ["git reset --hard"],
not_match = ["git status"],
)`)

	checks := Validate(target)
	if len(checks) != 3 {
		t.Fatalf("checks=%#v", checks)
	}
	for _, check := range checks {
		if check.Level != "ok" {
			t.Fatalf("unexpected check: %#v", check)
		}
	}
}

func TestValidateRejectsPlaceholderLifecycleAndLegacyKeys(t *testing.T) {
	target := t.TempDir()
	writeSurfaceFile(t, filepath.Join(target, ".codex", "config.toml"), "[features]\nmulti_agent = true\n\n[agents]\nmax_threads = 6\nmax_depth = 1\n")
	writeSurfaceFile(t, filepath.Join(target, ".codex", "hooks.json"), `{"hooks":{}}`)
	writeSurfaceFile(t, filepath.Join(target, ".codex", "rules", "lufy.rules"), "# placeholder\n")

	checks := Validate(target)
	if len(checks) != 3 {
		t.Fatalf("checks=%#v", checks)
	}
	for _, check := range checks {
		if check.Level != "fail" {
			t.Fatalf("placeholder accepted: %#v", check)
		}
	}
}

func writeSurfaceFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
