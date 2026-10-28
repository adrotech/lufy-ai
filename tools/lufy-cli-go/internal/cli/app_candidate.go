package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/evidence"
)

func runCandidate(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printCandidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
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
	evidencePath := fs.String("evidence", "", "Archivo JSON de comprobaciones declaradas")
	bindingPath := fs.String("binding", "", "Sidecar JSON; sólo verify")
	jsonOutput := fs.Bool("json", false, "Emitir JSON; capture siempre emite JSON")
	if err := fs.Parse(args[1:]); err != nil || len(fs.Args()) != 0 || *base == "" || *evidencePath == "" || (args[0] == "verify") != (*bindingPath != "") {
		printCandidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	record, err := readEvidence(*evidencePath)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err)
		return ExitResultRejected
	}
	fingerprint := evidence.Fingerprint(record)

	if args[0] == "capture" {
		binding, err := candidateevidence.Capture(backgroundContext(), *target, *base, fingerprint)
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
	if err := candidateevidence.Verify(backgroundContext(), *target, *base, fingerprint, binding); err != nil {
		fmt.Fprintln(deps.Stderr, err)
		return ExitResultRejected
	}
	if *jsonOutput {
		if err := json.NewEncoder(deps.Stdout).Encode(map[string]string{"schema_version": "candidate-evidence-verification/v1", "status": "current"}); err != nil {
			return ExitRuntimeErr
		}
	} else {
		fmt.Fprintln(deps.Stdout, "candidate verify: current; vinculación vigente; las comprobaciones siguen siendo declaradas")
	}
	return ExitOK
}

func printCandidateHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai candidate <capture|verify> --target <repo> --base <ref> --evidence <comprobaciones.json> [--binding <sidecar.json>] [--json]")
	fmt.Fprintln(out, "capture emite sidecar JSON por stdout; verify exige --binding. Requiere candidato comprometido limpio; no ejecuta comprobaciones ni concede autorización.")
}
