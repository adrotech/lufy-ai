package surfaceplan

import (
	"fmt"
	"sort"
	"strings"
)

type Resolution struct {
	Mode           ExecutionMode
	PrimarySurface string
	Surfaces       []SurfaceDefinition
	Contracts      []ContractEdge
	Decisions      []Decision
}

type SurfaceResolverStrategy struct{}

func (SurfaceResolverStrategy) Resolve(project ProjectDefinition, requested string, files []string) (Resolution, error) {
	if len(project.Surfaces) == 0 {
		return Resolution{}, fmt.Errorf("project_profile no declara superficies; ejecuta lufy-ai scan")
	}
	requested = strings.TrimSpace(requested)
	if requested != "" && requested != "auto" {
		return resolveExplicit(project.Surfaces, requested)
	}

	leaves := leafSurfaces(project.Surfaces)
	if len(leaves) == 0 {
		return Resolution{}, fmt.Errorf("project_profile no declara superficies ejecutables")
	}
	if len(files) == 0 {
		if len(leaves) == 1 {
			return singleResolution(leaves[0], Decision{Code: "single_surface_default", Reason: "el proyecto declara una única superficie hoja"}), nil
		}
		if composed, ok := findComposition(project.Surfaces, ids(leaves)); ok {
			return composedResolution(composed, leaves, "project_wide_composition", "sin archivos modificados se usa la composición declarada del proyecto", nil), nil
		}
		return Resolution{}, AmbiguousSurfaceError{Choices: ids(leaves)}
	}

	matched := map[string]SurfaceDefinition{}
	decisions := []Decision{}
	for _, file := range files {
		bestScore := 0
		var best []SurfaceDefinition
		for _, surface := range leaves {
			score := surfaceMatchScore(surface, file)
			if score == 0 || score < bestScore {
				continue
			}
			if score > bestScore {
				bestScore = score
				best = nil
			}
			best = append(best, surface)
		}
		for _, surface := range best {
			matched[surface.ID] = surface
		}
		if len(best) > 0 {
			decisions = append(decisions, Decision{Code: "root_match", Reason: "el archivo coincide con la root más específica", Evidence: []string{file, strings.Join(ids(best), ",")}})
		}
	}

	selected := mapValues(matched)
	if len(selected) == 0 {
		return Resolution{}, AmbiguousSurfaceError{Choices: ids(leaves)}
	}
	if len(selected) == 1 {
		if detectSignals(files)["contract_change"] {
			if composed, connected, ok := findContractComposition(project.Surfaces, selected[0]); ok {
				resolution := composedResolution(composed, connected, "contract_surface_composition", "un contrato modificado puede afectar a todas las superficies conectadas", files)
				resolution.Decisions = append(decisions, resolution.Decisions...)
				return resolution, nil
			}
		}
		resolution := singleResolution(selected[0], Decision{Code: "automatic_surface", Reason: "todos los archivos con alcance resoluble pertenecen a una superficie", Evidence: files})
		resolution.Decisions = append(decisions, resolution.Decisions...)
		return resolution, nil
	}
	if composed, ok := findComposition(project.Surfaces, ids(selected)); ok {
		resolution := composedResolution(composed, selected, "connected_surface_composition", "los archivos afectan superficies conectadas", files)
		resolution.Decisions = append(decisions, resolution.Decisions...)
		return resolution, nil
	}
	return Resolution{}, AmbiguousSurfaceError{Choices: ids(selected)}
}

func findContractComposition(surfaces []SurfaceDefinition, selected SurfaceDefinition) (SurfaceDefinition, []SurfaceDefinition, bool) {
	var candidates []SurfaceDefinition
	for _, surface := range surfaces {
		if surface.Type == "fullstack" && stringSet(surface.Connects)[selected.ID] {
			connected := surfacesByID(surfaces, surface.Connects)
			if len(surfaceIDsByType(connected, "frontend")) > 0 && len(surfaceIDsByType(connected, "backend")) > 0 {
				candidates = append(candidates, surface)
			}
		}
	}
	if len(candidates) == 0 {
		return SurfaceDefinition{}, nil, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if len(candidates[i].Connects) != len(candidates[j].Connects) {
			return len(candidates[i].Connects) < len(candidates[j].Connects)
		}
		return candidates[i].ID < candidates[j].ID
	})
	chosen := candidates[0]
	return chosen, surfacesByID(surfaces, chosen.Connects), true
}

func resolveExplicit(surfaces []SurfaceDefinition, requested string) (Resolution, error) {
	for _, surface := range surfaces {
		if surface.ID == requested {
			if surface.Type == "fullstack" {
				connected := surfacesByID(surfaces, surface.Connects)
				return composedResolution(surface, connected, "explicit_surface", "la selección explícita tiene precedencia", []string{requested}), nil
			}
			return singleResolution(surface, Decision{Code: "explicit_surface", Reason: "la selección explícita tiene precedencia", Evidence: []string{requested}}), nil
		}
	}
	var byType []SurfaceDefinition
	for _, surface := range surfaces {
		if surface.Type == requested {
			byType = append(byType, surface)
		}
	}
	if len(byType) == 1 {
		return resolveExplicit(surfaces, byType[0].ID)
	}
	if len(byType) > 1 {
		return Resolution{}, AmbiguousSurfaceError{Choices: ids(byType)}
	}
	return Resolution{}, fmt.Errorf("superficie %q no configurada; opciones: %v", requested, ids(surfaces))
}

func singleResolution(surface SurfaceDefinition, decision Decision) Resolution {
	return Resolution{Mode: ModeSingle, PrimarySurface: surface.ID, Surfaces: []SurfaceDefinition{surface}, Decisions: []Decision{decision}}
}

func composedResolution(composite SurfaceDefinition, leaves []SurfaceDefinition, code, reason string, evidence []string) Resolution {
	active := append([]SurfaceDefinition{}, leaves...)
	sort.Slice(active, func(i, j int) bool { return active[i].ID < active[j].ID })
	active = append(active, composite)
	return Resolution{
		Mode:           ModeComposed,
		PrimarySurface: composite.ID,
		Surfaces:       active,
		Contracts:      contractEdges(active, composite),
		Decisions:      []Decision{{Code: code, Reason: reason, Evidence: evidence}},
	}
}

func contractEdges(active []SurfaceDefinition, composite SurfaceDefinition) []ContractEdge {
	var frontend []SurfaceDefinition
	var backend []SurfaceDefinition
	for _, surface := range active {
		switch surface.Type {
		case "frontend":
			frontend = append(frontend, surface)
		case "backend":
			backend = append(backend, surface)
		}
	}
	var edges []ContractEdge
	for _, server := range backend {
		for _, client := range frontend {
			edges = append(edges, ContractEdge{From: server.ID, To: client.ID, Kind: "frontend_backend_contract"})
		}
	}
	if len(edges) == 0 {
		for _, connected := range composite.Connects {
			edges = append(edges, ContractEdge{From: composite.ID, To: connected, Kind: "surface_connection"})
		}
	}
	return edges
}

func findComposition(surfaces []SurfaceDefinition, selected []string) (SurfaceDefinition, bool) {
	selectedSet := stringSet(selected)
	var candidates []SurfaceDefinition
	for _, surface := range surfaces {
		if surface.Type != "fullstack" {
			continue
		}
		connected := stringSet(surface.Connects)
		containsAll := true
		for id := range selectedSet {
			if !connected[id] {
				containsAll = false
				break
			}
		}
		if containsAll {
			candidates = append(candidates, surface)
		}
	}
	if len(candidates) == 0 {
		return SurfaceDefinition{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if len(candidates[i].Connects) != len(candidates[j].Connects) {
			return len(candidates[i].Connects) < len(candidates[j].Connects)
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true
}

func surfaceMatchScore(surface SurfaceDefinition, file string) int {
	file = normalizeSlash(file)
	best := 0
	for _, root := range surface.Roots {
		root = strings.TrimSuffix(normalizeSlash(root), "/")
		score := 0
		switch {
		case root == "." || root == "":
			score = 1
		case file == root || strings.HasPrefix(file, root+"/"):
			score = len(root) + 1
		}
		if score > best {
			best = score
		}
	}
	return best
}

func leafSurfaces(surfaces []SurfaceDefinition) []SurfaceDefinition {
	var out []SurfaceDefinition
	for _, surface := range surfaces {
		if surface.Type != "fullstack" {
			out = append(out, surface)
		}
	}
	return out
}

func surfacesByID(surfaces []SurfaceDefinition, wanted []string) []SurfaceDefinition {
	wantedSet := stringSet(wanted)
	var out []SurfaceDefinition
	for _, surface := range surfaces {
		if wantedSet[surface.ID] && surface.Type != "fullstack" {
			out = append(out, surface)
		}
	}
	return out
}

func mapValues(values map[string]SurfaceDefinition) []SurfaceDefinition {
	out := make([]SurfaceDefinition, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func ids(surfaces []SurfaceDefinition) []string {
	out := make([]string, 0, len(surfaces))
	for _, surface := range surfaces {
		out = append(out, surface.ID)
	}
	sort.Strings(out)
	return out
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func normalizeSlash(value string) string {
	return strings.TrimPrefix(strings.ReplaceAll(value, "\\", "/"), "./")
}
