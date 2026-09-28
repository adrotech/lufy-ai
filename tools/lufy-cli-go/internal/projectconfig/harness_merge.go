package projectconfig

import (
	"fmt"
	"os"
	"reflect"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type HarnessMerge struct {
	Changed bool
	Created bool
	path    string
	before  []byte
	mode    os.FileMode
}

func (m HarnessMerge) Rollback() error {
	if !m.Changed {
		return nil
	}
	if m.Created {
		err := os.Remove(m.path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return platform.WriteFileAtomic(m.path, m.before, m.mode)
}

func (s Service) MergeHarnessSelection(targetValue string, harness domain.HarnessConfig) (HarnessMerge, error) {
	s = s.withDefaults()
	target, err := platform.ResolveTargetPath(targetValue)
	if err != nil {
		return HarnessMerge{}, err
	}
	harness = harness.WithDefaults()
	if err := harness.ValidateSupported(); err != nil {
		return HarnessMerge{}, err
	}
	if err := harness.MethodologyByTier.ValidateRoutingPolicy(domain.RoutingPolicyOptions{}); err != nil {
		return HarnessMerge{}, err
	}

	canonicalPath := Path(target)
	readPath, err := ExistingPath(target)
	if err != nil {
		return HarnessMerge{}, err
	}
	cfg, err := s.store.Load(readPath)
	if os.IsNotExist(err) {
		cfg, err = s.scanner.Scan(target)
	}
	if err != nil {
		return HarnessMerge{}, fmt.Errorf("preparar project config para harness: %w", err)
	}
	current := domain.HarnessConfig{Tool: cfg.Tool, MethodologyByTier: cfg.MethodologyByTier}.WithDefaults()
	canonicalBody, readErr := os.ReadFile(canonicalPath)
	canonicalExists := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return HarnessMerge{}, readErr
	}
	if canonicalExists && reflect.DeepEqual(current.Tool, harness.Tool) && reflect.DeepEqual(current.MethodologyByTier, harness.MethodologyByTier) {
		return HarnessMerge{path: canonicalPath}, nil
	}

	merge := HarnessMerge{Changed: true, Created: !canonicalExists, path: canonicalPath, before: canonicalBody, mode: 0o644}
	if canonicalExists {
		if info, statErr := os.Stat(canonicalPath); statErr == nil {
			merge.mode = info.Mode().Perm()
		}
	}
	cfg.Tool = harness.Tool
	cfg.MethodologyByTier = harness.MethodologyByTier
	if err := s.store.Write(canonicalPath, cfg); err != nil {
		return HarnessMerge{}, err
	}
	return merge, nil
}

func HarnessSelectionNeedsMerge(targetValue string, harness domain.HarnessConfig) (bool, error) {
	target, err := platform.ResolveTargetPath(targetValue)
	if err != nil {
		return false, err
	}
	canonical := Path(target)
	if _, err := os.Stat(canonical); os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}
	cfg, err := Load(canonical)
	if err != nil {
		return false, err
	}
	current := domain.HarnessConfig{Tool: cfg.Tool, MethodologyByTier: cfg.MethodologyByTier}.WithDefaults()
	expected := harness.WithDefaults()
	return current.Tool != expected.Tool || !reflect.DeepEqual(current.MethodologyByTier, expected.MethodologyByTier), nil
}
