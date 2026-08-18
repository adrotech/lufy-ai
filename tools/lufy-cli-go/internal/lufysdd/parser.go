package lufysdd

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	deltaHeading       = regexp.MustCompile(`^## (ADDED|MODIFIED|REMOVED) Requirements\s*$`)
	requirementHeading = regexp.MustCompile(`^### Requirement:\s*(.+?)\s*$`)
	scenarioHeading    = regexp.MustCompile(`^#### Scenario:\s*(.+?)\s*$`)
	whenClause         = regexp.MustCompile(`(?i)^\s*-?\s*\*\*WHEN\*\*(?:\s|:|$)`)
	thenClause         = regexp.MustCompile(`(?i)^\s*-?\s*\*\*THEN\*\*(?:\s|:|$)`)
	taskLine           = regexp.MustCompile(`^\s*-\s*\[([ xX])\]\s+`)
)

func ParseDelta(path, capability, body string) (Delta, []Diagnostic) {
	delta := Delta{Capability: capability, Path: path, Requirements: []Requirement{}}
	diagnostics := []Diagnostic{}
	marker := ""
	currentRequirement := -1
	currentScenario := -1

	scanner := bufio.NewScanner(strings.NewReader(body))
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if match := deltaHeading.FindStringSubmatch(text); match != nil {
			marker = match[1]
			currentRequirement = -1
			currentScenario = -1
			continue
		}
		if match := requirementHeading.FindStringSubmatch(text); match != nil {
			title := strings.TrimSpace(match[1])
			if marker == "" {
				diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "requirement_without_delta", Path: path, Line: line, Message: "requirement fuera de una sección delta"})
			}
			if title == "" {
				diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "empty_requirement", Path: path, Line: line, Message: "título de requirement vacío"})
			}
			delta.Requirements = append(delta.Requirements, Requirement{Marker: marker, Title: title, Line: line, Scenarios: []Scenario{}})
			currentRequirement = len(delta.Requirements) - 1
			currentScenario = -1
			continue
		}
		if match := scenarioHeading.FindStringSubmatch(text); match != nil {
			if currentRequirement < 0 {
				diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "scenario_without_requirement", Path: path, Line: line, Message: "scenario sin requirement padre"})
				continue
			}
			title := strings.TrimSpace(match[1])
			delta.Requirements[currentRequirement].Scenarios = append(delta.Requirements[currentRequirement].Scenarios, Scenario{Title: title, Line: line})
			currentScenario = len(delta.Requirements[currentRequirement].Scenarios) - 1
			continue
		}
		if currentRequirement >= 0 && currentScenario >= 0 {
			scenario := &delta.Requirements[currentRequirement].Scenarios[currentScenario]
			if whenClause.MatchString(text) {
				scenario.HasWhen = true
			}
			if thenClause.MatchString(text) {
				scenario.HasThen = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "read_delta", Path: path, Message: err.Error()})
	}

	if len(delta.Requirements) == 0 {
		diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "missing_requirements", Path: path, Message: "spec delta sin requirements"})
	}
	seen := map[string]int{}
	for _, requirement := range delta.Requirements {
		key := strings.ToLower(requirement.Title)
		if previous, ok := seen[key]; ok {
			diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "duplicate_requirement", Path: path, Line: requirement.Line, Message: fmt.Sprintf("requirement duplicado; primera aparición en línea %d", previous)})
		} else {
			seen[key] = requirement.Line
		}
		if requirement.Marker == "ADDED" || requirement.Marker == "MODIFIED" {
			if len(requirement.Scenarios) == 0 {
				diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "missing_scenario", Path: path, Line: requirement.Line, Message: "requirement agregado o modificado sin scenario"})
			}
			for _, scenario := range requirement.Scenarios {
				if !scenario.HasWhen {
					diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "missing_when", Path: path, Line: scenario.Line, Message: "scenario sin cláusula WHEN"})
				}
				if !scenario.HasThen {
					diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "missing_then", Path: path, Line: scenario.Line, Message: "scenario sin cláusula THEN"})
				}
			}
		}
	}
	sortDiagnostics(diagnostics)
	bodies := requirementBodies(body)
	for i := range delta.Requirements {
		delta.Requirements[i].Body = bodies[delta.Requirements[i].Line]
	}
	return delta, diagnostics
}

type requirementBlock struct {
	Title string
	Start int
	End   int
	Body  string
}

func requirementBlocks(body string) []requirementBlock {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	blocks := []requirementBlock{}
	start := -1
	title := ""
	flush := func(end int) {
		if start < 0 {
			return
		}
		text := strings.TrimRight(strings.Join(lines[start:end], "\n"), "\n") + "\n"
		startOffset := lineOffset(lines, start)
		endOffset := lineOffset(lines, end)
		if endOffset > len(normalized) {
			endOffset = len(normalized)
		}
		blocks = append(blocks, requirementBlock{Title: title, Start: startOffset, End: endOffset, Body: text})
	}
	for i, line := range lines {
		if match := requirementHeading.FindStringSubmatch(line); match != nil {
			flush(i)
			start = i
			title = strings.TrimSpace(match[1])
			continue
		}
		if start >= 0 && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			flush(i)
			start = -1
			title = ""
		}
	}
	flush(len(lines))
	return blocks
}

func requirementBodies(body string) map[int]string {
	out := map[int]string{}
	for _, block := range requirementBlocks(body) {
		line := 1 + strings.Count(strings.ReplaceAll(body, "\r\n", "\n")[:block.Start], "\n")
		out[line] = block.Body
	}
	return out
}

func lineOffset(lines []string, line int) int {
	offset := 0
	for i := 0; i < line && i < len(lines); i++ {
		offset += len(lines[i]) + 1
	}
	return offset
}

func TaskProgress(body string) Progress {
	progress := Progress{}
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		match := taskLine.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		progress.Total++
		if match[1] == "x" || match[1] == "X" {
			progress.Complete++
		}
	}
	return progress
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].Path != diagnostics[j].Path {
			return diagnostics[i].Path < diagnostics[j].Path
		}
		if diagnostics[i].Line != diagnostics[j].Line {
			return diagnostics[i].Line < diagnostics[j].Line
		}
		return diagnostics[i].Code < diagnostics[j].Code
	})
}
