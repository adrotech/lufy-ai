package verify

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/skillregistry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestVerifyDetectsMissingAndHashMismatch(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range fallbackRequiredManagedFiles {
		content := verifyFileContent(rel)
		writeVerifyFile(t, filepath.Join(target, rel), content)
	}
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), string(agentsref.MinimalContent()))
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == ".codex/hooks.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err != nil {
		t.Fatalf("Run(valid) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok: verify estructural completo") {
		t.Fatalf("valid verify output unexpected: %s", out.String())
	}

	writeVerifyFile(t, filepath.Join(target, "lufy-ia.harness.md"), "drift\n")
	out.Reset()
	if err := NewService().Run(Options{Target: target}, &out); err == nil {
		t.Fatalf("Run(drift) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: guardrail drift en lufy-ia.harness.md") {
		t.Fatalf("drift output unexpected: %s", out.String())
	}

	if err := os.Remove(filepath.Join(target, filepath.Join(".codex", "hooks.json"))); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target}, &out); err == nil {
		t.Fatalf("Run(missing) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta archivo crítico") {
		t.Fatalf("missing output unexpected: %s", out.String())
	}
}

func TestVerifyFailsWhenProjectHarnessAndInstallStateDisagree(t *testing.T) {
	target := validVerifyTarget(t)
	cfg := projectconfig.ProjectConfig{
		SchemaVersion: projectconfig.SchemaVersion,
		Tool:          domain.ToolCodex,
		Execution:     domain.ExecutionConfig{Mode: domain.ExecutionInline, Delegation: domain.DelegationOff},
	}
	if err := (projectconfig.ConfigStore{}).Write(projectconfig.Path(target), cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := NewService().Run(Options{Target: target}, &out)
	if err == nil || !strings.Contains(out.String(), "execution.delegation project=off install-state=auto") || !strings.Contains(out.String(), "--delegation") {
		t.Fatalf("expected harness mismatch failure, err=%v output=%s", err, out.String())
	}
}

func TestCheckBuilderFeedsSameReportToJSONPresenter(t *testing.T) {
	target := validVerifyTarget(t)
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		t.Fatal(err)
	}
	report := Report{TargetRoot: resolved, Scope: string(assets.ScopeProject)}

	if err := (CheckBuilder{}).Build(Options{Target: resolved, Scope: assets.ScopeProject}, &report); err != nil {
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

func TestVerifyReportsGuardrailDriftRecoveryAndDoesNotMutateUserOwnedState(t *testing.T) {
	target := validVerifyTarget(t)
	agentRel := filepath.Join(".codex", "config.toml")
	writeVerifyFile(t, filepath.Join(target, agentRel), "managed orchestrator\n")
	addVerifyManagedAsset(t, target, agentRel, "managed")
	agentsBefore := readVerifyFile(t, filepath.Join(target, "AGENTS.md"))
	memoryPath := filepath.Join(target, ".lufy", "memory", "knowledge", "private.md")
	contextPath := filepath.Join(target, ".lufy", "context", "graph.json")
	writeVerifyFile(t, memoryPath, "private memory\n")
	writeVerifyFile(t, contextPath, "{\"graph\":true}\n")
	memoryBefore := readVerifyFile(t, memoryPath)
	contextBefore := readVerifyFile(t, contextPath)

	writeVerifyFile(t, filepath.Join(target, agentRel), "local guardrail drift\n")
	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, JSON: true}, &out)
	if err == nil {
		t.Fatalf("Run(JSON guardrail drift) expected error, output=%s", out.String())
	}
	var report Report
	if json.Unmarshal(out.Bytes(), &report) != nil {
		t.Fatalf("invalid JSON report: %s", out.String())
	}
	found := false
	for _, check := range report.Checks {
		if check.Path == filepath.ToSlash(agentRel) && check.RecommendedAction == "lufy-ai sync --target <dir>" && strings.Contains(check.Message, "guardrail drift") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing guardrail drift recovery in report: %#v", report.Checks)
	}
	if got := readVerifyFile(t, filepath.Join(target, "AGENTS.md")); string(got) != string(agentsBefore) {
		t.Fatalf("verify mutated AGENTS.md: before=%q after=%q", agentsBefore, got)
	}
	if got := readVerifyFile(t, memoryPath); string(got) != string(memoryBefore) {
		t.Fatalf("verify mutated memory: before=%q after=%q", memoryBefore, got)
	}
	if got := readVerifyFile(t, contextPath); string(got) != string(contextBefore) {
		t.Fatalf("verify mutated context graph: before=%q after=%q", contextBefore, got)
	}
}

func TestReportRecorderUsesStableSlashPaths(t *testing.T) {
	report := Report{}
	recorder := reportRecorder{report: &report}

	recorder.emit("ok", `.opencode\agents\orchestrator.md`, "path estable")
	recorder.emitAsset("fail", `.opencode\skills\sdd-workflow\SKILL.md`, "managed", "lufy-ai sync --target <dir>", "asset con drift")

	if report.Checks[0].Path != ".opencode/agents/orchestrator.md" {
		t.Fatalf("emit path not normalized: %#v", report.Checks[0])
	}
	if report.Checks[1].Path != ".opencode/skills/sdd-workflow/SKILL.md" || report.Checks[1].RecommendedAction == "" {
		t.Fatalf("emitAsset path/policy fields unexpected: %#v", report.Checks[1])
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
		if rel == ".codex/hooks.json" {
			continue
		}
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == ".codex/hooks.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(target, filepath.Join(".codex", "rules"))); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err == nil {
		t.Fatalf("Run(invalid structure) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta directorio crítico: .codex/rules") {
		t.Fatalf("missing directory output unexpected: %s", out.String())
	}
	if !strings.Contains(out.String(), "fail: asset clave no está en manifest: .codex/hooks.json") {
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
	if err := os.RemoveAll(filepath.Join(target, filepath.Join(".codex", "rules"))); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err == nil {
		t.Fatalf("Run(missing templates) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta directorio crítico: .codex/rules") {
		t.Fatalf("missing templates output unexpected: %s", out.String())
	}
}

func TestVerifyFailsCorruptManifestAndMovedTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy", "managed-state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(target), []byte("{bad-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewService().Run(Options{Target: target}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "install-state.json inválido") {
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
		if rel == ".codex/hooks.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(actual, state.New(t.TempDir(), nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := NewService().Run(Options{Target: actual}, &out)
	if err == nil || !strings.Contains(out.String(), "targetRoot del manifest no coincide") {
		t.Fatalf("expected moved target failure, err=%v output=%s", err, out.String())
	}
}

func TestVerifyRejectsInvalidHooksJSON(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, ".codex", "hooks.json"), "{broken")
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err == nil || !strings.Contains(out.String(), "JSON inválido") {
		t.Fatalf("invalid hooks accepted: %s", out.String())
	}
}

func TestVerifyReportsExtraFilesInManagedDirsAsInfo(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, ".codex", "local-agent.md"), "local\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err != nil {
		t.Fatalf("Run() error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "info: archivo extra en directorio gestionado: .codex/local-agent.md") {
		t.Fatalf("extra managed dir file not reported: %s", out.String())
	}
}

func TestVerifyJSONReportsFailures(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "sin integración\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, JSON: true}, &out); err == nil {
		t.Fatalf("Run(JSON drift) expected error, output=%s", out.String())
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v body=%s", err, out.String())
	}
	if report.OK || report.Failures == 0 || report.TargetRoot == "" || report.Assets == 0 {
		t.Fatalf("unexpected JSON report: %#v", report)
	}
	foundMissingIntegration := false
	for _, check := range report.Checks {
		if check.Level == "fail" && check.Path == "AGENTS.md" && strings.Contains(check.Message, "no contiene integración LUFY") {
			foundMissingIntegration = true
		}
	}
	if !foundMissingIntegration {
		t.Fatalf("missing integration check not found in JSON: %#v", report.Checks)
	}
}

func TestVerifyAllowMissingAgentsRefDoesNotHideCriticalFailures(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "sin integración\n")
	writeVerifyFile(t, filepath.Join(target, "lufy-ia.harness.md"), "drift crítico\n")

	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, AllowMissingAgentsRef: true, JSON: true}, &out)
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
		if check.Level == "warn" && check.Path == "AGENTS.md" && strings.Contains(check.Message, "no contiene integración LUFY") {
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
	if err := NewService().Run(Options{Target: target, Quiet: true}, &out); err != nil {
		t.Fatalf("Run(quiet) error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("quiet output not suppressed: %s", out.String())
	}
}

func TestVerifyVerboseReportsDerivedRequirements(t *testing.T) {
	target := validVerifyTarget(t)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Verbose: true}, &out); err != nil {
		t.Fatalf("Run(verbose) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "info: requirements derivados") {
		t.Fatalf("verbose requirements info missing: %s", out.String())
	}
}

func TestVerifyDeepReportsMissingRegistryWithoutRepairingIt(t *testing.T) {
	target := validVerifyTarget(t)
	registryPath := filepath.Join(target, ".lufy", "skill-registry.json")
	if err := os.Remove(registryPath); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Deep: true}, &out); err == nil {
		t.Fatalf("deep verify should fail for missing registry: %s", out.String())
	}
	if !strings.Contains(out.String(), "skill registry not_available") || !os.IsNotExist(statError(registryPath)) {
		t.Fatalf("deep verify should report without repairing registry: %s", out.String())
	}
}

func TestVerifyDeepValidatesInitializedMemory(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, ".opencode/hooks/skills-ensure.sh"), "#!/usr/bin/env bash\n")
	writeVerifyFile(t, filepath.Join(target, ".opencode/hooks/memory-orient.sh"), "#!/usr/bin/env bash\n")
	writeVerifyFile(t, filepath.Join(target, ".opencode/hooks/memory-validate.sh"), "#!/usr/bin/env bash\n")
	writeVerifyFile(t, filepath.Join(target, ".opencode/plugins/lufy-memory-context.ts"), "export const LufyMemoryContextPlugin = async () => ({})\n")
	if err := memory.NewService().Init(memory.Options{Target: target}, &bytes.Buffer{}); err != nil {
		t.Fatalf("memory init fixture: %v", err)
	}
	writeVerifyFile(t, filepath.Join(target, ".lufy/memory/knowledge/valid.md"), `---
name: valid
description: Nota valida para verify deep.
type: lesson
status: active
---

Aprendizaje durable.
`)

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Deep: true}, &out); err != nil {
		t.Fatalf("Run(deep memory valid) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "memoria Obsidian schema=1 notas=1") {
		t.Fatalf("deep memory output unexpected: %s", out.String())
	}
	if !strings.Contains(out.String(), "context graph not_available") || !strings.Contains(out.String(), "SessionStart") {
		t.Fatalf("deep memory/context integration output unexpected: %s", out.String())
	}
}

func TestCatalogRequirementsIncludeRegisteredCatalogAssets(t *testing.T) {
	st := state.NewWithHarness(t.TempDir(), nil, []state.AssetState{
		{TargetRel: filepath.Join(".codex", "config.toml")},
	}, "test", domain.DefaultHarnessConfig())
	dirs, files := catalogRequirements(&st)
	if !containsString(files, filepath.Join(".codex", "config.toml")) {
		t.Fatalf("catalog registered file not required: %#v", files)
	}
	if !containsString(dirs, filepath.Join(".codex")) {
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
	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), string(agentsref.MinimalContent()))
	var states []state.AssetState
	for _, rel := range fallbackRequiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		policy := "managed"
		if rel == ".codex/hooks.json" {
			policy = "no-replace"
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	if err := skillregistry.NewService().Ensure(skillregistry.Options{Target: target, Tool: domain.ToolCodex}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	return target
}

func statError(path string) error {
	_, err := os.Stat(path)
	return err
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

func addVerifyManagedAsset(t *testing.T, target, rel, policy string) {
	t.Helper()
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := assets.FileSHA256(filepath.Join(target, rel))
	if err != nil {
		t.Fatal(err)
	}
	st.Assets = append(st.Assets, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, Policy: policy, Scope: "project", LastAction: "copy"})
	if err := state.WriteAtomic(target, *st); err != nil {
		t.Fatal(err)
	}
}

func readVerifyFile(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
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
	c, err := assets.BuildEmbeddedCatalog()
	if err == nil {
		b, e := assets.ReadSourceFile(c.SourceRoot, rel)
		if e == nil {
			return string(b)
		}
	}
	if rel == ".codex/hooks.json" {
		return "{}\n"
	}
	if rel == filepath.Join("openspec", "UPSTREAM.json") {
		return "{}\n"
	}
	return rel + "\n"
}
