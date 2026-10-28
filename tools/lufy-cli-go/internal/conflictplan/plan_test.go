package conflictplan

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
)

func TestBuildGroupsUnmanagedConflicts(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeFile(t, filepath.Join(target, ".codex", "config.toml"), "local agent\n")
	writeFile(t, filepath.Join(target, ".lufy", "contracts", "work.md"), "local spec\n")

	report, err := NewService().Build(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK || report.Summary.Conflicts != 2 || report.Summary.Groups != 2 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
	want := map[string]bool{".codex": false, ".lufy": false}
	for _, item := range report.Items {
		want[item.Category] = true
		if item.Recommendation != "merge" || len(item.AvailableActions) == 0 || item.ParallelGroup == "" {
			t.Fatalf("item missing recommendation/actions: %#v", item)
		}
	}
	for category, found := range want {
		if !found {
			t.Fatalf("missing category %s in %#v", category, report.Items)
		}
	}
}

func TestRunJSONIsParseable(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "lufy-ia.harness.md"), `{"local":true}`)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, JSON: true}, &out); err != nil {
		t.Fatal(err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json %q: %v", out.String(), err)
	}
	if len(report.Items) != 1 || report.Items[0].Category != "root/config" || report.Items[0].Recommendation != "block" {
		t.Fatalf("unexpected report: %#v", report.Items)
	}
}

func TestRunHumanNoConflictsReportsOKNextAction(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("No hay conflictos")) || !bytes.Contains(out.Bytes(), []byte("puedes continuar")) {
		t.Fatalf("unexpected human output: %s", out.String())
	}
}

func TestUnsafeTargetRecommendsBlock(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".codex", "config.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	report, err := NewService().Build(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Items) != 1 {
		t.Fatalf("expected one unsafe item: %#v", report.Items)
	}
	item := report.Items[0]
	if item.Status != "unsafe_target" || item.Recommendation != "block" || item.Risk != "high" {
		t.Fatalf("unexpected unsafe item: %#v", item)
	}
	if len(item.AvailableActions) != 2 {
		t.Fatalf("unsafe item should restrict actions: %#v", item.AvailableActions)
	}
}

func TestCategoryFallbacks(t *testing.T) {
	tests := map[string]string{
		".agents/skills/foo/SKILL.md":    ".agents/skills",
		".codex/agents/implementer.toml": ".codex",
		"lufy-ia.harness.md":             "root/config",
		"some-user-file.txt":             "other",
	}
	for path, want := range tests {
		if got := categoryFor(path); got != want {
			t.Fatalf("categoryFor(%q)=%q want %q", path, got, want)
		}
	}
}

func TestRecommendationAndStatusHelpers(t *testing.T) {
	if got := statusFor("archivo gestionado con drift local"); got != "managed_drift" {
		t.Fatalf("managed drift status = %s", got)
	}
	if got := statusFor("otra razon"); got != "blocked_conflict" {
		t.Fatalf("blocked status = %s", got)
	}
	if got := recommendationFor("AGENTS.md", "archivo existente no gestionado", ""); got != "block" {
		t.Fatalf("AGENTS recommendation = %s", got)
	}
	if got := recommendationFor("lufy-ia.harness.md", "archivo existente no gestionado", ""); got != "block" {
		t.Fatalf("root config recommendation = %s", got)
	}
	if got := recommendationFor("foo.md", "archivo", assets.PolicyMergeBlock); got != "merge" {
		t.Fatalf("merge-block recommendation = %s", got)
	}
	if got := riskFor("foo.md", "archivo existente no gestionado"); got != "medium" {
		t.Fatalf("medium risk = %s", got)
	}
	if got := riskRank("low"); got != 1 {
		t.Fatalf("low risk rank = %d", got)
	}
	if got := parallelGroupFor(".agents/skills"); got != "agents-skills" {
		t.Fatalf("parallel group = %s", got)
	}
	if got := reasonWithComponent("reason", "component"); got != "reason; component=component" {
		t.Fatalf("reason with component = %s", got)
	}
	if got := reasonWithComponent("reason", ""); got != "reason" {
		t.Fatalf("reason without component = %s", got)
	}
}

func TestPrintReportIncludesGroupDetails(t *testing.T) {
	report := Report{
		TargetRoot: "/tmp/project",
		Items: []Item{{
			Path:             ".agents/skills/foo/SKILL.md",
			Category:         ".agents/skills",
			Status:           "unmanaged_conflict",
			Risk:             "medium",
			Recommendation:   "merge",
			Reason:           "local plus managed",
			AvailableActions: []string{"keep-local", "merge"},
			ParallelGroup:    "agents-skills",
		}},
		Groups:      []Group{{Category: ".agents/skills", Risk: "medium", Count: 1, Paths: []string{".agents/skills/foo/SKILL.md"}, ParallelGroup: "agents-skills"}},
		NextActions: []string{"next"},
	}
	report.Summary = summaryFor(report)
	var out bytes.Buffer
	printReport(report, &out)
	for _, want := range []string{".agents/skills", "acciones: keep-local, merge", "next"} {
		if !bytes.Contains(out.Bytes(), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, out.String())
		}
	}
}

func minimalSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	paths := map[string]string{
		"lufy-ia.harness.md":                 "harness\n",
		".agents/skills/lufy-check/SKILL.md": "skill\n",
		".codex/config.toml":                 "config\n",
		".lufy/README.md":                    "readme\n",
		".lufy/contracts/work.md":            "work\n",
	}

	for path, body := range paths {
		writeFile(t, filepath.Join(root, path), body)
	}
	return root
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

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}
