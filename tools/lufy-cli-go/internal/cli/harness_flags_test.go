package cli

import (
	"flag"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"io"
	"testing"
)

func TestInlineHarnessFlags(t *testing.T) {
	for _, d := range []string{"auto", "on_request", "off"} {
		f := flag.NewFlagSet("test", flag.ContinueOnError)
		f.SetOutput(io.Discard)
		v := addHarnessFlags(f)
		if err := f.Parse([]string{"--tool", "codex", "--delegation", d}); err != nil {
			t.Fatal(err)
		}
		c, err := parseHarnessFlags(v)
		if err != nil {
			t.Fatal(err)
		}
		if c.Execution.Delegation != domain.DelegationPolicy(d) || !c.ExecutionIsExplicit() || !c.ToolIsExplicit() {
			t.Fatalf("flags: %#v", c)
		}
	}
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	v := addHarnessFlags(f)
	c, err := parseHarnessFlags(v)
	if err != nil || c.ExecutionIsExplicit() || c.ToolIsExplicit() {
		t.Fatalf("defaults: %#v %v", c, err)
	}
}
func TestRejectRetiredAndInvalidFlags(t *testing.T) {
	for _, args := range [][]string{{"--methodology-tier", "T1:openspec/full"}, {"--tool", "opencode"}, {"--delegation", ""}, {"--delegation", "always"}} {
		f := flag.NewFlagSet("test", flag.ContinueOnError)
		f.SetOutput(io.Discard)
		v := addHarnessFlags(f)
		err := f.Parse(args)
		if err == nil {
			_, err = parseHarnessFlags(v)
		}
		if err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if writableToolList() != "codex" || writableToolUsage() != "codex" || writableToolSupported(domain.ToolID("opencode")) {
		t.Fatal("retired writable tool")
	}
}
