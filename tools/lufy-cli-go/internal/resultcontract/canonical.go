package resultcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Canonical struct {
	JSON        []byte
	Fingerprint string
}

func Canonicalize(contract Contract) (Canonical, error) {
	if err := Validate(contract); err != nil {
		return Canonical{}, err
	}
	body, err := json.Marshal(contract)
	if err != nil {
		return Canonical{}, diagnostic("canonicalization_failed", "$", "corregir el contrato tipado")
	}
	digest := sha256.Sum256(body)
	return Canonical{JSON: body, Fingerprint: hex.EncodeToString(digest[:])}, nil
}
