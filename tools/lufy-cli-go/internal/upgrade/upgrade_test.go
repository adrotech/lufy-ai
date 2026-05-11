package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUpgradeDryRunDoesNotWrite(t *testing.T) {
	target := filepath.Join(t.TempDir(), "lufy-ai")
	var out bytes.Buffer
	if err := NewService().Run(Options{To: "v1.2.3", BaseURL: "file:///nope", ExecPath: target, DryRun: true}, &out); err != nil {
		t.Fatalf("Run(dry-run) error = %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote target, err=%v", err)
	}
	if !strings.Contains(out.String(), "Modo dry-run") {
		t.Fatalf("dry-run output unexpected: %s", out.String())
	}
}

func TestUpgradeReplacesExecutableFromFileRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test usa tar.gz para plataformas no Windows")
	}
	version := "v1.2.3"
	releaseRoot := t.TempDir()
	releaseDir := filepath.Join(releaseRoot, version)
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	artifact := artifactName(version, runtime.GOOS, runtime.GOARCH)
	body := tarGzBinary(t, "new-binary\n")
	if err := os.WriteFile(filepath.Join(releaseDir, artifact), body, 0o644); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(body)
	checksums := fmt.Sprintf("%x  %s\n", hash, artifact)
	if err := os.WriteFile(filepath.Join(releaseDir, fmt.Sprintf("lufy-ai_%s_checksums.txt", version)), []byte(checksums), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "lufy-ai")
	if err := os.WriteFile(target, []byte("old\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := NewService().Run(Options{To: version, BaseURL: "file://" + releaseRoot, ExecPath: target}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary\n" {
		t.Fatalf("target not replaced: %q", string(got))
	}
}

func TestUpgradeRejectsLatest(t *testing.T) {
	if err := NewService().Run(Options{To: "latest", DryRun: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "no acepta latest") {
		t.Fatalf("expected latest rejection, got %v", err)
	}
}

func tarGzBinary(t *testing.T, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte(content)
	if err := tw.WriteHeader(&tar.Header{Name: "lufy-ai_test/lufy-ai", Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
