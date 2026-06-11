package opsx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const (
	DefaultCacheRoot    = lufypaths.OpenSpecCache
	DefaultManifestName = "manifest.json"
	embeddedUpstreamRel = "openspec/UPSTREAM.json"
)

type ResolveOptions struct {
	Target string
}

type Service struct {
	lookPath      func(string) (string, error)
	commandOutput func(string, ...string) ([]byte, error)
}

func NewService() *Service {
	return &Service{lookPath: exec.LookPath, commandOutput: defaultCommandOutput}
}

func (s *Service) Resolve(opts ResolveOptions) (Resolution, error) {
	if s.lookPath == nil {
		s.lookPath = exec.LookPath
	}
	if s.commandOutput == nil {
		s.commandOutput = defaultCommandOutput
	}
	target := opts.Target
	if target == "" {
		target = "."
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return Resolution{}, err
	}

	baseline, diagnostics, err := loadBaseline(absTarget)
	if err != nil {
		return Resolution{Diagnostics: diagnostics}, err
	}
	minimum := baseline.MinimumCompatibleSpecVersion
	if minimum == "" {
		minimum = baseline.EffectiveOpenSpecVersion
	}

	res, diag := s.resolvePath(minimum)
	diagnostics = append(diagnostics, diag...)
	if res.Version != "" {
		res.Diagnostics = diagnostics
		return res, nil
	}

	res, diag = resolveCache(absTarget, minimum)
	diagnostics = append(diagnostics, diag...)
	if res.Version != "" {
		res.Diagnostics = diagnostics
		return res, nil
	}

	embedded, err := embeddedBaseline()
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{Layer: LayerEmbedded, Message: err.Error()})
		return Resolution{Diagnostics: diagnostics}, fmt.Errorf("no hay capas OpenSpec válidas; reinstala assets, limpia cache o revisa openspec/UPSTREAM.json")
	}
	version := embedded.EffectiveOpenSpecVersion
	if !CompatibleVersion(version, minimum) {
		msg := fmt.Sprintf("baseline embebida %s no cumple mínimo %s", version, minimum)
		diagnostics = append(diagnostics, Diagnostic{Layer: LayerEmbedded, Message: msg})
		return Resolution{Diagnostics: diagnostics}, fmt.Errorf("no hay capas OpenSpec válidas; %s", msg)
	}
	return Resolution{Layer: LayerEmbedded, Version: version, Source: embedded.Source, Path: assets.EmbeddedSourceRoot, Diagnostics: diagnostics}, nil
}

func (s *Service) resolvePath(minimum string) (Resolution, []Diagnostic) {
	path, err := s.lookPath("openspec")
	if err != nil {
		return Resolution{}, []Diagnostic{{Layer: LayerPath, Message: "openspec no está disponible en PATH"}}
	}
	out, err := s.commandOutput(path, "--version")
	if err != nil {
		return Resolution{}, []Diagnostic{{Layer: LayerPath, Message: fmt.Sprintf("no se pudo ejecutar %s --version: %v", path, err)}}
	}
	version, err := ParseVersion(string(out))
	if err != nil {
		return Resolution{}, []Diagnostic{{Layer: LayerPath, Message: err.Error()}}
	}
	if !CompatibleVersion(version, minimum) {
		return Resolution{}, []Diagnostic{{Layer: LayerPath, Message: fmt.Sprintf("openspec en PATH tiene versión %s menor que mínimo %s", version, minimum)}}
	}
	return Resolution{Layer: LayerPath, Version: version, Path: path, Source: Source{Type: "path", Description: "openspec CLI disponible en PATH"}}, nil
}

func loadBaseline(target string) (Upstream, []Diagnostic, error) {
	path, err := platform.SafeJoin(target, embeddedUpstreamRel)
	if err == nil {
		body, readErr := os.ReadFile(path)
		if readErr == nil {
			upstream, parseErr := parseUpstream(body)
			if parseErr == nil {
				return upstream, nil, nil
			}
			return Upstream{}, []Diagnostic{{Layer: LayerEmbedded, Message: fmt.Sprintf("openspec/UPSTREAM.json inválido: %v", parseErr)}}, parseErr
		}
		if !errors.Is(readErr, os.ErrNotExist) {
			return Upstream{}, []Diagnostic{{Layer: LayerEmbedded, Message: readErr.Error()}}, readErr
		}
	}
	upstream, err := embeddedBaseline()
	if err != nil {
		return Upstream{}, []Diagnostic{{Layer: LayerEmbedded, Message: err.Error()}}, err
	}
	return upstream, nil, nil
}

func embeddedBaseline() (Upstream, error) {
	body, err := assets.ReadSourceFile(assets.EmbeddedSourceRoot, embeddedUpstreamRel)
	if err != nil {
		return Upstream{}, err
	}
	return parseUpstream(body)
}

func parseUpstream(body []byte) (Upstream, error) {
	var upstream Upstream
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&upstream); err != nil {
		return Upstream{}, err
	}
	if upstream.SchemaVersion != 1 {
		return Upstream{}, fmt.Errorf("schemaVersion no soportado: %d", upstream.SchemaVersion)
	}
	if upstream.Workflow == "" || upstream.EffectiveOpenSpecVersion == "" || upstream.Source.Type == "" {
		return Upstream{}, fmt.Errorf("UPSTREAM.json incompleto")
	}
	if _, err := normalizeVersion(upstream.EffectiveOpenSpecVersion); err != nil {
		return Upstream{}, err
	}
	if upstream.MinimumCompatibleSpecVersion != "" {
		if _, err := normalizeVersion(upstream.MinimumCompatibleSpecVersion); err != nil {
			return Upstream{}, err
		}
	}
	return upstream, nil
}

func resolveCache(target, minimum string) (Resolution, []Diagnostic) {
	resolved, err := lufypaths.ResolveExisting(target, lufypaths.OpenSpecCache, lufypaths.LegacyOpenSpecCache)
	if err != nil {
		return Resolution{}, []Diagnostic{{Layer: LayerCache, Message: err.Error()}}
	}
	root := resolved.Path
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Resolution{}, []Diagnostic{{Layer: LayerCache, Message: "cache OpenSpec no existe"}}
		}
		return Resolution{}, []Diagnostic{{Layer: LayerCache, Message: err.Error()}}
	}
	sort.Slice(entries, func(i, j int) bool { return compareVersionSafe(entries[i].Name(), entries[j].Name()) > 0 })
	var diagnostics []Diagnostic
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, manifestPath, err := ReadCacheManifest(target, entry.Name())
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Layer: LayerCache, Message: err.Error()})
			continue
		}
		if !CompatibleVersion(manifest.Version, minimum) {
			diagnostics = append(diagnostics, Diagnostic{Layer: LayerCache, Message: fmt.Sprintf("cache %s menor que mínimo %s", manifest.Version, minimum)})
			continue
		}
		return Resolution{Layer: LayerCache, Version: manifest.Version, Path: filepath.Dir(manifestPath), Source: manifest.Source}, diagnostics
	}
	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, Diagnostic{Layer: LayerCache, Message: "cache OpenSpec sin versiones válidas"})
	}
	return Resolution{}, diagnostics
}

func compareVersionSafe(left, right string) int {
	l, err := normalizeVersion(left)
	if err != nil {
		return -1
	}
	r, err := normalizeVersion(right)
	if err != nil {
		return 1
	}
	return compareVersion(l, r)
}

func defaultCommandOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func joinDiagnosticMessages(diagnostics []Diagnostic) string {
	var parts []string
	for _, diagnostic := range diagnostics {
		parts = append(parts, string(diagnostic.Layer)+": "+diagnostic.Message)
	}
	return strings.Join(parts, "; ")
}
