package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufysdd"
)

func runSDD(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printSDDHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "new":
		return runSDDNew(args[1:], deps)
	case "status":
		return runSDDInspect("status", args[1:], deps)
	case "validate":
		return runSDDInspect("validate", args[1:], deps)
	case "sync":
		return runSDDInspect("sync", args[1:], deps)
	case "archive":
		return runSDDInspect("archive", args[1:], deps)
	case "-h", "--help", "help":
		printSDDHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando sdd desconocido: %s\n\n", args[0])
		printSDDHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runSDDNew(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("sdd new", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	change := fs.String("change", "", "ID kebab-case del change")
	mode := fs.String("mode", "full", "Mode del change: full o lite")
	capability := fs.String("capability", "", "Capability kebab-case; por default usa el change ID")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai sdd new --change <name> [--mode full|lite] [--capability <name>] [--target <dir>] [--json]")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		fs.Usage()
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *change == "" {
		if *change == "" {
			fmt.Fprintln(deps.Stderr, "sdd new requiere --change <name>")
		} else {
			fmt.Fprintln(deps.Stderr, "sdd new no acepta argumentos posicionales")
		}
		fs.Usage()
		return ExitUsageErr
	}
	report, err := lufysdd.NewService().NewWithMode(*target, *change, *capability, lufysdd.Mode(*mode))
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeSDDReport(deps, *jsonOutput, report)
}

func runSDDInspect(action string, args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("sdd "+action, flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	change := fs.String("change", "", "ID kebab-case del change")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	strict := false
	if action == "validate" {
		fs.BoolVar(&strict, "strict", false, "Exigir todos los artifacts y tasks del mode efectivo")
	}
	fs.Usage = func() {
		if action == "validate" {
			fmt.Fprintln(deps.Stderr, "Uso: lufy-ai sdd validate --change <name> [--strict] [--target <dir>] [--json]")
			return
		}
		fmt.Fprintf(deps.Stderr, "Uso: lufy-ai sdd %s --change <name> [--target <dir>] [--json]\n", action)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		fs.Usage()
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 || *change == "" {
		if *change == "" {
			fmt.Fprintf(deps.Stderr, "sdd %s requiere --change <name>\n", action)
		} else {
			fmt.Fprintf(deps.Stderr, "sdd %s no acepta argumentos posicionales\n", action)
		}
		fs.Usage()
		return ExitUsageErr
	}
	service := lufysdd.NewService()
	var (
		report lufysdd.Report
		err    error
	)
	switch action {
	case "validate":
		report, err = service.Validate(*target, *change, strict)
	case "status":
		report, err = service.Status(*target, *change)
	case "sync":
		report, err = service.Sync(*target, *change)
	case "archive":
		report, err = service.Archive(*target, *change)
	}
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	code := writeSDDReport(deps, *jsonOutput, report)
	if code == ExitOK && !report.Valid() {
		return ExitRuntimeErr
	}
	return code
}

func writeSDDReport(deps Dependencies, jsonOutput bool, report lufysdd.Report) int {
	if jsonOutput {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		fmt.Fprintln(deps.Stdout, string(body))
		return ExitOK
	}
	fmt.Fprintf(deps.Stdout, "Lufy SDD %s: %s\n", report.Action, report.Status)
	fmt.Fprintf(deps.Stdout, "change: %s\nmode: %s\ntasks: %d/%d\n", report.Change, report.Mode, report.Progress.Complete, report.Progress.Total)
	for _, diagnostic := range report.Diagnostics {
		location := diagnostic.Path
		if diagnostic.Line > 0 {
			location = fmt.Sprintf("%s:%d", location, diagnostic.Line)
		}
		fmt.Fprintf(deps.Stdout, "[%s] %s %s: %s\n", diagnostic.Level, diagnostic.Code, location, diagnostic.Message)
	}
	return ExitOK
}
