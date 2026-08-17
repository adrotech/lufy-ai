package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsRefreshAndStatusCommands(t *testing.T) {
	target := t.TempDir()
	skillPath := filepath.Join(target, ".agents", "skills", "reviewer", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("---\nname: reviewer\ndescription: Reviews changes.\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := Run([]string{"skills", "refresh", "--target", target, "--tool", "codex", "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), `"name": "reviewer"`) {
		t.Fatalf("refresh code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = Run([]string{"skills", "status", "--target", target, "--tool", "codex", "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), `"status": "ready"`) {
		t.Fatalf("status code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestSkillsRejectsUnsupportedTool(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run([]string{"skills", "refresh", "--tool", "claude-code"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr || !strings.Contains(errOut.String(), "disponibles: codex, opencode") {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}
