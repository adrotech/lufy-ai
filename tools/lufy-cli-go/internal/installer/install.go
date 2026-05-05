package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target   string
	DryRun   bool
	Yes      bool
	NoEngram bool
	Backup   bool
}

type Action struct {
	Kind   string
	Source string
	Target string
	Reason string
	Risk   string
}

type Conflict struct {
	Path   string
	Reason string
	Risk   string
}

type Plan struct {
	TargetRoot string
	Actions    []Action
	Conflicts  []Conflict
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	plan, err := s.BuildPlan(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Plan de instalación para %s\n", plan.TargetRoot)
	for _, a := range plan.Actions {
		fmt.Fprintf(stdout, "- [%s] %s (%s)\n", a.Kind, a.Target, a.Reason)
	}

	if opts.NoEngram {
		fmt.Fprintln(stdout, "Engram: omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		fmt.Fprintf(stdout, "Engram: detectado en PATH (%s)\n", path)
	} else {
		fmt.Fprintln(stdout, "Engram: no encontrado en PATH (instalación base continúa)")
	}

	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: sin mutaciones en filesystem")
		return nil
	}

	if err := s.applyInstall(plan.TargetRoot, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Install real completado (slice mínimo)")
	return nil
}

func (s Service) applyInstall(target string, stdout io.Writer) error {
	stateDir := filepath.Join(target, ".lufy-ai")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}

	statePath := filepath.Join(stateDir, "install-state.json")
	if _, err := os.Stat(statePath); err == nil {
		fmt.Fprintf(stdout, "- [skip] %s ya existe\n", statePath)
	} else {
		state := map[string]any{
			"schemaVersion": 1,
			"targetRoot":    target,
			"managedFiles":  []string{"AGENTS.md", ".lufy-ai/install-state.json"},
		}
		body, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(statePath, body, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "- [write] %s\n", statePath)
	}

	templatePath := filepath.Join(target, "AGENTS.md.template")
	agentsPath := filepath.Join(target, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		fmt.Fprintln(stdout, "- [skip] AGENTS.md ya existe")
	} else if content, readErr := os.ReadFile(templatePath); readErr == nil {
		if writeErr := os.WriteFile(agentsPath, content, 0o644); writeErr != nil {
			return writeErr
		}
		fmt.Fprintln(stdout, "- [copy] AGENTS.md desde AGENTS.md.template")
	}

	return nil
}

func (s Service) BuildPlan(opts Options) (Plan, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Plan{}, err
	}

	return Plan{
		TargetRoot: target,
		Actions: []Action{
			{Kind: "verify", Target: target, Reason: "Slice inicial: planificar sin escribir", Risk: "low"},
		},
	}, nil
}
