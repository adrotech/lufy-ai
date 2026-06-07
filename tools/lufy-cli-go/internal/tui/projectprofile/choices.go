package projectprofile

import "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"

type SurfaceChoice struct {
	Type        string
	Label       string
	Description string
}

var SurfaceChoices = []SurfaceChoice{
	{Type: "frontend", Label: "Frontend web", Description: "UI, accesibilidad, responsive, estados y consumo de APIs"},
	{Type: "backend", Label: "Backend/API", Description: "Contratos, dominio, persistencia, auth, idempotencia y observabilidad"},
	{Type: "fullstack", Label: "Fullstack", Description: "Contrato front/back, errores entre capas, E2E y rollout"},
	{Type: "mobile", Label: "Mobile", Description: "Navegación, estados offline/red, dispositivos y release channels"},
	{Type: "cli", Label: "CLI/tooling", Description: "Flags, exit codes, filesystem safety, idempotencia y scriptability"},
	{Type: "infra", Label: "Infra/DevOps", Description: "Drift, secrets, permisos, rollback, entornos y supply chain"},
	{Type: "library", Label: "Library/SDK", Description: "Contratos públicos, compatibilidad y uso por consumidores"},
}

func choiceIndex(surfaceType string) int {
	for i, choice := range SurfaceChoices {
		if choice.Type == surfaceType {
			return i
		}
	}
	return 0
}

func applySurfaceType(surface projectconfig.ProjectSurface, surfaceType string) projectconfig.ProjectSurface {
	surface.Type = surfaceType
	surface.AgentLens = projectconfig.DefaultAgentLens(surfaceType)
	return surface
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
			if seen[framework] {
				continue
			}
			seen[framework] = true
			out = append(out, framework)
		}
	}
	return out
}
