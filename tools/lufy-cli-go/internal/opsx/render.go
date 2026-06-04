package opsx

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type RenderOptions struct {
	Target string
	Change string
	Format string
	Theme  string
	Output string
}

type RenderResult struct {
	OutputPath string
	Artifacts  []RenderedArtifact
}

type RenderedArtifact struct {
	Title string
	Path  string
	State string
}

type ChangeRenderer struct{}

func NewChangeRenderer() ChangeRenderer {
	return ChangeRenderer{}
}

func (ChangeRenderer) Render(opts RenderOptions) (RenderResult, error) {
	format := opts.Format
	if format == "" {
		format = "html"
	}
	if format != "html" {
		return RenderResult{}, fmt.Errorf("formato no soportado: %s", format)
	}
	theme := opts.Theme
	if theme == "" {
		theme = "notion-dark"
	}
	if theme != "notion-dark" {
		return RenderResult{}, fmt.Errorf("tema no soportado: %s", theme)
	}
	if opts.Change == "" {
		return RenderResult{}, fmt.Errorf("opsx render requiere --change <name>")
	}
	target := opts.Target
	if target == "" {
		target = "."
	}
	changeRel, err := platform.EnsureRelativeSafe(filepath.Join("openspec", "changes", opts.Change))
	if err != nil {
		return RenderResult{}, err
	}
	changeDir, err := platform.SafeJoin(target, changeRel)
	if err != nil {
		return RenderResult{}, err
	}
	if info, err := os.Stat(changeDir); err != nil || !info.IsDir() {
		if err != nil {
			return RenderResult{}, fmt.Errorf("change no encontrado: %s", opts.Change)
		}
		return RenderResult{}, fmt.Errorf("change no es directorio: %s", opts.Change)
	}

	artifacts, err := collectChangeArtifacts(changeDir)
	if err != nil {
		return RenderResult{}, err
	}
	output := opts.Output
	if output == "" {
		output = filepath.Join(changeDir, "change-overview.html")
	} else if !filepath.IsAbs(output) {
		output = filepath.Join(target, output)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return RenderResult{}, err
	}
	body := renderNotionDarkHTML(opts.Change, artifacts)
	if err := os.WriteFile(output, []byte(body), 0o644); err != nil {
		return RenderResult{}, err
	}

	result := RenderResult{OutputPath: output}
	for _, artifact := range artifacts {
		result.Artifacts = append(result.Artifacts, RenderedArtifact{Title: artifact.Title, Path: artifact.DisplayPath, State: artifact.State})
	}
	return result, nil
}

type changeArtifact struct {
	Title       string
	DisplayPath string
	State       string
	Content     string
}

func collectChangeArtifacts(changeDir string) ([]changeArtifact, error) {
	base := []struct {
		title string
		rel   string
	}{
		{"Proposal", "proposal.md"},
		{"Design", "design.md"},
		{"Plan", "plan.md"},
		{"Tasks", "tasks.md"},
	}
	seen := make(map[string]bool, len(base))
	var artifacts []changeArtifact
	for _, item := range base {
		seen[item.rel] = true
		artifact, err := readArtifact(changeDir, item.title, item.rel)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	entries, err := os.ReadDir(changeDir)
	if err != nil {
		return nil, err
	}
	var extraRels []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".md" || seen[name] {
			continue
		}
		extraRels = append(extraRels, name)
	}
	sort.Strings(extraRels)
	for _, rel := range extraRels {
		artifact, err := readArtifact(changeDir, rel, rel)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func readArtifact(changeDir, title, rel string) (changeArtifact, error) {
	path, err := platform.SafeJoin(changeDir, rel)
	if err != nil {
		return changeArtifact{}, err
	}
	body, err := os.ReadFile(path)
	state := "No disponible"
	content := ""
	if err == nil {
		content = string(body)
		if strings.TrimSpace(content) != "" {
			state = "Disponible"
		}
	} else if !os.IsNotExist(err) {
		return changeArtifact{}, err
	}
	return changeArtifact{Title: title, DisplayPath: filepath.ToSlash(rel), State: state, Content: content}, nil
}

func renderNotionDarkHTML(change string, artifacts []changeArtifact) string {
	var controls, panels strings.Builder
	for idx, artifact := range artifacts {
		id := fmt.Sprintf("artifact-%d", idx+1)
		checked := ""
		if idx == 0 {
			checked = " checked"
		}
		fmt.Fprintf(&controls, `<input type="radio" name="artifact-tab" id="tab-%d"%s><label for="tab-%d"><span>%s</span><small>%s</small></label>`, idx+1, checked, idx+1, html.EscapeString(artifact.Title), html.EscapeString(artifact.State))
		fmt.Fprintf(&panels, `<section class="panel" id="%s"><div class="panel-head"><div><p class="eyebrow">%s</p><h2>%s</h2></div><span class="badge">%s</span></div><p class="path">%s</p><div class="markdown">%s</div></section>`, id, html.EscapeString(artifact.DisplayPath), html.EscapeString(artifact.Title), html.EscapeString(artifact.State), html.EscapeString(artifact.DisplayPath), markdownToHTML(artifact.Content))
	}
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + html.EscapeString(change) + ` - OpenSpec</title>
  <style>
    :root { color-scheme: dark; --bg: #141414; --panel: #191919; --card: #202020; --card2: #252525; --text: #ededeb; --muted: #a6a29b; --line: #33302b; --accent: #d4a96a; --accent2: #8f7351; --green: #6fbf8f; }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; background: radial-gradient(circle at top left, rgba(212,169,106,.13), transparent 28%), radial-gradient(circle at bottom right, rgba(111,191,143,.08), transparent 30%), var(--bg); color: var(--text); font: 15px/1.6 ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; overflow: hidden; }
    main { width: min(1240px, calc(100vw - 32px)); height: 100vh; margin: 0 auto; padding: 28px 0; display: grid; grid-template-rows: auto 1fr; gap: 16px; }
    .hero { border: 1px solid var(--line); background: linear-gradient(135deg, rgba(37,37,37,.98), rgba(25,25,25,.94)); border-radius: 24px; padding: 24px 28px; box-shadow: 0 24px 80px rgba(0,0,0,.35); }
    .crumb { color: var(--accent); font-size: 13px; letter-spacing: .08em; text-transform: uppercase; }
    h1 { margin: 8px 0 16px; font-size: clamp(34px, 6vw, 64px); line-height: .95; letter-spacing: -.055em; }
    .meta { display: flex; flex-wrap: wrap; gap: 10px; }
    .pill, .badge { border: 1px solid var(--line); border-radius: 999px; padding: 5px 11px; background: rgba(255,255,255,.035); color: var(--muted); font-size: 13px; }
    .workspace { min-height: 0; border: 1px solid var(--line); border-radius: 22px; background: rgba(25,25,25,.9); box-shadow: 0 18px 60px rgba(0,0,0,.28); display: grid; grid-template-rows: auto 1fr; overflow: hidden; }
    .tabs { display: flex; gap: 8px; overflow-x: auto; padding: 12px; border-bottom: 1px solid var(--line); background: rgba(255,255,255,.025); }
    .tabs input { position: absolute; opacity: 0; pointer-events: none; }
    .tabs label { flex: 0 0 auto; display: grid; gap: 1px; min-width: 128px; border: 1px solid var(--line); border-radius: 14px; padding: 9px 12px; color: var(--muted); background: rgba(255,255,255,.03); cursor: pointer; user-select: none; }
    .tabs label span { color: var(--text); font-weight: 650; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 220px; }
    .tabs label small { color: var(--muted); font-size: 12px; }
    .tabs label:hover { background: rgba(255,255,255,.06); }
    .panels { min-height: 0; overflow: auto; padding: 22px; }
    .panel { display: none; min-height: 100%; border: 1px solid var(--line); border-radius: 20px; background: linear-gradient(180deg, var(--card), var(--panel)); padding: 24px; box-shadow: 0 14px 42px rgba(0,0,0,.22); }
    .panel-head { display: flex; justify-content: space-between; align-items: start; gap: 16px; }
    ` + tabCSS(len(artifacts)) + `
	.eyebrow, .path { color: var(--muted); font: 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; margin: 0 0 6px; }
    h2 { margin: 0; font-size: 25px; letter-spacing: -.025em; }
    .markdown { margin-top: 18px; }
    .markdown h1, .markdown h2, .markdown h3, .markdown h4 { margin: 22px 0 8px; line-height: 1.2; letter-spacing: -.02em; }
    .markdown h1 { font-size: 26px; } .markdown h2 { font-size: 22px; } .markdown h3 { font-size: 18px; color: var(--accent); } .markdown h4 { font-size: 15px; color: var(--green); }
    .markdown p { margin: 9px 0; color: #ddd8cf; }
    .markdown ul { margin: 10px 0 14px 20px; padding: 0; }
    .markdown li { margin: 6px 0; }
    pre { overflow: auto; background: #0f0f0f; border: 1px solid var(--line); border-radius: 14px; padding: 14px; color: #f4ead9; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
    .empty { color: var(--muted); border-left: 3px solid var(--accent2); padding-left: 12px; }
    @media (max-width: 820px) { body { overflow: auto; } main { height: auto; min-height: 100vh; padding: 16px 0; } .hero { padding: 20px; } .workspace { min-height: 70vh; } .tabs label { min-width: 112px; } .panels { padding: 14px; } .panel { padding: 18px; } }
  </style>
</head>
<body>
  <main>
    <header class="hero">
      <div class="crumb">OpenSpec Change Overview</div>
      <h1>` + html.EscapeString(change) + `</h1>
    </header>
    <div class="workspace"><div class="tabs">` + controls.String() + `</div><div class="panels">` + panels.String() + `</div></div>
  </main>
</body>
</html>
`
}

func tabCSS(count int) string {
	var out strings.Builder
	for i := 1; i <= count; i++ {
		fmt.Fprintf(&out, ".tabs:has(#tab-%d:checked) label[for=\"tab-%d\"] { background: linear-gradient(180deg, rgba(212,169,106,.22), rgba(255,255,255,.04)); border-color: rgba(212,169,106,.55); color: var(--text); }\n", i, i)
		fmt.Fprintf(&out, ".workspace:has(#tab-%d:checked) .panels .panel:nth-child(%d) { display: block; }\n", i, i)
	}
	return out.String()
}

func markdownToHTML(markdown string) string {
	if strings.TrimSpace(markdown) == "" {
		return `<p class="empty">Este artefacto no existe o está vacío.</p>`
	}
	var out strings.Builder
	inList := false
	inPre := false
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(line, "```") {
			if inList {
				out.WriteString("</ul>")
				inList = false
			}
			if inPre {
				out.WriteString("</code></pre>")
			} else {
				out.WriteString("<pre><code>")
			}
			inPre = !inPre
			continue
		}
		if inPre {
			out.WriteString(html.EscapeString(line) + "\n")
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inList {
				out.WriteString("</ul>")
				inList = false
			}
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "#### "):
			closeList(&out, &inList)
			fmt.Fprintf(&out, "<h4>%s</h4>", renderInline(strings.TrimSpace(trimmed[5:])))
		case strings.HasPrefix(trimmed, "### "):
			closeList(&out, &inList)
			fmt.Fprintf(&out, "<h3>%s</h3>", renderInline(strings.TrimSpace(trimmed[4:])))
		case strings.HasPrefix(trimmed, "## "):
			closeList(&out, &inList)
			fmt.Fprintf(&out, "<h2>%s</h2>", renderInline(strings.TrimSpace(trimmed[3:])))
		case strings.HasPrefix(trimmed, "# "):
			closeList(&out, &inList)
			fmt.Fprintf(&out, "<h1>%s</h1>", renderInline(strings.TrimSpace(trimmed[2:])))
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			if !inList {
				out.WriteString("<ul>")
				inList = true
			}
			fmt.Fprintf(&out, "<li>%s</li>", renderListItem(strings.TrimSpace(trimmed[2:])))
		default:
			closeList(&out, &inList)
			fmt.Fprintf(&out, "<p>%s</p>", renderInline(trimmed))
		}
	}
	closeList(&out, &inList)
	if inPre {
		out.WriteString("</code></pre>")
	}
	return out.String()
}

func renderListItem(item string) string {
	for _, marker := range []struct {
		prefix  string
		checked bool
	}{
		{prefix: "[ ] ", checked: false},
		{prefix: "[x] ", checked: true},
		{prefix: "[X] ", checked: true},
	} {
		if strings.HasPrefix(item, marker.prefix) {
			checked := ""
			if marker.checked {
				checked = " checked"
			}
			return fmt.Sprintf(`<input type="checkbox" disabled%s> %s`, checked, renderInline(strings.TrimSpace(item[len(marker.prefix):])))
		}
	}
	return renderInline(item)
}

func renderInline(text string) string {
	var out strings.Builder
	for len(text) > 0 {
		codeStart := strings.Index(text, "`")
		linkStart := strings.Index(text, "[")
		next := firstInlineToken(codeStart, linkStart)
		if next < 0 {
			out.WriteString(html.EscapeString(text))
			break
		}
		out.WriteString(html.EscapeString(text[:next]))
		text = text[next:]
		if strings.HasPrefix(text, "`") {
			end := strings.Index(text[1:], "`")
			if end < 0 {
				out.WriteString(html.EscapeString("`"))
				text = text[1:]
				continue
			}
			code := text[1 : end+1]
			fmt.Fprintf(&out, "<code>%s</code>", html.EscapeString(code))
			text = text[end+2:]
			continue
		}
		linkHTML, consumed, ok := renderMarkdownLink(text)
		if ok {
			out.WriteString(linkHTML)
			text = text[consumed:]
			continue
		}
		out.WriteString(html.EscapeString("["))
		text = text[1:]
	}
	return out.String()
}

func firstInlineToken(indexes ...int) int {
	next := -1
	for _, idx := range indexes {
		if idx >= 0 && (next < 0 || idx < next) {
			next = idx
		}
	}
	return next
}

func renderMarkdownLink(text string) (string, int, bool) {
	closeText := strings.Index(text, "](")
	if closeText <= 0 {
		return "", 0, false
	}
	closeURL := strings.Index(text[closeText+2:], ")")
	if closeURL < 0 {
		return "", 0, false
	}
	closeURL += closeText + 2
	label := text[1:closeText]
	url := text[closeText+2 : closeURL]
	consumed := closeURL + 1
	if !isSafeMarkdownURL(url) {
		return html.EscapeString(text[:consumed]), consumed, true
	}
	return fmt.Sprintf(`<a href="%s" rel="noopener noreferrer">%s</a>`, html.EscapeString(url), renderInline(label)), consumed, true
}

func isSafeMarkdownURL(url string) bool {
	if strings.TrimSpace(url) != url || url == "" {
		return false
	}
	for _, r := range url {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	lower := strings.ToLower(url)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:")
}

func closeList(out *strings.Builder, inList *bool) {
	if *inList {
		out.WriteString("</ul>")
		*inList = false
	}
}
