package assets

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestBuildCatalogExpandsManagedAssetsAndExcludesOpenSpecChanges(t *testing.T) {
	source := minimalSource(t)
	catalog, err := BuildCatalog(source)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}

	want := map[string]bool{
		"lufy-ia.harness.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           false,
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            false,
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         false,
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         false,
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): false,
		filepath.Join(".opencode", "templates", "result-contract.md"):                     false,
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            false,
		filepath.Join("openspec", "config.yaml"):                                          false,
		filepath.Join("openspec", "UPSTREAM.json"):                                        false,
		filepath.Join(".lufy", "README.md"):                                               false,
		filepath.Join(".agents", "skills", "lufy-close", "SKILL.md"):                      false,
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"):                    false,
		filepath.Join(".codex", "config.toml"):                                            false,
		filepath.Join(".codex", "agents", "implementer.toml"):                             false,
		filepath.Join(".lufy", "workflows", "sdd", "README.md"):                           false,
		filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep"):                 false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/memory/") {
			t.Fatalf("catalog included user-owned memory asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("file asset %s missing hash", asset.TargetRel)
			}
			if expectedTool(asset.TargetRel) != asset.Tool || asset.Component == "" {
				t.Fatalf("asset %s missing ownership metadata: %#v", asset.TargetRel, asset)
			}
			if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), "openspec/") && asset.Methodology != domain.MethodologySpecWorkflow {
				t.Fatalf("openspec asset %s methodology = %s", asset.TargetRel, asset.Methodology)
			}
			if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/workflows/sdd/") && asset.Methodology != domain.MethodologyLufyWorkflow {
				t.Fatalf("lufy-sdd asset %s methodology = %s", asset.TargetRel, asset.Methodology)
			}
		}
	}
	for path, found := range want {
		if !found {
			t.Fatalf("catalog missing %s", path)
		}
	}
}

func TestBuildCatalogRejectsSourceSymlink(t *testing.T) {
	source := minimalSource(t)
	if err := os.Remove(filepath.Join(source, ".opencode", "agents", "orchestrator.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(source, "tui.json"), filepath.Join(source, ".opencode", "agents", "orchestrator.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if _, err := BuildCatalog(source); err == nil {
		t.Fatal("BuildCatalog() expected symlink error")
	}
}

func TestBuildEmbeddedCatalogIncludesManagedAssetsAndExcludesOpenSpecChanges(t *testing.T) {
	catalog, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatalf("BuildEmbeddedCatalog() error = %v", err)
	}
	if catalog.SourceRoot != EmbeddedSourceRoot {
		t.Fatalf("SourceRoot = %q, want %q", catalog.SourceRoot, EmbeddedSourceRoot)
	}

	want := map[string]bool{
		"lufy-ia.harness.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           false,
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            false,
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         false,
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         false,
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): false,
		filepath.Join(".opencode", "templates", "result-contract.md"):                     false,
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            false,
		filepath.Join("openspec", "config.yaml"):                                          false,
		filepath.Join("openspec", "UPSTREAM.json"):                                        false,
		filepath.Join(".lufy", "README.md"):                                               false,
		filepath.Join(".agents", "skills", "lufy-close", "SKILL.md"):                      false,
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"):                    false,
		filepath.Join(".codex", "config.toml"):                                            false,
		filepath.Join(".codex", "agents", "implementer.toml"):                             false,
		filepath.Join(".lufy", "workflows", "sdd", "README.md"):                           false,
		filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep"):                 false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("embedded catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/memory/") {
			t.Fatalf("embedded catalog included user-owned memory asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("embedded file asset %s missing hash", asset.TargetRel)
			}
			if expectedTool(asset.TargetRel) != asset.Tool || asset.Component == "" {
				t.Fatalf("embedded asset %s missing ownership metadata: %#v", asset.TargetRel, asset)
			}
		}
	}
	for path, found := range want {
		if !found {
			t.Fatalf("embedded catalog missing %s", path)
		}
	}
}

func TestCatalogFingerprintIsStableForSameFileAssets(t *testing.T) {
	catalog := Catalog{Assets: []Asset{
		{TargetRel: "b.txt", Kind: KindFile, SourceSHA256: "bbb"},
		{TargetRel: "dir", Kind: KindDir},
		{TargetRel: "a.txt", Kind: KindFile, SourceSHA256: "aaa"},
	}}
	reordered := Catalog{Assets: []Asset{
		{TargetRel: "a.txt", Kind: KindFile, SourceSHA256: "aaa"},
		{TargetRel: "b.txt", Kind: KindFile, SourceSHA256: "bbb"},
	}}
	one, err := catalog.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	two, err := reordered.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if one == "" || one != two {
		t.Fatalf("fingerprints mismatch: %q != %q", one, two)
	}
}

func TestEmbeddedCatalogMatchesRepositoryAssets(t *testing.T) {
	root := repoRoot(t)
	rootCatalog, err := BuildCatalog(root)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	embeddedCatalog, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatalf("BuildEmbeddedCatalog() error = %v", err)
	}
	rootAssets := comparableAssets(rootCatalog)
	embeddedAssets := comparableAssets(embeddedCatalog)
	if !reflect.DeepEqual(rootAssets, embeddedAssets) {
		t.Fatalf("root and embedded catalogs drifted\nroot=%#v\nembedded=%#v", rootAssets, embeddedAssets)
	}
}

func TestAgentAssetsContainT2FastPathApprovalGate(t *testing.T) {
	root := repoRoot(t)
	cases := []struct {
		path string
		want []string
	}{
		{
			path: filepath.Join(".opencode", "agents", "orchestrator.md"),
			want: []string{"fast_path_allowed: false", "post-plan user confirmation", "next_recommended.owner: implementer"},
		},
		{
			path: filepath.Join(".opencode", "agents", "implementer.md"),
			want: []string{"fast_path_allowed: false", "approved implementation after seeing a visible plan", "blocked` or `needs_decision"},
		},
		{
			path: filepath.Join(".codex", "agents", "orchestrator.toml"),
			want: []string{"fast_path_allowed=false", "explicit user approval", "next owner implementer or auto-chain is not approval"},
		},
		{
			path: filepath.Join(".codex", "agents", "implementer.toml"),
			want: []string{"fast_path_allowed=false", "explicit post-plan user approval", "blocked/needs_decision"},
		},
		{
			path: filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "orchestrator.md"),
			want: []string{"fast_path_allowed: false", "post-plan user confirmation", "next_recommended.owner: implementer"},
		},
		{
			path: filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "implementer.md"),
			want: []string{"fast_path_allowed: false", "approved implementation after seeing a visible plan", "blocked` or `needs_decision"},
		},
		{
			path: filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".codex", "agents", "orchestrator.toml"),
			want: []string{"fast_path_allowed=false", "explicit user approval", "next owner implementer or auto-chain is not approval"},
		},
		{
			path: filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".codex", "agents", "implementer.toml"),
			want: []string{"fast_path_allowed=false", "explicit post-plan user approval", "blocked/needs_decision"},
		},
	}

	for _, tc := range cases {
		t.Run(filepath.ToSlash(tc.path), func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join(root, tc.path))
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			text := string(body)
			for _, want := range tc.want {
				if !strings.Contains(text, want) {
					t.Fatalf("%s missing %q", tc.path, want)
				}
			}
		})
	}
}

func TestAgentAssetsCoverT2ApprovalConversationScenarios(t *testing.T) {
	root := repoRoot(t)
	read := func(rel string) string {
		t.Helper()
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", rel, err)
		}
		return string(body)
	}

	orchestrator := read(filepath.Join(".opencode", "agents", "orchestrator.md"))
	implementer := read(filepath.Join(".opencode", "agents", "implementer.md"))
	spec := read(filepath.Join("openspec", "specs", "sdd-harness-routing", "spec.md"))
	embeddedOrchestrator := read(filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "orchestrator.md"))
	embeddedImplementer := read(filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "implementer.md"))
	embeddedSpec := read(filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", "openspec", "specs", "sdd-harness-routing", "spec.md"))

	scenarios := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "orchestrator pauses before implementer without post-plan approval",
			text: orchestrator,
			want: []string{
				"Treat `T2` / `sdd_lite` feature or runtime/app changes with `fast_path_allowed: false` as an explicit user decision gate",
				"Before invoking `implementer`, present a visible SDD Lite plan",
				"Do not interpret `next_recommended.owner: implementer`",
				"`chain_strategy: auto-chain` as approval",
			},
		},
		{
			name: "orchestrator allows implementer only after explicit approval",
			text: orchestrator,
			want: []string{
				"Continue to `implementer` only after a post-plan user confirmation",
				"\"sí, implementa\"",
			},
		},
		{
			name: "implementer blocks incomplete handoff",
			text: implementer,
			want: []string{
				"do not edit until the handoff includes evidence",
				"approved implementation after seeing a visible plan",
				"return `blocked` or `needs_decision`",
			},
		},
		{
			name: "spec documents orchestrator regression",
			text: spec,
			want: []string{
				"T2 SDD Lite runtime work requires post-plan approval",
				"`orchestrator` SHALL NOT invoke `implementer`",
				"phrases that only express intent to generate or explore a feature SHALL NOT count",
			},
		},
		{
			name: "spec documents implementer regression",
			text: spec,
			want: []string{
				"Implementer blocks missing post-plan approval",
				"`implementer` SHALL return `blocked` or `needs_decision` instead of mutating the working tree",
			},
		},
		{
			name: "embedded orchestrator keeps same gate",
			text: embeddedOrchestrator,
			want: []string{
				"Treat `T2` / `sdd_lite` feature or runtime/app changes with `fast_path_allowed: false` as an explicit user decision gate",
				"Continue to `implementer` only after a post-plan user confirmation",
			},
		},
		{
			name: "embedded implementer keeps same defense",
			text: embeddedImplementer,
			want: []string{
				"do not edit until the handoff includes evidence",
				"return `blocked` or `needs_decision`",
			},
		},
		{
			name: "embedded spec keeps same regression",
			text: embeddedSpec,
			want: []string{
				"T2 SDD Lite runtime work requires post-plan approval",
				"`orchestrator` SHALL NOT invoke `implementer`",
				"Implementer blocks missing post-plan approval",
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for _, want := range scenario.want {
				if !strings.Contains(scenario.text, want) {
					t.Fatalf("scenario %q missing %q", scenario.name, want)
				}
			}
		})
	}
}

func TestAgentAssetsRequireRouterForSecuritySensitivePrompts(t *testing.T) {
	root := repoRoot(t)
	read := func(rel string) string {
		t.Helper()
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", rel, err)
		}
		return string(body)
	}

	cases := []struct {
		name string
		rel  string
		want []string
	}{
		{
			name: "opencode orchestrator routes sensitive prompts through router",
			rel:  filepath.Join(".opencode", "agents", "orchestrator.md"),
			want: []string{
				"Use `sdd-router` before implementation for security-sensitive runtime or global configuration requests",
				"CORS, authentication, authorization, JWT",
				"Do not classify these requests as direct `T3` from `orchestrator`",
				"documentation, tests, fixtures, comments, or a non-runtime/non-config mechanical change",
			},
		},
		{
			name: "embedded opencode orchestrator routes sensitive prompts through router",
			rel:  filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "orchestrator.md"),
			want: []string{
				"Use `sdd-router` before implementation for security-sensitive runtime or global configuration requests",
				"CORS, authentication, authorization, JWT",
				"Do not classify these requests as direct `T3` from `orchestrator`",
				"documentation, tests, fixtures, comments, or a non-runtime/non-config mechanical change",
			},
		},
		{
			name: "opencode router classifies sensitive prompts as non fast path",
			rel:  filepath.Join(".opencode", "agents", "sdd-router.md"),
			want: []string{
				"Security-sensitive routing guardrail",
				"CORS, authentication, authorization, JWT",
				"are not direct T3",
				"Set `fast_path_allowed: false` by default",
				"T3 is allowed only for explicitly non-runtime/non-config documentation, tests, fixtures, comments, or mechanical updates",
			},
		},
		{
			name: "embedded opencode router classifies sensitive prompts as non fast path",
			rel:  filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".opencode", "agents", "sdd-router.md"),
			want: []string{
				"Security-sensitive routing guardrail",
				"CORS, authentication, authorization, JWT",
				"are not direct T3",
				"Set `fast_path_allowed: false` by default",
				"T3 is allowed only for explicitly non-runtime/non-config documentation, tests, fixtures, comments, or mechanical updates",
			},
		},
		{
			name: "codex orchestrator routes sensitive prompts through router",
			rel:  filepath.Join(".codex", "agents", "orchestrator.toml"),
			want: []string{
				"Route security-sensitive runtime/global-config requests through sdd-router before implementation",
				"CORS, auth, authorization, JWT",
				"Do not classify them as direct T3 unless explicitly limited to docs/tests/fixtures/comments or non-runtime mechanical work",
			},
		},
		{
			name: "embedded codex orchestrator routes sensitive prompts through router",
			rel:  filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".codex", "agents", "orchestrator.toml"),
			want: []string{
				"Route security-sensitive runtime/global-config requests through sdd-router before implementation",
				"CORS, auth, authorization, JWT",
				"Do not classify them as direct T3 unless explicitly limited to docs/tests/fixtures/comments or non-runtime mechanical work",
			},
		},
		{
			name: "codex router marks sensitive prompts non direct T3",
			rel:  filepath.Join(".codex", "agents", "sdd-router.toml"),
			want: []string{
				"CORS, auth, authorization, JWT",
				"do not classify as direct T3",
				"set fast_path_allowed=false by default",
				"explicitly non-runtime docs/tests/fixtures/comments/mechanical work",
			},
		},
		{
			name: "embedded codex router marks sensitive prompts non direct T3",
			rel:  filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", ".codex", "agents", "sdd-router.toml"),
			want: []string{
				"CORS, auth, authorization, JWT",
				"do not classify as direct T3",
				"set fast_path_allowed=false by default",
				"explicitly non-runtime docs/tests/fixtures/comments/mechanical work",
			},
		},
		{
			name: "openspec captures sensitive routing scenarios",
			rel:  filepath.Join("openspec", "specs", "sdd-harness-routing", "spec.md"),
			want: []string{
				"Security-sensitive runtime request is not direct T3",
				"`orchestrator` SHALL route to `sdd-router` before implementation",
				"`sdd-router` SHALL set `fast_path_allowed: false` by default and classify at least T2",
				"Security keyword documentation-only exception",
			},
		},
		{
			name: "embedded openspec captures sensitive routing scenarios",
			rel:  filepath.Join("tools", "lufy-cli-go", "internal", "assets", "embedded", "openspec", "specs", "sdd-harness-routing", "spec.md"),
			want: []string{
				"Security-sensitive runtime request is not direct T3",
				"`orchestrator` SHALL route to `sdd-router` before implementation",
				"`sdd-router` SHALL set `fast_path_allowed: false` by default and classify at least T2",
				"Security keyword documentation-only exception",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := read(tc.rel)
			for _, want := range tc.want {
				if !strings.Contains(text, want) {
					t.Fatalf("%s missing %q", tc.rel, want)
				}
			}
		})
	}
}

type comparableAsset struct {
	TargetRel    string
	Kind         Kind
	Policy       Policy
	Scope        Scope
	SourceSHA256 string
}

func comparableAssets(c Catalog) []comparableAsset {
	out := make([]comparableAsset, 0, len(c.Assets))
	for _, asset := range c.Assets {
		out = append(out, comparableAsset{TargetRel: filepath.ToSlash(asset.TargetRel), Kind: asset.Kind, Policy: asset.Policy, Scope: asset.Scope, SourceSHA256: asset.SourceSHA256})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TargetRel < out[j].TargetRel })
	return out
}

func expectedTool(targetRel string) domain.ToolID {
	target := filepath.ToSlash(targetRel)
	if strings.HasPrefix(target, ".agents/") || strings.HasPrefix(target, ".codex/") || target == ".codex" {
		return domain.ToolCodex
	}
	return domain.ToolInitialDefault
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if fileExists(filepath.Join(dir, "AGENTS.md")) && fileExists(filepath.Join(dir, "tools", "lufy-cli-go", "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func minimalSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":          "agents root\n",
		"AGENTS.md.template": "agents template\n",
		"lufy-ia.harness.md": "agents template\n",
		"tui.json":           "{}\n",
		filepath.Join(".agents", "skills", "lufy-close", "SKILL.md"):                      "close skill\n",
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"):                    "sdd skill\n",
		filepath.Join(".codex", "README.md"):                                              "codex readme\n",
		filepath.Join(".codex", "config.toml"):                                            "project_doc_max_bytes = 32768\n\n[features]\nmulti_agent = true\n",
		filepath.Join(".codex", "lufy-agent-mapping.md"):                                  "agent_execution_mode\n",
		filepath.Join(".codex", "agents", "implementer.toml"):                             "name = \"implementer\"\n",
		filepath.Join(".codex", "hooks.json"):                                             "{\"hooks\":{}}\n",
		filepath.Join(".codex", "rules", "lufy.rules"):                                    "# rules\n",
		filepath.Join(".opencode", ".gitignore"):                                          "node_modules\n",
		filepath.Join(".opencode", "README.md"):                                           "readme\n",
		filepath.Join(".opencode", "package.json"):                                        "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                                   "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):                           "apply\n",
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            "sync\n",
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         "version\n",
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         "hook\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):                      "skill\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): "sync skill\n",
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            "lite\n",
		filepath.Join(".opencode", "templates", "result-contract.md"):                     "result\n",
		filepath.Join(".opencode", "policies", "delivery.md"):                             "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"):                    "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):                       "state\n",
		filepath.Join("openspec", "config.yaml"):                                          "config\n",
		filepath.Join("openspec", "UPSTREAM.json"):                                        "{}\n",
		filepath.Join("openspec", "README.md"):                                            "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                                    "",
		filepath.Join(".lufy", "README.md"):                                               "layout\n",
		filepath.Join(".lufy", "sdd", "README.md"):                                        "lufy-sdd\n",
		filepath.Join(".lufy", "sdd", "changes", ".gitkeep"):                              "",
		filepath.Join(".lufy", "sdd", "decisions", ".gitkeep"):                            "",
		filepath.Join(".lufy", "sdd", "specs", ".gitkeep"):                                "",
		filepath.Join(".lufy", "sdd", "verification", ".gitkeep"):                         "",
		filepath.Join(".lufy", "memory", "knowledge", "private.md"):                       "private\n",
		filepath.Join("tools", "lufy-cli-go", "go.mod"):                                   "module github.com/adrianrojas/lufy-ai/tools/lufy-cli-go\n",
		filepath.Join("openspec", "changes", "active", "proposal.md"):                     "must not copy\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
