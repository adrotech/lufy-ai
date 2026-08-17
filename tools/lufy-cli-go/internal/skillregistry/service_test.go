package skillregistry

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestRefreshAndStatusLifecycle(t *testing.T) {
	target := t.TempDir()
	path := writeSkill(t, filepath.Join(target, ".agents", "skills"), "reviewer", "reviewer", "review changes")
	service := Service{Env: ports.Env{}}
	opts := Options{Target: target, Tool: domain.ToolCodex}
	var out bytes.Buffer
	if err := service.Status(opts, &out); err != nil || !strings.Contains(out.String(), "not_available") {
		t.Fatalf("initial status: err=%v out=%s", err, out.String())
	}
	out.Reset()
	if err := service.Refresh(opts, &out); err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(target, lufypaths.SkillRegistry)
	if _, err := os.Stat(registryPath); err != nil {
		t.Fatalf("registry missing: %v", err)
	}
	out.Reset()
	if err := service.Status(opts, &out); err != nil || !strings.Contains(out.String(), "ready") {
		t.Fatalf("ready status: err=%v out=%s", err, out.String())
	}
	if err := os.WriteFile(path, []byte("---\nname: reviewer\ndescription: changed\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := service.Status(opts, &out); err != nil || !strings.Contains(out.String(), "stale") {
		t.Fatalf("stale status: err=%v out=%s", err, out.String())
	}
}
