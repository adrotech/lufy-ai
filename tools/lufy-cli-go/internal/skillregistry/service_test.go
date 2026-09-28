package skillregistry

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
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

func TestEnsureIsIdempotentAndRepairsMovedRepository(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "repo original")
	writeSkill(t, filepath.Join(target, ".agents", "skills"), "reviewer", "reviewer", "review changes")
	service := Service{Env: ports.Env{}}
	opts := Options{Target: target, Tool: domain.ToolCodex}
	var out bytes.Buffer
	if err := service.Ensure(opts, &out); err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(target, lufypaths.SkillRegistry)
	past := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(registryPath, past, past); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := service.Ensure(opts, &out); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(past) || strings.Contains(out.String(), "registry actualizado") {
		t.Fatalf("ready ensure rewrote registry: mtime=%s out=%s", info.ModTime(), out.String())
	}

	moved := filepath.Join(parent, "repo moved")
	if err := os.Rename(target, moved); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := service.Ensure(Options{Target: moved, Tool: domain.ToolCodex}, &out); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(moved, lufypaths.SkillRegistry))
	if err != nil {
		t.Fatal(err)
	}
	var index Index
	if err := json.Unmarshal(body, &index); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(moved, ".agents", "skills", "reviewer", "SKILL.md")
	if len(index.Skills) != 1 {
		t.Fatalf("moved registry paths not repaired: %+v", index.Skills)
	}
	indexedInfo, err := os.Stat(index.Skills[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	wantInfo, err := os.Stat(want)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(indexedInfo, wantInfo) {
		t.Fatalf("moved registry path points to another file: got=%s want=%s", index.Skills[0].Path, want)
	}
}

func TestInspectResolvesToolFromInstallStateAndBlocksUnresolvedMismatch(t *testing.T) {
	target := t.TempDir()
	writeSkill(t, filepath.Join(target, ".agents", "skills"), "reviewer", "reviewer", "review changes")
	harness := domain.HarnessConfig{Tool: domain.ToolCodex, MethodologyByTier: domain.DefaultMethodologyByTier()}
	st := state.NewWithHarness(target, nil, nil, "test", harness)
	if err := state.WriteAtomic(target, st); err != nil {
		t.Fatal(err)
	}
	service := Service{Env: ports.Env{}}
	report, err := service.Inspect(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if report.Tool != domain.ToolCodex {
		t.Fatalf("tool should come from install state, got %s", report.Tool)
	}
	cfg := projectconfig.ProjectConfig{SchemaVersion: projectconfig.SchemaVersion, Tool: domain.ToolInitialDefault, MethodologyByTier: domain.DefaultMethodologyByTier()}
	if err := (projectconfig.ConfigStore{}).Write(projectconfig.Path(target), cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Inspect(Options{Target: target}); err == nil || !strings.Contains(err.Error(), "tool project=opencode install-state=codex") {
		t.Fatalf("expected unresolved harness mismatch, got %v", err)
	}
	report, err = service.Inspect(Options{Target: target, Tool: domain.ToolCodex})
	if err != nil || report.Tool != domain.ToolCodex {
		t.Fatalf("explicit tool should resolve mismatch: report=%#v err=%v", report, err)
	}
}
