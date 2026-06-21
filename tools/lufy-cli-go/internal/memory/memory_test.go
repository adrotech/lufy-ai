package memory

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesStructureAndProjectConfigDefaults(t *testing.T) {
	target := t.TempDir()
	writeFile(t, target, "go.mod", "module example.com/app\n\ngo 1.22\n")

	var out bytes.Buffer
	if err := NewService().Init(Options{Target: target}, &out); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	for _, rel := range []string{
		".lufy/config/project.yaml",
		".lufy/memory/MEMORY.md",
		".lufy/memory/maps/_app-profile.md",
		".lufy/memory/index/backlinks.json",
		".lufy/memory/.gitignore",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	project := readFile(t, filepath.Join(target, ".lufy/config/project.yaml"))
	for _, want := range []string{"memory:", "provider: obsidian", "parallel_execution:", "strategy: independent_review_slices"} {
		if !strings.Contains(project, want) {
			t.Fatalf("project config missing %q:\n%s", want, project)
		}
	}
	if err := NewService().Init(Options{Target: target}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Init() second run error = %v", err)
	}
	out.Reset()
	if err := NewService().Init(Options{Target: target, JSON: true}, &out); err != nil {
		t.Fatalf("Init(JSON) error = %v", err)
	}
	if !strings.Contains(out.String(), `"initialized": true`) {
		t.Fatalf("Init(JSON) output unexpected: %s", out.String())
	}
}

func TestValidateFailsInvalidNotesAndPassesValidNotes(t *testing.T) {
	target := initializedTarget(t)
	writeFile(t, target, ".lufy/memory/knowledge/bad.md", `---
name: bad
description: bad
type: decision
status: active
---

No why.
`)
	var out bytes.Buffer
	if err := NewService().Validate(Options{Target: target}, &out); err == nil {
		t.Fatalf("Validate() expected invalid note error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "description debe aportar") || !strings.Contains(out.String(), "decision requiere") {
		t.Fatalf("Validate() output missing note failures: %s", out.String())
	}

	writeFile(t, target, ".lufy/memory/knowledge/bad.md", `---
name: good-decision
description: Decisión durable sobre memoria portable del proyecto.
type: decision
status: active
---

## Summary

Usar Obsidian como memoria canónica.

**Why:** Es portable y auditable por el equipo.
`)
	out.Reset()
	if err := NewService().Validate(Options{Target: target}, &out); err != nil {
		t.Fatalf("Validate() valid error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Memoria OK") {
		t.Fatalf("Validate() output unexpected: %s", out.String())
	}
	out.Reset()
	if err := NewService().Validate(Options{Target: target, JSON: true}, &out); err != nil {
		t.Fatalf("Validate(JSON) valid error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"ok": true`) {
		t.Fatalf("Validate(JSON) output unexpected: %s", out.String())
	}
}

func TestSearchReturnsActiveResults(t *testing.T) {
	target := initializedTarget(t)
	writeFile(t, target, ".lufy/memory/knowledge/memory-policy.md", `---
name: memory-policy
description: Regla durable sobre búsqueda de memoria.
type: rule
status: active
---

Buscar con rg antes de duplicar notas.
`)
	report, err := NewService().BuildSearch(Options{Target: target, Query: "duplicar"})
	if err != nil {
		t.Fatalf("BuildSearch() error = %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].Status != "active" || !strings.Contains(report.Results[0].Text, "duplicar") {
		t.Fatalf("unexpected search results: %#v", report.Results)
	}
	var out bytes.Buffer
	if err := NewService().Search(Options{Target: target, Query: "duplicar"}, &out); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if !strings.Contains(out.String(), "[active] knowledge/memory-policy.md") {
		t.Fatalf("Search() output unexpected: %s", out.String())
	}
	out.Reset()
	if err := NewService().Search(Options{Target: target, Query: "duplicar", JSON: true}, &out); err != nil {
		t.Fatalf("Search(JSON) error = %v", err)
	}
	if !strings.Contains(out.String(), `"results"`) {
		t.Fatalf("Search(JSON) output unexpected: %s", out.String())
	}
}

func TestCaptureConnectAndIndexMemoryNotes(t *testing.T) {
	target := initializedTarget(t)
	var out bytes.Buffer
	if err := NewService().Capture(CaptureOptions{Target: target, Title: "User Corrections", Type: "rule", Body: "Las correcciones explícitas del usuario deben persistirse como memoria durable.", Links: []string{"app-profile"}}, &out); err != nil {
		t.Fatalf("Capture() error = %v, output=%s", err, out.String())
	}
	notePath := filepath.Join(target, ".lufy/memory/knowledge/user-corrections.md")
	body := readFile(t, notePath)
	if !strings.Contains(body, "[[app-profile]]") || !strings.Contains(body, "correcciones explícitas") {
		t.Fatalf("captured note missing content/link:\n%s", body)
	}
	indexBody := readFile(t, filepath.Join(target, ".lufy/memory/index/backlinks.json"))
	if !strings.Contains(indexBody, `"app-profile"`) || !strings.Contains(indexBody, `"user-corrections"`) {
		t.Fatalf("index missing backlink:\n%s", indexBody)
	}

	out.Reset()
	if err := NewService().Capture(CaptureOptions{Target: target, Title: "AI Correction Lesson", Type: "lesson", Body: "No asumir decisiones técnicas cuando el usuario ya corrigió el criterio."}, &out); err != nil {
		t.Fatalf("Capture() second error = %v", err)
	}
	out.Reset()
	if err := NewService().Connect(ConnectOptions{Target: target, From: "ai-correction-lesson", To: "user-corrections", Bidirectional: true}, &out); err != nil {
		t.Fatalf("Connect() error = %v, output=%s", err, out.String())
	}
	lesson := readFile(t, filepath.Join(target, ".lufy/memory/knowledge/ai-correction-lesson.md"))
	rule := readFile(t, notePath)
	if !strings.Contains(lesson, "[[user-corrections]]") || !strings.Contains(rule, "[[ai-correction-lesson]]") {
		t.Fatalf("bidirectional links missing lesson=%s rule=%s", lesson, rule)
	}
	out.Reset()
	if err := NewService().Index(IndexOptions{Target: target, JSON: true}, &out); err != nil {
		t.Fatalf("Index(JSON) error = %v", err)
	}
	if !strings.Contains(out.String(), `"links"`) {
		t.Fatalf("Index(JSON) output unexpected: %s", out.String())
	}
	if err := NewService().Validate(Options{Target: target}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Validate() after capture/connect error = %v", err)
	}
}

func TestCaptureRejectsBrokenLinksAndDecisionRequiresWhy(t *testing.T) {
	target := initializedTarget(t)
	if err := NewService().Capture(CaptureOptions{Target: target, Title: "Broken Link", Type: "lesson", Body: "No debe crear backlinks rotos.", Links: []string{"missing"}}, &bytes.Buffer{}); err == nil {
		t.Fatalf("Capture() expected missing link error")
	}
	if err := NewService().Capture(CaptureOptions{Target: target, Title: "Decision Memory", Type: "decision", Body: "Persistir decisiones técnicas cuando cambian el criterio del proyecto."}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Capture(decision) error = %v", err)
	}
	decision := readFile(t, filepath.Join(target, ".lufy/memory/knowledge/decision-memory.md"))
	if !strings.Contains(decision, "**Why:**") {
		t.Fatalf("decision note missing why:\n%s", decision)
	}
}

func TestStatusAndValidateReportMissingMemory(t *testing.T) {
	target := t.TempDir()
	writeFile(t, target, "go.mod", "module example.com/app\n\ngo 1.22\n")
	if err := os.MkdirAll(filepath.Join(target, ".lufy"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, target, ".lufy/config/project.yaml", `schema_version: 1
detected_at: 2026-05-20T14:00:00Z
tool: opencode
methodology_by_tier: {}
project_profile:
  surfaces: []
stacks: []
ci:
  detected: false
  workflows: []
tdd:
  strict: true
  triangulate_required: true
  edge_case_categories: [boundary]
validation:
  allowed_commands:
    implementer: []
workflow_limits:
  sizing:
    loc_budget: 400
memory:
  provider: obsidian
  root: .lufy/memory
  git_policy: ignored
  schema_version: 1
  search: rg
  backlinks_index: .lufy/memory/index/backlinks.json
parallel_execution:
  enabled: true
  strategy: independent_review_slices
  max_parallel_agents: 3
  requires_independent_files: true
  requires_merge_plan: true
  validation_mode: grouped_after_join
`)

	report, err := NewService().BuildStatus(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildStatus() error = %v", err)
	}
	if !report.OK || report.Status.Initialized {
		t.Fatalf("status should warn without failing before init: %#v", report)
	}
	report, err = NewService().BuildValidate(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildValidate() error = %v", err)
	}
	if report.OK || !hasCheck(report.Checks, "fail", "memoria no inicializada") {
		t.Fatalf("validate should fail missing memory: %#v", report)
	}
	var out bytes.Buffer
	if err := NewService().Status(Options{Target: target, JSON: true}, &out); err != nil {
		t.Fatalf("Status(JSON) error = %v", err)
	}
	if !strings.Contains(out.String(), `"initialized": false`) {
		t.Fatalf("Status(JSON) output unexpected: %s", out.String())
	}
	out.Reset()
	if err := NewService().Status(Options{Target: target}, &out); err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !strings.Contains(out.String(), "Inicializada: no") {
		t.Fatalf("Status() output unexpected: %s", out.String())
	}
}

func TestValidateReportsBrokenBacklinksAndDrafts(t *testing.T) {
	target := initializedTarget(t)
	writeFile(t, target, ".lufy/memory/knowledge/draft-note.md", `---
name: draft-note
description: Nota draft con backlink roto.
type: lesson
status: draft
---

Relacionada con [[missing-note]].
`)
	status, err := NewService().BuildStatus(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildStatus() error = %v", err)
	}
	if status.Status.Drafts != 1 || status.Status.BrokenBacklinks != 1 || !status.OK {
		t.Fatalf("status should expose draft/broken backlink as warning: %#v", status)
	}
	report, err := NewService().BuildValidate(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildValidate() error = %v", err)
	}
	if report.OK || !hasCheck(report.Checks, "fail", "backlink roto") {
		t.Fatalf("validate should fail broken backlink: %#v", report)
	}
	var out bytes.Buffer
	if err := NewService().Validate(Options{Target: target, JSON: true}, &out); err == nil {
		t.Fatalf("Validate(JSON) expected broken backlink error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), `"ok": false`) {
		t.Fatalf("Validate(JSON) invalid output unexpected: %s", out.String())
	}
}

func TestUnsupportedProviderAndSearchErrors(t *testing.T) {
	target := initializedTarget(t)
	path := filepath.Join(target, ".lufy/config/project.yaml")
	body := strings.ReplaceAll(readFile(t, path), "provider: obsidian", "provider: other")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := NewService().BuildStatus(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildStatus() error = %v", err)
	}
	if report.OK || !hasCheck(report.Checks, "fail", "provider de memoria no soportado") {
		t.Fatalf("unsupported provider not reported: %#v", report)
	}
	if err := NewService().Search(Options{Target: target, Query: ""}, &bytes.Buffer{}); err == nil {
		t.Fatalf("Search() expected empty query error")
	}
}

func TestParseRGOutputAndHelpers(t *testing.T) {
	target := initializedTarget(t)
	notePath := filepath.Join(target, ".lufy/memory/knowledge/active-note.md")
	writeFile(t, target, ".lufy/memory/knowledge/active-note.md", `---
name: active-note
description: Nota activa para parsear rg.
type: concept
status: active
---

Texto durable.
`)
	root := filepath.Join(target, ".lufy/memory")
	results := parseRGOutput(root, notePath+":8:Texto durable.\n")
	if len(results) != 1 || results[0].Status != "active" || results[0].Line != 8 {
		t.Fatalf("parseRGOutput() unexpected: %#v", results)
	}
	windowsLine := `C:\repo\.lufy\memory\knowledge\active-note.md:8:Texto durable.` + "\n"
	path, line, text, ok := parseRGLine(strings.TrimSpace(windowsLine))
	if !ok || path != `C:\repo\.lufy\memory\knowledge\active-note.md` || line != 8 || text != "Texto durable." {
		t.Fatalf("parseRGLine() windows unexpected path=%q line=%d text=%q ok=%v", path, line, text, ok)
	}
	links := wikiLinks("[[Decision One|label]] y [[flow two]]")
	if len(links) != 2 || links[0] != "Decision One" || slugify(links[1]) != "flow-two" {
		t.Fatalf("wiki helper unexpected links=%#v slug=%s", links, slugify(links[1]))
	}
	if statusRank("active") >= statusRank("deprecated") {
		t.Fatalf("statusRank should sort active before deprecated")
	}
	fallback, err := searchMemoryFallback(root, []string{filepath.Join(root, "knowledge")}, "durable")
	if err != nil {
		t.Fatalf("searchMemoryFallback() error = %v", err)
	}
	if len(fallback) != 1 || fallback[0].Status != "active" {
		t.Fatalf("searchMemoryFallback() unexpected: %#v", fallback)
	}
	unsorted := []SearchResult{
		{Status: "deprecated", Path: "b.md", Line: 2},
		{Status: "active", Path: "a.md", Line: 3},
		{Status: "active", Path: "a.md", Line: 1},
	}
	sortSearchResults(unsorted)
	if unsorted[0].Line != 1 || unsorted[2].Status != "deprecated" {
		t.Fatalf("sortSearchResults() unexpected: %#v", unsorted)
	}
}

func TestReadNoteAndSchemaErrorBranches(t *testing.T) {
	target := t.TempDir()
	noFrontmatter := filepath.Join(target, "plain.md")
	writeFile(t, target, "plain.md", "plain\n")
	if _, err := readNote(noFrontmatter); err == nil || !strings.Contains(err.Error(), "frontmatter requerido") {
		t.Fatalf("readNote() expected missing frontmatter error, got %v", err)
	}
	writeFile(t, target, "open.md", "---\nname: open\n")
	if _, err := readNote(filepath.Join(target, "open.md")); err == nil || !strings.Contains(err.Error(), "frontmatter sin cierre") {
		t.Fatalf("readNote() expected open frontmatter error, got %v", err)
	}
	writeFile(t, target, "bad-yaml.md", "---\nname: [\n---\n")
	if _, err := readNote(filepath.Join(target, "bad-yaml.md")); err == nil || !strings.Contains(err.Error(), "frontmatter YAML inválido") {
		t.Fatalf("readNote() expected YAML error, got %v", err)
	}

	checks := []Check{}
	validateNoteSchema("bad.md", note{Type: "weird", Status: "stale"}, &checks)
	if !hasCheck(checks, "fail", "type inválido") || !hasCheck(checks, "fail", "status inválido") || !hasCheck(checks, "fail", "frontmatter requiere name") {
		t.Fatalf("validateNoteSchema() missing expected checks: %#v", checks)
	}
	if contains([]string{"a"}, "b") {
		t.Fatalf("contains() should return false for missing value")
	}
}

func initializedTarget(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	writeFile(t, target, "go.mod", "module example.com/app\n\ngo 1.22\n")
	if err := NewService().Init(Options{Target: target}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	return target
}

func hasCheck(checks []Check, level, text string) bool {
	for _, check := range checks {
		if check.Level == level && strings.Contains(check.Message, text) {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
