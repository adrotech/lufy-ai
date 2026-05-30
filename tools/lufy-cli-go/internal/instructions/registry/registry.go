package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"gopkg.in/yaml.v3"
)

type RoleDefinition struct {
	SchemaVersion    string          `yaml:"schema_version"`
	ID               domain.RoleID   `yaml:"id"`
	Kind             string          `yaml:"kind"`
	Purpose          string          `yaml:"purpose"`
	Permissions      RolePermissions `yaml:"permissions"`
	Delegation       RoleDelegation  `yaml:"delegation"`
	Responsibilities []string        `yaml:"responsibilities"`
	Boundaries       []string        `yaml:"boundaries"`
	Outputs          []string        `yaml:"outputs"`
	SkillSlots       RoleSkillSlots  `yaml:"skill_slots"`
	Output           RoleOutput      `yaml:"output_contract"`
}

type RolePermissions struct {
	Edit     any `yaml:"edit"`
	Shell    any `yaml:"shell"`
	Delivery any `yaml:"delivery"`
}

type RoleDelegation struct {
	Preferred string `yaml:"preferred"`
	Fallback  string `yaml:"fallback"`
}

type RoleSkillSlots struct {
	Direct     []domain.SkillSlot `yaml:"direct"`
	Delegated  []domain.SkillSlot `yaml:"delegated"`
	Referenced []domain.SkillSlot `yaml:"referenced"`
}

type RoleOutput struct {
	Schema          string   `yaml:"schema"`
	AllowedStatus   []string `yaml:"allowed_status"`
	CompactPayload  []string `yaml:"compact_payload"`
	MaxHandoffFocus []string `yaml:"max_handoff_focus"`
}

type SkillBinding struct {
	SchemaVersion string                         `yaml:"schema_version"`
	ID            string                         `yaml:"id"`
	Tool          domain.ToolID                  `yaml:"tool"`
	Methodology   domain.MethodologyID           `yaml:"methodology"`
	Skills        map[domain.SkillSlot]SkillSpec `yaml:"skills"`
	RoleContracts map[domain.RoleID]RoleBinding  `yaml:"role_contracts"`
}

type SkillSpec struct {
	Name            string   `yaml:"name"`
	Path            string   `yaml:"path"`
	Category        string   `yaml:"category"`
	RequiredFor     []string `yaml:"required_for"`
	LoadedBy        []string `yaml:"loaded_by"`
	MissingBehavior string   `yaml:"missing_behavior"`
}

type RoleBinding struct {
	DirectSlots    []domain.SkillSlot `yaml:"direct_slots"`
	DelegatedSlots []domain.SkillSlot `yaml:"delegated_slots"`
	HandoffRule    string             `yaml:"handoff_rule"`
}

type ResolvedSkill struct {
	Slot            domain.SkillSlot
	Name            string
	Path            string
	Category        string
	MissingBehavior string
}

func LoadRoleFile(path string) (RoleDefinition, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return RoleDefinition{}, err
	}
	var role RoleDefinition
	if err := yaml.Unmarshal(body, &role); err != nil {
		return RoleDefinition{}, fmt.Errorf("parse role %s: %w", path, err)
	}
	if err := role.Validate(); err != nil {
		return RoleDefinition{}, fmt.Errorf("invalid role %s: %w", path, err)
	}
	return role, nil
}

func LoadRoleDir(dir string) ([]RoleDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(paths)
	roles := make([]RoleDefinition, 0, len(paths))
	for _, path := range paths {
		role, err := LoadRoleFile(path)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func LoadSkillBindingFile(path string) (SkillBinding, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return SkillBinding{}, err
	}
	var binding SkillBinding
	if err := yaml.Unmarshal(body, &binding); err != nil {
		return SkillBinding{}, fmt.Errorf("parse skill binding %s: %w", path, err)
	}
	if err := binding.Validate(); err != nil {
		return SkillBinding{}, fmt.Errorf("invalid skill binding %s: %w", path, err)
	}
	return binding, nil
}

func ResolveDirectSkills(role RoleDefinition, binding SkillBinding) ([]ResolvedSkill, error) {
	out := make([]ResolvedSkill, 0, len(role.SkillSlots.Direct))
	for _, slot := range role.SkillSlots.Direct {
		spec, ok := binding.Skills[slot]
		if !ok {
			return nil, fmt.Errorf("role %s direct slot %s has no binding", role.ID, slot)
		}
		out = append(out, ResolvedSkill{
			Slot:            slot,
			Name:            spec.Name,
			Path:            spec.Path,
			Category:        spec.Category,
			MissingBehavior: spec.MissingBehavior,
		})
	}
	return out, nil
}

func RoleByID(roles []RoleDefinition, id domain.RoleID) (RoleDefinition, bool) {
	for _, role := range roles {
		if role.ID == id {
			return role, true
		}
	}
	return RoleDefinition{}, false
}

func (r RoleDefinition) Validate() error {
	if r.SchemaVersion != "lufy-role/v1" {
		return fmt.Errorf("schema_version esperado lufy-role/v1, recibido %q", r.SchemaVersion)
	}
	if r.ID == "" {
		return fmt.Errorf("id requerido")
	}
	if r.Kind == "" {
		return fmt.Errorf("kind requerido")
	}
	if r.Output.Schema != "result-contract/v1" {
		return fmt.Errorf("output_contract.schema esperado result-contract/v1")
	}
	if len(r.Output.AllowedStatus) == 0 {
		return fmt.Errorf("output_contract.allowed_status requerido")
	}
	return nil
}

func (b SkillBinding) Validate() error {
	if b.SchemaVersion != "lufy-role-skill-binding/v1" {
		return fmt.Errorf("schema_version esperado lufy-role-skill-binding/v1, recibido %q", b.SchemaVersion)
	}
	if b.ID == "" {
		return fmt.Errorf("id requerido")
	}
	if b.Tool == "" {
		return fmt.Errorf("tool requerido")
	}
	if b.Methodology == "" {
		return fmt.Errorf("methodology requerido")
	}
	for slot, spec := range b.Skills {
		if slot == "" {
			return fmt.Errorf("skill slot vacio")
		}
		if spec.Name == "" {
			return fmt.Errorf("skill %s sin name", slot)
		}
		if spec.Path == "" {
			return fmt.Errorf("skill %s sin path", slot)
		}
	}
	return nil
}
