package toolruntime

import (
	"fmt"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/config"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const OpenCodeProjectConfigFile = config.OpenCodeFile

type ProjectConfigResult struct {
	File    string
	Action  string
	Changed bool
}

func GlobalRoot(tool domain.ToolID) (string, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		return platform.ResolveOpenCodeConfigRoot()
	default:
		return "", unsupportedToolError(tool)
	}
}

func ProjectConfigFile(tool domain.ToolID) (string, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		return OpenCodeProjectConfigFile, nil
	default:
		return "", unsupportedToolError(tool)
	}
}

func PlanProjectConfig(tool domain.ToolID, targetRoot string) (ProjectConfigResult, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		result, err := config.NewService().Plan(config.Options{TargetRoot: targetRoot})
		return fromConfigResult(result), err
	default:
		return ProjectConfigResult{}, unsupportedToolError(tool)
	}
}

func EnsureProjectConfig(tool domain.ToolID, targetRoot string) (ProjectConfigResult, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		result, err := config.NewService().Ensure(config.Options{TargetRoot: targetRoot})
		return fromConfigResult(result), err
	default:
		return ProjectConfigResult{}, unsupportedToolError(tool)
	}
}

func ValidateProjectConfig(tool domain.ToolID, targetRoot string) (bool, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		result, err := config.NewService().ValidateManagedStructure(targetRoot)
		return result.Exists, err
	default:
		return false, unsupportedToolError(tool)
	}
}

func PluginConfigFiles(tool domain.ToolID) ([]string, error) {
	switch normalizeTool(tool) {
	case domain.ToolInitialDefault:
		return []string{"tui.json", OpenCodeProjectConfigFile}, nil
	default:
		return nil, unsupportedToolError(tool)
	}
}

func normalizeTool(tool domain.ToolID) domain.ToolID {
	if tool == "" {
		return domain.ToolInitialDefault
	}
	return tool
}

func unsupportedToolError(tool domain.ToolID) error {
	return fmt.Errorf("tool runtime no soporta configuración escribible para %s", normalizeTool(tool))
}

func fromConfigResult(result config.Result) ProjectConfigResult {
	return ProjectConfigResult{File: OpenCodeProjectConfigFile, Action: result.Action, Changed: result.Changed}
}
