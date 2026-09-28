package surfaceplan

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

type ProjectConfigAdapter struct{}

func (ProjectConfigAdapter) Load(target string) (ProjectDefinition, error) {
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return ProjectDefinition{}, fmt.Errorf("resolver project config: %w", err)
	}
	cfg, err := projectconfig.Load(path)
	if err != nil {
		return ProjectDefinition{}, fmt.Errorf("leer %s: %w; ejecuta lufy-ai init --target %s", path, err, target)
	}
	project := ProjectDefinition{Stacks: map[string]StackDefinition{}}
	for _, surface := range cfg.ProjectProfile.Surfaces {
		project.Surfaces = append(project.Surfaces, SurfaceDefinition{
			ID:                     surface.ID,
			Type:                   surface.Type,
			Roots:                  append([]string{}, surface.Roots...),
			Stacks:                 append([]string{}, surface.Stacks...),
			Frameworks:             append([]string{}, surface.Frameworks...),
			Connects:               append([]string{}, surface.Connects...),
			Capabilities:           append([]string{}, surface.Capabilities...),
			Architecture:           surface.Architecture.Preferred,
			PrimaryConcerns:        append([]string{}, surface.AgentLens.PrimaryConcerns...),
			StructuralExpectations: uniqueStrings(append(append([]string{}, surface.AgentLens.StructuralExpectations...), surface.Architecture.StructuralExpectations...)),
			ValidationExpectations: append([]string{}, surface.AgentLens.ValidationExpectations...),
		})
	}
	for _, stack := range cfg.Stacks {
		project.Stacks[stack.ID] = StackDefinition{
			ID:              stack.ID,
			TestCommand:     stack.TestRunner.Command,
			CoverageCommand: stack.TestRunner.CoverageCommand,
			LintCommand:     stack.Linter.Command,
			StaticCommand:   stack.StaticAnalysis.Command,
		}
	}
	return project, nil
}

type GitChangeAdapter struct{}

func (GitChangeAdapter) ChangedFiles(target, base string) ([]string, error) {
	if strings.TrimSpace(base) == "" {
		base = "HEAD"
	}
	if strings.HasPrefix(base, "-") {
		return nil, fmt.Errorf("referencia Git inválida: %s", base)
	}
	command := exec.Command("git", "-C", target, "diff", "--name-only", "--relative", base, "--")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("leer archivos modificados desde git diff %s: %w", base, err)
	}
	var files []string
	for _, line := range strings.Split(string(output), "\n") {
		if value := strings.TrimSpace(line); value != "" {
			files = append(files, value)
		}
	}
	return files, nil
}
