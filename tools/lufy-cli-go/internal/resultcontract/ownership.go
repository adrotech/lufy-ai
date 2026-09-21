package resultcontract

import "time"

type TransitionLease struct {
	TokenDigest string    `json:"token_digest" yaml:"token_digest"`
	ExpiresAt   time.Time `json:"expires_at" yaml:"expires_at"`
}

type TransitionOwnership struct {
	Actor TransitionActor `json:"actor" yaml:"actor"`
	Lease TransitionLease `json:"lease" yaml:"lease"`
}

func evaluateOwnership(previous TransitionState, intent TransitionIntent, now time.Time) *TransitionDecision {
	if previous.Ownership == nil {
		if intent.Lease != nil {
			decision := rejectedTransition("ownership_unavailable", "recargar ownership antes de usar un lease")
			return &decision
		}
		return nil
	}
	ownership := previous.Ownership
	if !fingerprintPattern.MatchString(ownership.Actor.OwnerRef) || !fingerprintPattern.MatchString(intent.Actor.OwnerRef) {
		decision := rejectedTransition("invalid_owner_ref", "usar una referencia de owner pseudonimizada")
		return &decision
	}
	if !fingerprintPattern.MatchString(ownership.Lease.TokenDigest) || intent.Lease == nil || !fingerprintPattern.MatchString(intent.Lease.TokenDigest) {
		decision := rejectedTransition("invalid_lease_token", "usar solamente el digest SHA-256 del lease token")
		return &decision
	}
	if ownership.Actor.OwnerRef != intent.Actor.OwnerRef {
		decision := conflictTransition("owner_mismatch", "recargar el owner actual antes de reintentar")
		return &decision
	}
	if ownership.Actor.Role != intent.Actor.Role {
		decision := conflictTransition("owner_role_mismatch", "usar el rol del owner actual")
		return &decision
	}
	if ownership.Lease.TokenDigest != intent.Lease.TokenDigest {
		decision := conflictTransition("lease_token_mismatch", "recargar el lease actual")
		return &decision
	}
	if ownership.Lease.ExpiresAt.IsZero() || intent.Lease.ExpiresAt.IsZero() || !ownership.Lease.ExpiresAt.Equal(intent.Lease.ExpiresAt) {
		decision := conflictTransition("lease_mismatch", "recargar la expiración del lease actual")
		return &decision
	}
	if !now.Before(ownership.Lease.ExpiresAt) {
		decision := rejectedTransition("lease_expired", "adquirir un lease nuevo sobre la versión actual")
		return &decision
	}
	return nil
}
