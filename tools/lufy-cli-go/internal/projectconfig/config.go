package projectconfig

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target        string
	Force         bool
	Rescan        bool
	ProfilePrompt ProfilePrompt
}

type ProfilePrompt func(ProjectConfig) (ProjectProfile, error)

type configScanner interface {
	Scan(string) (ProjectConfig, error)
}

type rescanPlanner interface {
	Build(ProjectConfig, ProjectConfig) RescanPlan
}

type configStore interface {
	Exists(string) (bool, error)
	Load(string) (ProjectConfig, error)
	Write(string, ProjectConfig) error
}

type Service struct {
	Now     func() time.Time
	scanner configScanner
	merger  rescanPlanner
	store   configStore
}

func NewService() Service {
	return Service{Now: func() time.Time { return time.Now().UTC() }}.withDefaults()
}

func (s Service) withDefaults() Service {
	now := s.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	s.Now = now
	if s.scanner == nil {
		s.scanner = Scanner{Now: now}
	}
	if s.merger == nil {
		s.merger = RescanMerger{}
	}
	if s.store == nil {
		s.store = ConfigStore{}
	}
	return s
}

func (s Service) Run(opts Options, out io.Writer) error {
	s = s.withDefaults()
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return fmt.Errorf("resolver target: %w", err)
	}
	configPath := filepath.Join(target, ProjectConfigPath)
	exists, err := s.store.Exists(configPath)
	if err != nil {
		return err
	}
	if exists && !opts.Force && !opts.Rescan {
		return fmt.Errorf("%s ya existe; usa --rescan para preservar overrides o --force para reemplazarlo", ProjectConfigPath)
	}

	detected, err := s.scanner.Scan(target)
	if err != nil {
		return err
	}
	finalConfig := detected
	var report *RescanPlan
	if opts.Rescan && exists {
		current, err := s.store.Load(configPath)
		if err != nil {
			return fmt.Errorf("leer config existente para rescan: corrige o respalda %s antes de reintentar: %w", ProjectConfigPath, err)
		}
		plan := s.merger.Build(current, detected)
		report = &plan
		finalConfig = plan.Merged
		if !plan.HasChanges && opts.ProfilePrompt == nil {
			printRescanReport(out, plan)
			return nil
		}
	}
	if opts.ProfilePrompt != nil {
		profile, err := opts.ProfilePrompt(finalConfig)
		if err != nil {
			return err
		}
		finalConfig.ProjectProfile = profile
	}
	if report != nil && !report.HasChanges {
		current, err := s.store.Load(configPath)
		if err != nil {
			return fmt.Errorf("leer config existente tras profile prompt: %w", err)
		}
		if !profileChanged(current.ProjectProfile, finalConfig.ProjectProfile) {
			printRescanReport(out, *report)
			return nil
		}
	}
	if err := s.store.Write(configPath, finalConfig); err != nil {
		return err
	}
	fmt.Fprintf(out, "Generado %s\n", configPath)
	fmt.Fprintf(out, "Stacks detectados: %s\n", stackSummary(finalConfig.Stacks))
	fmt.Fprintf(out, "Superficies detectadas: %s\n", surfaceSummary(finalConfig.ProjectProfile.Surfaces))
	if report != nil {
		printRescanReport(out, *report)
	}
	return nil
}

func profileChanged(current, next ProjectProfile) bool {
	return !reflect.DeepEqual(current, next)
}

func (s Service) Ensure(target string) (bool, error) {
	s = s.withDefaults()
	target, err := platform.ResolveTargetPath(target)
	if err != nil {
		return false, fmt.Errorf("resolver target: %w", err)
	}
	configPath := filepath.Join(target, ProjectConfigPath)
	if exists, err := s.store.Exists(configPath); err != nil {
		return false, err
	} else if exists {
		return false, nil
	}
	detected, err := s.scanner.Scan(target)
	if err != nil {
		return false, err
	}
	if err := s.store.Write(configPath, detected); err != nil {
		return false, err
	}
	return true, nil
}
