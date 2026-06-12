package agentsref

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMinimalContentAndRecommendedAction(t *testing.T) {
	body := string(MinimalContent())
	if !strings.Contains(body, "# AGENTS.md") || !strings.Contains(body, BeginMarker) || !strings.Contains(body, "Lufy AI Harness") {
		t.Fatalf("minimal content missing expected managed block: %q", body)
	}

	action := RecommendedInstallAction()
	if !strings.Contains(action, "lufy-ai install") || !strings.Contains(action, "bloque gestionado") {
		t.Fatalf("recommended action missing command/managed block: %q", action)
	}
}

func TestContainsReference(t *testing.T) {
	if !ContainsReference([]byte("before " + Reference + " after")) {
		t.Fatalf("expected legacy reference to be detected")
	}
	if !ContainsReference([]byte(ManagedBlock())) {
		t.Fatalf("expected managed block to be detected")
	}
	if ContainsReference([]byte("no harness reference")) {
		t.Fatalf("unexpected integration detected")
	}
}

func TestStatusMissingPresentAndUnsafe(t *testing.T) {
	target := t.TempDir()

	exists, hasReference, err := Status(target)
	if err != nil || exists || hasReference {
		t.Fatalf("missing status = exists:%v ref:%v err:%v", exists, hasReference, err)
	}

	if err := os.WriteFile(filepath.Join(target, AgentsFile), []byte("# Team\n"+Reference+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	exists, hasReference, err = Status(target)
	if err != nil || !exists || !hasReference {
		t.Fatalf("present status = exists:%v ref:%v err:%v", exists, hasReference, err)
	}

	unsafeRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(unsafeRoot, AgentsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	exists, hasReference, err = Status(unsafeRoot)
	if err == nil || !exists || hasReference {
		t.Fatalf("unsafe status = exists:%v ref:%v err:%v", exists, hasReference, err)
	}
}

func TestInsertReferenceCreatesAppendsAndIsIdempotent(t *testing.T) {
	target := t.TempDir()

	if err := InsertReference(target); err != nil {
		t.Fatalf("insert missing AGENTS.md: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(target, AgentsFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(MinimalContent()) {
		t.Fatalf("unexpected created body: %q", string(body))
	}

	if err := InsertReference(target); err != nil {
		t.Fatalf("idempotent insert: %v", err)
	}
	again, err := os.ReadFile(filepath.Join(target, AgentsFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(body) {
		t.Fatalf("insert was not idempotent: %q", string(again))
	}

	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(other, AgentsFile), []byte("# Existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := InsertReference(other); err != nil {
		t.Fatalf("append managed block: %v", err)
	}
	appended, err := os.ReadFile(filepath.Join(other, AgentsFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(appended) != "# Existing\n\n"+ManagedBlock() {
		t.Fatalf("unexpected appended body: %q", string(appended))
	}
}

func TestInsertReferenceRejectsSymlink(t *testing.T) {
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, AgentsFile)); err != nil {
		t.Skipf("symlink not available: %v", err)
	}
	if err := InsertReference(target); err == nil {
		t.Fatalf("expected symlink target to be rejected")
	}
}

func TestRemoveReferencePreservesUserContent(t *testing.T) {
	target := t.TempDir()
	original := "# Existing\n\nKeep this\n" + ManagedBlock() + "\n" + Reference + "\n\nKeep after\n"
	if err := os.WriteFile(filepath.Join(target, AgentsFile), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := RemoveReference(target)
	if err != nil {
		t.Fatalf("RemoveReference() error = %v", err)
	}
	if !changed {
		t.Fatal("RemoveReference() changed = false")
	}
	body, err := os.ReadFile(filepath.Join(target, AgentsFile))
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if strings.Contains(got, Reference) || strings.Contains(got, BeginMarker) || strings.Contains(got, EndMarker) {
		t.Fatalf("integration remained: %q", got)
	}
	for _, want := range []string{"# Existing", "Keep this", "Keep after"} {
		if !strings.Contains(got, want) {
			t.Fatalf("user content %q missing from %q", want, got)
		}
	}

	changed, err = RemoveReference(target)
	if err != nil {
		t.Fatalf("idempotent RemoveReference() error = %v", err)
	}
	if changed {
		t.Fatal("idempotent RemoveReference() changed = true")
	}
}
