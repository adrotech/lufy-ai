package resultcontract

import (
	"fmt"
	"testing"
)

func TestJoinAcceptsOnlyExplicitTerminalCausallyRelatedChildrenWithGroupedEvidence(t *testing.T) {
	t.Parallel()

	prior, intent, policy, _ := validJoinTransition(t)
	decision := EvaluateTransition(prior, intent, policy)
	if decision.Status != DecisionAccepted {
		t.Fatalf("valid join decision = %#v, want accepted", decision)
	}
}

func TestJoinRejectsMissingDeclarationOrDifferentRequiredSet(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*TransitionIntent)
		reason string
	}{
		{name: "join omitted", mutate: func(intent *TransitionIntent) { intent.Join = nil }, reason: "join_required"},
		{name: "required child omitted", mutate: func(intent *TransitionIntent) {
			intent.Join.RequiredRunIDs = intent.Join.RequiredRunIDs[:1]
			intent.Join.TerminalContractFingerprints = intent.Join.TerminalContractFingerprints[:1]
		}, reason: "join_children_mismatch"},
		{name: "undeclared child added", mutate: func(intent *TransitionIntent) {
			intent.Join.RequiredRunIDs = append(intent.Join.RequiredRunIDs, "run-child-extra")
			intent.Join.TerminalContractFingerprints = append(intent.Join.TerminalContractFingerprints, testDigest("child-extra"))
		}, reason: "join_children_mismatch"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy, _ := validJoinTransition(t)
			tc.mutate(&intent)
			decision := EvaluateTransition(prior, intent, policy)
			assertRejectedWithoutNextState(t, decision, DecisionRejected, tc.reason)
		})
	}
}

func TestJoinRejectsMissingBlockedOrCausallyUnrelatedChild(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*mapJoinResolver, []string)
		reason string
	}{
		{
			name:   "missing child",
			mutate: func(resolver *mapJoinResolver, ids []string) { delete(resolver.children, ids[1]) },
			reason: "join_incomplete",
		},
		{
			name: "blocked child",
			mutate: func(resolver *mapJoinResolver, ids []string) {
				child := resolver.children[ids[1]]
				child.Status = StatusBlocked
				resolver.children[ids[1]] = child
			},
			reason: "join_child_not_terminal",
		},
		{
			name: "causal mismatch",
			mutate: func(resolver *mapJoinResolver, ids []string) {
				child := resolver.children[ids[0]]
				child.ParentRunID = "unrelated-parent"
				resolver.children[ids[0]] = child
			},
			reason: "join_causal_mismatch",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy, resolver := validJoinTransition(t)
			tc.mutate(resolver, prior.RequiredJoinRunIDs)
			decision := EvaluateTransition(prior, intent, policy)
			assertRejectedWithoutNextState(t, decision, DecisionRejected, tc.reason)
		})
	}
}

func TestJoinRejectsFingerprintMismatchAndMissingGroupedEvidence(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*TransitionIntent)
		status DecisionStatus
		reason string
	}{
		{
			name: "terminal fingerprint mismatch",
			mutate: func(intent *TransitionIntent) {
				intent.Join.TerminalContractFingerprints[0] = testDigest("different-contract")
			},
			status: DecisionConflict,
			reason: "join_fingerprint_mismatch",
		},
		{
			name:   "missing grouped evidence",
			mutate: func(intent *TransitionIntent) { intent.Join.GroupedEvidenceFingerprint = "" },
			status: DecisionRejected,
			reason: "join_grouped_evidence_missing",
		},
		{
			name:   "invalid grouped evidence digest",
			mutate: func(intent *TransitionIntent) { intent.Join.GroupedEvidenceFingerprint = "raw-evidence" },
			status: DecisionRejected,
			reason: "join_grouped_evidence_invalid",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy, _ := validJoinTransition(t)
			tc.mutate(&intent)
			decision := EvaluateTransition(prior, intent, policy)
			assertRejectedWithoutNextState(t, decision, tc.status, tc.reason)
		})
	}
}

func validJoinTransition(t *testing.T) (TransitionState, TransitionIntent, TransitionPolicy, *mapJoinResolver) {
	t.Helper()
	prior, intent, policy := leasedTransition(t)
	childIDs := []string{"run-child-a", "run-child-b"}
	prior.RequiredJoinRunIDs = append([]string(nil), childIDs...)
	resolver := &mapJoinResolver{children: make(map[string]JoinChild, len(childIDs))}
	fingerprints := make([]string, 0, len(childIDs))
	for index, runID := range childIDs {
		fingerprint := testDigest(fmt.Sprintf("terminal-contract-%d", index))
		fingerprints = append(fingerprints, fingerprint)
		resolver.children[runID] = JoinChild{
			RunID:               runID,
			ParentRunID:         prior.RunID,
			Status:              StatusClosed,
			ContractFingerprint: fingerprint,
		}
	}
	intent.Join = &TransitionJoin{
		RequiredRunIDs:               append([]string(nil), childIDs...),
		TerminalContractFingerprints: fingerprints,
		GroupedEvidenceFingerprint:   testDigest("grouped-evidence"),
	}
	policy.Joins = resolver
	return prior, intent, policy, resolver
}

type mapJoinResolver struct {
	children map[string]JoinChild
}

func (resolver *mapJoinResolver) Resolve(runID string) (JoinChild, bool, error) {
	child, found := resolver.children[runID]
	return child, found, nil
}
