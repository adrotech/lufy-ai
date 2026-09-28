package codexsurface

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Check struct {
	Level   string
	Path    string
	Message string
}

type hookFile struct {
	Hooks map[string][]hookGroup `json:"hooks"`
}

type hookGroup struct {
	Hooks []hookHandler `json:"hooks"`
}

type hookHandler struct {
	Type           string `json:"type"`
	Command        string `json:"command"`
	CommandWindows string `json:"commandWindows"`
}

func Validate(target string) []Check {
	var checks []Check
	checks = append(checks, validateConfig(filepath.Join(target, ".codex", "config.toml"))...)
	checks = append(checks, validateHooks(filepath.Join(target, ".codex", "hooks.json"))...)
	checks = append(checks, validateRules(filepath.Join(target, ".codex", "rules", "lufy.rules"))...)
	return checks
}

func validateConfig(path string) []Check {
	const rel = ".codex/config.toml"
	body, err := readRegular(path)
	if err != nil {
		return []Check{{Level: "fail", Path: rel, Message: err.Error()}}
	}
	text := string(body)
	required := []string{"[features]", "multi_agent = true", "hooks = true"}
	for _, want := range required {
		if !strings.Contains(text, want) {
			return []Check{{Level: "fail", Path: rel, Message: fmt.Sprintf("config Codex incompleta: falta %q", want)}}
		}
	}
	for _, unsupported := range []string{"max_depth", "max_threads ="} {
		if strings.Contains(text, unsupported) {
			return []Check{{Level: "fail", Path: rel, Message: fmt.Sprintf("config Codex conserva key no canónica: %s", unsupported)}}
		}
	}
	return []Check{{Level: "ok", Path: rel, Message: "config Codex usa hooks y multi-agent sin límites opcionales incompatibles"}}
}

func validateHooks(path string) []Check {
	const rel = ".codex/hooks.json"
	body, err := readRegular(path)
	if err != nil {
		return []Check{{Level: "fail", Path: rel, Message: err.Error()}}
	}
	var config hookFile
	if err := json.Unmarshal(body, &config); err != nil {
		return []Check{{Level: "fail", Path: rel, Message: "hooks Codex JSON inválido: " + err.Error()}}
	}
	for _, event := range []string{"SessionStart", "SubagentStop", "Stop", "SessionEnd"} {
		groups := config.Hooks[event]
		if len(groups) == 0 {
			return []Check{{Level: "fail", Path: rel, Message: "lifecycle Codex incompleto: falta " + event}}
		}
		found := false
		for _, group := range groups {
			for _, hook := range group.Hooks {
				if hook.Type == "command" && strings.Contains(hook.Command, "lufy-ai lifecycle codex") && hook.CommandWindows != "" {
					found = true
				}
			}
		}
		if !found {
			return []Check{{Level: "fail", Path: rel, Message: event + " no declara comando portable lufy-ai lifecycle codex"}}
		}
	}
	return []Check{{Level: "ok", Path: rel, Message: "lifecycle Codex configurado para SessionStart, SubagentStop, Stop y SessionEnd"}}
}

func validateRules(path string) []Check {
	const rel = ".codex/rules/lufy.rules"
	body, err := readRegular(path)
	if err != nil {
		return []Check{{Level: "fail", Path: rel, Message: err.Error()}}
	}
	text := string(body)
	required := []string{
		"prefix_rule(",
		"decision = \"prompt\"",
		"decision = \"forbidden\"",
		"match = [",
		"not_match = [",
		"[\"commit\", \"push\", \"merge\", \"tag\"]",
		"[\"create\", \"merge\", \"ready\", \"review\"]",
		"pattern = [\"git\", \"reset\", \"--hard\"]",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			return []Check{{Level: "fail", Path: rel, Message: fmt.Sprintf("rules Codex incompletas: falta %q", want)}}
		}
	}
	return []Check{{Level: "ok", Path: rel, Message: "rules Codex conservadoras con ejemplos match/not_match"}}
}

func readRegular(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("asset Codex no disponible: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("asset Codex no es archivo regular")
	}
	return os.ReadFile(path)
}
