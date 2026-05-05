package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const OpenCodeFile = "opencode.json"

type Options struct {
	TargetRoot string
	NoEngram   bool
	Resolver   platform.CommandResolver
}

type Result struct {
	Path        string
	Action      string
	EngramPath  string
	EngramFound bool
	Changed     bool
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Plan(opts Options) (Result, error) {
	current, desired, path, engramPath, engramFound, err := loadAndMerge(opts)
	if err != nil {
		return Result{}, err
	}
	if current == nil {
		return Result{Path: path, Action: "merge-json", EngramPath: engramPath, EngramFound: engramFound, Changed: true}, nil
	}
	if reflect.DeepEqual(current, desired) {
		return Result{Path: path, Action: "skip", EngramPath: engramPath, EngramFound: engramFound}, nil
	}
	return Result{Path: path, Action: "merge-json", EngramPath: engramPath, EngramFound: engramFound, Changed: true}, nil
}

func (s Service) Ensure(opts Options) (Result, error) {
	_, desired, path, engramPath, engramFound, err := loadAndMerge(opts)
	if err != nil {
		return Result{}, err
	}
	body, err := json.MarshalIndent(desired, "", "  ")
	if err != nil {
		return Result{}, err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Result{}, err
	}
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, body) {
		return Result{Path: path, Action: "skip", EngramPath: engramPath, EngramFound: engramFound}, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".opencode-*.json.tmp")
	if err != nil {
		return Result{}, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return Result{}, err
	}
	if err := tmp.Close(); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return Result{}, err
	}
	return Result{Path: path, Action: "merge-json", EngramPath: engramPath, EngramFound: engramFound, Changed: true}, nil
}

func loadAndMerge(opts Options) (map[string]any, map[string]any, string, string, bool, error) {
	resolver := opts.Resolver
	if resolver == nil {
		resolver = platform.OSResolver{}
	}
	path, err := platform.SafeJoin(opts.TargetRoot, OpenCodeFile)
	if err != nil {
		return nil, nil, "", "", false, err
	}
	var current map[string]any
	body, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(body, &current); err != nil {
			return nil, nil, "", "", false, fmt.Errorf("opencode.json inválido; corrige JSON o respáldalo antes de instalar: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, nil, "", "", false, err
	}
	desired := cloneMap(current)
	if desired == nil {
		desired = map[string]any{}
	}
	if _, ok := desired["$schema"]; !ok {
		desired["$schema"] = "https://opencode.ai/config.json"
	}
	if _, ok := desired["plugin"]; !ok {
		desired["plugin"] = []any{}
	}
	engramPath, engramFound := platform.ResolveEngram(opts.NoEngram, resolver)
	if !opts.NoEngram && engramFound {
		mcp := objectAt(desired, "mcp")
		mcp["engram"] = map[string]any{
			"type":    "local",
			"command": []any{engramPath, "mcp", "--tools=agent"},
			"enabled": true,
			"timeout": float64(3000),
		}
	}
	return current, desired, path, engramPath, engramFound, nil
}

func objectAt(root map[string]any, key string) map[string]any {
	if existing, ok := root[key].(map[string]any); ok {
		return existing
	}
	obj := map[string]any{}
	root[key] = obj
	return obj
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	body, _ := json.Marshal(in)
	var out map[string]any
	_ = json.Unmarshal(body, &out)
	return out
}
