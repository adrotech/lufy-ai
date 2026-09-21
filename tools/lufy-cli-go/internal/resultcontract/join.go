package resultcontract

type TransitionJoin struct {
	RequiredRunIDs               []string `json:"required_run_ids" yaml:"required_run_ids"`
	TerminalContractFingerprints []string `json:"terminal_contract_fingerprints" yaml:"terminal_contract_fingerprints"`
	GroupedEvidenceFingerprint   string   `json:"grouped_evidence_fingerprint" yaml:"grouped_evidence_fingerprint"`
}

type JoinChild struct {
	RunID               string
	ParentRunID         string
	Status              Status
	ContractFingerprint string
}

type JoinResolver interface {
	Resolve(runID string) (JoinChild, bool, error)
}

func evaluateJoin(previous TransitionState, intent TransitionIntent, resolver JoinResolver) *TransitionDecision {
	required := previous.RequiredJoinRunIDs
	if len(required) == 0 {
		if intent.Join != nil {
			decision := rejectedTransition("join_children_mismatch", "eliminar el join no requerido")
			return &decision
		}
		return nil
	}
	if intent.Join == nil {
		decision := rejectedTransition("join_required", "declarar todos los children requeridos")
		return &decision
	}
	if !sameUniqueSet(required, intent.Join.RequiredRunIDs) || len(intent.Join.RequiredRunIDs) != len(intent.Join.TerminalContractFingerprints) {
		decision := rejectedTransition("join_children_mismatch", "declarar exactamente el conjunto requerido")
		return &decision
	}
	if intent.Join.GroupedEvidenceFingerprint == "" {
		decision := rejectedTransition("join_grouped_evidence_missing", "agregar el fingerprint de evidencia agrupada")
		return &decision
	}
	if !fingerprintPattern.MatchString(intent.Join.GroupedEvidenceFingerprint) {
		decision := rejectedTransition("join_grouped_evidence_invalid", "usar un digest SHA-256 para evidencia agrupada")
		return &decision
	}
	if resolver == nil {
		decision := rejectedTransition("join_unavailable", "proveer un resolver de children")
		return &decision
	}
	fingerprints := make(map[string]string, len(intent.Join.RequiredRunIDs))
	for index, runID := range intent.Join.RequiredRunIDs {
		fingerprints[runID] = intent.Join.TerminalContractFingerprints[index]
	}
	for _, runID := range required {
		child, found, err := resolver.Resolve(runID)
		if err != nil || !found {
			decision := rejectedTransition("join_incomplete", "resolver todos los children requeridos")
			return &decision
		}
		if child.RunID != runID || child.ParentRunID != previous.RunID {
			decision := rejectedTransition("join_causal_mismatch", "usar children causalmente vinculados al run padre")
			return &decision
		}
		if child.Status != StatusClosed {
			decision := rejectedTransition("join_child_not_terminal", "esperar un contrato terminal para cada child")
			return &decision
		}
		declared := fingerprints[runID]
		if !fingerprintPattern.MatchString(declared) || !fingerprintPattern.MatchString(child.ContractFingerprint) || declared != child.ContractFingerprint {
			decision := conflictTransition("join_fingerprint_mismatch", "recargar los fingerprints terminales de los children")
			return &decision
		}
	}
	return nil
}

func sameUniqueSet(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	seen := make(map[string]bool, len(expected))
	for _, value := range expected {
		if value == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	matched := make(map[string]bool, len(actual))
	for _, value := range actual {
		if !seen[value] || matched[value] {
			return false
		}
		matched[value] = true
	}
	return true
}
