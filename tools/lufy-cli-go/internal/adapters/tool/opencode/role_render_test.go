package opencode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/registry"
)

func TestRenderDeliveryAgentMatchesGolden(t *testing.T) {
	roles := loadRenderRepoRoles(t)
	binding := loadRenderCurrentBinding(t)
	delivery, ok := registry.RoleByID(roles, domain.RoleDelivery)
	if !ok {
		t.Fatalf("delivery role not found")
	}

	asset, err := RenderAgentAsset(delivery, binding, domain.HarnessConfig{}, domain.TierT1)
	if err != nil {
		t.Fatalf("render delivery: %v", err)
	}
	if asset.Path != ".opencode/agents/delivery.md" {
		t.Fatalf("asset path = %s", asset.Path)
	}
	assertGolden(t, "opencode-delivery.md", asset.Body)
}

func TestRenderRouterT3NoneMatchesGolden(t *testing.T) {
	roles := loadRenderRepoRoles(t)
	binding := loadRenderCurrentBinding(t)
	router, ok := registry.RoleByID(roles, domain.RoleRouter)
	if !ok {
		t.Fatalf("router role not found")
	}

	asset, err := RenderAgentAsset(router, binding, domain.HarnessConfig{}, domain.TierT3)
	if err != nil {
		t.Fatalf("render router: %v", err)
	}
	if asset.Path != ".opencode/agents/sdd-router.md" {
		t.Fatalf("asset path = %s", asset.Path)
	}
	assertGolden(t, "opencode-router-t3-none.md", asset.Body)
}

func loadRenderRepoRoles(t *testing.T) []registry.RoleDefinition {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	roles, err := registry.LoadRoleDir(filepath.Join(cwd, "..", "..", "..", "instructions", "roles"))
	if err != nil {
		t.Fatalf("load roles: %v", err)
	}
	return roles
}

func loadRenderCurrentBinding(t *testing.T) registry.SkillBinding {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binding, err := registry.LoadSkillBindingFile(filepath.Join(cwd, "..", "..", "..", "instructions", "bindings", "opencode-openspec-current", "role-skills.yaml"))
	if err != nil {
		t.Fatalf("load binding: %v", err)
	}
	return binding
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	want := normalizeGoldenLineEndings(string(body))
	got = normalizeGoldenLineEndings(got)
	if got != want {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func normalizeGoldenLineEndings(value string) string {
	return strings.ReplaceAll(value, "\r\n", "\n")
}
