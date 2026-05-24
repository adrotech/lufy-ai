package opsx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolvePrefersCompatiblePathOpenSpec(t *testing.T) {
	svc := &Service{
		lookPath: func(name string) (string, error) { return "/usr/local/bin/openspec", nil },
		commandOutput: func(name string, args ...string) ([]byte, error) {
			return []byte("OpenSpec 1.4.0\n"), nil
		},
	}
	res, err := svc.Resolve(ResolveOptions{Target: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if res.Layer != LayerPath || res.Version != "1.4.0" || res.Path == "" {
		t.Fatalf("resolution = %#v", res)
	}
}

func TestNewServiceUsesDefaultResolvers(t *testing.T) {
	svc := NewService()
	if svc.lookPath == nil || svc.commandOutput == nil {
		t.Fatalf("NewService did not initialize resolvers: %#v", svc)
	}
}

func TestResolveUsesValidCacheWhenPathUnavailable(t *testing.T) {
	target := t.TempDir()
	writeCacheAsset(t, target, "1.3.2", "bin/openspec", "cache binary")
	manifest := validManifest(t, target, "1.3.2")
	if _, err := WriteCacheManifest(target, manifest); err != nil {
		t.Fatal(err)
	}
	svc := &Service{lookPath: func(name string) (string, error) { return "", errors.New("missing") }}
	res, err := svc.Resolve(ResolveOptions{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if res.Layer != LayerCache || res.Version != "1.3.2" {
		t.Fatalf("resolution = %#v", res)
	}
}

func TestResolveFallsBackToEmbeddedWhenCacheCorrupt(t *testing.T) {
	target := t.TempDir()
	cacheDir := filepath.Join(target, DefaultCacheRoot, "1.3.2")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, DefaultManifestName), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := &Service{lookPath: func(name string) (string, error) { return "", errors.New("missing") }}
	res, err := svc.Resolve(ResolveOptions{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if res.Layer != LayerEmbedded || res.Version == "" {
		t.Fatalf("resolution = %#v", res)
	}
	if !hasDiagnostic(res.Diagnostics, LayerCache, "inválida") {
		t.Fatalf("expected corrupt cache diagnostic, got %#v", res.Diagnostics)
	}
}

func TestReadCacheManifestRejectsUnsafePaths(t *testing.T) {
	target := t.TempDir()
	manifest := CacheManifest{SchemaVersion: 1, Version: "1.3.2", Source: Source{Type: "fixture"}, CreatedAt: time.Now().UTC(), Assets: []CacheAsset{{Path: "../escape"}}}
	cacheDir := filepath.Join(target, DefaultCacheRoot, "1.3.2")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, DefaultManifestName), mustJSON(t, manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadCacheManifest(target, "1.3.2"); err == nil {
		t.Fatal("expected unsafe path error")
	}
}

func TestReadCacheManifestRejectsSymlinkEscape(t *testing.T) {
	target := t.TempDir()
	cacheDir := filepath.Join(target, DefaultCacheRoot, "1.3.2")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(target, "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(cacheDir, "link")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	manifest := CacheManifest{SchemaVersion: 1, Version: "1.3.2", Source: Source{Type: "fixture"}, CreatedAt: time.Now().UTC(), Assets: []CacheAsset{{Path: "link"}}}
	if err := os.WriteFile(filepath.Join(cacheDir, DefaultManifestName), mustJSON(t, manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadCacheManifest(target, "1.3.2"); err == nil {
		t.Fatal("expected symlink error")
	}
}

func TestWriteCacheManifestWritesReadableManifest(t *testing.T) {
	target := t.TempDir()
	writeCacheAsset(t, target, "1.3.2", "bin/openspec", "cache binary")
	manifest := validManifest(t, target, "1.3.2")
	path, err := WriteCacheManifest(target, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(filepath.Base(path), ".write-") {
		t.Fatalf("manifest path should be final path, got %s", path)
	}
	read, _, err := ReadCacheManifest(target, "1.3.2")
	if err != nil {
		t.Fatal(err)
	}
	if read.Version != manifest.Version || len(read.Assets) != 1 {
		t.Fatalf("manifest = %#v", read)
	}
}

func TestResolverHelpers(t *testing.T) {
	if got := compareVersionSafe("1.2.0", "1.1.9"); got <= 0 {
		t.Fatalf("expected left version greater, got %d", got)
	}
	if got := compareVersionSafe("bad", "1.1.9"); got != -1 {
		t.Fatalf("invalid left should sort low, got %d", got)
	}
	if got := compareVersionSafe("1.2.0", "bad"); got != 1 {
		t.Fatalf("invalid right should sort low, got %d", got)
	}

	joined := joinDiagnosticMessages([]Diagnostic{{Layer: LayerPath, Message: "missing"}, {Layer: LayerCache, Message: "stale"}})
	if !strings.Contains(joined, "PATH: missing") || !strings.Contains(joined, "cache: stale") || !strings.Contains(joined, "; ") {
		t.Fatalf("unexpected joined diagnostics: %q", joined)
	}
}

func TestDefaultCommandOutput(t *testing.T) {
	out, err := defaultCommandOutput("go", "version")
	if err != nil {
		t.Fatalf("go version failed: %v", err)
	}
	if !strings.Contains(string(out), "go version") {
		t.Fatalf("unexpected go version output: %s", string(out))
	}
}

func validManifest(t *testing.T, target, version string) CacheManifest {
	t.Helper()
	assetPath := filepath.Join(target, DefaultCacheRoot, version, "bin", "openspec")
	hash, err := fileSHA256(assetPath)
	if err != nil {
		t.Fatal(err)
	}
	return CacheManifest{SchemaVersion: 1, Version: version, Source: Source{Type: "fixture", Repository: "local"}, CreatedAt: time.Now().UTC(), Assets: []CacheAsset{{Path: "bin/openspec", SHA256: hash}}}
}

func writeCacheAsset(t *testing.T, target, version, rel, body string) {
	t.Helper()
	path := filepath.Join(target, DefaultCacheRoot, version, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func hasDiagnostic(diagnostics []Diagnostic, layer Layer, contains string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Layer == layer && strings.Contains(diagnostic.Message, contains) {
			return true
		}
	}
	return false
}
