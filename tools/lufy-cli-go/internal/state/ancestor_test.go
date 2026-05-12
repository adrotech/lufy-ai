package state

import "testing"

func TestAncestorRelMapsSafePath(t *testing.T) {
	got, err := AncestorRel(".opencode/agents/orchestrator.md")
	if err != nil {
		t.Fatalf("AncestorRel() error = %v", err)
	}
	want := ".lufy-ai/ancestors/.opencode/agents/orchestrator.md"
	if got != want {
		t.Fatalf("AncestorRel() = %q, want %q", got, want)
	}
}

func TestAncestorRelRejectsTraversal(t *testing.T) {
	if _, err := AncestorRel("../evil.md"); err == nil {
		t.Fatal("AncestorRel() expected traversal error")
	}
	if _, err := AncestorRel(`..\\evil.md`); err == nil {
		t.Fatal("AncestorRel() expected backslash traversal error")
	}
}
