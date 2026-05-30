package cli

import (
	"flag"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestParseHarnessFlagsDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := addHarnessFlags(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseHarnessFlags(flags)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tool != domain.ToolInitialDefault {
		t.Fatalf("tool = %s", cfg.Tool)
	}
	if cfg.MethodologyByTier[domain.TierT3].ID != domain.MethodologyNone {
		t.Fatalf("T3 methodology = %#v", cfg.MethodologyByTier[domain.TierT3])
	}
}

func TestParseHarnessFlagsComposesRepeatedMethodologyOverrides(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := addHarnessFlags(fs)
	if err := fs.Parse([]string{"--methodology-tier", "T3:openspec/full", "--methodology-tier", "T3:none"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseHarnessFlags(flags)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.MethodologyByTier[domain.TierT3]; got.ID != domain.MethodologyNone || got.Mode != domain.MethodologyModeNone || got.Required {
		t.Fatalf("T3 override = %#v", got)
	}
}

func TestParseHarnessFlagsRejectsUnsupportedTool(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := addHarnessFlags(fs)
	if err := fs.Parse([]string{"--tool", "codex"}); err != nil {
		t.Fatal(err)
	}

	if _, err := parseHarnessFlags(flags); err == nil {
		t.Fatalf("expected unsupported tool error")
	}
}

func TestParseHarnessFlagsRejectsEmptyTool(t *testing.T) {
	empty := ""
	if _, err := parseHarnessFlags(harnessFlagValues{Tool: &empty}); err == nil {
		t.Fatalf("expected empty tool error")
	}
}

func TestParseHarnessFlagsRejectsUnsafeNone(t *testing.T) {
	tests := []string{"T1:none", "T2:none"}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			flags := addHarnessFlags(fs)
			if err := fs.Parse([]string{"--methodology-tier", raw}); err != nil {
				t.Fatal(err)
			}
			if _, err := parseHarnessFlags(flags); err == nil {
				t.Fatalf("expected unsafe none error for %s", raw)
			}
		})
	}
}

func TestParseHarnessFlagsRejectsReservedLufySDD(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := addHarnessFlags(fs)
	if err := fs.Parse([]string{"--methodology-tier", "T3:lufy-sdd/lite"}); err != nil {
		t.Fatal(err)
	}

	if _, err := parseHarnessFlags(flags); err == nil {
		t.Fatalf("expected reserved lufy-sdd error")
	}
}

func TestMethodologyTierFlagRejectsEmptyValue(t *testing.T) {
	var values methodologyTierFlags
	if err := values.Set(""); err == nil {
		t.Fatalf("expected empty value error")
	}
}

func TestParseMethodologyTierValidationErrors(t *testing.T) {
	tests := []string{
		"",
		"T3",
		"T4:none",
		"T3:",
		"T3:openspec/",
		"T3:openspec/none",
		"T3:none/full",
	}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, _, err := parseMethodologyTier(raw); err == nil {
				t.Fatalf("expected parse error for %q", raw)
			}
		})
	}
}

func TestParseMethodologyTierInfersOpenSpecModes(t *testing.T) {
	tests := []struct {
		raw  string
		mode domain.MethodologyMode
	}{
		{raw: "T1:openspec", mode: domain.MethodologyModeFull},
		{raw: "T2:openspec", mode: domain.MethodologyModeLite},
		{raw: "T3:openspec", mode: domain.MethodologyModeLite},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			_, selection, err := parseMethodologyTier(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			if selection.ID != domain.MethodologySpecWorkflow || selection.Mode != tt.mode || !selection.Required {
				t.Fatalf("selection = %#v", selection)
			}
		})
	}
}

func TestParseMethodologyTierAcceptsExplicitNoneMode(t *testing.T) {
	tier, selection, err := parseMethodologyTier("T3:none/none")
	if err != nil {
		t.Fatal(err)
	}
	if tier != domain.TierT3 || selection.ID != domain.MethodologyNone || selection.Mode != domain.MethodologyModeNone || selection.Required {
		t.Fatalf("selection = %s %#v", tier, selection)
	}
}
