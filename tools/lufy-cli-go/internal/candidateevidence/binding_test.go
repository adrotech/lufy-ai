package candidateevidence_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
)

const fixtureEvidenceSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestDecodeRejectsCaseAliasedFields(t *testing.T) {
	root, base := candidateFixture(t)
	body, err := json.Marshal(captureFixture(t, root, base))
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range [][]byte{
		append([]byte(`{"role":"reviewer",`), body[1:]...),
		append([]byte(`{"Role":"reviewer",`), body[1:]...),
	} {
		if _, err := candidateevidence.Decode(bytes.NewReader(invalid)); err == nil {
			t.Error("sidecar no debe aceptar aliases de casing ni claves semánticamente duplicadas")
		}
	}
}

func isolatedGitEnv() []string {
	var result []string
	for _, entry := range os.Environ() {
		key := strings.SplitN(entry, "=", 2)[0]
		if !strings.HasPrefix(key, "GIT_") {
			result = append(result, entry)
		}
	}
	return append(result, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=Synthetic", "GIT_AUTHOR_EMAIL=synthetic@example.invalid",
		"GIT_COMMITTER_NAME=Synthetic", "GIT_COMMITTER_EMAIL=synthetic@example.invalid",
		"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
}

func fixtureGit(t *testing.T, target string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", target, "-c", "maintenance.auto=false", "-c", "gc.auto=0"}, args...)...)
	command.Env = isolatedGitEnv()
	body, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("local fixture git failed: %v (%s)", err, body)
	}
	return strings.TrimSpace(string(body))
}

func fixtureWrite(t *testing.T, root, relative, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func candidateFixture(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git local necesario para fixtures")
	}
	root := t.TempDir()
	fixtureGit(t, root, "init", "--initial-branch=fixture", "--object-format=sha1")
	fixtureGit(t, root, "config", "maintenance.auto", "false")
	fixtureGit(t, root, "config", "gc.auto", "0")
	fixtureGit(t, root, "config", "core.filemode", "false")
	fixtureWrite(t, root, ".lufy/contracts/delivery.md", "Política sintética de delivery.\n")
	fixtureWrite(t, root, ".lufy/contracts/work.md", "Política sintética de contratos.\n")
	fixtureWrite(t, root, ".lufy/contracts/evidence.md", "Política sintética de evidencia.\n")
	fixtureWrite(t, root, "candidate.txt", "baseline\n")
	fixtureGit(t, root, "add", ".")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "baseline")
	base := fixtureGit(t, root, "rev-parse", "HEAD")
	fixtureWrite(t, root, "candidate.txt", "candidate\n")
	fixtureGit(t, root, "add", "candidate.txt")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "candidate")
	return root, base
}

func captureFixture(t *testing.T, root, base string) candidateevidence.Binding {
	t.Helper()
	binding, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA)
	if err != nil {
		t.Fatalf("candidateevidence.Capture clean candidate failed: %v", err)
	}
	return binding
}

func fingerprintFixtureTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result[relative] = "directory"
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			result[relative] = "symlink:" + link
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		result[relative] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCaptureVerifyCurrentIsPathFreeAndDoesNotMutateGit(t *testing.T) {
	root, base := candidateFixture(t)
	before := fingerprintFixtureTree(t, root)
	binding := captureFixture(t, root, base)
	if binding.SchemaVersion != "candidate-evidence/v2" || binding.BaseCommit != base || binding.EvidenceSHA256 != fixtureEvidenceSHA {
		t.Fatalf("candidate binding lost identity: %#v", binding)
	}
	if err := candidateevidence.Verify(context.Background(), root, base, fixtureEvidenceSHA, binding); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(binding)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte(root)) || bytes.Contains(body, []byte("candidate.txt")) || bytes.Contains(body, []byte("Política sintética")) {
		t.Fatal("sidecar must expose only hashes/IDs/role, not private paths or contents")
	}
	if !reflect.DeepEqual(before, fingerprintFixtureTree(t, root)) {
		t.Fatal("capture/verify changed worktree or Git internals")
	}
}

func TestCaptureRejectsTrackedAndUntrackedProductChanges(t *testing.T) {
	for _, path := range []string{"candidate.txt", "new-source.txt", ".lufy/contracts/new-policy.md", ".lufy/config/project.yaml.extra"} {
		t.Run(path, func(t *testing.T) {
			root, base := candidateFixture(t)
			fixtureWrite(t, root, path, "dirty content\n")
			if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA); err == nil {
				t.Fatal("dirty product state must be rejected, including internal prefixes")
			}
		})
	}
}

func TestVerifyRejectsNewCommitEvenWhenTreeIsIdentical(t *testing.T) {
	root, base := candidateFixture(t)
	binding := captureFixture(t, root, base)
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "--allow-empty", "-m", "same-tree-new-identity")
	if fixtureGit(t, root, "rev-parse", "HEAD^{tree}") != binding.CandidateTree {
		t.Fatal("fixture must keep tree unchanged")
	}
	if err := candidateevidence.Verify(context.Background(), root, base, fixtureEvidenceSHA, binding); err == nil {
		t.Fatal("new commit identity must invalidate even an identical tree")
	}
}

func TestVerifyRejectsDifferentBaseEvidenceRoleAndRepository(t *testing.T) {
	root, base := candidateFixture(t)
	binding := captureFixture(t, root, base)
	for _, changed := range []struct{ base, evidence, role string }{
		{binding.CandidateCommit, fixtureEvidenceSHA, "reviewer"},
		{base, strings.Repeat("b", 64), "reviewer"},
	} {
		if err := candidateevidence.Verify(context.Background(), root, changed.base, changed.evidence, binding); err == nil {
			t.Fatal("changed binding inputs must reject")
		}
	}
	clone := t.TempDir()
	fixtureGit(t, root, "clone", "--local", "--no-hardlinks", root, clone)
	if err := candidateevidence.Verify(context.Background(), clone, base, fixtureEvidenceSHA, binding); err == nil {
		t.Fatal("a distinct repository must not reuse binding even with identical commits")
	}
}

func TestUntrackedExactConfigIsBoundButTrackedModifiedConfigIsDirty(t *testing.T) {
	root, base := candidateFixture(t)
	config := ".lufy/config/project.yaml"
	fixtureWrite(t, root, config, "schema_version: 2\ntool: codex\n")
	binding := captureFixture(t, root, base)
	fixtureWrite(t, root, config, "schema_version: 2\ntool: codex\n# changed\n")
	if err := candidateevidence.Verify(context.Background(), root, base, fixtureEvidenceSHA, binding); err == nil {
		t.Fatal("untracked exact config must be hash-bound")
	}
	fixtureGit(t, root, "add", config)
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "commit-config")
	fixtureWrite(t, root, config, "schema_version: 2\n# dirty tracked config\n")
	if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA); err == nil {
		t.Fatal("tracked config modifications are not the untracked exception")
	}
}

func TestPolicyAndCommittedFileModesAffectBinding(t *testing.T) {
	root, base := candidateFixture(t)
	initial := captureFixture(t, root, base)
	fixtureWrite(t, root, ".lufy/contracts/delivery.md", "Política modificada.\n")
	fixtureGit(t, root, "add", ".lufy/contracts/delivery.md")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "policy-change")
	changed := captureFixture(t, root, base)
	if changed.PolicySHA256 == initial.PolicySHA256 {
		t.Fatal("policy content change must affect policy hash")
	}
	if err := candidateevidence.Verify(context.Background(), root, base, fixtureEvidenceSHA, initial); err == nil {
		t.Fatal("old policy binding must reject")
	}
	fixtureGit(t, root, "update-index", "--chmod=+x", "candidate.txt")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "mode-change")
	mode := captureFixture(t, root, base)
	if mode.ChangedPathsModesSHA256 == changed.ChangedPathsModesSHA256 {
		t.Fatal("committed executable mode change must affect path/mode diff fingerprint")
	}
}

func TestCaptureRejectsMissingNonRegularAndSymlinkPolicies(t *testing.T) {
	for _, kind := range []string{"missing", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			root, base := candidateFixture(t)
			path := filepath.Join(root, ".lufy/contracts/delivery.md")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if kind == "directory" {
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
				fixtureWrite(t, root, ".lufy/contracts/delivery.md/file.txt", "not a regular policy\n")
			}
			if kind == "symlink" {
				if err := os.Symlink("work.md", path); err != nil {
					t.Skip("symlink unsupported in local fixture")
				}
			}
			fixtureGit(t, root, "add", "-A")
			fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "invalid-policy")
			if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA); err == nil {
				t.Fatal("required policy must be a regular non-symlink file")
			}
		})
	}
}

func TestGitEnvironmentOverridesCannotRedirectCaptureTarget(t *testing.T) {
	root, base := candidateFixture(t)
	foreign, _ := candidateFixture(t)
	expected := captureFixture(t, root, base)
	t.Setenv("GIT_DIR", filepath.Join(foreign, ".git"))
	t.Setenv("GIT_WORK_TREE", foreign)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(foreign, ".git", "index"))
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.worktree")
	t.Setenv("GIT_CONFIG_VALUE_0", foreign)
	actual := captureFixture(t, root, base)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatal("environment Git overrides redirected the target")
	}
}

func TestDecodeAcceptsCurrentBindingAndRejectsAmbiguousJSON(t *testing.T) {
	root, base := candidateFixture(t)
	binding := captureFixture(t, root, base)
	body, err := json.Marshal(binding)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := candidateevidence.Decode(bytes.NewReader(body))
	if err != nil || !reflect.DeepEqual(binding, decoded) {
		t.Fatalf("valid sidecar did not roundtrip: %v", err)
	}
	bad := [][]byte{
		[]byte("null"), []byte("[]"), []byte("{}"), []byte("{malformed"),
		append(append([]byte(nil), body...), []byte(" {}")...),
		append(append([]byte(nil), body...), []byte(" trailing")...),
		append([]byte(`{"unknown":true,`), body[1:]...),
		append([]byte(`{"schema_version":"candidate-evidence/v2",`), body[1:]...),
		append(append([]byte(nil), body...), bytes.Repeat([]byte(" "), 65537)...),
	}
	for index, invalid := range bad {
		if _, err := candidateevidence.Decode(bytes.NewReader(invalid)); err == nil {
			t.Fatalf("malformed/unknown/duplicate/trailing/oversized sidecar accepted: case %d", index)
		}
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for key, invalid := range map[string]any{
		"schema_version": "candidate-evidence/v999", "repository_sha256": "../../private",
		"base_commit": "not-a-commit", "candidate_tree": strings.Repeat("A", 40),
		"evidence_sha256": true, "policy_sha256": "", "role": "Reviewer With Spaces",
	} {
		original := fields[key]
		fields[key] = invalid
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := candidateevidence.Decode(bytes.NewReader(encoded)); err == nil {
			t.Fatalf("invalid field accepted: %s", key)
		}
		fields[key] = original
	}
	for key, original := range fields {
		fields[key] = nil
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := candidateevidence.Decode(bytes.NewReader(encoded)); err == nil {
			t.Fatalf("null required field accepted: %s", key)
		}
		fields[key] = original
	}

}

func TestCaptureRequiresTrackedCanonicalPoliciesEvenWhenIgnored(t *testing.T) {
	root, base := candidateFixture(t)
	fixtureGit(t, root, "rm", "--cached", "--", ".lufy/contracts/delivery.md")
	fixtureWrite(t, root, ".gitignore", "/.lufy/contracts/delivery.md\n")
	fixtureGit(t, root, "add", ".gitignore")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "ignore-untracked-policy")
	if fixtureGit(t, root, "status", "--porcelain") != "" {
		t.Fatal("fixture must be clean despite ignored policy")
	}
	if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA); err == nil {
		t.Fatal("required canonical policy must be tracked; ignored local policy cannot supply committed authority")
	}
}

func TestCaptureRejectsProductEditsHiddenByIndexFlags(t *testing.T) {
	for _, flag := range []string{"--assume-unchanged", "--skip-worktree"} {
		t.Run(flag, func(t *testing.T) {
			root, base := candidateFixture(t)
			fixtureGit(t, root, "update-index", flag, "--", "candidate.txt")
			fixtureWrite(t, root, "candidate.txt", "dirty product content hidden by index flags\n")
			if fixtureGit(t, root, "status", "--porcelain") != "" {
				t.Fatal("fixture must demonstrate status hiding modified source")
			}
			if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA); err == nil {
				t.Fatal("hidden product edits must not be attested as a clean committed candidate")
			}
		})
	}
}
