package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/surfaceplan"
)

func runPlan(args []string, deps Dependencies) int {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(deps.Stderr)
	target := flags.String("target", ".", "Directorio del proyecto")
	surface := flags.String("surface", "auto", "ID o tipo de superficie; auto usa roots y diff")
	base := flags.String("base", "HEAD", "Referencia base para git diff")
	files := flags.String("files", "", "Paths relativos separados por coma; reemplazan git diff")
	capabilities := flags.String("capabilities", "", "Capacidades separadas por coma")
	jsonOutput := flags.Bool("json", false, "Emitir salida JSON")
	flags.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai plan [--target <dir>] [--surface auto|<id>|<type>] [--base <ref>] [--files <a,b>] [--capabilities <a,b>] [--json]")
		fmt.Fprintln(deps.Stderr, "Genera un plan read-only de superficies, contratos y validaciones; no ejecuta comandos.")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		flags.Usage()
		return ExitUsageErr
	}
	if len(flags.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "plan no acepta argumentos posicionales")
		flags.Usage()
		return ExitUsageErr
	}
	plan, err := surfaceplan.NewService().Build(surfaceplan.Options{
		Target:           *target,
		RequestedSurface: *surface,
		Base:             *base,
		Files:            splitCSV(*files),
		Capabilities:     splitCSV(*capabilities),
	})
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if *jsonOutput {
		encoder := json.NewEncoder(deps.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(plan); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		return ExitOK
	}
	writeSurfacePlan(deps.Stdout, plan)
	return ExitOK
}

func writeSurfacePlan(out io.Writer, plan surfaceplan.ExecutionPlan) {
	fmt.Fprintln(out, "Plan de ejecución por superficies")
	fmt.Fprintf(out, "Superficie primaria: %s (%s)\n", plan.PrimarySurface, plan.Mode)
	fmt.Fprintf(out, "Superficies activas: %s\n", surfaceNames(plan.ActiveSurfaces))
	if len(plan.ChangedFiles) == 0 {
		fmt.Fprintln(out, "Archivos: sin cambios; planificación de alcance completo")
	} else {
		fmt.Fprintf(out, "Archivos: %s\n", strings.Join(plan.ChangedFiles, ", "))
	}
	if len(plan.Capabilities) > 0 {
		fmt.Fprintf(out, "Capacidades: %s\n", strings.Join(plan.Capabilities, ", "))
	}
	if len(plan.Signals) > 0 {
		fmt.Fprintf(out, "Señales: %s\n", strings.Join(plan.Signals, ", "))
	}
	fmt.Fprintln(out, "Decisiones:")
	for _, decision := range plan.Decisions {
		fmt.Fprintf(out, "  - %s: %s", decision.Code, decision.Reason)
		if len(decision.Evidence) > 0 {
			fmt.Fprintf(out, " [%s]", strings.Join(decision.Evidence, ", "))
		}
		fmt.Fprintln(out)
	}
	if len(plan.Contracts) > 0 {
		fmt.Fprintln(out, "Contratos:")
		for _, contract := range plan.Contracts {
			fmt.Fprintf(out, "  - %s -> %s (%s)\n", contract.From, contract.To, contract.Kind)
		}
	}
	fmt.Fprintln(out, "Validaciones requeridas:")
	for _, rule := range plan.ValidationRules {
		fmt.Fprintf(out, "  - %s [%s]: %s\n", rule.ID, rule.Category, rule.Rationale)
		if len(rule.Commands) > 0 {
			fmt.Fprintf(out, "    comandos sugeridos: %s\n", strings.Join(rule.Commands, " | "))
		}
		if len(rule.Evidence) > 0 {
			fmt.Fprintf(out, "    evidencia: %s\n", strings.Join(rule.Evidence, "; "))
		}
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func surfaceNames(surfaces []surfaceplan.SurfaceDefinition) string {
	values := make([]string, 0, len(surfaces))
	for _, surface := range surfaces {
		values = append(values, surface.ID+":"+surface.Type)
	}
	return strings.Join(values, ", ")
}
