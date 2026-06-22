package commandpalette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type stubModel struct{}

func (stubModel) Init() tea.Cmd                       { return nil }
func (stubModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return stubModel{}, nil }
func (stubModel) View() string                        { return "" }

func TestModelSelectsCommandAndBuildsArgs(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "setup", Title: "Setup", Description: "Setup", Args: []string{"setup"}, Params: []ParamSpec{
		{Name: "target", Flag: "--target", Kind: ParamText, Default: "."},
		{Name: "dry-run", Flag: "--dry-run", Kind: ParamBool},
	}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != screenParams || len(m.params) != 2 {
		t.Fatalf("expected params screen, got %#v", m)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != screenText {
		t.Fatalf("expected text screen, got %v", m.screen)
	}
	for _, r := range "/tmp/app" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.done {
		t.Fatalf("expected done model")
	}
	args := BuildArgs(m.commands[m.cmdIndex], m.params)
	want := "setup --target /tmp/app --dry-run"
	if strings.Join(args, " ") != want {
		t.Fatalf("args = %q want %q", strings.Join(args, " "), want)
	}
}

func TestModelBlocksMissingRequired(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "upgrade", Title: "Upgrade", Description: "Upgrade", Args: []string{"upgrade"}, Params: []ParamSpec{{Name: "to", Flag: "--to", Kind: ParamText, Required: true}}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.done || !strings.Contains(m.errMsg, "to") {
		t.Fatalf("expected required error, done=%t err=%q", m.done, m.errMsg)
	}
}

func TestModelCyclesChoiceAndViews(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "install", Title: "Install", Description: "Install", Args: []string{"install"}, Params: []ParamSpec{
		{Name: "scope", Flag: "--scope", Kind: ParamChoice, Default: "project", Choices: []string{"project", "global", "both"}},
	}}}
	if !strings.Contains(m.View(), "Install") {
		t.Fatalf("commands view missing command: %s", m.View())
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(model)
	if m.params[0].Value != "global" {
		t.Fatalf("choice not cycled: %#v", m.params[0])
	}
	if !strings.Contains(m.View(), "global") {
		t.Fatalf("params view missing choice: %s", m.View())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	if m.params[0].Value != "project" {
		t.Fatalf("choice not cycled back: %#v", m.params[0])
	}
}

func TestModelCancelAndBackspace(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "query", Title: "Query", Description: "Query", Args: []string{"context", "query"}, Params: []ParamSpec{{Name: "term", Kind: ParamArg}}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abc")})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(model)
	if m.input != "ab" {
		t.Fatalf("backspace failed, input=%q", m.input)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.screen != screenParams {
		t.Fatalf("esc should return to params")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if !m.cancel {
		t.Fatalf("expected cancel")
	}
}

func TestDisplayValueAndClamp(t *testing.T) {
	boolParam := ParamValue{Spec: ParamSpec{Name: "yes", Kind: ParamBool}, Value: "true"}
	if displayValue(boolParam) != "si" {
		t.Fatalf("bool display true mismatch")
	}
	boolParam.Value = ""
	if displayValue(boolParam) != "no" {
		t.Fatalf("bool display false mismatch")
	}
	required := ParamValue{Spec: ParamSpec{Name: "to", Kind: ParamText, Required: true}}
	if displayValue(required) != "<requerido>" {
		t.Fatalf("required display mismatch")
	}
	optional := ParamValue{Spec: ParamSpec{Name: "base", Kind: ParamText}}
	if displayValue(optional) != "<vacio>" {
		t.Fatalf("optional display mismatch")
	}
	if clamp(-1, 0, 2) != 0 || clamp(3, 0, 2) != 2 || clamp(1, 0, 2) != 1 {
		t.Fatalf("clamp mismatch")
	}
}

func TestBackReturnsToCommandScreen(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "version", Title: "Version", Description: "Version", Args: []string{"version"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != screenParams {
		t.Fatalf("expected params")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(model)
	if m.screen != screenCommands {
		t.Fatalf("expected commands after back")
	}
}

func TestTextViewAndSpaceEditing(t *testing.T) {
	m := newModel()
	m.commands = []CommandSpec{{ID: "search", Title: "Search", Description: "Search", Args: []string{"memory", "search"}, Params: []ParamSpec{{Name: "query", Kind: ParamArg, Description: "Query"}}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("world")})
	m = updated.(model)
	if !strings.Contains(m.viewText(), "hello world") {
		t.Fatalf("text view missing input: %s", m.viewText())
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.params[0].Value != "hello world" {
		t.Fatalf("text value mismatch: %#v", m.params[0])
	}
}

func TestResultFromModel(t *testing.T) {
	cancelled, err := resultFromModel(stubModel{})
	if err != nil || !cancelled.Cancelled {
		t.Fatalf("expected cancelled non-model, got %#v err=%v", cancelled, err)
	}
	m := newModel()
	m.commands = []CommandSpec{{ID: "version", Title: "Version", Args: []string{"version"}}}
	m.done = true
	result, err := resultFromModel(m)
	if err != nil || result.Cancelled || strings.Join(result.Args, " ") != "version" {
		t.Fatalf("unexpected result %#v err=%v", result, err)
	}
	m.done = false
	result, err = resultFromModel(m)
	if err != nil || !result.Cancelled {
		t.Fatalf("expected cancelled unfinished, got %#v err=%v", result, err)
	}
}

func TestModelIgnoresNonKeyAndEmptyParamMove(t *testing.T) {
	m := newModel()
	updated, _ := m.Update("not-key")
	m = updated.(model)
	if m.cmdIndex != 0 {
		t.Fatalf("non key changed model")
	}
	m.screen = screenParams
	m.params = nil
	m.move(1)
	if m.paramIdx != 0 {
		t.Fatalf("empty param move changed index")
	}
}
