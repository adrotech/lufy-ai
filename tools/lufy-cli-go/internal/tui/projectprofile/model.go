package projectprofile

import (
	"errors"
	"fmt"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	ErrCancelled       = errors.New("project_profile cancelado por el usuario")
	ErrNoActiveSurface = errors.New("project_profile requiere al menos una superficie activa")
)

type Model struct {
	cfg       projectconfig.ProjectConfig
	surfaces  []projectconfig.ProjectSurface
	active    []bool
	cursor    int
	confirmed bool
	cancelled bool
	keys      keyMap
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Prev    key.Binding
	Next    key.Binding
	Arch    key.Binding
	Toggle  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "bajar")),
		Prev:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "tipo anterior")),
		Next:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "tipo siguiente")),
		Arch:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "arquitectura")),
		Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("espacio", "activar/desactivar")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirmar")),
		Cancel:  key.NewBinding(key.WithKeys("esc", "ctrl+c", "q"), key.WithHelp("esc/q", "cancelar")),
	}
}

func NewModel(cfg projectconfig.ProjectConfig) Model {
	surfaces := copySurfaces(cfg)
	if len(surfaces) == 0 {
		surfaces = []projectconfig.ProjectSurface{{
			ID:           "main",
			Type:         "frontend",
			Roots:        []string{"."},
			Stacks:       stackIDs(cfg.Stacks),
			Frameworks:   stackFrameworks(cfg.Stacks),
			Architecture: projectconfig.DefaultArchitectureProfile("frontend"),
			AgentLens:    projectconfig.DefaultAgentLens("frontend"),
		}}
	}
	active := make([]bool, len(surfaces))
	for i := range active {
		active[i] = true
		if surfaces[i].Type == "" {
			surfaces[i] = applySurfaceType(surfaces[i], "frontend")
		}
		surfaces[i] = projectconfig.ApplySurfaceDefaults(surfaces[i])
	}
	return Model{cfg: cfg, surfaces: surfaces, active: active, keys: defaultKeyMap()}
}

func copySurfaces(cfg projectconfig.ProjectConfig) []projectconfig.ProjectSurface {
	surfaces := make([]projectconfig.ProjectSurface, len(cfg.ProjectProfile.Surfaces))
	copy(surfaces, cfg.ProjectProfile.Surfaces)
	for i := range surfaces {
		surfaces[i].Roots = append([]string{}, surfaces[i].Roots...)
		surfaces[i].Stacks = append([]string{}, surfaces[i].Stacks...)
		surfaces[i].Frameworks = append([]string{}, surfaces[i].Frameworks...)
		surfaces[i].Connects = append([]string{}, surfaces[i].Connects...)
		surfaces[i].Architecture.Detected = append([]string{}, surfaces[i].Architecture.Detected...)
		surfaces[i].Architecture.Options = append([]string{}, surfaces[i].Architecture.Options...)
		surfaces[i].AgentLens.PrimaryConcerns = append([]string{}, surfaces[i].AgentLens.PrimaryConcerns...)
		surfaces[i].AgentLens.ValidationExpectations = append([]string{}, surfaces[i].AgentLens.ValidationExpectations...)
	}
	return surfaces
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.keys.Cancel):
		m.cancelled = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Confirm):
		m.confirmed = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, m.keys.Down):
		if m.cursor < len(m.surfaces)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, m.keys.Prev):
		m.changeSelectedType(-1)
	case key.Matches(keyMsg, m.keys.Next):
		m.changeSelectedType(1)
	case key.Matches(keyMsg, m.keys.Arch):
		m.changeSelectedArchitecture(1)
	case key.Matches(keyMsg, m.keys.Toggle):
		if len(m.active) > 0 {
			m.active[m.cursor] = !m.active[m.cursor]
		}
	}
	return m, nil
}

func (m *Model) changeSelectedType(delta int) {
	if len(m.surfaces) == 0 {
		return
	}
	current := choiceIndex(m.surfaces[m.cursor].Type)
	next := (current + delta + len(SurfaceChoices)) % len(SurfaceChoices)
	m.surfaces[m.cursor] = applySurfaceType(m.surfaces[m.cursor], SurfaceChoices[next].Type)
	if len(m.surfaces[m.cursor].Stacks) == 0 {
		m.surfaces[m.cursor].Stacks = stackIDs(m.cfg.Stacks)
	}
	if len(m.surfaces[m.cursor].Frameworks) == 0 {
		m.surfaces[m.cursor].Frameworks = stackFrameworks(m.cfg.Stacks)
	}
}

func (m *Model) changeSelectedArchitecture(delta int) {
	if len(m.surfaces) == 0 {
		return
	}
	m.surfaces[m.cursor] = cycleArchitecture(m.surfaces[m.cursor], delta)
}

func (m Model) Result() (projectconfig.ProjectProfile, error) {
	if m.cancelled {
		return projectconfig.ProjectProfile{}, ErrCancelled
	}
	if !m.confirmed {
		return m.cfg.ProjectProfile, nil
	}
	profile := m.cfg.ProjectProfile
	profile.Surfaces = nil
	for i, surface := range m.surfaces {
		if i < len(m.active) && !m.active[i] {
			continue
		}
		profile.Surfaces = append(profile.Surfaces, surface)
	}
	if len(profile.Surfaces) == 0 {
		return projectconfig.ProjectProfile{}, ErrNoActiveSurface
	}
	return profile, nil
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Render("Project profile")
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Revisa las superficies que guiarán la mentalidad de los agentes.")
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n\n", title, subtitle)
	for i, surface := range m.surfaces {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		status := "activa"
		if i < len(m.active) && !m.active[i] {
			status = "omitida"
		}
		line := fmt.Sprintf("%s [%s] %s · %s · roots=%s · stacks=%s", cursor, status, surface.ID, surface.Type, strings.Join(surface.Roots, ","), strings.Join(surface.Stacks, ","))
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(line)
		}
		fmt.Fprintln(&b, line)
	}
	if len(m.surfaces) > 0 {
		selected := m.surfaces[m.cursor]
		if selected.Architecture.Preferred != "" {
			fmt.Fprintf(&b, "\nArquitectura: preferred=%s detected=%s options=%s\n", selected.Architecture.Preferred, strings.Join(selected.Architecture.Detected, ","), strings.Join(selected.Architecture.Options, ","))
		}
		fmt.Fprintf(&b, "\nLens: %s\n", strings.Join(selected.AgentLens.PrimaryConcerns, ", "))
		fmt.Fprintf(&b, "Validación: %s\n", strings.Join(selected.AgentLens.ValidationExpectations, ", "))
	}
	fmt.Fprintf(&b, "\n%s · %s · %s · %s · %s\n", m.keys.Up.Help().Key, m.keys.Next.Help().Key, m.keys.Arch.Help().Key, m.keys.Toggle.Help().Key, m.keys.Confirm.Help().Key)
	fmt.Fprintln(&b, "esc/q cancela sin escribir cambios")
	return b.String()
}
