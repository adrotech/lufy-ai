package toolruntime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestProjectConfigLifecycleSupportsCodexManagedConfig(t *testing.T) {
	target := t.TempDir()

	file, err := ProjectConfigFile(domain.ToolCodex)
	if err != nil {
		t.Fatalf("ProjectConfigFile(codex) error = %v", err)
	}
	if file != CodexProjectConfigFile {
		t.Fatalf("codex config file = %q", file)
	}

	plan, err := PlanProjectConfig(domain.ToolCodex, target)
	if err != nil {
		t.Fatalf("PlanProjectConfig(codex) error = %v", err)
	}
	if plan.File != CodexProjectConfigFile || plan.Action != "" {
		t.Fatalf("codex plan unexpected: %#v", plan)
	}

	if exists, err := ValidateProjectConfig(domain.ToolCodex, target); err != nil || exists {
		t.Fatalf("ValidateProjectConfig(codex missing) = %v, %v", exists, err)
	}
	if err := os.MkdirAll(filepath.Join(target, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, CodexProjectConfigFile), []byte("project_doc_max_bytes = 32768\n\n[features]\nmulti_agent = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if exists, err := ValidateProjectConfig(domain.ToolCodex, target); err != nil || !exists {
		t.Fatalf("ValidateProjectConfig(codex present) = %v, %v", exists, err)
	}
}

func TestRuntimeRejectsNonWritableTools(t *testing.T) {
	for _, tool := range []domain.ToolID{domain.ToolID("claude-code"), "other"} {
		if _, err := ProjectConfigFile(tool); err == nil || !strings.Contains(err.Error(), "no soporta configuración escribible") {
			t.Fatalf("ProjectConfigFile(%s) error = %v", tool, err)
		}
		if _, err := PlanProjectConfig(tool, t.TempDir()); err == nil || !strings.Contains(err.Error(), string(tool)) {
			t.Fatalf("PlanProjectConfig(%s) error = %v", tool, err)
		}
		if _, err := GlobalRoot(tool); err == nil || !strings.Contains(err.Error(), string(tool)) {
			t.Fatalf("GlobalRoot(%s) error = %v", tool, err)
		}
	}
}

func TestPluginConfigFilesUsesCodexFiles(t *testing.T) {
	files, err := PluginConfigFiles(domain.ToolCodex)
	if err != nil {
		t.Fatalf("PluginConfigFiles(codex) error = %v", err)
	}
	if len(files) != 1 || files[0] != CodexProjectConfigFile {
		t.Fatalf("PluginConfigFiles(codex) = %#v", files)
	}
}
