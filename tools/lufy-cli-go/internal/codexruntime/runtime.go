// Package codexruntime observa metadatos de Codex con probes read-only acotados.
package codexruntime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxOutput = 64 * 1024

type Options struct {
	Executable, Target string
	Timeout            time.Duration
}

type Report struct {
	SchemaVersion    string          `json:"schema_version"`
	Status           string          `json:"status"`
	Executable       string          `json:"executable,omitempty"`
	PathExecutable   string          `json:"path_executable,omitempty"`
	Selection        string          `json:"selection"`
	Version          string          `json:"version,omitempty"`
	FeaturesStatus   string          `json:"features_status"`
	NativeDelegation string          `json:"native_delegation"`
	Features         map[string]bool `json:"features,omitempty"`
	Reasons          []string        `json:"reasons,omitempty"`
}

type Service struct {
	LookPath func(string) (string, error)
	Run      func(context.Context, string, []string, string) ([]byte, error)
}

func NewService() Service { return Service{LookPath: exec.LookPath, Run: run} }

var versionPattern = regexp.MustCompile(`^codex-cli v?([0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?)$`)

func (s Service) Inspect(ctx context.Context, opts Options) Report {
	r := Report{SchemaVersion: "codex-runtime-report/v1", Status: "unavailable", Selection: "path", FeaturesStatus: "unavailable", NativeDelegation: "unknown"}
	if s.LookPath == nil || s.Run == nil {
		r.Reasons = append(r.Reasons, "service_unavailable")
		return r
	}
	if ctx.Err() != nil {
		r.Reasons = append(r.Reasons, "context_canceled")
		return r
	}
	path, _ := s.LookPath("codex")
	if path != "" {
		r.PathExecutable, _ = filepath.Abs(path)
	}
	name := "codex"
	if opts.Executable != "" {
		name = opts.Executable
		r.Selection = "explicit"
	}
	executable, err := s.LookPath(name)
	if err != nil {
		r.Reasons = append(r.Reasons, "executable_unavailable")
		return r
	}
	r.Executable, err = filepath.Abs(executable)
	if err != nil {
		r.Reasons = append(r.Reasons, "executable_unavailable")
		return r
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	probe := func(label string, args ...string) []byte {
		if ctx.Err() != nil {
			r.Reasons = append(r.Reasons, label+"_context_canceled")
			return nil
		}
		out, err := s.Run(ctx, r.Executable, args, opts.Target)
		if err != nil {
			r.Reasons = append(r.Reasons, label+"_probe_failed")
			return nil
		}
		if len(out) > maxOutput {
			r.Reasons = append(r.Reasons, label+"_output_limit")
			return nil
		}
		return out
	}
	if out := probe("version", "--version"); out != nil {
		for _, line := range strings.Split(string(out), "\n") {
			if match := versionPattern.FindStringSubmatch(strings.TrimSpace(line)); match != nil {
				if r.Version != "" {
					r.Version = ""
					break
				}
				r.Version = match[1]
			}
		}
		if r.Version == "" {
			r.Reasons = append(r.Reasons, "version_invalid")
		}
	}
	if out := probe("features", "features", "list"); out != nil {
		features, valid := parseFeatures(string(out))
		if valid {
			r.Features = features
			r.FeaturesStatus = "available"
		} else {
			r.Reasons = append(r.Reasons, "features_invalid")
		}
	}
	if r.Version != "" || r.FeaturesStatus == "available" {
		r.Status = "partial"
	}
	if r.Version != "" && r.FeaturesStatus == "available" {
		r.Status = "available"
	}
	return r
}

func parseFeatures(output string) (map[string]bool, bool) {
	features := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "multi_agent", "hooks", "apps", "plugins":
		default:
			continue
		}
		if _, duplicate := features[fields[0]]; duplicate || len(fields) < 3 {
			return nil, false
		}
		value := fields[len(fields)-1]
		if value != "true" && value != "false" {
			return nil, false
		}
		features[fields[0]] = value == "true"
	}
	return features, len(features) > 0
}

type limitedBuffer struct{ buffer bytes.Buffer }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := maxOutput - b.buffer.Len()
	if len(p) > remaining {
		_, _ = b.buffer.Write(p[:remaining])
		return remaining, errors.New("output_limit")
	}
	return b.buffer.Write(p)
}

func run(ctx context.Context, executable string, args []string, cwd string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = cwd
	cmd.WaitDelay = 100 * time.Millisecond
	var out limitedBuffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	err := cmd.Run()
	return out.buffer.Bytes(), err
}
