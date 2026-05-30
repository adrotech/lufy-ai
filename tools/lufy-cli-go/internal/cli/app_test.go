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

	code := Run([]string{"install", "--target", ".", "--dry-run", "--yes", "--no-engram"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d, stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Modo dry-run")) {
		t.Fatalf("expected dry-run message, got: %s", out.String())
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
}

func TestRunInstallPersistsHarnessSelection(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", target, "--yes", "--no-engram", "--tool", "opencode", "--methodology-tier", "T3:openspec/full"}, Dependencies{Stdout: &out, Stderr: &errOut})
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

func TestRunInstallRejectsUnsupportedHarnessFlags(t *testing.T) {
	tests := [][]string{
		{"install", "--target", t.TempDir(), "--dry-run", "--tool", "codex"},
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
	if !bytes.Contains(out.Bytes(), []byte("init")) || !bytes.Contains(out.Bytes(), []byte("install")) || !bytes.Contains(out.Bytes(), []byte("restore")) || !bytes.Contains(out.Bytes(), []byte("sync")) || !bytes.Contains(out.Bytes(), []byte("status")) || !bytes.Contains(out.Bytes(), []byte("upgrade")) {
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
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
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
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
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
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
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
	if !bytes.Contains(out.Bytes(), []byte("project.yaml")) || !bytes.Contains(out.Bytes(), []byte("go (supported)")) {
		t.Fatalf("init output unexpected: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".opencode", "project.yaml")); err != nil {
		t.Fatalf("project config not written: %v", err)
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
	code := Run([]string{"verify", "--target", t.TempDir(), "--quiet", "--no-engram"}, Dependencies{Stdout: &out, Stderr: &errOut})
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
	code := Run([]string{"verify", "--target", t.TempDir(), "--deep", "--no-engram"}, Dependencies{Stdout: &out, Stderr: &errOut})
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
	if !bytes.Contains(errOut.Bytes(), []byte("lufy-ai sync")) || !bytes.Contains(errOut.Bytes(), []byte("--target")) || !bytes.Contains(errOut.Bytes(), []byte("--scope")) || !bytes.Contains(errOut.Bytes(), []byte("--tool")) || !bytes.Contains(errOut.Bytes(), []byte("--dry-run")) || !bytes.Contains(errOut.Bytes(), []byte("--yes")) || !bytes.Contains(errOut.Bytes(), []byte("--no-engram")) {
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
		{"sync", "--target", t.TempDir(), "--dry-run", "--tool", "codex"},
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
