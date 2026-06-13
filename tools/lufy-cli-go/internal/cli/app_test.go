package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestRunInstallDryRun(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", ".", "--dry-run", "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d, stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Modo dry-run")) {
		t.Fatalf("expected dry-run message, got: %s", out.String())
	}
}

func TestRunMigrateLayoutDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{}\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"migrate-layout", "--target", target, "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("migrate-layout dry-run expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("[dry-run]")) {
		t.Fatalf("expected dry-run plan, got %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "managed-state", "install-state.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote install-state: %v", err)
	}
}

func TestRunMigrateLayoutRequiresYesForMutations(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{}\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"migrate-layout", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("migrate-layout without --yes expected ExitRuntimeErr, got %d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !bytes.Contains(errOut.Bytes(), []byte("requiere --yes")) {
		t.Fatalf("expected --yes error, got %s", errOut.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "managed-state", "install-state.json")); !os.IsNotExist(err) {
		t.Fatalf("run without --yes wrote install-state: %v", err)
	}
}

func TestRunMigrateLayoutJSONAppliesWithoutHumanLogs(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{}\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"migrate-layout", "--target", target, "--yes", "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("migrate-layout json expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json output %q: %v", out.String(), err)
	}
	if decoded["applied"] != true {
		t.Fatalf("expected applied=true, got %#v", decoded)
	}
	if bytes.Contains(out.Bytes(), []byte("[migrate-layout]")) {
		t.Fatalf("json output contains human log: %s", out.String())
	}
}

func TestRunInstallDryRunShowsLayoutMigrationWithoutWriting(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy", "project.yaml"), "schema_version: 1\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"install", "--target", target, "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install dry-run expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("[layout-migrate-copy] .lufy/project.yaml -> .lufy/config/project.yaml")) {
		t.Fatalf("expected layout migration plan, got %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "config", "project.yaml")); !os.IsNotExist(err) {
		t.Fatalf("install dry-run wrote project config: %v", err)
	}
}

func TestEnsureLayoutForMutationBlocksLegacyWithoutAllowApply(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{}\n")

	var out bytes.Buffer
	err := ensureLayoutForMutation(target, false, false, &out)
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("--yes")) {
		t.Fatalf("expected --yes layout error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".lufy", "managed-state", "install-state.json")); !os.IsNotExist(statErr) {
		t.Fatalf("preflight wrote canonical state: %v", statErr)
	}
}

func TestEnsureLayoutForMutationWithoutAllowApplyIgnoresReadmeOnlyPlan(t *testing.T) {
	target := t.TempDir()

	var out bytes.Buffer
	if err := ensureLayoutForMutation(target, false, false, &out); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".lufy", "README.md")); !os.IsNotExist(statErr) {
		t.Fatalf("preflight wrote README without confirmation: %v", statErr)
	}
}

func TestEnsureLayoutForMutationDryRunReportsConflicts(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy", "managed-state", "install-state.json"), "{\"new\":true}\n")
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{\"legacy\":true}\n")

	var out bytes.Buffer
	if err := ensureLayoutForMutation(target, true, false, &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("[layout-conflict]")) {
		t.Fatalf("expected layout conflict output, got %s", out.String())
	}
}

func TestReportLegacyLayoutForReadOnly(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, ".lufy", "managed-state", "install-state.json"), "{\"new\":true}\n")
	writeCLITestFile(t, filepath.Join(target, ".lufy-ai", "install-state.json"), "{\"legacy\":true}\n")
	writeCLITestFile(t, filepath.Join(target, ".lufy", "project.yaml"), "schema_version: 1\n")

	var out bytes.Buffer
	if err := reportLegacyLayoutForReadOnly(target, &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("legacy layout detectado: .lufy/project.yaml, .lufy-ai/install-state.json")) {
		t.Fatalf("expected legacy list, got %s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("legacy layout conflictivo")) {
		t.Fatalf("expected conflict warning, got %s", out.String())
	}
}

func TestReportLegacyLayoutForReadOnlyNoopWhenClean(t *testing.T) {
	var out bytes.Buffer
	if err := reportLegacyLayoutForReadOnly(t.TempDir(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output, got %s", out.String())
	}
}

func TestRunUninstallRequiresConfirmation(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"uninstall", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("uninstall without yes expected ExitRuntimeErr, got %d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !bytes.Contains(errOut.Bytes(), []byte("uninstall requiere --yes")) {
		t.Fatalf("uninstall confirmation error unexpected: %s", errOut.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"nope"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("expected ExitUsageErr, got %d", code)
	}
}

func TestRunHelpBranches(t *testing.T) {
	tests := [][]string{
		{"help"},
		{"opsx", "--help"},
		{"opsx", "render", "--help"},
		{"memory", "--help"},
		{"memory", "init", "--help"},
		{"memory", "search", "--help"},
		{"init", "--help"},
		{"scan", "--help"},
		{"migrate-layout", "--help"},
		{"merge", "--help"},
		{"pin", "--help"},
		{"unpin", "--help"},
		{"install", "--help"},
		{"sync", "--help"},
		{"uninstall", "--help"},
	}
	for _, args := range tests {
		t.Run(args[0]+"/"+args[len(args)-1], func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut})
			if code != ExitOK {
				t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, out.String(), errOut.String())
			}
		})
	}
}

func TestRunUsageBranches(t *testing.T) {
	tests := [][]string{
		{"opsx"},
		{"opsx", "unknown"},
		{"memory"},
		{"memory", "unknown"},
		{"memory", "status", "extra"},
		{"memory", "validate", "extra"},
		{"memory", "search"},
		{"init", "extra"},
		{"init", "--force", "--rescan"},
		{"scan", "extra"},
		{"migrate-layout", "extra"},
		{"merge"},
		{"merge", "--accept-theirs", "--accept-ours", "x"},
		{"pin"},
		{"unpin"},
		{"sync", "extra"},
		{"status", "extra"},
		{"info", "extra"},
		{"doctor", "extra"},
		{"version", "extra"},
		{"upgrade", "extra"},
		{"install", "extra"},
		{"uninstall", "extra"},
		{"restore"},
	}
	for _, args := range tests {
		t.Run(args[0]+"/usage", func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut})
			if code != ExitUsageErr {
				t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, out.String(), errOut.String())
			}
		})
	}
}

func TestRunMemoryCommands(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/app\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"memory", "init", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("memory init expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Memoria Obsidian inicializada")) {
		t.Fatalf("memory init output unexpected: %s", out.String())
	}
	writeCLITestFile(t, filepath.Join(target, ".lufy/memory/knowledge/searchable.md"), `---
name: searchable
description: Nota activa para buscar memoria.
type: rule
status: active
---

Lufy busca contexto durable.
`)

	out.Reset()
	errOut.Reset()
	code = Run([]string{"memory", "validate", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("memory validate expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"memory", "search", "--target", target, "durable"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("memory search expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("[active] knowledge/searchable.md")) {
		t.Fatalf("memory search output unexpected: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"memory", "status", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("memory status expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Inicializada: sí")) {
		t.Fatalf("memory status output unexpected: %s", out.String())
	}
}

func TestRunMemoryHelpAndUsageErrors(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"memory", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !bytes.Contains(out.Bytes(), []byte("Subcomandos")) {
		t.Fatalf("memory help unexpected code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"memory"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr || !bytes.Contains(errOut.Bytes(), []byte("Uso: lufy-ai memory")) {
		t.Fatalf("memory without subcommand unexpected code=%d stderr=%s", code, errOut.String())
	}

	errOut.Reset()
	code = Run([]string{"memory", "unknown"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr || !bytes.Contains(errOut.Bytes(), []byte("Subcomando memory desconocido")) {
		t.Fatalf("memory unknown unexpected code=%d stderr=%s", code, errOut.String())
	}

	errOut.Reset()
	code = Run([]string{"memory", "search"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("memory search without query expected usage, got %d", code)
	}
}

func TestRunVersionOutputAndRejectsArgs(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"version"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("version expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	for _, want := range []string{"lufy-ai", "commit:", "buildDate:", "goos:", "goarch:"} {
		if !bytes.Contains(out.Bytes(), []byte(want)) {
			t.Fatalf("version output missing %q: %s", want, out.String())
		}
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"version", "extra"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("version with positional arg expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("version no acepta argumentos posicionales")) {
		t.Fatalf("version positional arg error unexpected: %s", errOut.String())
	}
}

func TestRunInstallUnknownFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--unknown"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("expected ExitUsageErr, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"install", "--no-" + "en" + "gram"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("removed flag expected ExitUsageErr, got %d", code)
	}
}

func TestRunInstallPersistsHarnessSelection(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes", "--tool", "opencode", "--methodology-tier", "T3:openspec/full"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("expected install state")
	}
	if st.Tool != domain.ToolInitialDefault {
		t.Fatalf("tool = %s", st.Tool)
	}
	got := st.MethodologyByTier[domain.TierT3]
	if got.ID != domain.MethodologySpecWorkflow || got.Mode != domain.MethodologyModeFull || !got.Required {
		t.Fatalf("T3 methodology = %#v", got)
	}
}

func TestRunInstallPersistsLufySDDSelection(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{
		"install",
		"--target", target,
		"--yes",
		"--methodology-tier", "T1:lufy-sdd/full",
		"--methodology-tier", "T2:lufy-sdd/lite",
		"--methodology-tier", "T3:none",
	}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("expected install state")
	}
	got := st.MethodologyByTier[domain.TierT2]
	if got.ID != domain.MethodologyLufyWorkflow || got.Mode != domain.MethodologyModeLite || !got.Required {
		t.Fatalf("T2 methodology = %#v", got)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "workflows", "sdd", "changes", ".gitkeep")); err != nil {
		t.Fatalf("expected lufy-sdd changes asset: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "workflows", "sdd", "specs", ".gitkeep")); err != nil {
		t.Fatalf("expected lufy-sdd specs asset for full mode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "openspec", "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("openspec config should not be installed when no tier selects openspec, err=%v", err)
	}
}

func TestRunInstallWithCodexHarness(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes", "--tool", "codex"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install codex expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	for _, rel := range []string{
		filepath.Join(".codex", "lufy-agent-mapping.md"),
		filepath.Join(".codex", "agents", "implementer.toml"),
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"),
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("codex install missing %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(target, ".opencode")); !os.IsNotExist(err) {
		t.Fatalf("codex install should not create .opencode, err=%v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil || st.Tool != domain.ToolCodex {
		t.Fatalf("install state tool = %#v", st)
	}
}

func TestRunInstallRejectsUnsupportedHarnessFlags(t *testing.T) {
	tests := [][]string{
		{"install", "--target", t.TempDir(), "--dry-run", "--tool", "claude-code"},
		{"install", "--target", t.TempDir(), "--dry-run", "--methodology-tier", "T1:none"},
		{"install", "--target", t.TempDir(), "--dry-run", "--methodology-tier", "T3:spec-kit"},
	}
	for _, args := range tests {
		t.Run(args[len(args)-1], func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut})
			if code != ExitUsageErr {
				t.Fatalf("expected ExitUsageErr, got %d stderr=%s", code, errOut.String())
			}
		})
	}
}

func TestRunInstallHelpIncludesHarnessFlags(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"install", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install --help expected ExitOK, got %d", code)
	}
	for _, want := range []string{"--tool opencode", "--methodology-tier T3:none"} {
		if !bytes.Contains(errOut.Bytes(), []byte(want)) {
			t.Fatalf("install help missing %q: %s", want, errOut.String())
		}
	}
}

func TestRunHelpCommandsAndRestoreRequiresBackup(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("help expected ExitOK, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("init")) || !bytes.Contains(out.Bytes(), []byte("install")) || !bytes.Contains(out.Bytes(), []byte("restore")) || !bytes.Contains(out.Bytes(), []byte("sync")) || !bytes.Contains(out.Bytes(), []byte("status")) || !bytes.Contains(out.Bytes(), []byte("info")) || !bytes.Contains(out.Bytes(), []byte("doctor")) || !bytes.Contains(out.Bytes(), []byte("pin")) || !bytes.Contains(out.Bytes(), []byte("unpin")) || !bytes.Contains(out.Bytes(), []byte("upgrade")) {
		t.Fatalf("help output missing commands: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"restore", "--target", t.TempDir(), "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("restore without --backup expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("restore requiere --backup")) {
		t.Fatalf("restore missing backup output unexpected: %s", errOut.String())
	}
}

func TestRunInfoAndDoctor(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/app\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"info", "--target", target, "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("info expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"installed": true`)) || !bytes.Contains(out.Bytes(), []byte(`"catalogAssets"`)) || !bytes.Contains(out.Bytes(), []byte(`"stacks"`)) {
		t.Fatalf("info json unexpected: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"doctor", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("doctor expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Doctor OK")) {
		t.Fatalf("doctor output unexpected: %s", out.String())
	}
}

func TestRunPinAndUnpin(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("install expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"pin", "--target", target, "--reason", "local override", "lufy-ia.harness.md"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("pin expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	pinned := stateMustLoadForCLITest(t, target).AssetMap()["lufy-ia.harness.md"]
	if !pinned.Pinned || pinned.PinnedReason != "local override" || pinned.LastAction != "pin" {
		t.Fatalf("pin did not update state: %#v", pinned)
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"unpin", "--target", target, "lufy-ia.harness.md"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("unpin expected ExitOK, got %d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
	unpinned := stateMustLoadForCLITest(t, target).AssetMap()["lufy-ia.harness.md"]
	if unpinned.Pinned || unpinned.PinnedReason != "" || unpinned.LastAction != "unpin" {
		t.Fatalf("unpin did not update state: %#v", unpinned)
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"pin", "--target", target, "missing.md"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr || !bytes.Contains(errOut.Bytes(), []byte("asset no gestionado")) {
		t.Fatalf("pin missing asset expected runtime error, got code=%d stderr=%s stdout=%s", code, errOut.String(), out.String())
	}
}

func TestRunDoctorReportsMissingManifest(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"doctor", "--target", target, "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("doctor missing manifest expected ExitRuntimeErr, got %d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"ok": false`)) || !bytes.Contains(out.Bytes(), []byte("falta manifest")) {
		t.Fatalf("doctor missing manifest json unexpected: %s", out.String())
	}
}

func TestRunMergeHelpAndRequiresPath(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"merge", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("merge --help expected ExitOK, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai merge")) || !bytes.Contains(errOut.Bytes(), []byte("LUFY_MERGE_TOOL")) || !bytes.Contains(errOut.Bytes(), []byte("--accept-theirs")) || !bytes.Contains(errOut.Bytes(), []byte("--accept-ours")) {
		t.Fatalf("merge help unexpected: %s", errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"merge"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("merge without path expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai merge")) {
		t.Fatalf("merge missing path output unexpected: %s", errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"merge", "tui.json", "extra.json"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("merge with extra path expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai merge")) {
		t.Fatalf("merge extra path output unexpected: %s", errOut.String())
	}
}

func TestRunMergeRejectsConflictingAcceptFlags(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"merge", "--accept-theirs", "--accept-ours", "tui.json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("merge conflicting accept flags expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("no permite combinar")) {
		t.Fatalf("merge conflicting flags error unexpected: %s", errOut.String())
	}
}

func TestRunMergeDispatchesValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{
			name: "invalid relative path",
			args: []string{"merge", "--target", t.TempDir(), "../tui.json"},
			code: ExitRuntimeErr,
			want: "escapa del root permitido",
		},
		{
			name: "unknown flag",
			args: []string{"merge", "--unknown", "tui.json"},
			code: ExitUsageErr,
			want: "flag provided but not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(tt.args, Dependencies{Stdout: &out, Stderr: &errOut})
			if code != tt.code {
				t.Fatalf("expected code %d, got %d stderr=%s", tt.code, code, errOut.String())
			}
			if !bytes.Contains(errOut.Bytes(), []byte(tt.want)) {
				t.Fatalf("stderr missing %q: %s", tt.want, errOut.String())
			}
		})
	}
}

func TestRunMergeWithoutToolReturnsRuntimeErrorAfterDispatch(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, "tui.json"), "user\n")
	writeCLITestFile(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy", "managed-state", "ancestors", "tui.json")
	writeCLITestFile(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hashCLITestFile(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"merge", "--target", target, "tui.json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("merge without LUFY_MERGE_TOOL expected ExitRuntimeErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("LUFY_MERGE_TOOL")) {
		t.Fatalf("merge without tool stderr unexpected: %s", errOut.String())
	}
}

func TestRunMergeAcceptTheirsResolvesWithoutMergeTool(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, "tui.json"), "user\n")
	writeCLITestFile(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy", "managed-state", "ancestors", "tui.json")
	writeCLITestFile(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hashCLITestFile(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"merge", "--target", target, "--accept-theirs", "tui.json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("merge --accept-theirs expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if got := string(readCLITestFile(t, filepath.Join(target, "tui.json"))); got != "new\n" {
		t.Fatalf("target = %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("accept-theirs aplicado")) {
		t.Fatalf("merge accept-theirs output unexpected: %s", out.String())
	}
}

func TestRunMergeAcceptOursResolvesWithoutMergeTool(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, "tui.json"), "user\n")
	writeCLITestFile(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy", "managed-state", "ancestors", "tui.json")
	writeCLITestFile(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hashCLITestFile(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"merge", "--target", target, "--accept-ours", "tui.json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("merge --accept-ours expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if got := string(readCLITestFile(t, filepath.Join(target, "tui.json"))); got != "user\n" {
		t.Fatalf("target = %q", got)
	}
	if got := string(readCLITestFile(t, filepath.Join(target, ancestorRel))); got != "user\n" {
		t.Fatalf("ancestor = %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("accept-ours aplicado")) {
		t.Fatalf("merge accept-ours output unexpected: %s", out.String())
	}
}

func TestRunBackupAndScopeErrors(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"backup", "--bad"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("backup bad flag expected ExitUsageErr, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"status", "--scope", "invalid"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("status invalid scope expected ExitUsageErr, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"sync", "--scope", "invalid"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("sync invalid scope expected ExitUsageErr, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"install", "--scope", "invalid"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("install invalid scope expected ExitUsageErr, got %d", code)
	}
}

func TestRunBackupCreatesBackupForManagedFile(t *testing.T) {
	target := t.TempDir()
	writeCLITestFile(t, filepath.Join(target, "lufy-ia.harness.md"), "managed\n")
	hash := hashCLITestFile(t, filepath.Join(target, "lufy-ia.harness.md"))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "harness", SourceRel: "lufy-ia.harness.md", TargetRel: "lufy-ia.harness.md", SourceSHA256: hash, TargetSHA256: hash, Policy: "managed", Scope: "project", LastAction: "copy"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"backup", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("backup expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Backup creado")) {
		t.Fatalf("backup output unexpected: %s", out.String())
	}
}

func TestRunInitCreatesProjectConfig(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/app\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"init", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("init expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("project.yaml")) || !bytes.Contains(out.Bytes(), []byte("go (supported)")) || !bytes.Contains(out.Bytes(), []byte("Superficies detectadas")) {
		t.Fatalf("init output unexpected: %s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("project_profile: modo no interactivo")) {
		t.Fatalf("init did not attempt interactive profile fallback: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "config", "project.yaml")); err != nil {
		t.Fatalf("project config not written: %v", err)
	}
}

func TestRunInitCanDisableInteractiveProfilePrompt(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"init", "--target", target, "--interactive=false"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("init expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if bytes.Contains(out.Bytes(), []byte("project_profile: modo no interactivo")) {
		t.Fatalf("init should not run profile prompt when disabled: %s", out.String())
	}
}

func TestRunScanCreatesProjectProfileWithoutTTYPrompt(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "package.json"), []byte(`{"dependencies":{"react":"18.0.0","next":"14.0.0"},"devDependencies":{"typescript":"5.4.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "tsconfig.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"scan", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("scan expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	body, err := os.ReadFile(filepath.Join(target, ".lufy", "config", "project.yaml"))
	if err != nil {
		t.Fatalf("project config not written: %v", err)
	}
	if !bytes.Contains(body, []byte("project_profile:")) || !bytes.Contains(body, []byte("type: frontend")) {
		t.Fatalf("scan project profile unexpected:\n%s", string(body))
	}
	if !bytes.Contains(out.Bytes(), []byte("project_profile: modo no interactivo")) {
		t.Fatalf("scan did not report non-interactive profile fallback: %s", out.String())
	}
}

func TestRunInitHelpAndRejectsPositionals(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	if code := Run([]string{"init", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("init --help expected ExitOK, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai init")) || !bytes.Contains(errOut.Bytes(), []byte("--target")) || !bytes.Contains(errOut.Bytes(), []byte("reporta drift sin borrar")) {
		t.Fatalf("init help unexpected: %s", errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"init", "extra"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("init positional expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("init no acepta argumentos posicionales")) {
		t.Fatalf("init positional error unexpected: %s", errOut.String())
	}
}

func TestRunStatusJSON(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"status", "--target", t.TempDir(), "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("status --json expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("status output is not JSON: %v body=%s", err, out.String())
	}
	if decoded["installed"] != false {
		t.Fatalf("unexpected status JSON: %#v", decoded)
	}
}

func TestRunStatusVerbose(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"status", "--target", t.TempDir(), "--verbose"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("status --verbose expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("no instalado")) {
		t.Fatalf("status verbose output unexpected: %s", out.String())
	}
}

func TestRunVerifyQuietMissingState(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"verify", "--target", t.TempDir(), "--quiet"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("verify --quiet expected ExitRuntimeErr, got %d", code)
	}
	if out.Len() != 0 {
		t.Fatalf("quiet verify wrote stdout: %s", out.String())
	}
	if !bytes.Contains(errOut.Bytes(), []byte("verify falló")) {
		t.Fatalf("quiet verify stderr unexpected: %s", errOut.String())
	}
}

func TestRunVerifyAcceptsDeepFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"verify", "--target", t.TempDir(), "--deep"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("verify --deep expected runtime error for missing state, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("verify falló")) {
		t.Fatalf("verify --deep stderr unexpected: %s", errOut.String())
	}
}

func TestRunUpgradeRequiresTo(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"upgrade", "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitRuntimeErr {
		t.Fatalf("upgrade without --to expected ExitRuntimeErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("upgrade requiere --to")) {
		t.Fatalf("upgrade stderr unexpected: %s", errOut.String())
	}
}

func TestRunSyncHelpAndUnknownFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"sync", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("sync --help expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai sync")) || !bytes.Contains(errOut.Bytes(), []byte("--target")) || !bytes.Contains(errOut.Bytes(), []byte("--scope")) || !bytes.Contains(errOut.Bytes(), []byte("--tool")) || !bytes.Contains(errOut.Bytes(), []byte("--dry-run")) || !bytes.Contains(errOut.Bytes(), []byte("--yes")) {
		t.Fatalf("sync help missing flags: %s", errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"sync", "--unknown"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("sync unknown flag expected ExitUsageErr, got %d", code)
	}
}

func TestRunSyncAndVerifyRejectUnsupportedTool(t *testing.T) {
	tests := [][]string{
		{"sync", "--target", t.TempDir(), "--dry-run", "--tool", "claude-code"},
		{"verify", "--target", t.TempDir(), "--tool", "claude-code"},
	}
	for _, args := range tests {
		t.Run(args[0], func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut})
			if code != ExitUsageErr {
				t.Fatalf("expected ExitUsageErr, got %d stderr=%s", code, errOut.String())
			}
			if !bytes.Contains(errOut.Bytes(), []byte("tool adapter no soportado")) {
				t.Fatalf("stderr missing tool error: %s", errOut.String())
			}
		})
	}
}

func TestRunRejectsInvalidScope(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"install", "--scope", "elsewhere", "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("invalid scope expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("project, global, both")) {
		t.Fatalf("invalid scope output unexpected: %s", errOut.String())
	}
}

func TestRunRestoreDryRunParsesRequiredBackup(t *testing.T) {
	target := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest, err := json.Marshal(struct {
		SchemaVersion int      `json:"schemaVersion"`
		ToolVersion   string   `json:"toolVersion"`
		TargetRoot    string   `json:"targetRoot"`
		Files         []string `json:"files"`
	}{SchemaVersion: 1, ToolVersion: "dev", TargetRoot: target, Files: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"restore", "--target", target, "--backup", backupDir, "--dry-run", "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("restore dry-run expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
}

func TestRunOpsxRenderGeneratesHTML(t *testing.T) {
	target := t.TempDir()
	changeDir := filepath.Join(target, "openspec", "changes", "render-demo")
	writeCLITestFile(t, filepath.Join(changeDir, "proposal.md"), "# Proposal")
	writeCLITestFile(t, filepath.Join(changeDir, "design.md"), "# Design")
	writeCLITestFile(t, filepath.Join(changeDir, "tasks.md"), "# Tasks")
	writeCLITestFile(t, filepath.Join(changeDir, "specs", "demo", "spec.md"), "## ADDED Requirements\n\n### Requirement: Demo\n\n#### Scenario: View\nWHEN opened\nTHEN visible")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"opsx", "render", "--target", target, "--change", "render-demo"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("opsx render expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("HTML OpenSpec generado")) {
		t.Fatalf("unexpected stdout: %s", out.String())
	}
	body := readCLITestFile(t, filepath.Join(changeDir, "change-overview.html"))
	for _, want := range [][]byte{[]byte("Proposal"), []byte("Design"), []byte("Tasks"), []byte(`class="tabs"`)} {
		if !bytes.Contains(body, want) {
			t.Fatalf("generated html missing %q", want)
		}
	}
	for _, unwanted := range [][]byte{[]byte("Plan"), []byte("plan.md"), []byte("Notion dark"), []byte("Offline HTML"), []byte("Artifacts disponibles:"), []byte("Sin recursos remotos")} {
		if bytes.Contains(body, unwanted) {
			t.Fatalf("generated html should not include %q", unwanted)
		}
	}
	if bytes.Contains(body, []byte("specs/demo/spec.md")) {
		t.Fatalf("generated html should not include nested spec artifact")
	}
}

func TestRunOpsxRenderRejectsPositionalArgs(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"opsx", "render", "extra"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("opsx render positional expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("opsx render no acepta argumentos posicionales")) {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRunOpsxHelpAndUnknownSubcommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"opsx", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("opsx help expected ExitOK, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("render")) {
		t.Fatalf("opsx help missing render: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"opsx", "unknown"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("unknown opsx expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("Subcomando opsx desconocido")) {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func writeCLITestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readCLITestFile(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func hashCLITestFile(t *testing.T, path string) string {
	t.Helper()
	h := sha256.Sum256(readCLITestFile(t, path))
	return hex.EncodeToString(h[:])
}

func stateMustLoadForCLITest(t *testing.T, target string) *state.InstallState {
	t.Helper()
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("missing install-state")
	}
	return st
}
