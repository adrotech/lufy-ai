package lufysdd

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const overviewFile = "change-overview.html"

type overviewArtifact struct {
	Title   string
	Path    string
	Content string
}

func (s Service) renderOverview(root, change string, mode Mode) error {
	changeRoot, err := platform.SafeJoin(root, filepath.Join("changes", change))
	if err != nil {
		return err
	}
	artifacts, err := collectOverviewArtifacts(changeRoot, mode)
	if err != nil {
		return err
	}
	output, err := platform.SafeJoin(changeRoot, overviewFile)
	if err != nil {
		return err
	}
	if info, statErr := os.Lstat(output); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("overview no es un archivo regular: %s", output)
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	body := renderOverviewHTML(change, mode, artifacts)
	write := s.write
	if write == nil {
		write = atomicWrite
	}
	if err := write(output, []byte(body), 0o644); err != nil {
		return fmt.Errorf("no se pudo materializar %s: %w", overviewFile, err)
	}
	return nil
}

func collectOverviewArtifacts(changeRoot string, mode Mode) ([]overviewArtifact, error) {
	base := []struct {
		title string
		rel   string
	}{
		{title: "Proposal", rel: "proposal.md"},
	}
	if mode == ModeFull {
		base = append(base, struct {
			title string
			rel   string
		}{title: "Design", rel: "design.md"})
	}
	base = append(base, struct {
		title string
		rel   string
	}{title: "Tasks", rel: "tasks.md"})

	artifacts := make([]overviewArtifact, 0, len(base)+1)
	for _, item := range base {
		body, err := readRegular(changeRoot, item.rel)
		if err != nil {
			continue
		}
		artifacts = append(artifacts, overviewArtifact{Title: item.title, Path: filepath.ToSlash(item.rel), Content: string(body)})
	}
	if mode != ModeFull {
		return artifacts, nil
	}

	specRoot, err := platform.SafeJoin(changeRoot, "specs")
	if err != nil {
		return nil, err
	}
	if err := requireDirectory(specRoot); err != nil {
		return artifacts, nil
	}
	entries, err := os.ReadDir(specRoot)
	if os.IsNotExist(err) {
		return artifacts, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if !entry.IsDir() || !safeID.MatchString(entry.Name()) {
			continue
		}
		rel := filepath.Join("specs", entry.Name(), "spec.md")
		body, readErr := readRegular(changeRoot, rel)
		if readErr != nil {
			continue
		}
		artifacts = append(artifacts, overviewArtifact{Title: "Spec: " + entry.Name(), Path: filepath.ToSlash(rel), Content: string(body)})
	}
	return artifacts, nil
}

func renderOverviewHTML(change string, mode Mode, artifacts []overviewArtifact) string {
	var navigation strings.Builder
	var sections strings.Builder
	for index, artifact := range artifacts {
		id := fmt.Sprintf("artifact-%d", index+1)
		fmt.Fprintf(&navigation, `<a href="#%s"><strong>%s</strong><small>%s</small></a>`, id, html.EscapeString(artifact.Title), html.EscapeString(artifact.Path))
		fmt.Fprintf(&sections, `<article id="%s"><header><div><p class="eyebrow">%s</p><h2>%s</h2></div><span class="badge">Markdown</span></header><div class="markdown">%s</div></article>`, id, html.EscapeString(artifact.Path), html.EscapeString(artifact.Title), overviewMarkdownHTML(artifact.Content))
	}
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + html.EscapeString(change) + ` · Lufy SDD</title>
  <style>
    :root { color-scheme: light; --navy:#0a1530; --deep:#070f24; --paper:#fff; --canvas:#f5f4f2; --line:#e4e1dc; --ink:#24231f; --muted:#6d6a63; --violet:#5b48d6; --green:#138a36; --lavender:#ece8ff; }
    * { box-sizing:border-box; }
    html { scroll-behavior:smooth; }
    body { margin:0; color:var(--ink); background:linear-gradient(180deg,var(--navy) 0,var(--deep) 310px,var(--canvas) 310px); font:16px/1.6 Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; }
    main { width:min(1180px,calc(100% - 32px)); margin:auto; padding:46px 0 72px; }
    .hero { color:#fff; text-align:center; margin-bottom:28px; }
    .hero p { margin:0; color:#b9c2da; font-size:13px; font-weight:700; letter-spacing:.08em; text-transform:uppercase; }
    h1 { margin:8px 0 14px; font-size:clamp(34px,6vw,58px); line-height:1.08; }
    .mode { display:inline-block; padding:5px 12px; border:1px solid #56617b; border-radius:999px; color:#dce3f4; font-size:13px; }
    .workspace { display:grid; grid-template-columns:240px minmax(0,1fr); gap:20px; align-items:start; }
    nav { position:sticky; top:18px; display:grid; gap:8px; padding:12px; border:1px solid var(--line); border-radius:14px; background:var(--paper); box-shadow:0 16px 35px rgba(6,13,30,.16); }
    nav a { display:grid; gap:2px; padding:10px 12px; border-radius:9px; color:var(--ink); text-decoration:none; }
    nav a:hover { background:var(--lavender); }
    nav small,.eyebrow { color:var(--muted); font:12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace; overflow-wrap:anywhere; }
    .content { display:grid; gap:18px; }
    article { scroll-margin-top:18px; padding:26px; border:1px solid var(--line); border-radius:14px; background:var(--paper); box-shadow:0 10px 28px rgba(20,18,14,.08); }
    article > header { display:flex; justify-content:space-between; gap:16px; align-items:start; padding-bottom:14px; border-bottom:1px solid var(--line); }
    .eyebrow { margin:0 0 4px; }
    h2 { margin:0; font-size:28px; line-height:1.25; }
    .badge { padding:4px 10px; border-radius:999px; background:var(--lavender); color:var(--violet); font-size:12px; font-weight:700; }
    .markdown { margin-top:18px; }
    .markdown h1,.markdown h2,.markdown h3,.markdown h4 { margin:22px 0 8px; line-height:1.25; }
    .markdown h1 { font-size:28px; }.markdown h2 { font-size:23px; }.markdown h3 { color:var(--violet); font-size:19px; }.markdown h4 { color:var(--green); font-size:16px; }
    .markdown p { margin:8px 0; }.markdown ul { margin:9px 0 14px; padding-left:24px; }.markdown li { margin:5px 0; }
    code { padding:2px 6px; border:1px solid var(--line); border-radius:5px; background:var(--canvas); font:13px ui-monospace,SFMono-Regular,Menlo,monospace; }
    pre { overflow:auto; padding:15px; border-radius:9px; background:var(--deep); color:#e6ebf7; } pre code { padding:0; border:0; background:none; color:inherit; }
    input[type="checkbox"] { accent-color:var(--violet); }
    @media (max-width:760px) { main { padding-top:30px; }.workspace { grid-template-columns:1fr; }nav { position:static; grid-template-columns:repeat(auto-fit,minmax(130px,1fr)); }article { padding:20px; } }
  </style>
</head>
<body>
  <main>
    <header class="hero"><p>Lufy SDD Change Overview</p><h1>` + html.EscapeString(change) + `</h1><span class="mode">mode: ` + html.EscapeString(string(mode)) + `</span></header>
    <div class="workspace"><nav aria-label="Artifacts">` + navigation.String() + `</nav><section class="content">` + sections.String() + `</section></div>
  </main>
</body>
</html>
`
}

func overviewMarkdownHTML(markdown string) string {
	if strings.TrimSpace(markdown) == "" {
		return `<p>Artifact vacío.</p>`
	}
	var output strings.Builder
	inList := false
	inCode := false
	closeList := func() {
		if inList {
			output.WriteString("</ul>")
			inList = false
		}
	}
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			closeList()
			if inCode {
				output.WriteString("</code></pre>")
			} else {
				output.WriteString("<pre><code>")
			}
			inCode = !inCode
			continue
		}
		if inCode {
			output.WriteString(html.EscapeString(line))
			output.WriteByte('\n')
			continue
		}
		if trimmed == "" {
			closeList()
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "#### "):
			closeList()
			fmt.Fprintf(&output, "<h4>%s</h4>", overviewInline(trimmed[5:]))
		case strings.HasPrefix(trimmed, "### "):
			closeList()
			fmt.Fprintf(&output, "<h3>%s</h3>", overviewInline(trimmed[4:]))
		case strings.HasPrefix(trimmed, "## "):
			closeList()
			fmt.Fprintf(&output, "<h2>%s</h2>", overviewInline(trimmed[3:]))
		case strings.HasPrefix(trimmed, "# "):
			closeList()
			fmt.Fprintf(&output, "<h1>%s</h1>", overviewInline(trimmed[2:]))
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			if !inList {
				output.WriteString("<ul>")
				inList = true
			}
			fmt.Fprintf(&output, "<li>%s</li>", overviewListItem(strings.TrimSpace(trimmed[2:])))
		default:
			closeList()
			fmt.Fprintf(&output, "<p>%s</p>", overviewInline(trimmed))
		}
	}
	closeList()
	if inCode {
		output.WriteString("</code></pre>")
	}
	return output.String()
}

func overviewListItem(item string) string {
	checked := false
	checkbox := false
	for _, marker := range []string{"[ ] ", "[x] ", "[X] "} {
		if strings.HasPrefix(item, marker) {
			checkbox = true
			checked = marker != "[ ] "
			item = strings.TrimSpace(item[len(marker):])
			break
		}
	}
	if !checkbox {
		return overviewInline(item)
	}
	state := ""
	if checked {
		state = " checked"
	}
	return fmt.Sprintf(`<input type="checkbox" disabled%s> %s`, state, overviewInline(item))
}

func overviewInline(value string) string {
	escaped := html.EscapeString(value)
	for strings.Contains(escaped, "**") {
		start := strings.Index(escaped, "**")
		endOffset := strings.Index(escaped[start+2:], "**")
		if endOffset < 0 {
			break
		}
		end := start + 2 + endOffset
		escaped = escaped[:start] + "<strong>" + escaped[start+2:end] + "</strong>" + escaped[end+2:]
	}
	for strings.Contains(escaped, "`") {
		start := strings.Index(escaped, "`")
		endOffset := strings.Index(escaped[start+1:], "`")
		if endOffset < 0 {
			break
		}
		end := start + 1 + endOffset
		escaped = escaped[:start] + "<code>" + escaped[start+1:end] + "</code>" + escaped[end+1:]
	}
	return escaped
}
