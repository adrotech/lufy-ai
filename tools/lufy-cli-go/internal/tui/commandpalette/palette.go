package commandpalette

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Input  io.Reader
	Output io.Writer
}

type Result struct {
	Args      []string
	Command   string
	Cancelled bool
}

type screen int

const (
	screenCommands screen = iota
	screenParams
	screenText
)

type model struct {
	commands []CommandSpec
	cmdIndex int
	screen   screen
	params   []ParamValue
	paramIdx int
	input    string
	done     bool
	cancel   bool
	errMsg   string
	keys     keyMap
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Back   key.Binding
	Toggle key.Binding
	Cancel key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "bajar")),
		Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "anterior")),
		Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "siguiente")),
		Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "seleccionar/ejecutar")),
		Back:   key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "volver")),
		Toggle: key.NewBinding(key.WithKeys(" "), key.WithHelp("espacio", "activar/editar")),
		Cancel: key.NewBinding(key.WithKeys("esc", "ctrl+c", "q"), key.WithHelp("esc/q", "cancelar")),
	}
}

func Run(opts Options) (Result, error) {
	input := opts.Input
	if input == nil {
		input = os.Stdin
	}
	output := opts.Output
	if output == nil {
		output = os.Stdout
	}
	m := newModel()
	program := tea.NewProgram(m, tea.WithInput(input), tea.WithOutput(output))
	finalModel, err := program.Run()
	if err != nil {
		return Result{}, err
	}
	return resultFromModel(finalModel)
}

func resultFromModel(finalModel tea.Model) (Result, error) {
	result, ok := finalModel.(model)
	if !ok || result.cancel {
		return Result{Cancelled: true}, nil
	}
	if !result.done {
		return Result{Cancelled: true}, nil
	}
	spec := result.commands[result.cmdIndex]
	return Result{Args: BuildArgs(spec, result.params), Command: spec.Title}, nil
}

func newModel() model {
	return model{commands: Registry(), keys: defaultKeys()}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.screen == screenText {
		return m.updateText(keyMsg)
	}
	switch {
	case key.Matches(keyMsg, m.keys.Cancel):
		m.cancel = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Back):
		if m.screen == screenParams {
			m.screen = screenCommands
			m.paramIdx = 0
		}
	case key.Matches(keyMsg, m.keys.Up):
		m.move(-1)
	case key.Matches(keyMsg, m.keys.Down):
		m.move(1)
	case key.Matches(keyMsg, m.keys.Left):
		m.cycleChoice(-1)
	case key.Matches(keyMsg, m.keys.Right):
		m.cycleChoice(1)
	case key.Matches(keyMsg, m.keys.Toggle):
		m.toggleOrEdit()
	case key.Matches(keyMsg, m.keys.Enter):
		return m.enter()
	}
	return m, nil
}

func (m model) updateText(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(keyMsg, m.keys.Cancel):
		m.screen = screenParams
		return m, nil
	case key.Matches(keyMsg, m.keys.Enter):
		m.params[m.paramIdx].Value = strings.TrimSpace(m.input)
		m.screen = screenParams
		return m, nil
	}
	switch keyMsg.Type {
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case tea.KeySpace:
		m.input += " "
	case tea.KeyRunes:
		m.input += keyMsg.String()
	}
	return m, nil
}

func (m *model) move(delta int) {
	if m.screen == screenCommands {
		m.cmdIndex = clamp(m.cmdIndex+delta, 0, len(m.commands)-1)
		return
	}
	max := len(m.params)
	if max == 0 {
		return
	}
	// Extra row is the Execute row.
	m.paramIdx = clamp(m.paramIdx+delta, 0, max)
}

func (m *model) enter() (tea.Model, tea.Cmd) {
	if m.screen == screenCommands {
		m.params = InitialValues(m.commands[m.cmdIndex])
		m.paramIdx = 0
		m.screen = screenParams
		return *m, nil
	}
	if m.paramIdx >= len(m.params) {
		if missing := MissingRequired(m.params); len(missing) > 0 {
			m.errMsg = "Faltan parametros requeridos: " + strings.Join(missing, ", ")
			return *m, nil
		}
		m.done = true
		return *m, tea.Quit
	}
	m.errMsg = ""
	m.toggleOrEdit()
	return *m, nil
}

func (m *model) toggleOrEdit() {
	if m.screen != screenParams || m.paramIdx >= len(m.params) {
		return
	}
	param := m.params[m.paramIdx]
	switch param.Spec.Kind {
	case ParamBool:
		if param.Value == "true" {
			m.params[m.paramIdx].Value = ""
		} else {
			m.params[m.paramIdx].Value = "true"
		}
	case ParamChoice:
		m.cycleChoice(1)
	case ParamText, ParamArg:
		m.input = editableInitialValue(param)
		m.screen = screenText
	}
}

func (m *model) cycleChoice(delta int) {
	if m.screen != screenParams || m.paramIdx >= len(m.params) {
		return
	}
	param := m.params[m.paramIdx]
	if param.Spec.Kind != ParamChoice || len(param.Spec.Choices) == 0 {
		return
	}
	idx := 0
	for i, choice := range param.Spec.Choices {
		if choice == param.Value {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(param.Spec.Choices)) % len(param.Spec.Choices)
	m.params[m.paramIdx].Value = param.Spec.Choices[idx]
}

func (m model) View() string {
	if m.screen == screenText {
		return m.viewText()
	}
	if m.screen == screenParams {
		return m.viewParams()
	}
	return m.viewCommands()
}

func (m model) viewCommands() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Render("Lufy Command Palette")
	fmt.Fprintf(&b, "%s\nSelecciona un comando:\n\n", title)
	for i, command := range m.commands {
		cursor := " "
		line := fmt.Sprintf("%s %s - %s", cursor, command.Title, command.Description)
		if i == m.cmdIndex {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("> " + command.Title + " - " + command.Description)
		}
		fmt.Fprintln(&b, line)
	}
	fmt.Fprintf(&b, "\n%s · %s · %s\n", m.keys.Up.Help().Key, m.keys.Enter.Help().Key, m.keys.Cancel.Help().Key)
	return b.String()
}

func (m model) viewParams() string {
	var b strings.Builder
	command := m.commands[m.cmdIndex]
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Render(command.Title)
	fmt.Fprintf(&b, "%s\n%s\n\n", title, command.Description)
	for i, param := range m.params {
		cursor := " "
		if i == m.paramIdx {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s = %s", cursor, param.Spec.Name, displayValue(param))
		if param.Spec.Description != "" {
			line += " - " + param.Spec.Description
		}
		if i == m.paramIdx {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(line)
		}
		fmt.Fprintln(&b, line)
	}
	execCursor := " "
	if m.paramIdx >= len(m.params) {
		execCursor = ">"
	}
	line := fmt.Sprintf("%s Ejecutar: lufy-ai %s", execCursor, strings.Join(BuildArgs(command, m.params), " "))
	if m.paramIdx >= len(m.params) {
		line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render(line)
	}
	fmt.Fprintln(&b, line)
	if m.errMsg != "" {
		fmt.Fprintln(&b, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMsg))
	}
	fmt.Fprintf(&b, "\nenter/espacio edita · ←/→ choices · backspace vuelve · esc cancela\n")
	return b.String()
}

func (m model) viewText() string {
	param := m.params[m.paramIdx]
	return fmt.Sprintf("%s\n%s\n\n> %s\n\nenter guarda · esc cancela edicion · backspace borra\n", param.Spec.Name, param.Spec.Description, m.input)
}

func displayValue(param ParamValue) string {
	switch param.Spec.Kind {
	case ParamBool:
		if param.Value == "true" {
			return "si"
		}
		return "no"
	case ParamText, ParamArg, ParamChoice:
		if param.Value == "" {
			if param.Spec.Required {
				return "<requerido>"
			}
			return "<vacio>"
		}
		return param.Value
	default:
		return param.Value
	}
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func editableInitialValue(param ParamValue) string {
	if param.Value == param.Spec.Default {
		return ""
	}
	return param.Value
}
