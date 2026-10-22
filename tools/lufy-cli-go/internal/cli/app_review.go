package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/reviewevents"
)

func runReview(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printReviewHelp(deps.Stderr)
		return ExitUsageErr
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printReviewHelp(deps.Stdout)
		return ExitOK
	}
	action := args[0]
	if action != "start" && action != "complete" && action != "rework" && action != "reopen" {
		printReviewHelp(deps.Stderr)
		return ExitUsageErr
	}
	fs := flag.NewFlagSet("review "+action, flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	run := fs.String("run", "", "Run ID opaco")
	review := fs.String("review", "", "ID opaco del intento de review")
	operation := fs.String("operation", "", "ID estable de ocurrencia; por defecto review ID")
	result := fs.String("result", "", "Sólo complete: passed o failed, declarado por el host")
	evidence := fs.String("evidence", "", "Result Contract del reviewer; requerido salvo start")
	jsonOutput := fs.Bool("json", false, "Emitir resultado JSON content-free")
	if err := fs.Parse(args[1:]); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) != 0 || *run == "" || *review == "" {
		printReviewHelp(deps.Stderr)
		return ExitUsageErr
	}
	opts := reviewevents.Options{RunID: *run, ReviewID: *review, Action: action, OperationID: *operation, Result: *result}
	if action == "start" {
		if *evidence != "" {
			fmt.Fprintln(deps.Stderr, "start no acepta evidencia")
			return ExitUsageErr
		}
	} else {
		if *evidence == "" {
			fmt.Fprintln(deps.Stderr, "review requiere --evidence")
			return ExitUsageErr
		}
		input, closeInput, err := resultInput(false, *evidence, deps)
		if err != nil {
			fmt.Fprintln(deps.Stderr, "evidencia review no disponible")
			return ExitRuntimeErr
		}
		defer closeInput()
		contract, err := resultcontract.Decode(input)
		if err != nil {
			fmt.Fprintln(deps.Stderr, "Result Contract review inválido")
			return ExitResultRejected
		}
		claim := resultcontract.EvaluateClaim(contract, "reviewer", resultRolePolicy(), resultEvidencePolicy())
		if !claim.Accepted || *result == "passed" && (contract.Status == "blocked" || contract.Status == "escalated") {
			fmt.Fprintln(deps.Stderr, "claim reviewer u outcome no permitido")
			return ExitResultRejected
		}
		canonical, err := resultcontract.Canonicalize(contract)
		if err != nil {
			fmt.Fprintln(deps.Stderr, "evidencia review no canonicalizable")
			return ExitResultRejected
		}
		opts.EvidenceSHA256 = canonical.Fingerprint
	}
	store, _, err := newRunStore(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	appended, err := reviewevents.NewService(store).Record(backgroundContext(), opts)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeRunResult(deps, *jsonOutput, appended, func() {
		fmt.Fprintf(deps.Stdout, "review %s: %s event=%s; metadata sin avance de gates\n", action, appended.Status, appended.Event.EventID)
	})
}

func printReviewHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai review <start|complete|rework|reopen> --run <id> --review <id> [--operation <id>] [--target <repo>] [--json]")
	fmt.Fprintln(out, "Salvo start requiere --evidence <Result Contract>; complete requiere --result passed|failed. Registra sólo metadata; no ejecuta la review ni avanza gates.")
}
