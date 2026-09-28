package codexlifecycle

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufysdd"
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
		return writeOutput(output, Output{Continue: true, SystemMessage: "LUFY lifecycle recibió un payload JSON inválido; no se avanzó ningún gate."})
	}
	target, err := platform.ResolveTargetPath(event.CWD)
	if err != nil {
		return writeOutput(output, Output{Continue: true, SystemMessage: "LUFY lifecycle no pudo resolver el repositorio; ejecuta lufy-ai doctor manualmente."})
	}

	var result Output
	switch event.HookEventName {
	case "SessionStart":
		result = sessionStart(target)
	case "SubagentStart":
		result = Output{Continue: true, SuppressOutput: true}
	case "SubagentStop":
		result = subagentStop(event)
	case "Stop":
		result = stop(target)
	case "SessionEnd":
		result = sessionEnd(target)
	default:
		result = Output{Continue: true, SystemMessage: "LUFY lifecycle ignoró un evento Codex no soportado."}
	}
	if supportsLedgerEvent(event.HookEventName) {
		if err := recordLifecycleEvent(target, event); err != nil {
			result.SystemMessage = joinMessages(result.SystemMessage, "LUFY Run Ledger no pudo registrar metadata estable; el lifecycle continúa sin avanzar gates. Recovery: ejecuta lufy-ai run verify manualmente.")
			result.SuppressOutput = false
		}
	}
	return writeOutput(output, result)
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

func subagentStop(event Input) Output {
	if event.LastAssistantMessage == nil || strings.TrimSpace(*event.LastAssistantMessage) == "" {
		return Output{
			Continue: true, SuppressOutput: false,
			SystemMessage: "LUFY SubagentStop: resultado vacío; el orchestrator debe recuperar o reasignar y ningún gate fue avanzado.",
		}
	}
	if _, err := extractResultContract(*event.LastAssistantMessage); err != nil {
		return Output{
			Continue: true, SuppressOutput: false,
			SystemMessage: "LUFY SubagentStop: Result Contract ausente, ambiguo o inválido; requiere revisión del orchestrator y ningún gate fue avanzado.",
		}
	}
	return Output{Continue: true, SuppressOutput: true}
}

func stop(target string) Output {
	changes, err := activeChanges(target)
	if err != nil || len(changes) == 0 {
		return Output{Continue: true, SuppressOutput: true}
	}
	if len(changes) > 1 {
		return Output{Continue: true, SystemMessage: fmt.Sprintf("LUFY Stop: hay %d changes activos; selecciona uno antes de inferir estado. Ningún gate fue avanzado.", len(changes))}
	}
	report, err := lufysdd.NewService().Status(target, changes[0])
	if err != nil {
		return Output{Continue: true, SystemMessage: "LUFY Stop: change activo no evaluable; ejecuta lufy-ai sdd status manualmente. Ningún gate fue avanzado."}
	}
	message := fmt.Sprintf("LUFY Stop: change=%s status=%s tasks=%d/%d; %s. Diagnóstico read-only: ningún gate fue avanzado.", report.Change, report.Status, report.Progress.Complete, report.Progress.Total, nextAction(report.Status, report.Change))
	return Output{Continue: true, SystemMessage: bounded(message)}
}

func sessionEnd(target string) Output {
	report, err := memory.NewService().BuildStatus(memory.Options{Target: target})
	if err != nil || !report.Status.Initialized {
		return Output{Continue: true, SuppressOutput: true}
	}
	validated, err := memory.NewService().BuildValidate(memory.Options{Target: target})
	if err != nil || !validated.OK {
		return Output{Continue: true, SystemMessage: "LUFY SessionEnd: memoria inicializada con advertencias; ejecuta lufy-ai memory validate --target <repo>."}
	}
	return Output{Continue: true, SuppressOutput: true}
}

func activeChanges(target string) ([]string, error) {
	root, err := lufypaths.ResolveExisting(target, lufypaths.LufySDD, lufypaths.LegacyLufySDD)
	if err != nil || !root.Exists {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(root.Path, "changes"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	changes := []string{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			changes = append(changes, entry.Name())
		}
	}
	sort.Strings(changes)
	return changes, nil
}

func nextAction(status, change string) string {
	switch status {
	case "in_progress":
		return "continúa la siguiente task pendiente y conserva evidencia por slice"
	case "verification":
		return "registra evidencia bajo .lufy/workflows/sdd/verification/" + change
	case "sync_pending":
		return "ejecuta lufy-ai sdd sync sólo después de validación proporcional"
	case "completed":
		return "solicita autorización explícita antes de delivery o archive"
	case "blocked":
		return "revisa diagnostics con lufy-ai sdd status --change " + change
	default:
		return "consulta lufy-ai sdd status --change " + change
	}
}

func bounded(value string) string {
	if len(value) <= maxContextBytes {
		return value
	}
	return value[:maxContextBytes] + "…"
}

func supportsLedgerEvent(name string) bool {
	switch name {
	case "SessionStart", "SubagentStart", "SubagentStop", "Stop", "SessionEnd":
		return true
	default:
		return false
	}
}

func joinMessages(current, next string) string {
	if strings.TrimSpace(current) == "" {
		return next
	}
	return current + " " + next
}

func writeOutput(output io.Writer, result Output) error {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}
