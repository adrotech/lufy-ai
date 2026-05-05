package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	SchemaVersion  = 1
	ToolVersion    = "dev"
	SourceChangeID = "install-managed-assets-with-hash-idempotency"
)

type InstallState struct {
	SchemaVersion         int          `json:"schemaVersion"`
	ToolVersion           string       `json:"toolVersion"`
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
	tmp, err := os.CreateTemp(filepath.Dir(path), ".install-state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func New(targetRoot string, previous *InstallState, assets []AssetState) InstallState {
	now := time.Now().UTC().Format(time.RFC3339)
	installedAt := now
	if previous != nil && previous.InstalledAt != "" {
		installedAt = previous.InstalledAt
	}
	return InstallState{SchemaVersion: SchemaVersion, ToolVersion: ToolVersion, SourceChangeID: SourceChangeID, SourceRootFingerprint: "dev-checkout", InstalledAt: installedAt, UpdatedAt: now, TargetRoot: targetRoot, Assets: assets}
}

func (s InstallState) AssetMap() map[string]AssetState {
	out := make(map[string]AssetState, len(s.Assets))
	for _, asset := range s.Assets {
		out[asset.TargetRel] = asset
	}
	return out
}
