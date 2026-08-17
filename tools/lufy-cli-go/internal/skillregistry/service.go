package skillregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/registry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

type Options struct {
	Target string
	Tool   domain.ToolID
	JSON   bool
}

type StatusReport struct {
	Status     string        `json:"status"`
	Path       string        `json:"path"`
	Tool       domain.ToolID `json:"tool"`
	SkillCount int           `json:"skillCount"`
	RootCount  int           `json:"rootCount"`
	Warnings   []Warning     `json:"warnings,omitempty"`
	Recovery   string        `json:"recovery,omitempty"`
	Updated    bool          `json:"updated,omitempty"`
}

type Service struct {
	Env ports.Env
}

func NewService() Service {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return Service{Env: ports.Env{"HOME": home, "XDG_CONFIG_HOME": os.Getenv("XDG_CONFIG_HOME")}}
}

func (s Service) Refresh(opts Options, out io.Writer) error {
	index, path, err := s.build(opts)
	if err != nil {
		return err
	}
	data, err := Marshal(index)
	if err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(path, data, 0o644); err != nil {
		return fmt.Errorf("escribir registry de skills: %w", err)
	}
	if opts.JSON {
		_, err = out.Write(data)
		return err
	}
	fmt.Fprintf(out, "Registry de skills actualizado: %s\n", path)
	fmt.Fprintf(out, "Tool: %s; skills: %d; roots: %d; warnings: %d\n", index.Tool, len(index.Skills), len(index.Roots), len(index.Warnings))
	return nil
}

func (s Service) Status(opts Options, out io.Writer) error {
	report, err := s.Inspect(opts)
	if err != nil {
		return err
	}
	return presentStatus(report, opts.JSON, out)
}

// Inspect returns the current registry state without mutating the filesystem.
func (s Service) Inspect(opts Options) (StatusReport, error) {
	report, _, err := s.inspect(opts)
	return report, err
}

func (s Service) inspect(opts Options) (StatusReport, []byte, error) {
	index, path, err := s.build(opts)
	if err != nil {
		return StatusReport{}, nil, err
	}
	expected, err := Marshal(index)
	if err != nil {
		return StatusReport{}, nil, err
	}
	report := StatusReport{Status: "ready", Path: path, Tool: index.Tool, SkillCount: len(index.Skills), RootCount: len(index.Roots), Warnings: index.Warnings}
	current, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		report.Status = "not_available"
		report.Recovery = recoveryCommand(opts.Target, index.Tool)
	} else if err != nil {
		return StatusReport{}, nil, fmt.Errorf("leer registry de skills: %w", err)
	} else if !bytes.Equal(current, expected) {
		report.Status = "stale"
		report.Recovery = recoveryCommand(opts.Target, index.Tool)
	}
	return report, expected, nil
}

// Ensure refreshes an absent or stale registry and leaves a ready registry untouched.
func (s Service) Ensure(opts Options, out io.Writer) error {
	report, expected, err := s.inspect(opts)
	if err != nil {
		return err
	}
	if report.Status != "ready" {
		if err := platform.WriteFileAtomic(report.Path, expected, 0o644); err != nil {
			return fmt.Errorf("escribir registry de skills: %w", err)
		}
		report.Status = "ready"
		report.Recovery = ""
		report.Updated = true
	}
	return presentStatus(report, opts.JSON, out)
}

func presentStatus(report StatusReport, jsonOutput bool, out io.Writer) error {
	if jsonOutput {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "%s\n", data)
		return err
	}
	fmt.Fprintf(out, "Skill registry: %s\n", report.Status)
	fmt.Fprintf(out, "Path: %s\nTool: %s; skills: %d; roots: %d; warnings: %d\n", report.Path, report.Tool, report.SkillCount, report.RootCount, len(report.Warnings))
	if report.Updated {
		fmt.Fprintln(out, "Acción: registry actualizado")
	}
	if report.Recovery != "" {
		fmt.Fprintf(out, "Recovery: %s\n", report.Recovery)
	}
	return nil
}

func recoveryCommand(target string, tool domain.ToolID) string {
	if target == "" {
		target = "."
	}
	return fmt.Sprintf("lufy-ai skills ensure --target \"%s\" --tool %s", target, tool)
}

func (s Service) build(opts Options) (Index, string, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Index{}, "", fmt.Errorf("resolver target: %w", err)
	}
	tool, err := resolveTool(target, opts.Tool)
	if err != nil {
		return Index{}, "", err
	}
	adapter, err := registry.Default().Tool(tool)
	if err != nil {
		return Index{}, "", err
	}
	if adapter.Capabilities().DryRunOnly || !adapter.Capabilities().Skills {
		return Index{}, "", fmt.Errorf("tool adapter no soportado por skills registry: %s; disponibles: opencode, codex", tool)
	}
	provider, ok := adapter.(ports.SkillRootProvider)
	if !ok {
		return Index{}, "", fmt.Errorf("tool adapter %s no declara raíces de skills", tool)
	}
	index, err := Build(tool, provider.SkillRoots(ports.Target{Root: target}, s.Env))
	if err != nil {
		return Index{}, "", err
	}
	path, err := lufypaths.WritePath(target, lufypaths.SkillRegistry)
	return index, path, err
}

func resolveTool(target string, requested domain.ToolID) (domain.ToolID, error) {
	if requested != "" {
		return requested, nil
	}
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return domain.ToolInitialDefault, nil
	} else if err != nil {
		return "", err
	}
	cfg, err := projectconfig.Load(path)
	if err != nil {
		return "", fmt.Errorf("leer tool desde project config: %w", err)
	}
	if cfg.Tool == "" {
		return domain.ToolInitialDefault, nil
	}
	return cfg.Tool, nil
}
