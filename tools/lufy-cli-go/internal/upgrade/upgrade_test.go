package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestUpgradeRejectsMissingAndUnversionedTargets(t *testing.T) {
	for _, tc := range []struct {
		to   string
		want string
	}{
		{"", "requiere --to"},
		{"1.2.3", "prefijo v"},
	} {
		err := NewService().Run(Options{To: tc.to, DryRun: true}, &bytes.Buffer{})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("Run(%q) error = %v, want %q", tc.to, err, tc.want)
		}
	}
}

func TestFetchURLReadsFileAndRejectsHTTPError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "asset.txt")
	if err := os.WriteFile(file, []byte("asset"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := fetchURL("file://" + file)
	if err != nil || string(body) != "asset" {
		t.Fatalf("fetch file body=%q err=%v", string(body), err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	}))
	t.Cleanup(server.Close)
	if _, err := fetchURL(server.URL); err == nil || !strings.Contains(err.Error(), "HTTP 418") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
}

func TestVerifyChecksumErrors(t *testing.T) {
	artifact := "lufy-ai_v1_test.tar.gz"
	if err := verifyChecksum(artifact, []byte("body"), []byte("")); err == nil || !strings.Contains(err.Error(), "no contiene entrada") {
		t.Fatalf("expected missing checksum entry, got %v", err)
	}
	if err := verifyChecksum(artifact, []byte("body"), []byte("deadbeef  "+artifact+"\n")); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}

func TestExtractBinaryZipAndMissingArtifacts(t *testing.T) {
	zipBody := zipBinary(t, "lufy-ai.exe", "exe-body")
	got, err := extractBinary("lufy-ai_v1_windows_amd64.zip", zipBody)
	if err != nil || string(got) != "exe-body" {
		t.Fatalf("extract zip body=%q err=%v", string(got), err)
	}

	if _, err := extractZipBinary(zipBinary(t, "other.txt", "x")); err == nil || !strings.Contains(err.Error(), "lufy-ai.exe") {
		t.Fatalf("expected missing zip binary, got %v", err)
	}
	if _, err := extractTarBinary(tarGzNamed(t, "dir/other", "x")); err == nil || !strings.Contains(err.Error(), "binario lufy-ai") {
		t.Fatalf("expected missing tar binary, got %v", err)
	}
	if _, err := extractTarBinary([]byte("not gzip")); err == nil {
		t.Fatalf("expected invalid gzip error")
	}
}

func tarGzBinary(t *testing.T, content string) []byte {
	return tarGzNamed(t, "lufy-ai_test/lufy-ai", content)
}

func tarGzNamed(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte(content)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
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

func zipBinary(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
