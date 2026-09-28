package surfaceplan

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Service struct {
	projects    ProjectSource
	changes     ChangeSource
	resolver    SurfaceResolverStrategy
	validations ValidationPlanFactory
}

func NewService() Service {
	return NewServiceWith(ProjectConfigAdapter{}, GitChangeAdapter{})
}

func NewServiceWith(projects ProjectSource, changes ChangeSource) Service {
	return Service{projects: projects, changes: changes, resolver: SurfaceResolverStrategy{}, validations: ValidationPlanFactory{}}
}

func (s Service) Build(options Options) (ExecutionPlan, error) {
	target, err := platform.ResolveTargetPath(options.Target)
	if err != nil {
		return ExecutionPlan{}, fmt.Errorf("resolver target: %w", err)
	}
	project, err := s.projects.Load(target)
	if err != nil {
		return ExecutionPlan{}, err
	}
	base := strings.TrimSpace(options.Base)
	if base == "" {
		base = "HEAD"
	}
	if strings.HasPrefix(base, "-") {
		return ExecutionPlan{}, fmt.Errorf("--base no acepta opciones Git: %s", base)
	}
	explicitFiles := len(options.Files) > 0
	files, err := normalizeFiles(options.Files)
	if err != nil {
		return ExecutionPlan{}, err
	}
	if len(files) == 0 {
		files, err = s.changes.ChangedFiles(target, base)
		if err != nil {
			return ExecutionPlan{}, err
		}
		files, err = normalizeFiles(files)
		if err != nil {
			return ExecutionPlan{}, err
		}
	}
	requested := strings.TrimSpace(options.RequestedSurface)
	if requested == "" {
		requested = "auto"
	}
	resolution, err := s.resolver.Resolve(project, requested, files)
	if err != nil {
		return ExecutionPlan{}, err
	}
	rules, signals, capabilities := s.validations.Build(project, resolution, files, options.Capabilities)
	source := "git_diff"
	if explicitFiles {
		source = "files"
	}
	if requested != "auto" {
		source = "explicit"
	}
	return ExecutionPlan{
		SchemaVersion:    SchemaVersion,
		Target:           target,
		Base:             base,
		Source:           source,
		RequestedSurface: requested,
		Mode:             resolution.Mode,
		PrimarySurface:   resolution.PrimarySurface,
		ActiveSurfaces:   resolution.Surfaces,
		ChangedFiles:     files,
		Capabilities:     capabilities,
		Signals:          signals,
		Contracts:        resolution.Contracts,
		Decisions:        resolution.Decisions,
		ValidationRules:  rules,
	}, nil
}

func normalizeFiles(files []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		if filepath.IsAbs(file) {
			return nil, fmt.Errorf("--files sólo acepta paths relativos al target: %s", file)
		}
		clean := filepath.ToSlash(filepath.Clean(file))
		if clean == ".." || strings.HasPrefix(clean, "../") {
			return nil, fmt.Errorf("path fuera del target no permitido: %s", file)
		}
		clean = strings.TrimPrefix(clean, "./")
		if clean == "." || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	sort.Strings(out)
	return out, nil
}
