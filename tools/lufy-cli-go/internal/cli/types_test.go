package cli

import "testing"

func TestActionableErrorError(t *testing.T) {
	if got := (ActionableError{Message: "fallo"}).Error(); got != "fallo" {
		t.Fatalf("unexpected message-only error: %q", got)
	}
	got := (ActionableError{Message: "fallo", Hint: "haz algo"}).Error()
	if got != "fallo\nSugerencia: haz algo" {
		t.Fatalf("unexpected actionable error: %q", got)
	}
}
