package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestLoadCurrentRoleContracts(t *testing.T) {
	roles := loadRepoRoles(t)

	if len(roles) != 8 {
		t.Fatalf("roles = %d, want 8", len(roles))
	}
	for _, id := range []domain.RoleID{domain.RoleOrchestrator, domain.RoleRouter, domain.RoleExplorer, domain.RoleImplementer, domain.RoleTestWriter, domain.RoleValidator, domain.RoleReviewer, domain.RoleDelivery} {
		if _, ok := RoleByID(roles, id); !ok {
			t.Fatalf("missing role %s", id)
		}
	}
	implementer, ok := RoleByID(roles, domain.RoleImplementer)
	if !ok {
		t.Fatalf("implementer role not found")
	}
	if implementer.Permissions.Edit != true || implementer.Permissions.Shell != "bounded" || implementer.Delegation.Fallback != "inline_implementation" {
		t.Fatalf("implementer metadata not decoded: %#v", implementer)
	}
	if len(implementer.Responsibilities) == 0 || len(implementer.Boundaries) == 0 || len(implementer.Outputs) == 0 {
		t.Fatalf("implementer missing rendered metadata: %#v", implementer)
	}
}

func TestResolveDeliveryDirectSkills(t *testing.T) {
	roles := loadRepoRoles(t)
	binding := loadCurrentBinding(t)
	delivery, ok := RoleByID(roles, domain.RoleDelivery)
	if !ok {
		t.Fatalf("delivery role not found")
	}

	resolved, err := ResolveDirectSkills(delivery, binding)
	if err != nil {
		t.Fatalf("resolve delivery skills: %v", err)
	}

	paths := map[domain.SkillSlot]string{}
	for _, skill := range resolved {
		paths[skill.Slot] = skill.Path
	}
	if paths[domain.SkillSlotDeliveryPRContent] != ".opencode/skills/pr.creator/SKILL.md" {
		t.Fatalf("pr content path = %q", paths[domain.SkillSlotDeliveryPRContent])
	}
	if paths[domain.SkillSlotDeliveryGit] != ".opencode/skills/git-delivery/SKILL.md" {
		t.Fatalf("git delivery path = %q", paths[domain.SkillSlotDeliveryGit])
	}
}

func TestRouterDoesNotResolveMethodologySkillsDirectly(t *testing.T) {
	roles := loadRepoRoles(t)
	binding := loadCurrentBinding(t)
	router, ok := RoleByID(roles, domain.RoleRouter)
	if !ok {
		t.Fatalf("router role not found")
	}

	resolved, err := ResolveDirectSkills(router, binding)
	if err != nil {
		t.Fatalf("resolve router skills: %v", err)
	}
	if len(resolved) != 0 {
		t.Fatalf("router direct skills = %d, want 0", len(resolved))
	}
}

func loadRepoRoles(t *testing.T) []RoleDefinition {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	roles, err := LoadRoleDir(filepath.Join(cwd, "..", "roles"))
	if err != nil {
		t.Fatalf("load roles: %v", err)
	}
	return roles
}

func loadCurrentBinding(t *testing.T) SkillBinding {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binding, err := LoadSkillBindingFile(filepath.Join(cwd, "..", "bindings", "opencode-openspec-current", "role-skills.yaml"))
	if err != nil {
		t.Fatalf("load binding: %v", err)
	}
	return binding
}
