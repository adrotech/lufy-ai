package agentsref

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMinimalContentAndRecommendedAction(t *testing.T) {
	body := string(MinimalContent())
	if !strings.Contains(body, "# AGENTS.md") || !strings.Contains(body, Reference) {
		t.Fatalf("minimal content missing expected reference: %q", body)
	}

	action := RecommendedInstallAction()
	if !strings.Contains(action, "lufy-ai install") || !strings.Contains(action, Reference) {
		t.Fatalf("recommended action missing command/reference: %q", action)
	}
}

func TestContainsReference(t *testing.T) {
	if !ContainsReference([]byte("before " + Reference + " after")) {
		t.Fatalf("expected reference to be detected")
	}
	if ContainsReference([]byte("no harness reference")) {
		t.Fatalf("unexpected reference detected")
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
		t.Fatalf("append reference: %v", err)
	}
	appended, err := os.ReadFile(filepath.Join(other, AgentsFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(appended) != "# Existing\n\n"+Reference+"\n" {
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
