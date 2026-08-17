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
	if len(index.Skills) != 1 || filepath.Clean(index.Skills[0].Path) != filepath.Clean(want) {
		t.Fatalf("moved registry paths not repaired: %+v", index.Skills)
	}
}
