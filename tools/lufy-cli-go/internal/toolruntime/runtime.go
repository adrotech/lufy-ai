package toolruntime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

const CodexProjectConfigFile = ".codex/config.toml"

type ProjectConfigResult struct {
	File    string
	Action  string
	Changed bool
}

func GlobalRoot(tool domain.ToolID) (string, error) {
	switch normalizeTool(tool) {
	default:
		return "", unsupportedToolError(tool)
	}
}

func ProjectConfigFile(tool domain.ToolID) (string, error) {
	switch normalizeTool(tool) {
	case domain.ToolCodex:
		return CodexProjectConfigFile, nil
	default:
		return "", unsupportedToolError(tool)
	}
}

func PlanProjectConfig(tool domain.ToolID, targetRoot string) (ProjectConfigResult, error) {
	switch normalizeTool(tool) {
	case domain.ToolCodex:
		return ProjectConfigResult{File: CodexProjectConfigFile}, nil
	default:
		return ProjectConfigResult{}, unsupportedToolError(tool)
	}
}

func EnsureProjectConfig(tool domain.ToolID, targetRoot string) (ProjectConfigResult, error) {
	switch normalizeTool(tool) {
	case domain.ToolCodex:
		return ProjectConfigResult{File: CodexProjectConfigFile}, nil
	default:
		return ProjectConfigResult{}, unsupportedToolError(tool)
	}
}

func ValidateProjectConfig(tool domain.ToolID, targetRoot string) (bool, error) {
	switch normalizeTool(tool) {
	case domain.ToolCodex:
		info, err := os.Lstat(filepath.Join(targetRoot, CodexProjectConfigFile))
		if os.IsNotExist(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0, nil
	default:
		return false, unsupportedToolError(tool)
	}
}

func PluginConfigFiles(tool domain.ToolID) ([]string, error) {
	switch normalizeTool(tool) {
	case domain.ToolCodex:
		return []string{CodexProjectConfigFile}, nil
	default:
		return nil, unsupportedToolError(tool)
	}
}

func normalizeTool(tool domain.ToolID) domain.ToolID {
	if tool == "" {
		return domain.ToolCodex
	}
	return tool
}

func unsupportedToolError(tool domain.ToolID) error {
	return fmt.Errorf("tool runtime no soporta configuración escribible para %s", normalizeTool(tool))
}
