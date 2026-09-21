package resultcontract

import (
	"io"
	"reflect"
)

const LegacySchemaVersion = "lufy-result-legacy/v1"

type legacyResult struct {
	SchemaVersion string `json:"schema_version" yaml:"schema_version"`
	Status        Status `json:"status" yaml:"status"`
	Summary       string `json:"summary" yaml:"summary"`
	ProvenanceRef string `json:"provenance_ref" yaml:"provenance_ref"`
}

func NormalizeLegacy(input io.Reader) (Contract, error) {
	root, err := decodeDocument(input)
	if err != nil {
		return Contract{}, err
	}
	if err := validateNodeShape(root, reflect.TypeOf(legacyResult{}), ""); err != nil {
		return Contract{}, err
	}
	var legacy legacyResult
	if err := root.Decode(&legacy); err != nil {
		return Contract{}, diagnostic("legacy_decode_failed", "$", "usar lufy-result-legacy/v1 estructurado")
	}
	if legacy.SchemaVersion != LegacySchemaVersion {
		return Contract{}, diagnostic("unsupported_legacy_schema", "schema_version", "usar lufy-result-legacy/v1")
	}
	if !oneOf(string(legacy.Status), "ready", "implemented", "blocked", "escalated") {
		return Contract{}, diagnostic("legacy_status_unsupported", "status", "normalizar solo claims no avanzados")
	}
	if legacy.Summary == "" {
		return Contract{}, required("summary")
	}
	if !fingerprintPattern.MatchString(legacy.ProvenanceRef) {
		return Contract{}, diagnostic("invalid_provenance", "provenance_ref", "usar un digest SHA-256 de procedencia")
	}

	contract := Contract{
		SchemaVersion:    SchemaVersion,
		Status:           legacy.Status,
		LegacyFallback:   true,
		ExecutiveSummary: legacy.Summary,
		Artifacts: Artifacts{
			Changed:    []string{"none"},
			Referenced: []string{"legacy:" + legacy.ProvenanceRef},
		},
		Evidence: Evidence{
			Commands: []CommandEvidence{{Command: "none", Result: EvidenceNotRun, Notes: "not_available"}},
			Static:   []string{"not_available"},
		},
		SurfaceExecution: SurfaceExecution{
			SchemaVersion: "not_applicable", Source: "not_applicable", PrimarySurface: "not_available",
			Mode: "not_applicable", ActiveSurfaces: []string{"not_applicable"}, ValidationRuleIDs: []string{"not_applicable"},
		},
		WorkflowDecision: WorkflowDecision{
			Tier: "not_applicable", ProgramTier: "not_applicable", SliceTier: "not_applicable", FastPathAllowed: BoolNotApplicable,
			AdapterContext: AdapterContext{
				ToolID: "not_applicable", MethodologyID: "not_applicable", MethodologyMode: "not_applicable",
				MethodologyRequired: BoolNotApplicable, ExecutionMode: "not_applicable",
			},
			WorkflowLimitsSource: "not_available",
			WorkflowLimitsPaths: WorkflowLimitsPaths{
				Sizing: "not_available", Routing: "not_available", ProposalSlicing: "not_available",
				DeliveryBatching: "not_applicable", Preflight: "not_available", StopRules: "not_available",
			},
			WorkloadDecisionNeeded: BoolFalse, ReviewSlices: []string{"not_applicable"},
			PreflightStatus: "not_applicable", StopRuleStatus: "not_applicable", DeliveryBatchingGuidance: "not_applicable",
			ArtifactBranching: ArtifactBranching{
				Status: "not_applicable", Stage: "not_applicable", CandidateCount: CountNotApplicable, Reason: "not_applicable",
				ParallelAllowed: BoolFalse, RequiresJoin: BoolFalse, CandidateIsolation: "not_applicable", MergePlanRequired: BoolFalse,
				HumanEscalationTriggers: []string{"not_applicable"}, CandidatePaths: []string{"not_applicable"}, CanonicalArtifactSet: []string{"not_applicable"},
			},
		},
		Risks:           []string{"none"},
		NextRecommended: NextRecommended{Owner: "none", Action: "Revalidar evidencia antes de avanzar el gate."},
		SkillResolution: SkillResolution{
			LocalSkillsUsed: []string{"none"}, SelectedSkillPaths: []string{"none"}, BootstrapRecommended: false, Notes: "legacy normalization",
		},
	}
	if err := Validate(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}
