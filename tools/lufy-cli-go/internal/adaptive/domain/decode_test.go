package domain

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeEvaluationAcceptsStrictYAML(t *testing.T) {
	t.Parallel()
	body := `schema_version: lufy-adaptive-evaluation/v1
demand:
  schema_version: lufy-demand-signal/v1
  demand_id: demand-1
  task_ref: task-1
  snapshot_version: 1
  priority: 70
  required_budget: 30
  risk: 20
  coordination_cost: 10
  required_capabilities:
    - name: implementation
      level: 80
      weight: 5
profiles:
  - schema_version: lufy-capability-profile/v1
    actor_ref: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    capabilities:
      - name: implementation
        level: 90
    available_budget: 60
    risk_tolerance: 50
    coordination_cost: 10
    role_hints: [implementer]
`
	request, err := DecodeEvaluation(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if request.Demand.DemandID != "demand-1" || len(request.Profiles) != 1 {
		t.Fatalf("request = %#v", request)
	}
}

func TestDecodeEvaluationRejectsUnsafeOrUnboundedInput(t *testing.T) {
	t.Parallel()
	base := `schema_version: lufy-adaptive-evaluation/v1
demand:
  schema_version: lufy-demand-signal/v1
  demand_id: demand-1
  task_ref: task-1
  snapshot_version: 1
  priority: 70
  required_budget: 30
  risk: 20
  coordination_cost: 10
  required_capabilities:
    - name: implementation
      level: 80
      weight: 5
profiles:
  - schema_version: lufy-capability-profile/v1
    actor_ref: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    capabilities:
      - name: implementation
        level: 90
    available_budget: 60
    risk_tolerance: 50
    coordination_cost: 10
    role_hints: [implementer]
`
	tests := map[string][]byte{
		"unknown content": []byte(base + "prompt: super-secret-canary\n"),
		"duplicate":       []byte(strings.Replace(base, "  priority: 70", "  priority: 70\n  priority: 71", 1)),
		"invalid utf8":    {0xff, 0xfe, 0xfd},
		"oversized":       bytes.Repeat([]byte("x"), maxInputBytes+1),
		"yaml alias":      []byte("schema_version: &v lufy-adaptive-evaluation/v1\ndemand: *v\nprofiles: []\n"),
	}
	for name, body := range tests {
		name, body := name, body
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeEvaluation(bytes.NewReader(body)); err == nil {
				t.Fatal("se esperaba rechazo")
			} else if strings.Contains(err.Error(), "super-secret-canary") {
				t.Fatalf("diagnóstico filtró contenido: %v", err)
			}
		})
	}
}

func TestValidateEvaluationRejectsInvalidRangesAndRefs(t *testing.T) {
	t.Parallel()
	tests := []func(*EvaluationRequest){
		func(r *EvaluationRequest) { r.Demand.Priority = 101 },
		func(r *EvaluationRequest) { r.Demand.RequiredCapabilities[0].Weight = 0 },
		func(r *EvaluationRequest) { r.Profiles[0].ActorRef = "raw-agent-name" },
		func(r *EvaluationRequest) { r.Demand.ProtectedBoundaries = []string{"unknown"} },
		func(r *EvaluationRequest) {
			r.Profiles[0].Capabilities = append(r.Profiles[0].Capabilities, r.Profiles[0].Capabilities[0])
		},
	}
	for index, mutate := range tests {
		request := validEvaluationRequest()
		mutate(&request)
		if err := ValidateEvaluation(request); err == nil {
			t.Fatalf("case %d: se esperaba error", index)
		}
	}
}
