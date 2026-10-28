package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

const SchemaVersion = 4

type InstallState struct {
	SchemaVersion         int                    `json:"schemaVersion"`
	ToolVersion           string                 `json:"toolVersion"`
	ToolCommit            string                 `json:"toolCommit,omitempty"`
	ToolBuildDate         string                 `json:"toolBuildDate,omitempty"`
	SourceChangeID        string                 `json:"sourceChangeID"`
	SourceRootFingerprint string                 `json:"sourceRootFingerprint"`
	Tool                  domain.ToolID          `json:"tool"`
	Execution             domain.ExecutionConfig `json:"execution"`
	InstalledAt           string                 `json:"installedAt"`
	UpdatedAt             string                 `json:"updatedAt"`
	TargetRoot            string                 `json:"targetRoot"`
	Assets                []AssetState           `json:"assets"`
}

type AssetState struct {
	ID           string `json:"id"`
	SourceRel    string `json:"sourceRel"`
	TargetRel    string `json:"targetRel"`
	SourceSHA256 string `json:"sourceSHA256"`
	TargetSHA256 string `json:"targetSHA256"`
	Policy       string `json:"policy,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Tool         string `json:"tool,omitempty"`
	Component    string `json:"component,omitempty"`
	AncestorRel  string `json:"ancestorRel,omitempty"`
	AncestorHash string `json:"ancestorSHA256,omitempty"`
	Pinned       bool   `json:"pinned,omitempty"`
	PinnedAt     string `json:"pinnedAt,omitempty"`
	PinnedReason string `json:"pinnedReason,omitempty"`
	InstalledAt  string `json:"installedAt"`
	LastAction   string `json:"lastAction"`
}

func Path(targetRoot string) string {
	return filepath.Join(targetRoot, lufypaths.InstallState)
}

func ExistingPath(targetRoot string) (string, error) {
	if _, err := os.Lstat(filepath.Join(targetRoot, lufypaths.RetiredInstallState)); err == nil {
		return "", fmt.Errorf("install-state en ruta retirada; respalda la instalación antes de sustituirla")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return Path(targetRoot), nil
}

func Load(targetRoot string) (*InstallState, error) {
	path, err := ExistingPath(targetRoot)
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return Decode(body)
}

// Decode valida exclusivamente el formato actual, sin traducir generaciones previas.
func Decode(body []byte) (*InstallState, error) {
	var st InstallState
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&st); err != nil {
		return nil, fmt.Errorf("install-state.json inválido: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("install-state.json debe contener un único objeto")
	}
	if st.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("schema de install-state.json no soportado: %d", st.SchemaVersion)
	}
	if st.Tool == "" || st.Execution.Mode == "" || st.Execution.Delegation == "" {
		return nil, fmt.Errorf("install-state.json requiere tool y execution explícitos")
	}
	if err := normalize(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

func normalize(st *InstallState) error {
	st.SchemaVersion = SchemaVersion
	cfg := domain.HarnessConfig{Tool: st.Tool, Execution: st.Execution}.WithDefaults()
	if err := cfg.ValidateSupported(); err != nil {
		return err
	}
	st.Tool = cfg.Tool
	st.Execution = cfg.Execution
	for i := range st.Assets {
		asset := &st.Assets[i]
		if asset.Policy == "" {
			asset.Policy = string(assets.PolicyManaged)
		}
		if !assets.Policy(asset.Policy).Valid() {
			return fmt.Errorf("policy de install-state.json no soportada para %s: %s", asset.TargetRel, asset.Policy)
		}
		if asset.Scope == "" {
			asset.Scope = string(assets.ScopeProject)
		}
		if !assets.Scope(asset.Scope).Valid() {
			return fmt.Errorf("scope de install-state.json no soportado para %s: %s", asset.TargetRel, asset.Scope)
		}
		if asset.Tool == "" {
			asset.Tool = string(cfg.Tool)
		}
		if !domain.ToolID(asset.Tool).Valid() {
			return fmt.Errorf("tool de install-state.json no soportada para %s: %s", asset.TargetRel, asset.Tool)
		}
		if asset.Component == "" {
			asset.Component = "managed"
		}
		if strings.HasPrefix(filepath.ToSlash(asset.AncestorRel), lufypaths.RetiredAncestors+"/") {
			return fmt.Errorf("ancestor en ruta retirada: %s", asset.AncestorRel)
		}
	}
	return nil
}

func WriteAtomic(targetRoot string, st InstallState) error {
	if st.SchemaVersion == 0 {
		st.SchemaVersion = SchemaVersion
	}
	if st.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema de install-state.json no soportado: %d", st.SchemaVersion)
	}
	if err := normalize(&st); err != nil {
		return err
	}
	path := Path(targetRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return platform.WriteFileAtomic(path, body, 0o644)
}

func New(targetRoot string, previous *InstallState, assets []AssetState, sourceRootFingerprint string) InstallState {
	return NewWithHarness(targetRoot, previous, assets, sourceRootFingerprint, domain.DefaultHarnessConfig())
}

func NewWithHarness(targetRoot string, previous *InstallState, assets []AssetState, sourceRootFingerprint string, cfg domain.HarnessConfig) InstallState {
	now := time.Now().UTC().Format(time.RFC3339)
	installedAt := now
	if previous != nil && previous.InstalledAt != "" {
		installedAt = previous.InstalledAt
	}
	info := version.Current()
	cfg = cfg.WithDefaults()
	st := InstallState{SchemaVersion: SchemaVersion, ToolVersion: info.Version, ToolCommit: info.Commit, ToolBuildDate: info.BuildDate, SourceChangeID: sourceRootFingerprint, SourceRootFingerprint: sourceRootFingerprint, Tool: cfg.Tool, Execution: cfg.Execution, InstalledAt: installedAt, UpdatedAt: now, TargetRoot: targetRoot, Assets: assets}
	_ = normalize(&st)
	return st
}

func (s InstallState) AssetMap() map[string]AssetState {
	out := make(map[string]AssetState, len(s.Assets))
	for _, asset := range s.Assets {
		out[asset.TargetRel] = asset
	}
	return out
}
