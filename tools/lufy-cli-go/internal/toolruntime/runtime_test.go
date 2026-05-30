package toolruntime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestProjectConfigLifecycleUsesOpenCodeRuntime(t *testing.T) {
	target := t.TempDir()

	plan, err := PlanProjectConfig(domain.ToolInitialDefault, target, true)
	if err != nil {
		t.Fatalf("PlanProjectConfig() error = %v", err)
	}
	if plan.File != OpenCodeProjectConfigFile || plan.Action != "merge-json" {
		t.Fatalf("plan unexpected: %#v", plan)
	}

	ensure, err := EnsureProjectConfig(domain.ToolInitialDefault, target, true)
	if err != nil {
		t.Fatalf("EnsureProjectConfig() error = %v", err)
	}
	if ensure.File != OpenCodeProjectConfigFile || ensure.Action != "merge-json" {
		t.Fatalf("ensure unexpected: %#v", ensure)
	}
	if _, err := os.Stat(filepath.Join(target, OpenCodeProjectConfigFile)); err != nil {
		t.Fatalf("opencode project config not written: %v", err)
	}

	exists, err := ValidateProjectConfig(domain.ToolInitialDefault, target)
	if err != nil {
		t.Fatalf("ValidateProjectConfig() error = %v", err)
	}
	if !exists {
		t.Fatal("ValidateProjectConfig() exists = false")
	}
}

func TestRuntimeRejectsNonWritableTools(t *testing.T) {
	for _, tool := range []domain.ToolID{domain.ToolCodex, domain.ToolClaudeCode, "other"} {
		if _, err := ProjectConfigFile(tool); err == nil || !strings.Contains(err.Error(), "no soporta configuración escribible") {
			t.Fatalf("ProjectConfigFile(%s) error = %v", tool, err)
		}
		if _, err := PlanProjectConfig(tool, t.TempDir(), true); err == nil || !strings.Contains(err.Error(), string(tool)) {
			t.Fatalf("PlanProjectConfig(%s) error = %v", tool, err)
		}
		if _, err := GlobalRoot(tool); err == nil || !strings.Contains(err.Error(), string(tool)) {
			t.Fatalf("GlobalRoot(%s) error = %v", tool, err)
		}
	}
}

func TestPluginConfigFilesUsesOpenCodeFiles(t *testing.T) {
	files, err := PluginConfigFiles(domain.ToolInitialDefault)
	if err != nil {
		t.Fatalf("PluginConfigFiles() error = %v", err)
	}
	if len(files) != 2 || files[0] != "tui.json" || files[1] != OpenCodeProjectConfigFile {
		t.Fatalf("PluginConfigFiles() = %#v", files)
	}
}
