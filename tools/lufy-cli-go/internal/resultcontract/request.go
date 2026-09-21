package resultcontract

import (
	"io"
	"reflect"
)

const TransitionRequestSchemaVersion = "lufy-result-transition-request/v1"

type TransitionRequest struct {
	SchemaVersion string           `json:"schema_version" yaml:"schema_version"`
	Prior         TransitionState  `json:"prior" yaml:"prior"`
	Intent        TransitionIntent `json:"intent" yaml:"intent"`
}

func DecodeTransitionRequest(input io.Reader) (TransitionRequest, error) {
	root, err := decodeDocument(input)
	if err != nil {
		return TransitionRequest{}, err
	}
	if err := validateNodeShape(root, reflect.TypeOf(TransitionRequest{}), ""); err != nil {
		return TransitionRequest{}, err
	}
	var request TransitionRequest
	if err := root.Decode(&request); err != nil {
		return TransitionRequest{}, diagnostic("transition_request_decode_failed", "$", "corregir tipos según el request tipado")
	}
	if request.SchemaVersion != TransitionRequestSchemaVersion {
		return TransitionRequest{}, diagnostic("unsupported_request_schema", "schema_version", "usar lufy-result-transition-request/v1")
	}
	if request.Prior.Version == 0 || !fingerprintPattern.MatchString(request.Prior.Fingerprint) || request.Prior.RunID == "" {
		return TransitionRequest{}, diagnostic("invalid_prior", "prior", "proveer version, fingerprint y run_id válidos")
	}
	if err := Validate(request.Prior.Contract); err != nil {
		return TransitionRequest{}, diagnostic("invalid_prior_contract", "prior.contract", "corregir el contrato previo")
	}
	priorCanonical, err := Canonicalize(request.Prior.Contract)
	if err != nil || priorCanonical.Fingerprint != request.Prior.Fingerprint {
		return TransitionRequest{}, diagnostic("prior_fingerprint_mismatch", "prior.fingerprint", "recalcular el fingerprint del contrato previo")
	}
	if err := Validate(request.Intent.NextContract); err != nil {
		return TransitionRequest{}, diagnostic("invalid_next_contract", "intent.next_contract", "corregir el contrato siguiente")
	}
	if request.Intent.SchemaVersion != TransitionSchemaVersion || !isValidTransitionIntent(request.Intent) {
		return TransitionRequest{}, diagnostic("invalid_transition", "intent", "completar el intent result-transition/v1")
	}
	if !fingerprintPattern.MatchString(request.Intent.Actor.OwnerRef) {
		return TransitionRequest{}, diagnostic("invalid_owner_ref", "intent.actor.owner_ref", "usar una referencia SHA-256 pseudonimizada")
	}
	if request.Intent.Lease != nil && (!fingerprintPattern.MatchString(request.Intent.Lease.TokenDigest) || request.Intent.Lease.ExpiresAt.IsZero()) {
		return TransitionRequest{}, diagnostic("invalid_lease", "intent.lease", "usar token digest y expiración válidos")
	}
	return request, nil
}
