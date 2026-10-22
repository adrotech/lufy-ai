package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
)

func runCandidate(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printCandidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	if resultHelpRequested(args) {
		printCandidateHelp(deps.Stdout)
		return ExitOK
	}
	if args[0] != "capture" && args[0] != "verify" {
		printCandidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	fs := flag.NewFlagSet("candidate "+args[0], flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	base := fs.String("base", "", "Base Git explícita")
	evidence := fs.String("evidence", "", "Result Contract de evidencia")
	role := fs.String("role", "reviewer", "Rol registrado del claim")
	bindingPath := fs.String("binding", "", "Sidecar JSON; sólo verify")
	jsonOutput := fs.Bool("json", false, "Emitir JSON; capture siempre emite JSON")
	if err := fs.Parse(args[1:]); err != nil || len(fs.Args()) != 0 || *base == "" || *evidence == "" || (args[0] == "verify") != (*bindingPath != "") {
		printCandidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	input, closeInput, err := resultInput(false, *evidence, deps)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "evidencia no disponible")
		return ExitRuntimeErr
	}
	defer closeInput()
	contract, err := resultcontract.Decode(input)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "Result Contract inválido")
		return ExitResultRejected
	}
	claim := resultcontract.EvaluateClaim(contract, *role, resultRolePolicy(), resultEvidencePolicy())
	if !claim.Accepted {
		fmt.Fprintln(deps.Stderr, "claim del rol rechazado")
		return ExitResultRejected
	}
	canonical, err := resultcontract.Canonicalize(contract)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "evidencia no canonicalizable")
		return ExitResultRejected
	}
	if args[0] == "capture" {
		binding, err := candidateevidence.Capture(backgroundContext(), *target, *base, canonical.Fingerprint, *role)
		if err != nil {
			fmt.Fprintln(deps.Stderr, err)
			return ExitResultRejected
		}
		if err := json.NewEncoder(deps.Stdout).Encode(binding); err != nil {
			return ExitRuntimeErr
		}
		return ExitOK
	}
	f, err := os.Open(*bindingPath)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "binding no disponible")
		return ExitRuntimeErr
	}
	defer f.Close()
	binding, err := candidateevidence.Decode(f)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err)
		return ExitResultRejected
	}
	if err := candidateevidence.Verify(backgroundContext(), *target, *base, canonical.Fingerprint, *role, binding); err != nil {
		fmt.Fprintln(deps.Stderr, err)
		return ExitResultRejected
	}
	if *jsonOutput {
		if err := json.NewEncoder(deps.Stdout).Encode(map[string]string{"schema_version": "candidate-evidence-verification/v1", "status": "current"}); err != nil {
			return ExitRuntimeErr
		}
	} else {
		fmt.Fprintln(deps.Stdout, "candidate verify: current; sin avance de gates")
	}
	return ExitOK
}

func printCandidateHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai candidate <capture|verify> --target <repo> --base <ref> --evidence <Result Contract> [--role reviewer] [--binding <sidecar.json>] [--json]")
	fmt.Fprintln(out, "capture emite sidecar JSON por stdout; verify exige --binding. Requiere candidato comprometido limpio; no avanza gates.")
}
