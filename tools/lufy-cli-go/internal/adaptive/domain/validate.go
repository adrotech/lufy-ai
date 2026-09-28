package domain

import (
	"fmt"
	"regexp"
)

const (
	maxInputBytes   = 64 * 1024
	maxCapabilities = 32
	maxProfiles     = 32
	maxAssignments  = 32
	maxRoleHints    = 8
)

var (
	idPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{1,127}$`)
	tokenPattern  = regexp.MustCompile(`^[a-z][a-z0-9._:-]{0,63}$`)
	digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

func ValidateEvaluation(request EvaluationRequest) error {
	if request.SchemaVersion != EvaluationSchemaVersion {
		return fieldError("schema_version", "usar lufy-adaptive-evaluation/v1")
	}
	if err := validateDemand(request.Demand); err != nil {
		return err
	}
	if len(request.Profiles) == 0 || len(request.Profiles) > maxProfiles {
		return fieldError("profiles", "usar entre 1 y 32 perfiles")
	}
	actors := map[string]bool{}
	for _, profile := range request.Profiles {
		if err := validateProfile(profile); err != nil {
			return err
		}
		if actors[profile.ActorRef] {
			return fieldError("profiles.actor_ref", "eliminar actor_ref duplicado")
		}
		actors[profile.ActorRef] = true
	}
	return nil
}

func validateDemand(demand DemandSignal) error {
	if demand.SchemaVersion != DemandSignalSchemaVersion {
		return fieldError("demand.schema_version", "usar lufy-demand-signal/v1")
	}
	if !idPattern.MatchString(demand.DemandID) {
		return fieldError("demand.demand_id", "usar un identificador bounded")
	}
	if !idPattern.MatchString(demand.TaskRef) {
		return fieldError("demand.task_ref", "usar una referencia bounded")
	}
	if demand.SnapshotVersion == 0 {
		return fieldError("demand.snapshot_version", "usar una versión positiva")
	}
	if err := rangeValue("demand.priority", demand.Priority, 0, 100); err != nil {
		return err
	}
	if err := rangeValue("demand.required_budget", demand.RequiredBudget, 1, 100); err != nil {
		return err
	}
	if err := rangeValue("demand.risk", demand.Risk, 0, 100); err != nil {
		return err
	}
	if err := rangeValue("demand.coordination_cost", demand.CoordinationCost, 0, 100); err != nil {
		return err
	}
	if len(demand.RequiredCapabilities) == 0 || len(demand.RequiredCapabilities) > maxCapabilities {
		return fieldError("demand.required_capabilities", "usar entre 1 y 32 capabilities")
	}
	seen := map[string]bool{}
	for _, capability := range demand.RequiredCapabilities {
		if !tokenPattern.MatchString(capability.Name) {
			return fieldError("demand.required_capabilities.name", "usar un token documentado")
		}
		if seen[capability.Name] {
			return fieldError("demand.required_capabilities.name", "eliminar capability duplicada")
		}
		seen[capability.Name] = true
		if err := rangeValue("demand.required_capabilities.level", capability.Level, 1, 100); err != nil {
			return err
		}
		if err := rangeValue("demand.required_capabilities.weight", capability.Weight, 1, 10); err != nil {
			return err
		}
	}
	if len(demand.ProtectedBoundaries) > 8 {
		return fieldError("demand.protected_boundaries", "usar hasta 8 boundaries")
	}
	for _, boundary := range demand.ProtectedBoundaries {
		if !oneOf(boundary, "delivery", "security", "public_contract", "database_schema", "destructive_migration") {
			return fieldError("demand.protected_boundaries", "usar una frontera protegida documentada")
		}
	}
	if demand.CausedByEventID != "" && !idPattern.MatchString(demand.CausedByEventID) {
		return fieldError("demand.caused_by_event_id", "usar una referencia bounded")
	}
	return nil
}

func validateProfile(profile CapabilityProfile) error {
	if profile.SchemaVersion != CapabilityProfileSchemaVersion {
		return fieldError("profiles.schema_version", "usar lufy-capability-profile/v1")
	}
	if !digestPattern.MatchString(profile.ActorRef) {
		return fieldError("profiles.actor_ref", "usar una referencia SHA-256")
	}
	if len(profile.Capabilities) == 0 || len(profile.Capabilities) > maxCapabilities {
		return fieldError("profiles.capabilities", "usar entre 1 y 32 capabilities")
	}
	seen := map[string]bool{}
	for _, capability := range profile.Capabilities {
		if !tokenPattern.MatchString(capability.Name) {
			return fieldError("profiles.capabilities.name", "usar un token documentado")
		}
		if seen[capability.Name] {
			return fieldError("profiles.capabilities.name", "eliminar capability duplicada")
		}
		seen[capability.Name] = true
		if err := rangeValue("profiles.capabilities.level", capability.Level, 0, 100); err != nil {
			return err
		}
	}
	for path, value := range map[string]int{
		"profiles.available_budget":  profile.AvailableBudget,
		"profiles.risk_tolerance":    profile.RiskTolerance,
		"profiles.coordination_cost": profile.CoordinationCost,
	} {
		if err := rangeValue(path, value, 0, 100); err != nil {
			return err
		}
	}
	if len(profile.ActiveAssignments) > maxAssignments {
		return fieldError("profiles.active_assignments", "usar hasta 32 assignments")
	}
	for _, assignment := range profile.ActiveAssignments {
		if !idPattern.MatchString(assignment) {
			return fieldError("profiles.active_assignments", "usar referencias bounded")
		}
	}
	if len(profile.RoleHints) > maxRoleHints {
		return fieldError("profiles.role_hints", "usar hasta 8 role hints")
	}
	for _, role := range profile.RoleHints {
		if !tokenPattern.MatchString(role) {
			return fieldError("profiles.role_hints", "usar tokens documentados")
		}
	}
	return nil
}

func validatePolicy(policy Policy) error {
	if policy.Version != PolicyDeterministicV1 {
		return fieldError("policy.version", "usar deterministic-v1")
	}
	if policy.Mode != ModeShadow && policy.Mode != ModeAdvisory {
		return fieldError("policy.mode", "usar shadow o advisory")
	}
	if policy.MaxCandidates < 1 || policy.MaxCandidates > maxProfiles {
		return fieldError("policy.max_candidates", "usar un valor entre 1 y 32")
	}
	return nil
}

func fieldError(path, recovery string) error {
	return fmt.Errorf("adaptive input inválido: %s; %s", path, recovery)
}

func rangeValue(path string, value, min, max int) error {
	if value < min || value > max {
		return fieldError(path, fmt.Sprintf("usar un entero entre %d y %d", min, max))
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
