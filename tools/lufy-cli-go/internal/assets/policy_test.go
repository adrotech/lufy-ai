package assets

import "testing"

func TestPolicyAndScopeHelpers(t *testing.T) {
	for _, policy := range []Policy{PolicyManaged, PolicyNoReplace, PolicyMergeBlock, PolicyMergeJSON, PolicyMetadata} {
		if !policy.Valid() {
			t.Fatalf("policy %q should be valid", policy)
		}
	}
	if Policy("unknown").Valid() {
		t.Fatalf("unknown policy should be invalid")
	}

	for _, policy := range []Policy{PolicyManaged, PolicyNoReplace, PolicyMergeBlock} {
		if !policy.SupportsAncestor() {
			t.Fatalf("policy %q should support ancestors", policy)
		}
	}
	for _, policy := range []Policy{PolicyMergeJSON, PolicyMetadata, Policy("unknown")} {
		if policy.SupportsAncestor() {
			t.Fatalf("policy %q should not support ancestors", policy)
		}
	}

	for _, scope := range []Scope{ScopeProject, ScopeGlobal, ScopeBoth} {
		if !scope.Valid() {
			t.Fatalf("scope %q should be valid", scope)
		}
		parsed, err := ParseScope(string(scope))
		if err != nil || parsed != scope {
			t.Fatalf("ParseScope(%q) = %q, %v", scope, parsed, err)
		}
	}
	parsed, err := ParseScope("")
	if err != nil || parsed != ScopeProject {
		t.Fatalf("empty scope = %q, %v", parsed, err)
	}
	if _, err := ParseScope("invalid"); err == nil {
		t.Fatalf("invalid scope should fail")
	}
}
