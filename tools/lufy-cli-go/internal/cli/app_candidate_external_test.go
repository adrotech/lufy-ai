package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/cli"
)

func candidateCLIGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.SplitN(entry, "=", 2)[0], "GIT_") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=Synthetic", "GIT_AUTHOR_EMAIL=synthetic@example.invalid",
		"GIT_COMMITTER_NAME=Synthetic", "GIT_COMMITTER_EMAIL=synthetic@example.invalid")
	body, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("fixture Git failed: %v (%s)", err, body)
	}
	return strings.TrimSpace(string(body))
}

func candidateCLIWrite(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
}

func candidateCLIFixture(t *testing.T) (string, string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git local requerido")
	}
	root := t.TempDir()
	candidateCLIGit(t, root, "init", "--initial-branch=fixture", "--object-format=sha1")
	for _, name := range []string{"delivery.md", "result-contract.md"} {
		candidateCLIWrite(t, filepath.Join(root, ".lufy/contracts", name), []byte("Política sintética.\n"))
	}
	candidateCLIWrite(t, filepath.Join(root, "product.txt"), []byte("baseline\n"))
	candidateCLIGit(t, root, "add", ".")
	candidateCLIGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "baseline")
	base := candidateCLIGit(t, root, "rev-parse", "HEAD")
	candidateCLIWrite(t, filepath.Join(root, "product.txt"), []byte("candidate\n"))
	candidateCLIGit(t, root, "add", "product.txt")
	candidateCLIGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "candidate")
	body, err := os.ReadFile(filepath.Join("..", "resultcontract", "testdata", "valid-minimal.json"))
	if err != nil {
		t.Fatal(err)
	}
	evidence := filepath.Join(t.TempDir(), "evidence.json")
	candidateCLIWrite(t, evidence, body)
	return root, base, evidence
}

func candidateCLIInvoke(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, cli.Dependencies{Stdout: &stdout, Stderr: &stderr})
	return code, stdout.String(), stderr.String()
}

func candidateFileSHA(t *testing.T, path string) [32]byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(body)
}

func TestCandidateCLICaptureVerifyJSONAndRejectStaleCommitWithoutRewritingEvidence(t *testing.T) {
	root, base, evidence := candidateCLIFixture(t)
	evidenceSHA := candidateFileSHA(t, evidence)
	code, stdout, stderr := candidateCLIInvoke("candidate", "capture", "--target", root, "--base", base, "--evidence", evidence, "--role", "reviewer", "--json")
	if code != cli.ExitOK || !json.Valid([]byte(stdout)) {
		t.Fatalf("capture failed: %d %s", code, stderr)
	}
	var binding map[string]any
	if err := json.Unmarshal([]byte(stdout), &binding); err != nil {
		t.Fatal(err)
	}
	if binding["schema_version"] != "candidate-evidence/v1" {
		t.Fatal("capture schema missing")
	}
	if strings.Contains(stdout, root) || strings.Contains(stdout, evidence) {
		t.Fatal("capture leaked private paths")
	}
	path := filepath.Join(t.TempDir(), "binding.json")
	candidateCLIWrite(t, path, []byte(stdout))
	bindingSHA := candidateFileSHA(t, path)
	verify := []string{"candidate", "verify", "--target", root, "--base", base, "--evidence", evidence, "--role", "reviewer", "--binding", path, "--json"}
	code, stdout, stderr = candidateCLIInvoke(verify...)
	if code != cli.ExitOK {
		t.Fatalf("current binding rejected: %d %s", code, stderr)
	}
	var current map[string]string
	if err := json.Unmarshal([]byte(stdout), &current); err != nil {
		t.Fatal(err)
	}
	if current["status"] != "current" || current["schema_version"] != "candidate-evidence-verification/v1" {
		t.Fatal("verify JSON must identify current binding")
	}
	candidateCLIGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "--allow-empty", "-m", "same-tree-new-commit")
	code, _, _ = candidateCLIInvoke(verify...)
	if code != cli.ExitResultRejected {
		t.Fatal("new committed identity must reject the stale sidecar")
	}
	if evidenceSHA != candidateFileSHA(t, evidence) || bindingSHA != candidateFileSHA(t, path) {
		t.Fatal("capture/verify must never rewrite evidence or binding files")
	}
}

func TestCandidateCLIRejectsUnknownRoleAndMalformedEvidenceWithoutLeakingBody(t *testing.T) {
	root, base, evidence := candidateCLIFixture(t)
	evidenceSHA := candidateFileSHA(t, evidence)
	code, _, _ := candidateCLIInvoke("candidate", "capture", "--target", root, "--base", base, "--evidence", evidence, "--role", "unregistered-role", "--json")
	if code != cli.ExitResultRejected {
		t.Fatal("unregistered claim role must not be accepted")
	}
	canary := "PRIVATE CANDIDATE EVIDENCE CANARY"
	malformed := filepath.Join(t.TempDir(), "invalid.txt")
	candidateCLIWrite(t, malformed, []byte(canary))
	malformedSHA := candidateFileSHA(t, malformed)
	code, stdout, stderr := candidateCLIInvoke("candidate", "capture", "--target", root, "--base", base, "--evidence", malformed, "--role", "reviewer", "--json")
	if code != cli.ExitResultRejected || strings.Contains(stdout+stderr, canary) {
		t.Fatal("malformed evidence must reject without disclosing input")
	}
	if evidenceSHA != candidateFileSHA(t, evidence) || malformedSHA != candidateFileSHA(t, malformed) {
		t.Fatal("rejection rewrote an evidence file")
	}
}
