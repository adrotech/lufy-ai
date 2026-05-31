package ports

import (
	"context"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

type DetectionResult struct {
	Detected bool
	Reason   string
}

type AssetSpec struct {
	ID        string
	TargetRel string
	Policy    string
	Scope     string
}

type Check struct {
	Level   string
	Path    string
	Message string
}

type Target struct {
	Root string
}

type Env map[string]string

type HarnessModel struct {
	Tool              domain.ToolID
	MethodologyByTier domain.MethodologyByTier
	Roles             []domain.RoleID
}

type WorkflowModel struct {
	Tier      domain.Tier
	Selection domain.MethodologySelection
}

type ToolAdapter interface {
	ID() domain.ToolID
	Capabilities() domain.ToolCapabilities
	Detect(context.Context, Env) DetectionResult
	RenderSurface(HarnessModel) ([]AssetSpec, error)
	Verify(Target) ([]Check, error)
}

type MethodologyAdapter interface {
	ID() domain.MethodologyID
	SupportedModes() []domain.MethodologyMode
	RenderWorkflow(WorkflowModel) ([]AssetSpec, error)
	VerifyWorkflow(Target, domain.Tier) ([]Check, error)
}
