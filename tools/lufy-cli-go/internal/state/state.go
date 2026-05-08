package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

const (
	SchemaVersion = 1
)

type InstallState struct {
	SchemaVersion         int          `json:"schemaVersion"`
	ToolVersion           string       `json:"toolVersion"`
	ToolCommit            string       `json:"toolCommit,omitempty"`
	ToolBuildDate         string       `json:"toolBuildDate,omitempty"`
	SourceChangeID        string       `json:"sourceChangeID"`
	SourceRootFingerprint string       `json:"sourceRootFingerprint"`
	InstalledAt           string       `json:"installedAt"`
	UpdatedAt             string       `json:"updatedAt"`
	TargetRoot            string       `json:"targetRoot"`
	Assets                []AssetState `json:"assets"`
}

type AssetState struct {
	ID           string `json:"id"`
	SourceRel    string `json:"sourceRel"`
	TargetRel    string `json:"targetRel"`
	SourceSHA256 string `json:"sourceSHA256"`
	TargetSHA256 string `json:"targetSHA256"`
	InstalledAt  string `json:"installedAt"`
	LastAction   string `json:"lastAction"`
}

func Path(targetRoot string) string {
	return filepath.Join(targetRoot, ".lufy-ai", "install-state.json")
}

func Load(targetRoot string) (*InstallState, error) {
	body, err := os.ReadFile(Path(targetRoot))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var st InstallState
	if err := json.Unmarshal(body, &st); err != nil {
		return nil, fmt.Errorf("install-state.json inválido: %w", err)
	}
	if st.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("schema de install-state.json no soportado: %d", st.SchemaVersion)
	}
	return &st, nil
}

func WriteAtomic(targetRoot string, st InstallState) error {
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
	now := time.Now().UTC().Format(time.RFC3339)
	installedAt := now
	if previous != nil && previous.InstalledAt != "" {
		installedAt = previous.InstalledAt
	}
	info := version.Current()
	return InstallState{SchemaVersion: SchemaVersion, ToolVersion: info.Version, ToolCommit: info.Commit, ToolBuildDate: info.BuildDate, SourceChangeID: sourceRootFingerprint, SourceRootFingerprint: sourceRootFingerprint, InstalledAt: installedAt, UpdatedAt: now, TargetRoot: targetRoot, Assets: assets}
}

func (s InstallState) AssetMap() map[string]AssetState {
	out := make(map[string]AssetState, len(s.Assets))
	for _, asset := range s.Assets {
		out[asset.TargetRel] = asset
	}
	return out
}
