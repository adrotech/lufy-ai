package verify

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestVerifyDetectsMissingAndHashMismatch(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range fallbackRequiredManagedFiles {
		content := verifyFileContent(rel)
		writeVerifyFile(t, filepath.Join(target, rel), content)
	}
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "# Proyecto\n\n@lufy-ia.harness.md\n")
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == "tui.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(valid) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok: verify estructural completo") {
		t.Fatalf("valid verify output unexpected: %s", out.String())
	}

	writeVerifyFile(t, filepath.Join(target, "lufy-ia.harness.md"), "drift\n")
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(drift) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: drift en lufy-ia.harness.md") {
		t.Fatalf("drift output unexpected: %s", out.String())
	}

	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "plugins", "agent-observatory.tsx"))); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(missing) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta archivo crítico") {
		t.Fatalf("missing output unexpected: %s", out.String())
	}
}

func TestCheckBuilderFeedsSameReportToJSONPresenter(t *testing.T) {
	target := validVerifyTarget(t)
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		t.Fatal(err)
	}
	report := Report{TargetRoot: resolved, Scope: string(assets.ScopeProject)}

	if err := (CheckBuilder{}).Build(Options{Target: resolved, NoEngram: true, Scope: assets.ScopeProject}, &report); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if report.Failures != 0 || len(report.Checks) == 0 {
		t.Fatalf("expected complete report model without failures, got %#v", report)
	}
	var out bytes.Buffer
	if err := (ReportPresenter{}).Present(report, Options{JSON: true}, &out, nil); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	var rendered Report
	if err := json.Unmarshal(out.Bytes(), &rendered); err != nil {
		t.Fatal(err)
	}
	if !rendered.OK || len(rendered.Checks) != len(report.Checks) {
		t.Fatalf("presenter did not render same report model: %#v", rendered)
	}
}

func TestVerifyWarnsForNoReplaceDriftWithLufyNew(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeVerifyFile(t, filepath.Join(target, "tui.json.lufy-new"), "{\"new\":true}\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true, JSON: true}, &out); err != nil {
		t.Fatalf("Run() error = %v body=%s", err, out.String())
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.OK || report.Warnings == 0 {
		t.Fatalf("expected ok report with warning, got %#v", report)
	}
	found := false
	for _, check := range report.Checks {
		if check.Path == "tui.json" && check.Policy == "no-replace" && check.RecommendedAction == "review-lufy-new" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing no-replace lufy-new check: %#v", report.Checks)
	}
}

func TestVerifyDetectsMissingCriticalDirectoryAndManifestEntry(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range fallbackRequiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		if rel == "tui.json" {
			continue
		}
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == "tui.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "skills"))); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid structure) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta directorio crítico: "+filepath.Join(".opencode", "skills")) {
		t.Fatalf("missing directory output unexpected: %s", out.String())
	}
	if !strings.Contains(out.String(), "fail: asset clave no está en manifest: tui.json") {
		t.Fatalf("missing manifest output unexpected: %s", out.String())
	}
}

func TestVerifyDetectsMissingTemplatesDirectory(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range fallbackRequiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: "managed", Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "templates"))); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(missing templates) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta directorio crítico: "+filepath.Join(".opencode", "templates")) {
		t.Fatalf("missing templates output unexpected: %s", out.String())
	}
}

func TestVerifyFailsCorruptManifestAndMovedTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(target), []byte("{bad-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "install-state.json inválido") {
		t.Fatalf("expected corrupt manifest error, got %v", err)
	}

	actual := t.TempDir()
	writeVerifyDirs(t, actual)
	for _, rel := range fallbackRequiredManagedFiles {
		content := verifyFileContent(rel)
		writeVerifyFile(t, filepath.Join(actual, rel), content)
	}
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(actual, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == "tui.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(actual, state.New(t.TempDir(), nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := NewService().Run(Options{Target: actual, NoEngram: true}, &out)
	if err == nil || !strings.Contains(out.String(), "targetRoot del manifest no coincide") {
		t.Fatalf("expected moved target failure, err=%v output=%s", err, out.String())
	}
}

func TestVerifyFailsInvalidTUIJSON(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range fallbackRequiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	writeVerifyFile(t, filepath.Join(target, "tui.json"), "tui.json\n")

	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == "tui.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid tui.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: JSON inválido en tui.json") {
		t.Fatalf("invalid tui.json output unexpected: %s", out.String())
	}
}

func TestVerifyFailsInvalidOrIncompleteMergeManagedOpenCode(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{bad-json`)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: JSON inválido en opencode.json") || !strings.Contains(out.String(), "estructura gestionada inválida") {
		t.Fatalf("invalid opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json"}`)
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(incomplete opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "falta clave gestionada mínima plugin") {
		t.Fatalf("incomplete opencode.json output unexpected: %s", out.String())
	}
}

func TestVerifyFailsUnsafeOrInvalidManagedOpenCodeStructure(t *testing.T) {
	target := validVerifyTarget(t)
	outside := filepath.Join(t.TempDir(), "opencode.json")
	writeVerifyFile(t, outside, `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
	if err := os.Remove(filepath.Join(target, "opencode.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, "opencode.json")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(symlink opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "opencode.json") || !strings.Contains(out.String(), "symlink no permitido") {
		t.Fatalf("symlink opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	if err := os.Remove(filepath.Join(target, "opencode.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(target, "opencode.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(directory opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "archivo regular seguro") {
		t.Fatalf("directory opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":123,"plugin":[]}`)
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid managed type opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "$schema debe ser string") {
		t.Fatalf("invalid managed type output unexpected: %s", out.String())
	}
}

func TestVerifyReportsExtraFilesInManagedDirsAsInfo(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, ".opencode", "agents", "local-agent.md"), "local\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run() error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "info: archivo extra en directorio gestionado: "+filepath.Join(".opencode", "agents", "local-agent.md")) {
		t.Fatalf("extra managed dir file not reported: %s", out.String())
	}
}

func TestVerifyJSONReportsFailures(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "sin referencia\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true, JSON: true}, &out); err == nil {
		t.Fatalf("Run(JSON drift) expected error, output=%s", out.String())
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v body=%s", err, out.String())
	}
	if report.OK || report.Failures == 0 || report.TargetRoot == "" || report.Assets == 0 {
		t.Fatalf("unexpected JSON report: %#v", report)
	}
	foundMissingReference := false
	for _, check := range report.Checks {
		if check.Level == "fail" && check.Path == "AGENTS.md" && strings.Contains(check.Message, "no referencia") {
			foundMissingReference = true
		}
	}
	if !foundMissingReference {
		t.Fatalf("missing reference check not found in JSON: %#v", report.Checks)
	}
}

func TestVerifyAllowMissingAgentsRefDoesNotHideCriticalFailures(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "sin referencia\n")
	writeVerifyFile(t, filepath.Join(target, "lufy-ia.harness.md"), "drift crítico\n")

	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, NoEngram: true, AllowMissingAgentsRef: true, JSON: true}, &out)
	if err == nil {
		t.Fatalf("Run(allow missing agents ref with critical drift) expected error, output=%s", out.String())
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v body=%s", err, out.String())
	}
	foundAgentsWarn := false
	foundHarnessFail := false
	for _, check := range report.Checks {
		if check.Level == "warn" && check.Path == "AGENTS.md" && strings.Contains(check.Message, "no referencia") {
			foundAgentsWarn = true
		}
		if check.Level == "fail" && strings.Contains(check.Message, "drift en lufy-ia.harness.md") {
			foundHarnessFail = true
		}
	}
	if report.OK || report.Failures == 0 || !foundAgentsWarn || !foundHarnessFail {
		t.Fatalf("allow missing AGENTS ref masked critical failure: report=%#v", report)
	}
}

func TestVerifyQuietSuppressesHumanOutput(t *testing.T) {
	target := validVerifyTarget(t)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true, Quiet: true}, &out); err != nil {
		t.Fatalf("Run(quiet) error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("quiet output not suppressed: %s", out.String())
	}
}

func TestVerifyVerboseReportsDerivedRequirements(t *testing.T) {
	target := validVerifyTarget(t)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true, Verbose: true}, &out); err != nil {
		t.Fatalf("Run(verbose) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "info: requirements derivados") {
		t.Fatalf("verbose requirements info missing: %s", out.String())
	}
}

func TestVerifyDeepValidatesPluginReferences(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "tui.json"), `{"plugin":["./.opencode/plugins/agent-observatory.tsx"]}`)
	refreshVerifyAssetHash(t, target, "tui.json")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true, Deep: true}, &out); err != nil {
		t.Fatalf("Run(deep valid) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "plugin referenciado por tui.json") {
		t.Fatalf("deep output missing plugin ok: %s", out.String())
	}

	writeVerifyFile(t, filepath.Join(target, "tui.json"), `{"plugin":["../evil.ts"]}`)
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true, Deep: true}, &out); err == nil {
		t.Fatalf("Run(deep invalid) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "plugin path inseguro") {
		t.Fatalf("deep invalid output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "tui.json"), `{"plugin":{}}`)
	refreshVerifyAssetHash(t, target, "tui.json")
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true, Deep: true}, &out); err == nil {
		t.Fatalf("Run(deep non-array plugin) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "plugin debe ser array") {
		t.Fatalf("deep non-array output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "tui.json"), `{"plugin":[123,"./.opencode/plugins/missing.ts"]}`)
	refreshVerifyAssetHash(t, target, "tui.json")
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true, Deep: true}, &out); err == nil {
		t.Fatalf("Run(deep bad plugin entries) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "plugin contiene entrada no string") || !strings.Contains(out.String(), "plugin referenciado no existe") {
		t.Fatalf("deep bad entries output unexpected: %s", out.String())
	}
}

func TestCatalogRequirementsIncludeRegisteredCatalogAssets(t *testing.T) {
	st := state.NewWithHarness(t.TempDir(), nil, []state.AssetState{
		{TargetRel: filepath.Join(".opencode", "agents", "orchestrator.md")},
	}, "test", domain.DefaultHarnessConfig())
	dirs, files := catalogRequirements(&st)
	if !containsString(files, filepath.Join(".opencode", "agents", "orchestrator.md")) {
		t.Fatalf("catalog registered file not required: %#v", files)
	}
	if !containsString(dirs, filepath.Join(".opencode", "agents")) {
		t.Fatalf("catalog registered file parent dir not required: %#v", dirs)
	}
}

func validVerifyTarget(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	writeVerifyDirs(t, target)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
	for _, rel := range fallbackRequiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "# Proyecto\n\n@lufy-ia.harness.md\n")
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == "tui.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	return target
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func refreshVerifyAssetHash(t *testing.T, target, rel string) {
	t.Helper()
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := assets.FileSHA256(filepath.Join(target, rel))
	if err != nil {
		t.Fatal(err)
	}
	for i := range st.Assets {
		if st.Assets[i].TargetRel == rel {
			st.Assets[i].SourceSHA256 = hash
			st.Assets[i].TargetSHA256 = hash
		}
	}
	if err := state.WriteAtomic(target, *st); err != nil {
		t.Fatal(err)
	}
}

func writeVerifyDirs(t *testing.T, target string) {
	t.Helper()
	for _, rel := range fallbackRequiredDirs {
		if err := os.MkdirAll(filepath.Join(target, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
}

func writeVerifyFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func verifyFileContent(rel string) string {
	if rel == "tui.json" {
		return "{}\n"
	}
	if rel == filepath.Join("openspec", "UPSTREAM.json") {
		return "{}\n"
	}
	return rel + "\n"
}
