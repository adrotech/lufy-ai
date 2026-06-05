package opencode

import (
	"fmt"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/registry"
	corerender "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/render"
)

type InstructionAsset struct {
	Path string
	Body string
}

func RenderAgentAsset(role registry.RoleDefinition, binding registry.SkillBinding, cfg domain.HarnessConfig, tier domain.Tier) (InstructionAsset, error) {
	surface, err := corerender.BuildRoleSurfaceForTier(role, binding, cfg, tier)
	if err != nil {
		return InstructionAsset{}, err
	}
	return InstructionAsset{
		Path: agentPath(role.ID),
		Body: RenderAgentMarkdown(role, surface),
	}, nil
}

func RenderAgentMarkdown(role registry.RoleDefinition, surface corerender.RoleSurface) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "description: %s\n", role.Purpose)
	fmt.Fprintf(&b, "mode: %s\n", agentMode(role.ID))
	fmt.Fprintf(&b, "x-lufy-role: %s\n", role.ID)
	fmt.Fprintf(&b, "x-lufy-kind: %s\n", role.Kind)
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "You are **%s**.\n\n", role.ID)
	writeSection(&b, "Purpose", []string{role.Purpose})
	writeSection(&b, "Adapter Context", []string{
		fmt.Sprintf("tool=%s", surface.ResultContractContext.Tool),
		fmt.Sprintf("tier=%s", surface.ResultContractContext.Tier),
		fmt.Sprintf("methodology=%s", surface.ResultContractContext.Methodology),
		fmt.Sprintf("methodology_mode=%s", surface.ResultContractContext.MethodologyMode),
		fmt.Sprintf("methodology_required=%t", surface.ResultContractContext.MethodologyRequired),
	})
	permissions := []string{
		fmt.Sprintf("edit=%s", scalar(role.Permissions.Edit)),
		fmt.Sprintf("shell=%s", scalar(role.Permissions.Shell)),
		fmt.Sprintf("delivery=%s", scalar(role.Permissions.Delivery)),
	}
	if hasPermissionPolicy(role.Permissions.ShellPolicy) {
		permissions = append(permissions,
			fmt.Sprintf("shell_policy.default=%s", permissionPolicyDefault(role.Permissions.ShellPolicy)),
			fmt.Sprintf("shell_policy.ask=%s", permissionPolicyList(role.Permissions.ShellPolicy.Ask)),
			fmt.Sprintf("shell_policy.deny=%s", permissionPolicyList(role.Permissions.ShellPolicy.Deny)),
		)
	}
	writeSection(&b, "Permissions", permissions)
	writeSection(&b, "Delegation", []string{
		fmt.Sprintf("preferred=%s", role.Delegation.Preferred),
		fmt.Sprintf("fallback=%s", role.Delegation.Fallback),
	})
	writeSection(&b, "Responsibilities", role.Responsibilities)
	writeSection(&b, "Boundaries", role.Boundaries)
	writeSection(&b, "Outputs", role.Outputs)
	writeSkillSection(&b, role, surface)
	writeSection(&b, "Result Contract", []string{
		fmt.Sprintf("schema=%s", surface.OutputSchema),
		fmt.Sprintf("allowed_status=%s", strings.Join(surface.AllowedStatus, ", ")),
		fmt.Sprintf("compact_payload=%s", strings.Join(surface.CompactPayload, ", ")),
		fmt.Sprintf("max_handoff_focus=%s", strings.Join(surface.MaxHandoffFocus, ", ")),
	})
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func agentPath(id domain.RoleID) string {
	if id == domain.RoleRouter {
		return ".opencode/agents/sdd-router.md"
	}
	return fmt.Sprintf(".opencode/agents/%s.md", id)
}

func agentMode(id domain.RoleID) string {
	if id == domain.RoleOrchestrator {
		return "primary"
	}
	return "subagent"
}

func writeSection(b *strings.Builder, title string, values []string) {
	fmt.Fprintf(b, "## %s\n\n", title)
	if len(values) == 0 {
		fmt.Fprintln(b, "- none")
		fmt.Fprintln(b)
		return
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		fmt.Fprintf(b, "- %s\n", value)
	}
	fmt.Fprintln(b)
}

func writeSkillSection(b *strings.Builder, role registry.RoleDefinition, surface corerender.RoleSurface) {
	fmt.Fprintln(b, "## Skill Resolution")
	fmt.Fprintln(b)
	if len(surface.DirectSkills) == 0 {
		fmt.Fprintln(b, "- direct_skills=none")
	} else {
		for _, skill := range surface.DirectSkills {
			missing := skill.MissingBehavior
			if missing == "" {
				missing = "error"
			}
			fmt.Fprintf(b, "- direct %s -> %s (%s), category=%s, missing=%s\n", skill.Slot, skill.Name, skill.Path, skill.Category, missing)
		}
	}
	for _, slot := range role.SkillSlots.Delegated {
		fmt.Fprintf(b, "- delegated %s\n", slot)
	}
	for _, slot := range role.SkillSlots.Referenced {
		fmt.Fprintf(b, "- referenced %s\n", slot)
	}
	fmt.Fprintln(b)
}

func scalar(value any) string {
	if value == nil {
		return "not_configured"
	}
	return fmt.Sprint(value)
}

func permissionPolicyDefault(policy registry.RoleShellPolicy) string {
	if policy.Default == "" {
		return "not_configured"
	}
	return policy.Default
}

func hasPermissionPolicy(policy registry.RoleShellPolicy) bool {
	return policy.Default != "" || len(policy.Ask) > 0 || len(policy.Deny) > 0
}

func permissionPolicyList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}
