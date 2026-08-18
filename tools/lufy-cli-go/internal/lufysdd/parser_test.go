package lufysdd

import "testing"

func TestParseDeltaAcceptsTestableRequirement(t *testing.T) {
	body := `# Demo

## ADDED Requirements

### Requirement: Audit log

#### Scenario: Entry is recorded

- **WHEN** an action succeeds
- **THEN** an entry is recorded
`
	delta, diagnostics := ParseDelta("spec.md", "audit-log", body)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(delta.Requirements) != 1 || delta.Requirements[0].Title != "Audit log" {
		t.Fatalf("delta = %#v", delta)
	}
}

func TestParseDeltaReportsMissingThenAndDuplicate(t *testing.T) {
	body := `## ADDED Requirements
### Requirement: Audit log
#### Scenario: Entry
- **WHEN** action succeeds
### Requirement: Audit log
`
	_, diagnostics := ParseDelta("spec.md", "audit-log", body)
	if !hasCode(diagnostics, "missing_then") || !hasCode(diagnostics, "duplicate_requirement") || !hasCode(diagnostics, "missing_scenario") {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestTaskProgress(t *testing.T) {
	progress := TaskProgress("- [x] done\n- [ ] pending\n- [X] done too\n")
	if progress.Complete != 2 || progress.Total != 3 {
		t.Fatalf("progress = %#v", progress)
	}
}

func hasCode(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
