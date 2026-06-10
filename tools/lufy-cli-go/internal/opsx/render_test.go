package opsx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderChangeHTMLIncludesArtifactsAndEscapesContent(t *testing.T) {
	target := t.TempDir()
	changeDir := filepath.Join(target, "openspec", "changes", "add-widget")
	writeFile(t, filepath.Join(changeDir, "proposal.md"), "# Proposal\n\nDiseño y acción\n\n<script>alert(1)</script>\n\n```\n<script>fence()</script>\n```")
	writeFile(t, filepath.Join(changeDir, "design.md"), "## Design\n\nDark `cards <raw>`\n\n[Seguro](https://example.com/ok?x=1&y=2)\n\n[Inseguro](javascript:alert(1))")
	writeFile(t, filepath.Join(changeDir, "tasks.md"), "## Tasks\n\n- [ ] Pendiente\n- [x] Hecho\n* [X] También hecho")
	writeFile(t, filepath.Join(changeDir, "notes.md"), "## Notes\n\nTop-level context")
	writeFile(t, filepath.Join(changeDir, "specs", "widget", "spec.md"), "## ADDED Requirements\n\n### Requirement: Widget\n\n#### Scenario: Render\nWHEN open\nTHEN show")

	out := filepath.Join(target, "overview.html")
	res, err := NewChangeRenderer().Render(RenderOptions{Target: target, Change: "add-widget", Output: out})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if res.OutputPath != out || len(res.Artifacts) != 5 {
		t.Fatalf("unexpected result: %#v", res)
	}
	body := readFile(t, out)
	for _, want := range []string{"Proposal", "Design", "Plan", "Tasks", "notes.md"} {
		if !strings.Contains(body, want) {
			t.Fatalf("html missing %q", want)
		}
	}
	for _, unwanted := range []string{"Spec: specs/widget/spec.md", "specs/widget/spec.md", "Notion dark", "Offline HTML", "Artifacts disponibles:", "Sin recursos remotos"} {
		if strings.Contains(body, unwanted) {
			t.Fatalf("html should not include %q", unwanted)
		}
	}
	for _, want := range []string{`class="workspace"`, `class="tabs"`, `id="tab-1"`, `class="panels"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("html missing tab layout marker %q", want)
		}
	}
	if strings.Contains(body, "<nav>") || strings.Contains(body, `class="layout"`) {
		t.Fatalf("html should not render sidebar/navigation layout")
	}
	for _, want := range []string{"Diseño", "acción", "está vacío"} {
		if !strings.Contains(body, want) {
			t.Fatalf("html should preserve unicode text %q", want)
		}
	}
	for _, want := range []string{`<input type="checkbox" disabled> Pendiente`, `<input type="checkbox" disabled checked> Hecho`, `<input type="checkbox" disabled checked> También hecho`} {
		if !strings.Contains(body, want) {
			t.Fatalf("html missing checkbox %q", want)
		}
	}
	if !strings.Contains(body, `<code>cards &lt;raw&gt;</code>`) {
		t.Fatalf("html did not render escaped inline code")
	}
	if !strings.Contains(body, `<a href="https://example.com/ok?x=1&amp;y=2" rel="noopener noreferrer">Seguro</a>`) {
		t.Fatalf("html did not render safe markdown link")
	}
	if strings.Contains(body, `href="javascript:alert(1)"`) || !strings.Contains(body, `[Inseguro](javascript:alert(1))`) {
		t.Fatalf("html should render unsafe markdown link as escaped text")
	}
	if strings.Contains(body, "<script>alert(1)</script>") || !strings.Contains(body, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("html did not escape script content")
	}
	if strings.Contains(body, "<script>fence()</script>") || !strings.Contains(body, "&lt;script&gt;fence()&lt;/script&gt;") {
		t.Fatalf("html did not escape fenced code content")
	}
}

func TestRenderChangeHTMLAllowsMissingPlan(t *testing.T) {
	target := t.TempDir()
	changeDir := filepath.Join(target, "openspec", "changes", "without-plan")
	writeFile(t, filepath.Join(changeDir, "proposal.md"), "# Proposal")
	writeFile(t, filepath.Join(changeDir, "tasks.md"), "# Tasks")

	res, err := NewChangeRenderer().Render(RenderOptions{Target: target, Change: "without-plan"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	body := readFile(t, res.OutputPath)
	if !strings.Contains(body, "Plan") || !strings.Contains(body, "No disponible") {
		t.Fatalf("html should include unavailable Plan section: %s", body)
	}
}

func TestRenderChangeRejectsUnknownFormat(t *testing.T) {
	_, err := NewChangeRenderer().Render(RenderOptions{Target: t.TempDir(), Change: "x", Format: "pdf"})
	if err == nil || !strings.Contains(err.Error(), "formato no soportado") {
		t.Fatalf("expected format error, got %v", err)
	}
}

func TestRenderChangeWritesRelativeOutputUnderTarget(t *testing.T) {
	target := t.TempDir()
	changeDir := filepath.Join(target, "openspec", "changes", "relative-output")
	writeFile(t, filepath.Join(changeDir, "proposal.md"), "# Proposal")
	writeFile(t, filepath.Join(changeDir, "tasks.md"), "# Tasks")

	res, err := NewChangeRenderer().Render(RenderOptions{Target: target, Change: "relative-output", Output: filepath.Join("reports", "overview.html")})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := filepath.Join(target, "reports", "overview.html")
	if res.OutputPath != want {
		t.Fatalf("OutputPath = %s, want %s", res.OutputPath, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRenderChangeRejectsMissingChangeAndUnknownTheme(t *testing.T) {
	target := t.TempDir()
	_, err := NewChangeRenderer().Render(RenderOptions{Target: target, Theme: "notion-dark"})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Fatalf("expected missing change error, got %v", err)
	}
	_, err = NewChangeRenderer().Render(RenderOptions{Target: target, Change: "x", Theme: "light"})
	if err == nil || !strings.Contains(err.Error(), "tema no soportado") {
		t.Fatalf("expected theme error, got %v", err)
	}
}

func TestRenderInlineRendersBoldSafely(t *testing.T) {
	got := renderInline("Antes **fuerte** después")
	want := "Antes <strong>fuerte</strong> después"
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
}

func TestRenderInlineEscapesBoldContent(t *testing.T) {
	got := renderInline("**<raw>**")
	want := "<strong>&lt;raw&gt;</strong>"
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
}

func TestRenderInlineCodeKeepsBoldMarkersLiteral(t *testing.T) {
	got := renderInline("`**no bold**`")
	want := "<code>**no bold**</code>"
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
}

func TestRenderInlineSafeLinksAllowBoldLabels(t *testing.T) {
	got := renderInline("[**Seguro**](https://example.com/ok) y **adyacente** [link](mailto:test@example.com)")
	want := `<a href="https://example.com/ok" rel="noopener noreferrer"><strong>Seguro</strong></a> y <strong>adyacente</strong> <a href="mailto:test@example.com" rel="noopener noreferrer">link</a>`
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
}

func TestRenderInlineUnsafeLinksStayEscapedText(t *testing.T) {
	got := renderInline("[**Inseguro**](javascript:alert(1))")
	want := `[**Inseguro**](javascript:alert(1))`
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
	if strings.Contains(got, "<a ") || strings.Contains(got, "href=") {
		t.Fatalf("unsafe link should not become clickable: %q", got)
	}
}

func TestRenderInlineUnclosedBoldStaysEscapedText(t *testing.T) {
	got := renderInline("**sin cerrar <raw> [link](https://example.com)")
	want := "**sin cerrar &lt;raw&gt; [link](https://example.com)"
	if got != want {
		t.Fatalf("renderInline() = %q, want %q", got, want)
	}
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
