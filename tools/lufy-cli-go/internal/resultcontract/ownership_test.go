package resultcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var sliceCFixedNow = time.Date(2026, 9, 20, 15, 0, 0, 0, time.UTC)

func TestEvaluateTransitionAcceptsMatchingOwnerAndActiveLease(t *testing.T) {
	t.Parallel()

	prior, intent, policy := leasedTransition(t)
	decision := EvaluateTransition(prior, intent, policy)
	if decision.Status != DecisionAccepted {
		t.Fatalf("active matching lease decision = %#v, want accepted", decision)
	}
	if decision.NextVersion != prior.Version+1 {
		t.Fatalf("NextVersion = %d, want %d", decision.NextVersion, prior.Version+1)
	}
}

func TestEvaluateTransitionRejectsOwnerOrRoleMismatchWithoutNextState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*TransitionIntent)
		reason string
	}{
		{
			name: "owner mismatch",
			mutate: func(intent *TransitionIntent) {
				intent.Actor.OwnerRef = testDigest("another-owner")
			},
			reason: "owner_mismatch",
		},
		{
			name: "role mismatch",
			mutate: func(intent *TransitionIntent) {
				intent.Actor.Role = "delivery"
			},
			reason: "owner_role_mismatch",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy := leasedTransition(t)
			tc.mutate(&intent)
			decision := EvaluateTransition(prior, intent, policy)
			assertRejectedWithoutNextState(t, decision, DecisionConflict, tc.reason)
		})
	}
}

func TestEvaluateTransitionValidatesLeaseExpiryAndTokenDigest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*TransitionState, *TransitionIntent)
		status DecisionStatus
		reason string
	}{
		{
			name: "expired lease",
			mutate: func(prior *TransitionState, intent *TransitionIntent) {
				prior.Ownership.Lease.ExpiresAt = sliceCFixedNow.Add(-time.Second)
				intent.Lease.ExpiresAt = prior.Ownership.Lease.ExpiresAt
			},
			status: DecisionRejected,
			reason: "lease_expired",
		},
		{
			name: "lease digest mismatch",
			mutate: func(_ *TransitionState, intent *TransitionIntent) {
				intent.Lease.TokenDigest = testDigest("another-token")
			},
			status: DecisionConflict,
			reason: "lease_token_mismatch",
		},
		{
			name: "lease expiry mismatch",
			mutate: func(_ *TransitionState, intent *TransitionIntent) {
				intent.Lease.ExpiresAt = intent.Lease.ExpiresAt.Add(time.Minute)
			},
			status: DecisionConflict,
			reason: "lease_mismatch",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy := leasedTransition(t)
			tc.mutate(&prior, &intent)
			decision := EvaluateTransition(prior, intent, policy)
			assertRejectedWithoutNextState(t, decision, tc.status, tc.reason)
		})
	}
}

func TestExpectedVersionActsAsFencingTokenEvenWithFormerValidLease(t *testing.T) {
	t.Parallel()

	prior, intent, policy := leasedTransition(t)
	intent.ExpectedVersion = prior.Version - 1
	decision := EvaluateTransition(prior, intent, policy)
	assertRejectedWithoutNextState(t, decision, DecisionConflict, "stale_version")
}

func TestOwnershipDiagnosticsDoNotLeakRawRefsOrTokens(t *testing.T) {
	t.Parallel()

	const rawOwner = "owner@example.invalid::PRIVATE_OWNER_CANARY"
	const rawToken = "PRIVATE_LEASE_TOKEN_CANARY"
	cases := []struct {
		name   string
		mutate func(*TransitionState, *TransitionIntent)
		reason string
	}{
		{
			name: "raw owner ref",
			mutate: func(prior *TransitionState, intent *TransitionIntent) {
				prior.Ownership.Actor.OwnerRef = rawOwner
				intent.Actor.OwnerRef = rawOwner
			},
			reason: "invalid_owner_ref",
		},
		{
			name: "raw lease token",
			mutate: func(prior *TransitionState, intent *TransitionIntent) {
				prior.Ownership.Lease.TokenDigest = rawToken
				intent.Lease.TokenDigest = rawToken
			},
			reason: "invalid_lease_token",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior, intent, policy := leasedTransition(t)
			tc.mutate(&prior, &intent)
			decision := EvaluateTransition(prior, intent, policy)
			if decision.Status != DecisionRejected || decision.Reason != tc.reason {
				t.Fatalf("raw refs decision = %#v, want rejected/%s", decision, tc.reason)
			}
			body, err := json.Marshal(decision)
			if err != nil {
				t.Fatal(err)
			}
			observed := string(body) + decision.Recovery + decision.Reason
			for _, canary := range []string{rawOwner, rawToken} {
				if strings.Contains(observed, canary) {
					t.Fatalf("decision leaked raw ownership canary %q: %s", canary, observed)
				}
			}
		})
	}
}

func leasedTransition(t *testing.T) (TransitionState, TransitionIntent, TransitionPolicy) {
	t.Helper()
	prior := transitionState(t, StatusImplemented, 11)
	prior.RunID = "run-parent-slice-c"
	prior.Ownership = &TransitionOwnership{
		Actor: TransitionActor{Role: "validator", OwnerRef: testDigest("owner-validator")},
		Lease: TransitionLease{TokenDigest: testDigest("lease-token"), ExpiresAt: sliceCFixedNow.Add(5 * time.Minute)},
	}
	next := contractWithStatus(t, StatusValidated)
	next.Evidence.Commands[0].Result = EvidencePassed
	intent := validTransitionIntent(prior, next, "validator")
	intent.Actor = prior.Ownership.Actor
	lease := prior.Ownership.Lease
	intent.Lease = &lease
	policy := transitionTestPolicy()
	policy.Now = func() time.Time { return sliceCFixedNow }
	return prior, intent, policy
}

func assertRejectedWithoutNextState(t *testing.T, decision TransitionDecision, status DecisionStatus, reason string) {
	t.Helper()
	if decision.Status != status || decision.Reason != reason {
		t.Fatalf("decision = %#v, want %s/%s", decision, status, reason)
	}
	if decision.NextVersion != 0 || decision.NextFingerprint != "" || decision.Recovery == "" {
		t.Fatalf("rejected decision exposed next state or lacks recovery: %#v", decision)
	}
}

func testDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
