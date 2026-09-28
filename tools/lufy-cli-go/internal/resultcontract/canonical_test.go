package resultcontract

import (
	"bytes"
	"strings"
	"testing"
)

func TestCanonicalIdentityIgnoresMapOrderAndLineEndings(t *testing.T) {
	t.Parallel()

	base := readFixture(t, "testdata/valid-minimal.yaml")
	reordered := readFixture(t, "testdata/valid-minimal-reordered.yaml")
	crlf := []byte(strings.ReplaceAll(string(base), "\n", "\r\n"))

	inputs := map[string][]byte{
		"different key order": reordered,
		"CRLF":                crlf,
	}
	baseline, err := Decode(bytes.NewReader(base))
	if err != nil {
		t.Fatalf("Decode(baseline) error = %v", err)
	}
	want := mustCanonicalize(t, baseline)
	for name, input := range inputs {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			contract, err := Decode(bytes.NewReader(input))
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			got := mustCanonicalize(t, contract)
			if !bytes.Equal(got.JSON, want.JSON) {
				t.Fatalf("canonical JSON differs:\ngot:  %s\nwant: %s", got.JSON, want.JSON)
			}
			if got.Fingerprint != want.Fingerprint {
				t.Fatalf("fingerprint = %q, want %q", got.Fingerprint, want.Fingerprint)
			}
		})
	}
}

func TestCanonicalFingerprintChangesWhenMeaningChanges(t *testing.T) {
	t.Parallel()

	contract := decodeFixture(t, "testdata/valid-minimal.yaml")
	before := mustCanonicalize(t, contract)

	contract.Status = StatusImplemented
	afterStatus := mustCanonicalize(t, contract)
	if before.Fingerprint == afterStatus.Fingerprint {
		t.Fatal("semantic status change did not change fingerprint")
	}

	contract = decodeFixture(t, "testdata/valid-minimal.yaml")
	contract.Evidence.Commands[0].Result = EvidencePassed
	afterEvidence := mustCanonicalize(t, contract)
	if before.Fingerprint == afterEvidence.Fingerprint {
		t.Fatal("semantic evidence result change did not change fingerprint")
	}
}
