package opsx

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

func ReadCacheManifest(target, version string) (CacheManifest, string, error) {
	cacheDir, err := cacheVersionDir(target, version)
	if err != nil {
		return CacheManifest{}, "", err
	}
	manifestPath, err := platform.SafeJoin(cacheDir, DefaultManifestName)
	if err != nil {
		return CacheManifest{}, "", err
	}
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		return CacheManifest{}, "", err
	}
	manifest, err := ParseCacheManifest(body)
	if err != nil {
		return CacheManifest{}, "", fmt.Errorf("cache %s inválida: %w", version, err)
	}
	if err := ValidateCacheManifest(cacheDir, version, manifest); err != nil {
		return CacheManifest{}, "", err
	}
	return manifest, manifestPath, nil
}

func ParseCacheManifest(body []byte) (CacheManifest, error) {
	var manifest CacheManifest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&manifest); err != nil {
		return CacheManifest{}, err
	}
	return manifest, nil
}

func ValidateCacheManifest(cacheDir, expectedVersion string, manifest CacheManifest) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("schemaVersion de cache no soportado: %d", manifest.SchemaVersion)
	}
	version, err := normalizeVersion(manifest.Version)
	if err != nil {
		return err
	}
	expected, err := normalizeVersion(expectedVersion)
	if err != nil {
		return err
	}
	if version != expected {
		return fmt.Errorf("manifest version %s no coincide con directorio %s", manifest.Version, expectedVersion)
	}
	if manifest.Source.Type == "" {
		return fmt.Errorf("manifest cache sin source.type")
	}
	if manifest.CreatedAt.IsZero() {
		return fmt.Errorf("manifest cache sin createdAt")
	}
	for _, asset := range manifest.Assets {
		if err := validateCacheAsset(cacheDir, asset); err != nil {
			return err
		}
	}
	return nil
}

func WriteCacheManifest(target string, manifest CacheManifest) (string, error) {
	if manifest.CreatedAt.IsZero() {
		manifest.CreatedAt = time.Now().UTC()
	}
	cacheDir, err := cacheVersionDir(target, manifest.Version)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	if err := ValidateCacheManifest(cacheDir, manifest.Version, manifest); err != nil {
		return "", err
	}
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	body = append(body, '\n')
	manifestPath, err := platform.SafeJoin(cacheDir, DefaultManifestName)
	if err != nil {
		return "", err
	}
	if err := platform.WriteFileAtomic(manifestPath, body, 0o644); err != nil {
		return "", err
	}
	return manifestPath, nil
}

func cacheVersionDir(target, version string) (string, error) {
	if _, err := normalizeVersion(version); err != nil {
		return "", err
	}
	if strings.ContainsAny(version, `/\\`) {
		return "", fmt.Errorf("versión cache insegura: %q", version)
	}
	root, err := platform.SafeJoin(target, lufypaths.OpenSpecCache)
	if err != nil {
		return "", err
	}
	return platform.SafeJoin(root, version)
}

func validateCacheAsset(cacheDir string, asset CacheAsset) error {
	path, err := platform.SafeJoin(cacheDir, asset.Path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("asset cache inválido: %s", filepath.ToSlash(asset.Path))
	}
	if asset.SHA256 == "" {
		return nil
	}
	got, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if got != asset.SHA256 {
		return fmt.Errorf("sha256 no coincide para asset cache %s", filepath.ToSlash(asset.Path))
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
