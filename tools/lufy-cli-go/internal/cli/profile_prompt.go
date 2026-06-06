package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

type surfaceChoice struct {
	Type        string
	Label       string
	Description string
}

var surfaceChoices = []surfaceChoice{
	{Type: "frontend", Label: "Frontend web", Description: "UI, accesibilidad, responsive, estados y consumo de APIs"},
	{Type: "backend", Label: "Backend/API", Description: "Contratos, dominio, persistencia, auth, idempotencia y observabilidad"},
	{Type: "fullstack", Label: "Fullstack", Description: "Contrato front/back, errores entre capas, E2E y rollout"},
	{Type: "mobile", Label: "Mobile", Description: "Navegación, estados offline/red, dispositivos y release channels"},
	{Type: "cli", Label: "CLI/tooling", Description: "Flags, exit codes, filesystem safety, idempotencia y scriptability"},
	{Type: "infra", Label: "Infra/DevOps", Description: "Drift, secrets, permisos, rollback, entornos y supply chain"},
	{Type: "library", Label: "Library/SDK", Description: "Contratos públicos, compatibilidad y uso por consumidores"},
}

func surfaceProfilePrompt(deps Dependencies) projectconfig.ProfilePrompt {
	return func(cfg projectconfig.ProjectConfig) (projectconfig.ProjectProfile, error) {
		input := deps.Stdin
		if input == nil {
			input = os.Stdin
		}
		output := deps.Stdout
		if output == nil {
			output = os.Stdout
		}
		if !interactiveIO(input, output) {
			fmt.Fprintln(output, "project_profile: modo no interactivo; se conserva la detección automática.")
			return cfg.ProjectProfile, nil
		}
		surfaceType, err := promptPrimarySurface(input, output, cfg)
		if err != nil {
			return projectconfig.ProjectProfile{}, err
		}
		if surfaceType == "" {
			return cfg.ProjectProfile, nil
		}
		return applyPrimarySurfaceType(cfg, surfaceType), nil
	}
}

func interactiveIO(input io.Reader, output io.Writer) bool {
	inFile, inOK := input.(*os.File)
	outFile, outOK := output.(*os.File)
	if !inOK || !outOK {
		return false
	}
	inInfo, inErr := inFile.Stat()
	outInfo, outErr := outFile.Stat()
	if inErr != nil || outErr != nil {
		return false
	}
	return inInfo.Mode()&os.ModeCharDevice != 0 && outInfo.Mode()&os.ModeCharDevice != 0
}

func promptPrimarySurface(input io.Reader, output io.Writer, cfg projectconfig.ProjectConfig) (string, error) {
	defaultIndex := defaultSurfaceIndex(cfg)
	fmt.Fprintln(output, "Selecciona la mentalidad principal del proyecto:")
	for i, choice := range surfaceChoices {
		marker := " "
		if i == defaultIndex {
			marker = "*"
		}
		fmt.Fprintf(output, "%s %d. %s - %s\n", marker, i+1, choice.Label, choice.Description)
	}
	fmt.Fprintf(output, "Opción [%d]: ", defaultIndex+1)
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", fmt.Errorf("leer selección project_profile: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return surfaceChoices[defaultIndex].Type, nil
	}
	index, err := strconv.Atoi(line)
	if err != nil || index < 1 || index > len(surfaceChoices) {
		return "", fmt.Errorf("selección project_profile inválida: %q", line)
	}
	return surfaceChoices[index-1].Type, nil
}

func defaultSurfaceIndex(cfg projectconfig.ProjectConfig) int {
	if len(cfg.ProjectProfile.Surfaces) > 0 {
		for i, choice := range surfaceChoices {
			if choice.Type == cfg.ProjectProfile.Surfaces[0].Type {
				return i
			}
		}
	}
	return 0
}

func applyPrimarySurfaceType(cfg projectconfig.ProjectConfig, surfaceType string) projectconfig.ProjectProfile {
	profile := cfg.ProjectProfile
	if len(profile.Surfaces) == 0 {
		profile.Surfaces = []projectconfig.ProjectSurface{{
			ID:         "main",
			Type:       surfaceType,
			Roots:      []string{"."},
			Stacks:     stackIDs(cfg.Stacks),
			Frameworks: stackFrameworks(cfg.Stacks),
			AgentLens:  projectconfig.DefaultAgentLens(surfaceType),
		}}
		return profile
	}
	profile.Surfaces[0].Type = surfaceType
	profile.Surfaces[0].AgentLens = projectconfig.DefaultAgentLens(surfaceType)
	if len(profile.Surfaces[0].Stacks) == 0 {
		profile.Surfaces[0].Stacks = stackIDs(cfg.Stacks)
	}
	if len(profile.Surfaces[0].Frameworks) == 0 {
		profile.Surfaces[0].Frameworks = stackFrameworks(cfg.Stacks)
	}
	return profile
}

func stackIDs(stacks []projectconfig.Stack) []string {
	var out []string
	for _, stack := range stacks {
		out = append(out, stack.ID)
	}
	return out
}

func stackFrameworks(stacks []projectconfig.Stack) []string {
	seen := map[string]bool{}
	var out []string
	for _, stack := range stacks {
		for _, framework := range stack.Frameworks {
			if !seen[framework] {
				seen[framework] = true
				out = append(out, framework)
			}
		}
	}
	return out
}
