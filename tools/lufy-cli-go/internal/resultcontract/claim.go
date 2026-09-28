package resultcontract

import "strings"

type RolePolicy map[string][]Status

type EvidenceCategory string

const (
	EvidenceCategoryPassedCommand     EvidenceCategory = "passed_command"
	EvidenceCategoryArtifactReference EvidenceCategory = "artifact_reference"
	EvidenceCategoryStatic            EvidenceCategory = "static"
)

type EvidenceRequirement struct {
	Categories []EvidenceCategory
	NextOwner  string
}

type EvidencePolicy struct {
	ByStatus map[Status]EvidenceRequirement
	Recovery EvidenceRequirement
}

type ClaimDecision struct {
	Accepted        bool
	Reason          string
	Recovery        string
	MissingEvidence []EvidenceCategory
	NextOwner       string
}

func EvaluateClaim(contract Contract, role string, roles RolePolicy, evidence EvidencePolicy) ClaimDecision {
	allowedStatuses, knownRole := roles[role]
	if !knownRole {
		return rejectedClaim("unknown_role", "usar un rol registrado", nil, "orchestrator")
	}
	if !statusIn(allowedStatuses, contract.Status) {
		return rejectedClaim("role_status_not_allowed", "usar un status permitido para el rol", nil, role)
	}

	requirement, configured := evidence.ByStatus[contract.Status]
	if advancedStatus(contract.Status) && !configured {
		return rejectedClaim("evidence_policy_unavailable", "configurar evidencia para el claim avanzado", nil, role)
	}
	if configured {
		missing := missingEvidence(contract, requirement)
		if len(missing) > 0 {
			nextOwner := requirement.NextOwner
			if nextOwner == "" {
				nextOwner = role
			}
			return rejectedClaim("missing_evidence", "agregar evidencia observable requerida", missing, nextOwner)
		}
	}
	return ClaimDecision{Accepted: true}
}

func rejectedClaim(reason, recovery string, missing []EvidenceCategory, nextOwner string) ClaimDecision {
	return ClaimDecision{
		Accepted:        false,
		Reason:          reason,
		Recovery:        recovery,
		MissingEvidence: missing,
		NextOwner:       nextOwner,
	}
}

func advancedStatus(status Status) bool {
	switch status {
	case StatusValidated, StatusDelivered, StatusClosed:
		return true
	default:
		return false
	}
}

func missingEvidence(contract Contract, requirement EvidenceRequirement) []EvidenceCategory {
	missing := make([]EvidenceCategory, 0, len(requirement.Categories))
	seen := make(map[EvidenceCategory]bool, len(requirement.Categories))
	for _, category := range requirement.Categories {
		if seen[category] {
			continue
		}
		seen[category] = true
		if !hasEvidence(contract, category) {
			missing = append(missing, category)
		}
	}
	return missing
}

func hasEvidence(contract Contract, category EvidenceCategory) bool {
	switch category {
	case EvidenceCategoryPassedCommand:
		for _, command := range contract.Evidence.Commands {
			if command.Result == EvidencePassed {
				return true
			}
		}
	case EvidenceCategoryArtifactReference:
		return hasNonPlaceholder(contract.Artifacts.Referenced)
	case EvidenceCategoryStatic:
		return hasNonPlaceholder(contract.Evidence.Static)
	}
	return false
}

func hasNonPlaceholder(values []string) bool {
	for _, value := range values {
		switch strings.TrimSpace(value) {
		case "", "none", "not_applicable", "not_available":
			continue
		default:
			return true
		}
	}
	return false
}

func statusIn(statuses []Status, target Status) bool {
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}
