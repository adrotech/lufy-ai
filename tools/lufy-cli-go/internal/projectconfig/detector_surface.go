package projectconfig

import (
	"fmt"
	"sort"
	"strings"
)

func detectProjectProfile(root string, stacks []Stack) ProjectProfile {
	surfaces := detectStackSurfaces(root, stacks)
	if hasInfraEvidence(root) {
		surfaces = append(surfaces, ProjectSurface{ID: "infra", Type: "infra", Roots: existingRoots(root, []string{"infra", "terraform", "deployments", "k8s", "."}), Stacks: []string{}, Frameworks: []string{}, AgentLens: DefaultAgentLens("infra")})
	}
	surfaces = compactSurfaces(surfaces)
	if hasSurfaceType(surfaces, "frontend") && hasSurfaceType(surfaces, "backend") {
		surfaces = append(surfaces, ProjectSurface{ID: "fullstack-flow", Type: "fullstack", Roots: []string{"."}, Stacks: surfaceStackIDs(surfaces), Frameworks: surfaceFrameworks(surfaces), Connects: surfaceIDs(surfaces), AgentLens: DefaultAgentLens("fullstack")})
	}
	sort.SliceStable(surfaces, func(i, j int) bool { return surfaces[i].ID < surfaces[j].ID })
	return ProjectProfile{Surfaces: surfaces}
}

func detectJSSurface(root string, stack Stack) ProjectSurface {
	pkg := readPackageJSON(root)
	if hasDep(pkg, "expo") || hasDep(pkg, "react-native") {
		return ProjectSurface{ID: "mobile-app", Type: "mobile", Roots: existingRoots(root, []string{"app", "src", "mobile", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("mobile")}
	}
	if hasAny(stack.Frameworks, []string{"react", "next", "remix", "vue", "svelte"}) {
		return ProjectSurface{ID: "web-app", Type: "frontend", Roots: existingRoots(root, []string{"app", "src", "components", "pages", "frontend", "web", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("frontend")}
	}
	return ProjectSurface{ID: stack.ID + "-package", Type: "library", Roots: existingRoots(root, []string{"src", "packages", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("library")}
}

func detectGoSurface(root string, stack Stack) ProjectSurface {
	if existsAny(root, "api", "internal/api", "internal/server") {
		return ProjectSurface{ID: "api", Type: "backend", Roots: existingRoots(root, []string{"api", "internal", "cmd", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("backend")}
	}
	if exists(root, "cmd") {
		return ProjectSurface{ID: "cli", Type: "cli", Roots: existingRoots(root, []string{"cmd", "internal", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("cli")}
	}
	return ProjectSurface{ID: "go-library", Type: "library", Roots: existingRoots(root, []string{"internal", "pkg", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("library")}
}

func detectServerSurface(fallbackID, root string, stack Stack) ProjectSurface {
	if hasAny(stack.Frameworks, []string{"fastapi", "django", "flask", "spring-boot"}) || existsAny(root, "api", "server", "src/main") {
		return ProjectSurface{ID: "api", Type: "backend", Roots: existingRoots(root, []string{"api", "server", "src", "src/main", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("backend")}
	}
	return ProjectSurface{ID: fallbackID, Type: "library", Roots: existingRoots(root, []string{"src", "."}), Stacks: []string{stack.ID}, Frameworks: stack.Frameworks, AgentLens: DefaultAgentLens("library")}
}

func hasSurfaceType(surfaces []ProjectSurface, surfaceType string) bool {
	for _, surface := range surfaces {
		if surface.Type == surfaceType {
			return true
		}
	}
	return false
}

func hasInfraEvidence(root string) bool {
	return existsGlob(root, "*.tf") || existsAny(root, "terraform", "infra", "k8s", "deployments") || exists(root, "Dockerfile")
}

func existingRoots(root string, candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		if candidate == "." || exists(root, candidate) {
			out = append(out, candidate)
		}
	}
	if len(out) == 0 {
		return []string{"."}
	}
	return unique(out)
}

func compactSurfaces(surfaces []ProjectSurface) []ProjectSurface {
	byID := map[string]ProjectSurface{}
	for _, surface := range surfaces {
		if surface.ID == "" {
			continue
		}
		existing, ok := byID[surface.ID]
		if !ok {
			surface.Roots = unique(surface.Roots)
			surface.Stacks = unique(surface.Stacks)
			surface.Frameworks = unique(surface.Frameworks)
			byID[surface.ID] = surface
			continue
		}
		existing.Roots = unique(append(existing.Roots, surface.Roots...))
		existing.Stacks = unique(append(existing.Stacks, surface.Stacks...))
		existing.Frameworks = unique(append(existing.Frameworks, surface.Frameworks...))
		byID[surface.ID] = existing
	}
	out := make([]ProjectSurface, 0, len(byID))
	for _, surface := range byID {
		out = append(out, surface)
	}
	return out
}

func surfaceIDs(surfaces []ProjectSurface) []string {
	var out []string
	for _, surface := range surfaces {
		if surface.Type == "fullstack" {
			continue
		}
		out = append(out, surface.ID)
	}
	return unique(out)
}

func surfaceStackIDs(surfaces []ProjectSurface) []string {
	var out []string
	for _, surface := range surfaces {
		out = append(out, surface.Stacks...)
	}
	return unique(out)
}

func surfaceFrameworks(surfaces []ProjectSurface) []string {
	var out []string
	for _, surface := range surfaces {
		out = append(out, surface.Frameworks...)
	}
	return unique(out)
}

func surfaceSummary(surfaces []ProjectSurface) string {
	if len(surfaces) == 0 {
		return "ninguna"
	}
	parts := make([]string, 0, len(surfaces))
	for _, surface := range surfaces {
		parts = append(parts, fmt.Sprintf("%s:%s", surface.ID, surface.Type))
	}
	return strings.Join(parts, ", ")
}
