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
		{"Tasks", "tasks.md"},
	}
	seen := map[string]bool{"plan.md": true}
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
    :root { color-scheme: light; --primary: #5645d4; --primary-pressed: #4534b3; --navy: #0a1530; --navy-deep: #070f24; --link: #0075de; --orange: #dd5b00; --green: #1aae39; --peach: #ffe8d4; --rose: #fde0ec; --mint: #d9f3e1; --lavender: #e6e0f5; --sky: #dcecfa; --canvas: #ffffff; --surface: #f6f5f4; --surface-soft: #fafaf9; --hairline: #e5e3df; --hairline-soft: #ede9e4; --hairline-strong: #c8c4be; --ink: #1a1a1a; --charcoal: #37352f; --slate: #5d5b54; --steel: #787671; --stone: #a4a097; --muted: #bbb8b1; --error: #e03131; --shadow-card: rgba(15,15,15,.08) 0 4px 12px; --shadow-hero: rgba(15,15,15,.20) 0 24px 48px -8px; }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; background: linear-gradient(180deg, var(--navy) 0, var(--navy-deep) 340px, var(--surface) 340px); color: var(--charcoal); font: 16px/1.55 "Notion Sans", Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; }
    main { max-width: 1180px; margin: 0 auto; padding: 48px 32px 72px; }
    .hero { color: #fff; text-align: center; padding: 16px 0 30px; }
    .crumb { color: var(--stone); font-size: 13px; font-weight: 600; letter-spacing: 0; text-transform: uppercase; }
    h1 { margin: 8px auto 16px; max-width: 900px; font-size: 56px; line-height: 1.08; letter-spacing: 0; font-weight: 600; }
    .meta { display: flex; flex-wrap: wrap; gap: 10px; }
    .pill, .badge { border: 1px solid var(--hairline); border-radius: 999px; padding: 5px 11px; background: var(--surface); color: var(--slate); font-size: 13px; font-weight: 600; }
    .workspace { border: 1px solid var(--hairline); border-radius: 12px; background: var(--canvas); box-shadow: var(--shadow-hero); display: grid; grid-template-rows: auto 1fr; overflow: hidden; }
    .tabs { display: flex; gap: 8px; overflow-x: auto; padding: 12px; border-bottom: 1px solid var(--hairline); background: var(--surface); }
    .tabs input { position: absolute; opacity: 0; pointer-events: none; }
    .tabs label { flex: 0 0 auto; display: grid; gap: 1px; min-width: 128px; border: 1px solid var(--hairline); border-radius: 8px; padding: 9px 12px; color: var(--slate); background: var(--canvas); cursor: pointer; user-select: none; }
    .tabs label span { color: var(--ink); font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 220px; }
    .tabs label small { color: var(--slate); font-size: 12px; }
    .tabs label:hover { background: var(--surface-soft); border-color: var(--hairline-strong); }
    .panels { overflow: auto; padding: 22px; background: var(--surface-soft); }
    .panel { display: none; min-height: 100%; border: 1px solid var(--hairline); border-radius: 12px; background: var(--canvas); padding: 24px; box-shadow: var(--shadow-card); }
    .panel-head { display: flex; justify-content: space-between; align-items: start; gap: 16px; }
    ` + tabCSS(len(artifacts)) + `
	.eyebrow, .path { color: var(--slate); font: 12px/1.4 ui-monospace, SFMono-Regular, Menlo, monospace; margin: 0 0 6px; }
    h2 { margin: 0; font-size: 28px; line-height: 1.25; letter-spacing: 0; color: var(--ink); }
    .markdown { margin-top: 18px; }
    .markdown h1, .markdown h2, .markdown h3, .markdown h4 { margin: 22px 0 8px; line-height: 1.2; letter-spacing: 0; color: var(--ink); }
    .markdown h1 { font-size: 28px; } .markdown h2 { font-size: 24px; } .markdown h3 { font-size: 18px; color: var(--primary); } .markdown h4 { font-size: 15px; color: var(--green); }
    .markdown p { margin: 9px 0; color: var(--charcoal); }
    .markdown ul { margin: 10px 0 14px 20px; padding: 0; }
    .markdown li { margin: 6px 0; }
    pre { overflow: auto; background: var(--navy-deep); border: 1px solid var(--hairline); border-radius: 8px; padding: 14px; color: #e2e8f0; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; background: var(--surface); border: 1px solid var(--hairline-soft); border-radius: 6px; padding: 2px 6px; font-size: 13px; overflow-wrap: anywhere; }
    pre code { background: none; border: 0; padding: 0; color: inherit; }
    a { color: var(--link); }
    input[type="checkbox"] { accent-color: var(--primary); }
    .empty { color: var(--slate); border-left: 3px solid var(--orange); padding-left: 12px; }
    @media (max-width: 820px) { body { background: linear-gradient(180deg, var(--navy) 0, var(--navy-deep) 300px, var(--surface) 300px); } main { min-height: 100vh; padding: 32px 16px 56px; } .hero { text-align: left; padding: 0 0 24px; } h1 { font-size: 36px; line-height: 1.12; } .workspace { min-height: 70vh; } .tabs label { min-width: 112px; } .panels { padding: 14px; } .panel { padding: 18px; } .panel-head { display: grid; } h2 { font-size: 24px; } }
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
		fmt.Fprintf(&out, ".tabs:has(#tab-%d:checked) label[for=\"tab-%d\"] { background: var(--lavender); border-color: var(--primary); color: var(--ink); }\n", i, i)
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
		boldStart := strings.Index(text, "**")
		next := firstInlineToken(codeStart, linkStart, boldStart)
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
		if strings.HasPrefix(text, "**") {
			end := strings.Index(text[2:], "**")
			if end < 0 {
				out.WriteString(html.EscapeString(text))
				break
			}
			content := text[2 : end+2]
			fmt.Fprintf(&out, "<strong>%s</strong>", renderInline(content))
			text = text[end+4:]
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
