package codexlifecycle

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/skillregistry"
)

const maxContextBytes = 1800

type Input struct {
	CWD                  string  `json:"cwd"`
	HookEventName        string  `json:"hook_event_name"`
	SessionID            string  `json:"session_id,omitempty"`
	TurnID               string  `json:"turn_id,omitempty"`
	AgentID              string  `json:"agent_id,omitempty"`
	AgentType            string  `json:"agent_type,omitempty"`
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type Output struct {
	Continue           bool                `json:"continue"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
	SuppressOutput     bool                `json:"suppressOutput"`
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type Service struct{}

func NewService() Service { return Service{} }

func (Service) Run(input io.Reader, output io.Writer) error {
	var event Input
	if err := json.NewDecoder(input).Decode(&event); err != nil {
		return writeOutput(output, Output{Continue: true, SystemMessage: "LUFY lifecycle recibió un payload JSON inválido."})
	}
	target, err := platform.ResolveTargetPath(event.CWD)
	if err != nil {
		return writeOutput(output, Output{Continue: true, SystemMessage: "LUFY lifecycle no pudo resolver el repositorio; ejecuta lufy-ai doctor manualmente."})
	}

	if event.HookEventName != "SessionStart" {
		return writeOutput(output, Output{Continue: true, SuppressOutput: true})
	}
	return writeOutput(output, sessionStart(target))
}

func sessionStart(target string) Output {
	parts := []string{"LUFY orientation (metadata only)"}
	warnings := []string{}
	skills := skillregistry.NewService()
	if err := skills.Ensure(skillregistry.Options{Target: target, Tool: domain.ToolCodex}, io.Discard); err != nil {
		parts = append(parts, "skills=unavailable recovery=lufy-ai skills ensure --target <repo> --tool codex")
		warnings = append(warnings, "skill registry no disponible")
	} else if report, err := skills.Inspect(skillregistry.Options{Target: target, Tool: domain.ToolCodex}); err != nil {
		parts = append(parts, "skills=unavailable recovery=lufy-ai skills ensure --target <repo> --tool codex")
		warnings = append(warnings, "skill registry no evaluable")
	} else {
		parts = append(parts, fmt.Sprintf("skills=%s count=%d path=.lufy/skill-registry.json", report.Status, report.SkillCount))
	}

	if report, err := memory.NewService().BuildStatus(memory.Options{Target: target}); err != nil {
		parts = append(parts, "memory=unavailable recovery=lufy-ai memory status --target <repo>")
		warnings = append(warnings, "memoria no evaluable")
	} else if report.Status.Initialized {
		parts = append(parts, fmt.Sprintf("memory=ready notes=%d root=%s", report.Status.Notes, report.Root))
	} else {
		parts = append(parts, "memory=not_initialized recovery=lufy-ai memory init --target <repo>")
	}

	contextStatus := contextapp.NewService().Status(target)
	switch contextStatus.Status {
	case "ready":
		parts = append(parts, fmt.Sprintf("context=ready nodes=%d edges=%d", contextStatus.Nodes, contextStatus.Edges))
	case "stale":
		parts = append(parts, "context=stale recovery=lufy-ai context build --target <repo>")
	default:
		parts = append(parts, "context=not_available recovery=lufy-ai context build --target <repo>")
	}

	message := ""
	if len(warnings) > 0 {
		message = "LUFY SessionStart: " + strings.Join(warnings, "; ") + ". El inicio continúa sin avanzar gates."
	}
	return Output{
		Continue:       true,
		SystemMessage:  message,
		SuppressOutput: false,
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: bounded(strings.Join(parts, "; ")),
		},
	}
}

func bounded(value string) string {
	if len(value) <= maxContextBytes {
		return value
	}
	return value[:maxContextBytes] + "…"
}

func writeOutput(output io.Writer, result Output) error {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}
