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

	adaptiveadapters "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/adapters"
	adaptiveapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/application"
	adaptivedomain "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func runAdaptive(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printAdaptiveHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "recommend":
		return runAdaptiveRecommend(args[1:], deps)
	case "assign":
		return runAdaptiveAssign(args[1:], deps)
	case "yield":
		return runAdaptiveYield(args[1:], deps)
	case "status":
		return runAdaptiveStatus(args[1:], deps)
	case "-h", "--help", "help":
		printAdaptiveHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando adaptive desconocido: %s\n\n", args[0])
		printAdaptiveHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runAdaptiveRecommend(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("adaptive recommend", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID requerido para recording/capacity")
	file := fs.String("file", "", "Archivo YAML/JSON; por default stdin")
	record := fs.Bool("record", false, "Registrar demand y recommendation en advisory")
	key := fs.String("idempotency-key", "", "Clave estable requerida con --record")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai adaptive recommend [--file <path>] [--target <dir>] [--run <id>] [--record --idempotency-key <key>] [--json]")
	}
	if err := fs.Parse(args); err != nil {
		return adaptiveFlagError(err)
	}
	if len(fs.Args()) > 0 || *record && (*runID == "" || strings.TrimSpace(*key) == "") {
		fs.Usage()
		return ExitUsageErr
	}
	input, closeInput, err := adaptiveInput(*file, deps.Stdin)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "adaptive input no disponible")
		return ExitRuntimeErr
	}
	defer closeInput()
	evaluation, err := adaptivedomain.DecodeEvaluation(input)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	config, service, ledgerErr, err := newAdaptiveService(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if *record && ledgerErr != nil {
		fmt.Fprintln(deps.Stderr, "adaptive ledger no disponible")
		return ExitResultUnavailable
	}
	result, err := service.Recommend(backgroundContext(), adaptiveapp.RecommendRequest{
		RunID: *runID, Evaluation: evaluation, Config: config, Record: *record, IdempotencyKey: *key,
	})
	if err != nil {
		return writeAdaptiveError(deps, err)
	}
	return writeAdaptiveResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "adaptive recommend: action=%s mode=%s actor=%s role=%s durability=%s gate_advanced=%t capacity=%s\n",
			result.Decision.Action, result.Decision.Mode, result.Decision.SelectedActorRef,
			result.Decision.SelectedRoleHint, result.Durability, result.Decision.GateAdvanced, result.Capacity.Availability)
	})
}

func runAdaptiveAssign(args []string, deps Dependencies) int {
	return runAdaptiveMutation("assign", args, deps)
}

func runAdaptiveYield(args []string, deps Dependencies) int {
	return runAdaptiveMutation("yield", args, deps)
}

func runAdaptiveMutation(action string, args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("adaptive "+action, flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID")
	file := fs.String("file", "", "Archivo YAML/JSON; por default stdin")
	record := fs.Bool("record", false, "Confirmar efecto durable explícito")
	key := fs.String("idempotency-key", "", "Clave estable de idempotencia")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintf(deps.Stderr, "Uso: lufy-ai adaptive %s --run <id> --record --idempotency-key <key> [--file <path>] [--target <dir>] [--json]\n", action)
	}
	if err := fs.Parse(args); err != nil {
		return adaptiveFlagError(err)
	}
	if len(fs.Args()) > 0 || *runID == "" || strings.TrimSpace(*key) == "" {
		fs.Usage()
		return ExitUsageErr
	}
	input, closeInput, err := adaptiveInput(*file, deps.Stdin)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "adaptive input no disponible")
		return ExitRuntimeErr
	}
	defer closeInput()
	config, service, ledgerErr, err := newAdaptiveService(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if ledgerErr != nil {
		fmt.Fprintln(deps.Stderr, "adaptive ledger no disponible")
		return ExitResultUnavailable
	}
	var result adaptivedomain.LedgerOperation
	if action == "assign" {
		value, decodeErr := adaptivedomain.DecodeAssignment(input)
		if decodeErr != nil {
			fmt.Fprintln(deps.Stderr, decodeErr.Error())
			return ExitUsageErr
		}
		result, err = service.Assign(backgroundContext(), adaptiveapp.MutationRequest[adaptivedomain.Assignment]{
			RunID: *runID, Value: value, Config: config, Record: *record, IdempotencyKey: *key,
		})
	} else {
		value, decodeErr := adaptivedomain.DecodeYieldCheckpoint(input)
		if decodeErr != nil {
			fmt.Fprintln(deps.Stderr, decodeErr.Error())
			return ExitUsageErr
		}
		result, err = service.Yield(backgroundContext(), adaptiveapp.MutationRequest[adaptivedomain.YieldCheckpoint]{
			RunID: *runID, Value: value, Config: config, Record: *record, IdempotencyKey: *key,
		})
	}
	if err != nil {
		return writeAdaptiveError(deps, err)
	}
	return writeAdaptiveResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "adaptive %s: status=%s event=%s sequence=%d gate_advanced=false\n",
			action, result.Status, result.EventID, result.EventSequence)
	})
}

func runAdaptiveStatus(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("adaptive status", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	runID := fs.String("run", "", "Run ID")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return adaptiveFlagError(err)
	}
	if len(fs.Args()) > 0 || *runID == "" {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai adaptive status --run <id> [--target <dir>] [--json]")
		return ExitUsageErr
	}
	config, service, ledgerErr, err := newAdaptiveService(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if ledgerErr != nil {
		fmt.Fprintln(deps.Stderr, "adaptive ledger no disponible")
		return ExitResultUnavailable
	}
	result, err := service.Status(backgroundContext(), *runID, config)
	if err != nil {
		return writeAdaptiveError(deps, err)
	}
	return writeAdaptiveResult(deps, *jsonOutput, result, func() {
		fmt.Fprintf(deps.Stdout, "adaptive status: run=%s version=%d active=%d waiting=%d truncated=%t gate_advanced=false\n",
			result.RunID, result.Version, len(result.ActiveAssignments), len(result.Waiting), result.Truncated)
	})
}

func newAdaptiveService(target string) (adaptiveapp.RuntimeConfig, *adaptiveapp.Service, error, error) {
	config, err := loadAdaptiveProjectConfig(target)
	if err != nil {
		return adaptiveapp.RuntimeConfig{}, nil, nil, err
	}
	runtimeConfig := adaptiveapp.RuntimeConfig{
		Enabled: config.AdaptiveRouting.Enabled, Mode: adaptivedomain.Mode(config.AdaptiveRouting.Mode),
		PolicyVersion: config.AdaptiveRouting.PolicyVersion, LeaseTTLSeconds: config.AdaptiveRouting.LeaseTTLSeconds,
		MaxCandidates: config.AdaptiveRouting.MaxCandidates, MaxWaitingItems: config.AdaptiveRouting.MaxWaitingItems,
		StarvationAfterCycles: config.AdaptiveRouting.StarvationAfterCycles,
		ParallelEnabled:       config.ParallelExecution.Enabled,
		MaxParallelAgents:     config.ParallelExecution.MaxParallelAgents,
		MaxConcurrentSlices:   config.WorkflowLimits.Review.MaxConcurrentSlices,
	}
	store, _, ledgerErr := newRunStore(target)
	if ledgerErr != nil {
		return runtimeConfig, adaptiveapp.NewService(adaptiveapp.LedgerPort{}), ledgerErr, nil
	}
	adapter := adaptiveadapters.NewLedgerAdapter(store, nil)
	port := adaptiveapp.LedgerPort{
		RecordDemand: func(ctx context.Context, runID string, value adaptivedomain.DemandSignal, key string) (adaptivedomain.LedgerOperation, error) {
			result, callErr := adapter.RecordDemand(ctx, adaptiveadapters.RecordDemandRequest{RunID: runID, Demand: value, IdempotencyKey: key})
			return adaptiveLedgerOperation(result), callErr
		},
		RecordRecommendation: func(ctx context.Context, runID string, value adaptivedomain.Recommendation, key string) (adaptivedomain.LedgerOperation, error) {
			result, callErr := adapter.RecordRecommendation(ctx, adaptiveadapters.RecordRecommendationRequest{RunID: runID, Recommendation: value, IdempotencyKey: key})
			return adaptiveLedgerOperation(result), callErr
		},
		Assign: func(ctx context.Context, runID string, value adaptivedomain.Assignment, key string) (adaptivedomain.LedgerOperation, error) {
			result, callErr := adapter.Assign(ctx, adaptiveadapters.AssignRequest{RunID: runID, Assignment: value, IdempotencyKey: key})
			return adaptiveLedgerOperation(result), callErr
		},
		Yield: func(ctx context.Context, runID string, value adaptivedomain.YieldCheckpoint, key string) (adaptivedomain.LedgerOperation, error) {
			result, callErr := adapter.Yield(ctx, adaptiveadapters.YieldRequest{RunID: runID, Checkpoint: value, IdempotencyKey: key})
			return adaptiveLedgerOperation(result), callErr
		},
		Status: adapter.Status,
	}
	return runtimeConfig, adaptiveapp.NewService(port), nil, nil
}

func loadAdaptiveProjectConfig(target string) (projectconfig.ProjectConfig, error) {
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return projectconfig.ProjectConfig{}, err
	}
	if _, err := os.Stat(path); err == nil {
		return projectconfig.Load(path)
	} else if !os.IsNotExist(err) {
		return projectconfig.ProjectConfig{}, err
	}
	return projectconfig.ProjectConfig{
		AdaptiveRouting:   projectconfig.DefaultAdaptiveRoutingConfig(),
		ParallelExecution: projectconfig.DefaultParallelExecutionConfig(),
		RunLedger:         projectconfig.DefaultRunLedgerConfig(),
	}, nil
}

func adaptiveLedgerOperation(result adaptiveadapters.OperationResult) adaptivedomain.LedgerOperation {
	return adaptivedomain.LedgerOperation{
		Status: string(result.Status), EventID: result.EventID,
		EventSequence: result.EventSequence, State: result.State,
	}
}

func adaptiveInput(path string, stdin io.Reader) (io.Reader, func(), error) {
	if strings.TrimSpace(path) == "" {
		if stdin == nil {
			stdin = os.Stdin
		}
		return stdin, func() {}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	return file, func() { _ = file.Close() }, nil
}

func adaptiveFlagError(err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return ExitOK
	}
	return ExitUsageErr
}

func writeAdaptiveError(deps Dependencies, err error) int {
	fmt.Fprintln(deps.Stderr, err.Error())
	switch {
	case errors.Is(err, adaptiveapp.ErrRecordingRequired), errors.Is(err, adaptiveapp.ErrMutationDisabled),
		errors.Is(err, adaptiveapp.ErrLeaseOutOfBounds):
		return ExitResultRejected
	case errors.Is(err, runledger.ErrIdempotencyConflict), errors.Is(err, runledger.ErrVersionConflict),
		errors.Is(err, adaptiveadapters.ErrAssignmentUnavailable), errors.Is(err, adaptiveadapters.ErrLeaseMismatch),
		errors.Is(err, adaptiveadapters.ErrLeaseExpired):
		return ExitResultConflict
	case errors.Is(err, adaptiveapp.ErrLedgerUnavailable):
		return ExitResultUnavailable
	default:
		return ExitResultUnavailable
	}
}

func writeAdaptiveResult(deps Dependencies, jsonOutput bool, value any, human func()) int {
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

func printAdaptiveHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai adaptive <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos:")
	fmt.Fprintln(out, "  recommend  Evalúa una demanda en disabled/shadow/advisory; read-only por default")
	fmt.Fprintln(out, "  assign     Confirma una recommendation vigente con lease fenced y --record")
	fmt.Fprintln(out, "  yield      Registra checkpoint content-free y libera recursos con --record")
	fmt.Fprintln(out, "  status     Reconstruye assignments, budget y waiting pool desde Run Ledger")
	fmt.Fprintln(out, "Exit: 0=ok 2=usage 4=rejected 5=conflict 6=unavailable")
}
