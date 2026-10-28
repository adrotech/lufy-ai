package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/evidence"
)

func readEvidence(path string) (evidence.Record, error) {
	input, err := os.Open(path)
	if err != nil {
		return evidence.Record{}, fmt.Errorf("evidencia no disponible")
	}
	defer input.Close()
	return evidence.Decode(input)
}

func runEvidence(args []string, deps Dependencies) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "help" || args[0] == "-h") {
		fmt.Fprintln(deps.Stdout, "Uso: lufy-ai evidence inspect --file <evidencia.json>")
		return ExitOK
	}
	if len(args) == 0 || args[0] != "inspect" {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai evidence inspect --file <evidencia.json>")
		return ExitUsageErr
	}
	fs := flag.NewFlagSet("evidence inspect", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	file := fs.String("file", "", "Comprobaciones declaradas en JSON; no es un reporte obligatorio")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 0 || *file == "" {
		return ExitUsageErr
	}
	record, err := readEvidence(*file)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err)
		return ExitResultInvalid
	}
	if err := json.NewEncoder(deps.Stdout).Encode(evidence.Assess(record)); err != nil {
		return ExitRuntimeErr
	}
	return ExitOK
}
