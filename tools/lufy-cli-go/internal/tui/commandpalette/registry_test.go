package commandpalette

import (
	"reflect"
	"testing"
)

func TestBuildArgsIncludesOnlySetValues(t *testing.T) {
	spec := CommandSpec{Args: []string{"setup"}, Params: []ParamSpec{
		{Name: "target", Flag: "--target", Kind: ParamText, Default: "."},
		{Name: "dry-run", Flag: "--dry-run", Kind: ParamBool},
		{Name: "scope", Flag: "--scope", Kind: ParamChoice, Default: "project", Choices: []string{"project", "global"}},
	}}
	values := []ParamValue{
		{Spec: spec.Params[0], Value: "/tmp/app"},
		{Spec: spec.Params[1], Value: "true"},
		{Spec: spec.Params[2], Value: "project"},
	}
	got := BuildArgs(spec, values)
	want := []string{"setup", "--target", "/tmp/app", "--dry-run"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %#v want %#v", got, want)
	}
}

func TestBuildArgsUsesEqualsForBoolChoices(t *testing.T) {
	spec := CommandSpec{Args: []string{"init"}, Params: []ParamSpec{{Name: "interactive", Flag: "--interactive", Kind: ParamChoice, Default: "true", Choices: []string{"true", "false"}}}}
	values := []ParamValue{{Spec: spec.Params[0], Value: "false"}}
	got := BuildArgs(spec, values)
	want := []string{"init", "--interactive=false"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArgs() = %#v want %#v", got, want)
	}
}

func TestMissingRequired(t *testing.T) {
	values := []ParamValue{{Spec: ParamSpec{Name: "to", Required: true, Kind: ParamText}}, {Spec: ParamSpec{Name: "target", Required: false, Kind: ParamText}}}
	got := MissingRequired(values)
	want := []string{"to"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MissingRequired() = %#v want %#v", got, want)
	}
}

func TestRegistryContainsSetupAndUpgrade(t *testing.T) {
	seen := map[string]bool{}
	for _, spec := range Registry() {
		seen[spec.ID] = true
	}
	for _, id := range []string{"setup", "upgrade", "context-build", "memory-search", "skills-refresh", "skills-status", "pr-guard"} {
		if !seen[id] {
			t.Fatalf("registry missing %s", id)
		}
	}
}

func TestSkillsCommandsOnlyOfferWritableSkillAdapters(t *testing.T) {
	for _, spec := range Registry() {
		if spec.ID != "skills-refresh" && spec.ID != "skills-status" {
			continue
		}
		for _, param := range spec.Params {
			if param.Name == "tool" && !reflect.DeepEqual(param.Choices, []string{"opencode", "codex"}) {
				t.Fatalf("%s tool choices = %v", spec.ID, param.Choices)
			}
		}
	}
}

func TestSDDNewOffersFullAndLiteModes(t *testing.T) {
	for _, spec := range Registry() {
		if spec.ID != "sdd-new" {
			continue
		}
		for _, param := range spec.Params {
			if param.Name == "mode" {
				if param.Default != "full" || !reflect.DeepEqual(param.Choices, []string{"full", "lite"}) {
					t.Fatalf("sdd-new mode = %#v", param)
				}
				return
			}
		}
		t.Fatal("sdd-new missing mode parameter")
	}
	t.Fatal("registry missing sdd-new")
}
