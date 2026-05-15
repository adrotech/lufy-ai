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

type ValidationResult struct {
	Exists bool
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
	if err := validateRegularOpenCode(path); err != nil {
		return nil, nil, "", "", false, err
	}
	var current map[string]any
	body, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(body, &current); err != nil {
			return nil, nil, "", "", false, fmt.Errorf("opencode.json inválido; corrige JSON o respáldalo antes de instalar: %w", err)
		}
		if err := validateManagedKeyTypes(current); err != nil {
			return nil, nil, "", "", false, err
		}
	} else if !os.IsNotExist(err) {
		return nil, nil, "", "", false, err
	}
	desired := cloneMap(current)
	if desired == nil {
		desired = map[string]any{}
	}
	// OpenCode rejects unknown top-level keys, so do not persist lufy metadata
	// inside opencode.json. Existing installs from older versions are cleaned up
	// during the next merge.
	delete(desired, "x-lufy-ai")
	if _, ok := desired["$schema"]; !ok {
		desired["$schema"] = "https://opencode.ai/config.json"
	}
	if _, ok := desired["plugin"]; !ok {
		desired["plugin"] = []any{}
	}
	engramPath, engramFound := platform.ResolveEngram(opts.NoEngram, resolver)
	if !opts.NoEngram && engramFound {
		mcp, err := objectAt(desired, "mcp")
		if err != nil {
			return nil, nil, "", "", false, err
		}
		mcp["engram"] = mergeEngramConfig(mcp["engram"], engramPath)
	}
	return current, desired, path, engramPath, engramFound, nil
}

func (s Service) ValidateManagedStructure(targetRoot string) (ValidationResult, error) {
	path, err := platform.SafeJoin(targetRoot, OpenCodeFile)
	if err != nil {
		return ValidationResult{}, err
	}
	if err := validateRegularOpenCode(path); err != nil {
		return ValidationResult{}, err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ValidationResult{}, fmt.Errorf("falta %s merge-managed", OpenCodeFile)
	}
	if err != nil {
		return ValidationResult{}, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ValidationResult{}, err
	}
	if err := validateManagedKeyTypes(decoded); err != nil {
		return ValidationResult{Exists: true}, err
	}
	if _, ok := decoded["$schema"].(string); !ok {
		return ValidationResult{Exists: true}, fmt.Errorf("falta clave gestionada mínima $schema")
	}
	if _, ok := decoded["plugin"].([]any); !ok {
		return ValidationResult{Exists: true}, fmt.Errorf("falta clave gestionada mínima plugin")
	}
	return ValidationResult{Exists: true}, nil
}

func validateRegularOpenCode(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%s debe ser un archivo regular seguro; symlinks, directorios y archivos especiales no están permitidos", OpenCodeFile)
	}
	return nil
}

func validateManagedKeyTypes(decoded map[string]any) error {
	if schema, ok := decoded["$schema"]; ok {
		if _, valid := schema.(string); !valid {
			return fmt.Errorf("opencode.json inválido: clave gestionada $schema debe ser string; corrige o respalda el archivo antes de instalar/sincronizar")
		}
	}
	if plugin, ok := decoded["plugin"]; ok {
		if _, valid := plugin.([]any); !valid {
			return fmt.Errorf("opencode.json inválido: clave gestionada plugin debe ser array; corrige o respalda el archivo antes de instalar/sincronizar")
		}
	}
	return nil
}

func objectAt(root map[string]any, key string) (map[string]any, error) {
	if existing, ok := root[key].(map[string]any); ok {
		return existing, nil
	}
	if value, exists := root[key]; exists && value == nil {
		obj := map[string]any{}
		root[key] = obj
		return obj, nil
	}
	if _, exists := root[key]; exists {
		return nil, fmt.Errorf("opencode.json inválido: clave %s debe ser object para configurar Engram; corrige o respalda el archivo antes de instalar/sincronizar", key)
	}
	obj := map[string]any{}
	root[key] = obj
	return obj, nil
}

func mergeEngramConfig(existing any, engramPath string) map[string]any {
	engram, ok := existing.(map[string]any)
	if !ok {
		engram = map[string]any{}
	} else {
		engram = cloneMap(engram)
	}
	if _, ok := engram["enabled"]; !ok {
		engram["enabled"] = true
	}
	if _, ok := engram["timeout"]; !ok {
		engram["timeout"] = float64(3000)
	}
	engram["type"] = "local"
	engram["command"] = []any{engramPath, "mcp", "--tools=agent"}
	return engram
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
