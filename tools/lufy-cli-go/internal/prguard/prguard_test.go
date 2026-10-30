package prguard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardDetectsTrackedIgnoredFileInDiff(t *testing.T) {
	repo := initRepo(t)
	writeFile(t, filepath.Join(repo, ".gitignore"), ".lufy/private/\n")
	writeFile(t, filepath.Join(repo, "README.md"), "base\n")
	git(t, repo, "add", ".gitignore", "README.md")
	git(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))

	writeFile(t, filepath.Join(repo, ".lufy", "private", "changes", "demo", "proposal.md"), "ignored but tracked\n")
	git(t, repo, "add", "-f", ".lufy/private/changes/demo/proposal.md")
	git(t, repo, "commit", "-m", "add ignored openspec file")

	report, err := NewService().Build(Options{Target: repo, Base: base})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if report.OK {
		t.Fatalf("guard should fail: %#v", report)
	}
	if len(report.IgnoredMatches) != 1 || report.IgnoredMatches[0].Pattern != ".lufy/private/" || report.IgnoredMatches[0].Path != ".lufy/private/changes/demo/proposal.md" {
		t.Fatalf("ignored match unexpected: %#v", report.IgnoredMatches)
	}
	if len(report.InternalMatches) != 1 || report.InternalMatches[0].Pattern != ".lufy/" {
		t.Fatalf("internal match unexpected: %#v", report.InternalMatches)
	}

	var out bytes.Buffer
	err = NewService().Run(Options{Target: repo, Base: base}, &out)
	if err == nil {
		t.Fatalf("Run() expected error, output=%s", out.String())
	}
	for _, want := range []string{"git check-ignore -v --no-index --stdin", ".lufy/private/", ".gitignore", "cherry-pick", "git rm --cached"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("Run() output missing %q: %s", want, out.String())
		}
	}
}

func TestGuardPassesWithoutIgnoredOrInternalPaths(t *testing.T) {
	repo := initRepo(t)
	writeFile(t, filepath.Join(repo, ".gitignore"), ".lufy/private/\n")
	writeFile(t, filepath.Join(repo, "README.md"), "base\n")
	git(t, repo, "add", ".gitignore", "README.md")
	git(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))

	writeFile(t, filepath.Join(repo, "src", "app.go"), "package app\n")
	git(t, repo, "add", "src/app.go")
	git(t, repo, "commit", "-m", "add app")

	report, err := NewService().Build(Options{Target: repo, Base: base})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !report.OK || len(report.ChangedFiles) != 1 || report.ChangedFiles[0] != "src/app.go" {
		t.Fatalf("guard should pass with src/app.go: %#v", report)
	}
}

func TestInternalFilesAllowsOnlyCurrentPublicLufyArtifacts(t *testing.T) {
	files := []string{".lufy/README.md", ".lufy/contracts/work.md", ".lufy/managed-state/install-state.json", ".lufy/context/graph.json"}
	matches := internalFiles(files)
	if len(matches) != 2 || matches[0].Path != files[2] || matches[1].Path != files[3] {
		t.Fatalf("matches: %#v", matches)
	}
	if got := internalFileEntries([]changedFile{{Path: files[2], Status: "D"}}); len(got) != 0 {
		t.Fatalf("deletion blocked: %#v", got)
	}
}

func TestGuardStillBlocksExplicitlyIgnoredPublicContract(t *testing.T) {
	repo := initRepo(t)
	writeFile(t, filepath.Join(repo, ".gitignore"), ".lufy/\n")
	writeFile(t, filepath.Join(repo, "README.md"), "base\n")
	git(t, repo, "add", ".gitignore", "README.md")
	git(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))

	path := filepath.Join(repo, ".lufy", "contracts", "work.md")
	writeFile(t, path, "# Proposal\n")
	git(t, repo, "add", "-f", ".lufy/contracts/work.md")
	git(t, repo, "commit", "-m", "add ignored contract")

	report, err := NewService().Build(Options{Target: repo, Base: base})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK || len(report.IgnoredMatches) != 1 || len(report.InternalMatches) != 0 {
		t.Fatalf("explicit ignore must block without internal false positive: %#v", report)
	}
}

func TestGuardIncludeWorktreeDetectsPendingIgnoredFile(t *testing.T) {
	repo := initRepo(t)
	writeFile(t, filepath.Join(repo, ".gitignore"), ".lufy/private/\n")
	writeFile(t, filepath.Join(repo, "README.md"), "base\n")
	git(t, repo, "add", ".gitignore", "README.md")
	git(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))

	writeFile(t, filepath.Join(repo, ".lufy", "private", "changes", "demo", "tasks.md"), "pending ignored\n")
	git(t, repo, "add", "-f", ".lufy/private/changes/demo/tasks.md")

	report, err := NewService().Build(Options{Target: repo, Base: base, IncludeWorktree: true})
	if err != nil {
		t.Fatalf("Build(include worktree) error = %v", err)
	}
	if report.OK || report.Range != base || len(report.IgnoredMatches) != 1 {
		t.Fatalf("include worktree should detect pending ignored file: %#v", report)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	git(t, repo, "init")
	git(t, repo, "config", "user.email", "test@example.com")
	git(t, repo, "config", "user.name", "Test User")
	return repo
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, repo string, args ...string) {
	t.Helper()
	_ = gitOutput(t, repo, args...)
}

func gitOutput(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}
