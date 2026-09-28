package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func runRun(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printRunHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "record":
		return runRunRecord(args[1:], deps)
	case "checkpoint":
		return runRunCheckpoint(args[1:], deps)
	case "status":
		return runRunQuery("status", args[1:], deps)
	case "summary":
		return runRunQuery("summary", args[1:], deps)
	case "verify":
		return runRunVerify(args[1:], deps)
	case "prune":
		return runRunPrune(args[1:], deps)
	case "-h", "--help", "help":
		printRunHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando run desconocido: %s\n\n", args[0])
		printRunHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runRunRecord(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("run record", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	idempotencyKey := fs.String("idempotency-key", "", "Clave estable de idempotencia")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai run record --idempotency-key <key> [--target <dir>] [--json]")
		fmt.Fprintln(deps.Stderr, "Lee un EventDraft JSON estricto desde stdin.")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		fs.Usage()
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || strings.TrimSpace(*idempotencyKey) == "" {
		fmt.Fprintln(deps.Stderr, "run record requiere --idempotency-key y no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	input := deps.Stdin
	if input == nil {
		input = os.Stdin
	}
	draft, err := runledger.DecodeDraft(input)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	store, _, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	result, err := store.Append(backgroundContext(), runledger.AppendRequest{Draft: draft, IdempotencyKey: *idempotencyKey})
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeRunResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "run event: %s run=%s event=%s clock=%d sequence=%d\n", result.Status, result.Event.RunID, result.Event.EventID, result.Event.LamportClock, result.Event.LocalSequence)
	})
}

func runRunCheckpoint(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("run checkpoint", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID")
	parentRunID := fs.String("parent-run", "", "Parent run ID opcional")
	causedBy := fs.String("caused-by", "", "Event ID causal opcional")
	status := fs.String("status", "", "Estado Result Contract")
	gate := fs.String("gate", "", "Gate opcional")
	nextOwner := fs.String("next-owner", "", "Siguiente owner opcional")
	taskRef := fs.String("task", "", "Task ref opcional")
	adapter := fs.String("adapter", "manual", "Adapter productor")
	eventName := fs.String("event", "Checkpoint", "Nombre de evento productor")
	idempotencyKey := fs.String("idempotency-key", "", "Clave estable de idempotencia")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai run checkpoint --run <id> --status <status> --idempotency-key <key> [flags]")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		fs.Usage()
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *runID == "" || *status == "" || strings.TrimSpace(*idempotencyKey) == "" {
		fmt.Fprintln(deps.Stderr, "run checkpoint requiere --run, --status y --idempotency-key; no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	kind := runledger.KindCheckpoint
	if *status == "blocked" {
		kind = runledger.KindBlocked
	} else if *status == "closed" || *status == "delivered" {
		kind = runledger.KindFinish
	}
	draft := runledger.EventDraft{
		RunID: *runID, ParentRunID: *parentRunID, CausedByEventID: *causedBy, Kind: kind,
		Source: runledger.Source{Adapter: *adapter, EventName: *eventName}, TaskRef: *taskRef,
		Checkpoint: &runledger.Checkpoint{Status: *status, Gate: *gate, NextOwner: *nextOwner},
	}
	store, _, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	result, err := store.Append(backgroundContext(), runledger.AppendRequest{Draft: draft, IdempotencyKey: *idempotencyKey})
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeRunResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "run checkpoint: %s run=%s status=%s event=%s\n", result.Status, result.Event.RunID, *status, result.Event.EventID)
	})
}

func runRunQuery(action string, args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("run "+action, flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *runID == "" {
		fmt.Fprintf(deps.Stderr, "Uso: lufy-ai run %s --run <id> [--target <dir>] [--json]\n", action)
		return ExitUsageErr
	}
	store, _, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	projector := runledger.NewProjector(store)
	summary, err := projector.Build(backgroundContext(), *runID)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if action == "summary" {
		return writeRunResult(deps, *jsonOutput, summary, func() { printRunSummary(deps.Stdout, summary, "") })
	}
	verification, err := projector.Verify(backgroundContext(), *runID, false)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	status := runStatusView{
		SchemaVersion: "lufy-run-status/v1", RunID: summary.RunID, Status: summary.Status,
		Terminal: summary.Terminal, EventCount: summary.EventCount, ChildCount: len(summary.Children),
		MetricsAvailability: summary.Metrics.Availability, SourceHealthy: verification.SourceHealthy,
		ProjectionPresent: verification.ProjectionPresent, ProjectionFresh: verification.ProjectionFresh,
	}
	return writeRunResult(deps, *jsonOutput, status, func() {
		fmt.Fprintf(deps.Stdout, "run %s: status=%s terminal=%t events=%d children=%d metrics=%s source_healthy=%t projection=%s\n",
			status.RunID, status.Status, status.Terminal, status.EventCount, status.ChildCount, status.MetricsAvailability,
			status.SourceHealthy, projectionLabel(status.ProjectionPresent, status.ProjectionFresh))
	})
}

func runRunVerify(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("run verify", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID")
	repair := fs.Bool("repair", false, "Reconstruir solo proyecciones derivadas")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *runID == "" {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai run verify --run <id> [--repair] [--target <dir>] [--json]")
		return ExitUsageErr
	}
	store, _, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	result, err := runledger.NewProjector(store).Verify(backgroundContext(), *runID, *repair)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	code := writeRunResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "run verify: source_healthy=%t projection=%s repaired=%t\n", result.SourceHealthy, projectionLabel(result.ProjectionPresent, result.ProjectionFresh), result.Repaired)
		for _, report := range result.Reports {
			for _, issue := range report.Issues {
				fmt.Fprintf(deps.Stdout, "[%s] %s ref=%s\n", issue.Severity, issue.Code, issue.Ref)
			}
		}
	})
	if code == ExitOK && !result.SourceHealthy {
		return ExitRuntimeErr
	}
	return code
}

func runRunPrune(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("run prune", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	dryRun := fs.Bool("dry-run", false, "Previsualizar sin eliminar")
	yes := fs.Bool("yes", false, "Confirmar eliminación de runs terminales elegibles")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *dryRun == *yes {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai run prune (--dry-run|--yes) [--target <dir>] [--json]")
		return ExitUsageErr
	}
	store, config, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	policy := runledger.RetentionPolicy{
		MaxAge:          time.Duration(config.Retention.MaxAgeDays) * 24 * time.Hour,
		MaxTerminalRuns: config.Retention.MaxTerminalRuns,
		MaxBytes:        config.Retention.MaxBytes,
	}
	report, err := store.Prune(backgroundContext(), policy, time.Now().UTC(), *yes)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeRunResult(deps, *jsonOutput, report, func() {
		fmt.Fprintf(deps.Stdout, "run prune: dry_run=%t candidates=%d deleted=%d reclaimed_bytes=%d protected_active=%d\n",
			report.DryRun, len(report.Candidates), len(report.Deleted), report.ReclaimedBytes, len(report.ProtectedActive))
		for _, candidate := range report.Candidates {
			fmt.Fprintf(deps.Stdout, "- %s bytes=%d reasons=%s\n", candidate.RunID, candidate.Bytes, strings.Join(candidate.Reasons, ","))
		}
	})
}

type runStatusView struct {
	SchemaVersion       string `json:"schema_version"`
	RunID               string `json:"run_id"`
	Status              string `json:"status"`
	Terminal            bool   `json:"terminal"`
	EventCount          int    `json:"event_count"`
	ChildCount          int    `json:"child_count"`
	MetricsAvailability string `json:"metrics_availability"`
	SourceHealthy       bool   `json:"source_healthy"`
	ProjectionPresent   bool   `json:"projection_present"`
	ProjectionFresh     bool   `json:"projection_fresh"`
}

func newRunStore(target string) (*runledger.FileStore, projectconfig.RunLedgerConfig, error) {
	config := projectconfig.DefaultRunLedgerConfig()
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return nil, config, err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		loaded, loadErr := projectconfig.Load(path)
		if loadErr != nil {
			return nil, config, fmt.Errorf("leer run_ledger config: %w", loadErr)
		}
		config = loaded.RunLedger
	} else if !os.IsNotExist(statErr) {
		return nil, config, statErr
	}
	if !config.IsEnabled() {
		return nil, config, fmt.Errorf("run ledger deshabilitado en project.yaml")
	}
	store, err := runledger.NewFileStore(target, runledger.Options{RuntimeRoot: config.Root})
	return store, config, err
}

func writeRunResult(deps Dependencies, jsonOutput bool, value any, human func()) int {
	if jsonOutput {
		body, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		fmt.Fprintln(deps.Stdout, string(body))
		return ExitOK
	}
	human()
	return ExitOK
}

func printRunSummary(out io.Writer, summary runledger.RunSummary, indent string) {
	fmt.Fprintf(out, "%s- run=%s status=%s terminal=%t events=%d metrics=%s tasks=%s artifacts=%d evidence=%d blockers=%d\n",
		indent, summary.RunID, summary.Status, summary.Terminal, summary.EventCount, summary.Metrics.Availability,
		strings.Join(summary.Tasks, ","), len(summary.ArtifactRefs), len(summary.EvidenceRefs), len(summary.BlockerEventIDs))
	for _, child := range summary.Children {
		printRunSummary(out, child, indent+"  ")
	}
}

func projectionLabel(present, fresh bool) string {
	if !present {
		return "absent"
	}
	if fresh {
		return "fresh"
	}
	return "stale"
}

func printRunHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai run <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos:")
	fmt.Fprintln(out, "  record      Registra un EventDraft JSON estricto desde stdin")
	fmt.Fprintln(out, "  checkpoint  Registra un checkpoint tipado")
	fmt.Fprintln(out, "  status      Resume estado, integridad y freshness")
	fmt.Fprintln(out, "  summary     Muestra el árbol causal completo")
	fmt.Fprintln(out, "  verify      Verifica fuente y opcionalmente repara derivados")
	fmt.Fprintln(out, "  prune       Previsualiza o aplica retención de runs terminales")
}

func backgroundContext() context.Context {
	return context.Background()
}
