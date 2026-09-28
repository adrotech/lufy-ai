package resultcontract

import (
	"testing"
)

func TestDeriveStateVectorForEveryStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status Status
		want   StateVector
	}{
		{StatusReady, StateVector{Work: "ready", Delivery: "not_required", Sync: "not_required", Attention: "none"}},
		{StatusImplemented, StateVector{Work: "implemented", Delivery: "not_required", Sync: "not_required", Attention: "none"}},
		{StatusValidated, StateVector{Work: "validated", Delivery: "not_required", Sync: "not_required", Attention: "none"}},
		{StatusDeliveryPending, StateVector{Work: "validated", Delivery: "pending", Sync: "not_required", Attention: "none"}},
		{StatusSyncPending, StateVector{Work: "validated", Delivery: "not_required", Sync: "pending", Attention: "none"}},
		{StatusDelivered, StateVector{Work: "validated", Delivery: "delivered", Sync: "not_required", Attention: "none"}},
		{StatusClosed, StateVector{Work: "terminal", Delivery: "delivered", Sync: "synced", Attention: "none", Terminal: true}},
	}
	for _, tc := range cases {
		got, err := DeriveState(tc.status, nil)
		if err != nil {
			t.Fatalf("DeriveState(%q) error = %v", tc.status, err)
		}
		if got != tc.want {
			t.Fatalf("DeriveState(%q) = %#v, want %#v", tc.status, got, tc.want)
		}
	}
}

func TestDeriveStatePreservesDeliveryAndSyncWhenBlockedOrEscalated(t *testing.T) {
	t.Parallel()

	prior := StateVector{Work: "validated", Delivery: "pending", Sync: "synced", Attention: "none"}
	for _, tc := range []struct {
		status    Status
		attention string
	}{
		{StatusBlocked, "recovery_required"},
		{StatusEscalated, "escalated"},
	} {
		got, err := DeriveState(tc.status, &prior)
		if err != nil {
			t.Fatalf("DeriveState(%q) error = %v", tc.status, err)
		}
		if got.Work != "blocked" || got.Delivery != prior.Delivery || got.Sync != prior.Sync || got.Attention != tc.attention {
			t.Fatalf("DeriveState(%q) = %#v, prior = %#v", tc.status, got, prior)
		}
	}
}

func TestAllowedTransitionTableIsExplicitAndClosedIsTerminal(t *testing.T) {
	t.Parallel()

	all := []Status{
		StatusReady, StatusImplemented, StatusValidated, StatusDeliveryPending,
		StatusSyncPending, StatusBlocked, StatusEscalated, StatusDelivered, StatusClosed,
	}
	allowed := map[Status][]Status{
		StatusReady:           {StatusImplemented, StatusBlocked, StatusEscalated},
		StatusImplemented:     {StatusValidated, StatusBlocked, StatusEscalated},
		StatusValidated:       {StatusDeliveryPending, StatusSyncPending, StatusDelivered, StatusClosed, StatusBlocked, StatusEscalated},
		StatusDeliveryPending: {StatusSyncPending, StatusDelivered, StatusBlocked, StatusEscalated},
		StatusSyncPending:     {StatusDeliveryPending, StatusDelivered, StatusBlocked, StatusEscalated},
		StatusBlocked:         {StatusReady, StatusImplemented, StatusValidated, StatusDeliveryPending, StatusSyncPending, StatusEscalated},
		StatusEscalated:       {StatusReady, StatusImplemented, StatusValidated, StatusDeliveryPending, StatusSyncPending, StatusBlocked},
		StatusDelivered:       {StatusSyncPending, StatusClosed, StatusBlocked, StatusEscalated},
		StatusClosed:          {},
	}
	for _, from := range all {
		for _, to := range all {
			want := containsStatus(allowed[from], to)
			if got := IsAllowedTransition(from, to); got != want {
				t.Fatalf("IsAllowedTransition(%q,%q) = %t, want %t", from, to, got, want)
			}
		}
	}
}

func TestEvaluateTransitionAcceptsCurrentCASAndReturnsNextIdentity(t *testing.T) {
	t.Parallel()

	prior := transitionState(t, StatusImplemented, 4)
	next := contractWithStatus(t, StatusValidated)
	next.Evidence.Commands[0].Result = EvidencePassed
	wantCanonical := mustCanonicalize(t, next)
	intent := validTransitionIntent(prior, next, "validator")
	decision := EvaluateTransition(prior, intent, transitionTestPolicy())

	if decision.Status != DecisionAccepted {
		t.Fatalf("EvaluateTransition() = %#v, want accepted", decision)
	}
	if decision.NextVersion != prior.Version+1 || decision.NextFingerprint != wantCanonical.Fingerprint {
		t.Fatalf("next identity = version %d fingerprint %q, want %d/%q",
			decision.NextVersion, decision.NextFingerprint, prior.Version+1, wantCanonical.Fingerprint)
	}
	if decision.State.Work != "validated" || decision.State.Terminal {
		t.Fatalf("state = %#v, want validated non-terminal", decision.State)
	}
}

func TestEvaluateTransitionRejectsStaleVersionOrFingerprintAsConflict(t *testing.T) {
	t.Parallel()

	prior := transitionState(t, StatusImplemented, 4)
	next := contractWithStatus(t, StatusValidated)
	next.Evidence.Commands[0].Result = EvidencePassed
	cases := []struct {
		name   string
		mutate func(*TransitionIntent)
		reason string
	}{
		{name: "stale version", mutate: func(intent *TransitionIntent) { intent.ExpectedVersion-- }, reason: "stale_version"},
		{name: "stale fingerprint", mutate: func(intent *TransitionIntent) { intent.PreviousFingerprint = stringsOf('a', 64) }, reason: "stale_fingerprint"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			intent := validTransitionIntent(prior, next, "validator")
			tc.mutate(&intent)
			decision := EvaluateTransition(prior, intent, transitionTestPolicy())
			if decision.Status != DecisionConflict || decision.Reason != tc.reason {
				t.Fatalf("decision = %#v, want conflict/%s", decision, tc.reason)
			}
			if decision.NextVersion != 0 || decision.NextFingerprint != "" || decision.Recovery == "" {
				t.Fatalf("conflict exposed a next state or lacks recovery: %#v", decision)
			}
		})
	}
}

func TestEvaluateTransitionRejectsInvalidProtocolAndDisallowedEdges(t *testing.T) {
	t.Parallel()

	prior := transitionState(t, StatusReady, 2)
	next := contractWithStatus(t, StatusClosed)
	next.Evidence.Commands[0].Result = EvidencePassed
	cases := []struct {
		name   string
		mutate func(*TransitionIntent)
		reason string
	}{
		{name: "unsupported protocol", mutate: func(intent *TransitionIntent) { intent.SchemaVersion = "result-transition/v2" }, reason: "unsupported_transition_schema"},
		{name: "missing transition id", mutate: func(intent *TransitionIntent) { intent.TransitionID = "" }, reason: "invalid_transition"},
		{name: "ready cannot skip to closed", mutate: func(intent *TransitionIntent) {}, reason: "transition_not_allowed"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			intent := validTransitionIntent(prior, next, "orchestrator")
			tc.mutate(&intent)
			decision := EvaluateTransition(prior, intent, transitionTestPolicy())
			if decision.Status != DecisionRejected || decision.Reason != tc.reason || decision.Recovery == "" {
				t.Fatalf("decision = %#v, want rejected/%s with recovery", decision, tc.reason)
			}
		})
	}

	closed := transitionState(t, StatusClosed, 9)
	blocked := contractWithStatus(t, StatusBlocked)
	decision := EvaluateTransition(closed, validTransitionIntent(closed, blocked, "delivery"), transitionTestPolicy())
	if decision.Status != DecisionRejected || decision.Reason != "terminal_state" {
		t.Fatalf("transition leaving closed = %#v, want terminal rejection", decision)
	}
}

func TestRecoveryFromBlockedOrEscalatedRequiresConfiguredEvidenceAndNewVersion(t *testing.T) {
	t.Parallel()

	policy := transitionTestPolicy()
	policy.Evidence.Recovery = EvidenceRequirement{
		Categories: []EvidenceCategory{EvidenceCategoryPassedCommand},
		NextOwner:  "implementer",
	}
	for _, priorStatus := range []Status{StatusBlocked, StatusEscalated} {
		prior := transitionState(t, priorStatus, 7)
		next := contractWithStatus(t, StatusImplemented)
		intent := validTransitionIntent(prior, next, "implementer")

		missing := EvaluateTransition(prior, intent, policy)
		if missing.Status != DecisionRejected || missing.Reason != "recovery_evidence_missing" ||
			!containsEvidenceCategory(missing.MissingEvidence, EvidenceCategoryPassedCommand) {
			t.Fatalf("recovery without evidence = %#v", missing)
		}

		next.Evidence.Commands[0].Result = EvidencePassed
		intent = validTransitionIntent(prior, next, "implementer")
		accepted := EvaluateTransition(prior, intent, policy)
		if accepted.Status != DecisionAccepted || accepted.NextVersion != prior.Version+1 {
			t.Fatalf("recovery with evidence = %#v, want accepted at version %d", accepted, prior.Version+1)
		}
	}
}

func transitionState(t *testing.T, status Status, version uint64) TransitionState {
	t.Helper()
	contract := contractWithStatus(t, status)
	canonical := mustCanonicalize(t, contract)
	vector, err := DeriveState(status, nil)
	if status == StatusBlocked || status == StatusEscalated {
		prior := StateVector{Work: "validated", Delivery: "pending", Sync: "not_required", Attention: "none"}
		vector, err = DeriveState(status, &prior)
	}
	if err != nil {
		t.Fatalf("DeriveState(%q) error = %v", status, err)
	}
	return TransitionState{Version: version, Fingerprint: canonical.Fingerprint, Contract: contract, State: vector}
}

func contractWithStatus(t *testing.T, status Status) Contract {
	t.Helper()
	contract := decodeFixture(t, "testdata/valid-minimal.yaml")
	contract.Status = status
	return contract
}

func validTransitionIntent(prior TransitionState, next Contract, role string) TransitionIntent {
	return TransitionIntent{
		SchemaVersion:       TransitionSchemaVersion,
		TransitionID:        "transition-slice-b",
		IdempotencyKey:      "transition-slice-b-key",
		ExpectedVersion:     prior.Version,
		PreviousFingerprint: prior.Fingerprint,
		Actor: TransitionActor{
			Role:     role,
			OwnerRef: "owner-ref-pseudonymous",
		},
		NextContract: next,
	}
}

func transitionTestPolicy() TransitionPolicy {
	return TransitionPolicy{
		Roles: currentRolePolicy(),
		Evidence: EvidencePolicy{ByStatus: map[Status]EvidenceRequirement{
			StatusValidated: {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand}, NextOwner: "validator"},
			StatusDelivered: {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference}, NextOwner: "delivery"},
			StatusClosed:    {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference, EvidenceCategoryStatic}, NextOwner: "orchestrator"},
		}},
	}
}

func containsStatus(statuses []Status, target Status) bool {
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}

func containsEvidenceCategory(categories []EvidenceCategory, target EvidenceCategory) bool {
	for _, category := range categories {
		if category == target {
			return true
		}
	}
	return false
}

func stringsOf(value byte, count int) string {
	result := make([]byte, count)
	for index := range result {
		result[index] = value
	}
	return string(result)
}
