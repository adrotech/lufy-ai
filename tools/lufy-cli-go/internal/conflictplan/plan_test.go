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
	writeFile(t, filepath.Join(target, ".opencode", "agents", "orchestrator.md"), "local agent\n")
	writeFile(t, filepath.Join(target, "openspec", "specs", "sdd-harness-routing", "spec.md"), "local spec\n")

	report, err := NewService().Build(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK || report.Summary.Conflicts != 2 || report.Summary.Groups != 2 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
	want := map[string]bool{".opencode/agents": false, "openspec/specs": false}
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
	writeFile(t, filepath.Join(target, ".opencode", "package.json"), `{"local":true}`)
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

func TestLegacyDeprecatedLayoutIsReported(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeFile(t, filepath.Join(target, ".lufy-ai", "backups", "legacy.txt"), "legacy\n")
	report, err := NewService().Build(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.LegacyDeprecated) == 0 || report.LegacyDeprecated[0].Recommendation != "migrate-layout" {
		t.Fatalf("missing legacy deprecated item: %#v", report.LegacyDeprecated)
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
	if err := os.MkdirAll(filepath.Join(target, ".opencode", "agents", "orchestrator.md"), 0o755); err != nil {
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
		".opencode/commands/foo.md":      ".opencode/commands",
		".opencode/skills/foo/SKILL.md":  ".opencode/skills",
		".opencode/templates/foo.md":     ".opencode/templates",
		".agents/skills/foo/SKILL.md":    ".agents/skills",
		".codex/agents/implementer.toml": ".codex",
		".lufy/workflows/sdd/README.md":  ".lufy",
		"openspec/config.yaml":           "openspec",
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
	if got := recommendationFor("AGENTS.md", "archivo existente no gestionado", ""); got != "merge" {
		t.Fatalf("AGENTS recommendation = %s", got)
	}
	if got := recommendationFor(".opencode/README.md", "archivo existente no gestionado", ""); got != "block" {
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
	if got := parallelGroupFor(".opencode/skills"); got != "opencode-skills" {
		t.Fatalf("parallel group = %s", got)
	}
	if got := reasonWithComponent("reason", "component"); got != "reason; component=component" {
		t.Fatalf("reason with component = %s", got)
	}
	if got := reasonWithComponent("reason", ""); got != "reason" {
		t.Fatalf("reason without component = %s", got)
	}
}

func TestPrintReportIncludesLegacyAndGroupDetails(t *testing.T) {
	report := Report{
		TargetRoot: "/tmp/project",
		Items: []Item{{
			Path:             ".opencode/skills/foo/SKILL.md",
			Category:         ".opencode/skills",
			Status:           "unmanaged_conflict",
			Risk:             "medium",
			Recommendation:   "merge",
			Reason:           "local plus managed",
			AvailableActions: []string{"keep-local", "merge"},
			ParallelGroup:    "opencode-skills",
		}},
		Groups:           []Group{{Category: ".opencode/skills", Risk: "medium", Count: 1, Paths: []string{".opencode/skills/foo/SKILL.md"}, ParallelGroup: "opencode-skills"}},
		LegacyDeprecated: []LegacyItem{{Path: ".lufy-ai/backups", CanonicalPath: ".lufy/managed-state/backups", Status: "deprecated_layout", Recommendation: "migrate-layout"}},
		NextActions:      []string{"next"},
	}
	report.Summary = summaryFor(report)
	var out bytes.Buffer
	printReport(report, &out)
	for _, want := range []string{".opencode/skills", "acciones: keep-local, merge", "Legacy/deprecated", "next"} {
		if !bytes.Contains(out.Bytes(), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, out.String())
		}
	}
}

func minimalSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	paths := map[string]string{
		"lufy-ia.harness.md":                                                 "harness\n",
		"tui.json":                                                           "{}\n",
		filepath.Join(".opencode", "README.md"):                              "readme\n",
		filepath.Join(".opencode", "package.json"):                           "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                      "{}\n",
		filepath.Join(".opencode", ".gitignore"):                             "node_modules\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):              "agent\n",
		filepath.Join(".opencode", "commands", "opsx-sync.md"):               "command\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "SKILL.md"):     "skill\n",
		filepath.Join(".opencode", "templates", "result-contract.md"):        "template\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"):       "plugin\n",
		filepath.Join(".opencode", "policies", "delivery.md"):                "policy\n",
		filepath.Join("openspec", "config.yaml"):                             "version: 2\n",
		filepath.Join("openspec", "UPSTREAM.json"):                           "{}\n",
		filepath.Join("openspec", "specs", "sdd-harness-routing", "spec.md"): "spec\n",
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"):       "skill\n",
		filepath.Join(".codex", "config.toml"):                               "config\n",
		filepath.Join(".lufy", "README.md"):                                  "readme\n",
		filepath.Join(".lufy", "workflows", "sdd", "README.md"):              "sdd\n",
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
