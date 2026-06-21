package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	contextstore "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/adapters"
	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/governance"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/layout"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/merger"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/opsx"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/status"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/syncer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/uninstaller"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/upgrade"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

func Run(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printGeneralHelp(deps.Stderr)
		return ExitUsageErr
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], deps)
	case "scan":
		return runScan(args[1:], deps)
	case "install":
		return runInstall(args[1:], deps)
	case "uninstall":
		return runUninstall(args[1:], deps)
	case "verify":
		return runVerify(args[1:], deps)
	case "backup":
		return runBackup(args[1:], deps)
	case "restore":
		return runRestore(args[1:], deps)
	case "merge":
		return runMerge(args[1:], deps)
	case "migrate-layout":
		return runMigrateLayout(args[1:], deps)
	case "memory":
		return runMemory(args[1:], deps)
	case "sync":
		return runSync(args[1:], deps)
	case "status":
		return runStatus(args[1:], deps)
	case "info":
		return runInfo(args[1:], deps)
	case "doctor":
		return runDoctor(args[1:], deps)
	case "pin":
		return runPin(args[1:], deps)
	case "unpin":
		return runUnpin(args[1:], deps)
	case "upgrade":
		return runUpgrade(args[1:], deps)
	case "opsx":
		return runOpsx(args[1:], deps)
	case "context":
		return runContext(args[1:], deps)
	case "version":
		return runVersion(args[1:], deps)
	case "-h", "--help", "help":
		printGeneralHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Comando desconocido: %s\n\n", args[0])
		printGeneralHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runContext(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printContextHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "scan":
		return runContextScan(args[1:], deps)
	case "build":
		return runContextBuild(args[1:], deps)
	case "status":
		return runContextStatus(args[1:], deps)
	case "query":
		return runContextQuery(args[1:], deps)
	case "path":
		return runContextPath(args[1:], deps)
	case "explain":
		return runContextExplain(args[1:], deps)
	case "diff":
		return runContextDiff(args[1:], deps)
	case "-h", "--help", "help":
		printContextHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando context desconocido: %s\n\n", args[0])
		printContextHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func contextFlagSet(name string, deps Dependencies) (*flag.FlagSet, *string, *bool) {
	fs := flag.NewFlagSet("context "+name, flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	return fs, target, jsonOutput
}

func runContextScan(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("scan", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "context scan no acepta argumentos posicionales")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Scan(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeContextResult(deps, *jsonOutput, res, func() {
		fmt.Fprintf(deps.Stdout, "context scan: %d sources, %d nodes, %d edges (no persistido)\n", res.Sources, res.Nodes, res.Edges)
	})
}

func runContextBuild(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("build", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "context build no acepta argumentos posicionales")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Build(*target)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return writeContextResult(deps, *jsonOutput, res, func() {
		fmt.Fprintf(deps.Stdout, "context graph ready: %s (%d sources, %d nodes, %d edges, changed=%t)\n", res.GraphPath, res.Sources, res.Nodes, res.Edges, res.Changed)
	})
}

func runContextStatus(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("status", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "context status no acepta argumentos posicionales")
		return ExitUsageErr
	}
	res := contextapp.NewService().Status(*target)
	return writeContextResult(deps, *jsonOutput, res, func() {
		fmt.Fprintf(deps.Stdout, "context graph: %s\n", res.Status)
		if res.Reason != "" {
			fmt.Fprintf(deps.Stdout, "reason: %s\n", res.Reason)
		}
		if res.Recovery != "" {
			fmt.Fprintf(deps.Stdout, "recovery: %s\n", res.Recovery)
		}
	})
}

func runContextQuery(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("query", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai context query [--target <dir>] [--json] <term>")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Query(*target, fs.Args()[0])
	if err != nil {
		return contextGraphErr(deps, err, *jsonOutput)
	}
	return writeContextResult(deps, *jsonOutput, res, func() {
		for _, m := range res.Matches {
			fmt.Fprintf(deps.Stdout, "%s [%s] %s\n", m.Node.ID, m.Node.Type, m.Node.Label)
		}
	})
}

func runContextPath(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("path", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) != 2 {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai context path [--target <dir>] [--json] <from> <to>")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Path(*target, fs.Args()[0], fs.Args()[1])
	if err != nil {
		return contextGraphErr(deps, err, *jsonOutput)
	}
	return writeContextResult(deps, *jsonOutput, res, func() {
		if !res.Found {
			fmt.Fprintln(deps.Stdout, "path: not_found")
			return
		}
		fmt.Fprintf(deps.Stdout, "path: %s\n", joinComma(res.Nodes))
	})
}

func runContextExplain(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("explain", deps)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai context explain [--target <dir>] [--json] <node-or-edge>")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Explain(*target, fs.Args()[0])
	if err != nil {
		return contextGraphErr(deps, err, *jsonOutput)
	}
	return writeContextResult(deps, *jsonOutput, res, func() { fmt.Fprintf(deps.Stdout, "%s: %s\n", res.ID, res.Explanation) })
}

func runContextDiff(args []string, deps Dependencies) int {
	fs, target, jsonOutput := contextFlagSet("diff", deps)
	base := fs.String("base", "", "Referencia Git base")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if *base == "" {
		fmt.Fprintln(deps.Stderr, "context diff requiere --base <ref>")
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "context diff no acepta argumentos posicionales")
		return ExitUsageErr
	}
	res, err := contextapp.NewService().Diff(*target, *base)
	if err != nil {
		return contextGraphErr(deps, err, *jsonOutput)
	}
	return writeContextResult(deps, *jsonOutput, res, func() {
		fmt.Fprintf(deps.Stdout, "changed files: %d\nimpact nodes: %d\n", len(res.ChangedFiles), len(res.Impact))
	})
}

func writeContextResult(deps Dependencies, jsonOutput bool, value interface{}, human func()) int {
	if jsonOutput {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		fmt.Fprintln(deps.Stdout, string(data))
		return ExitOK
	}
	human()
	return ExitOK
}

func contextGraphErr(deps Dependencies, err error, jsonOutput bool) int {
	if errors.Is(err, contextstore.ErrGraphNotAvailable) {
		value := map[string]string{"status": "not_available", "recovery": "lufy-ai context build"}
		return writeContextResult(deps, jsonOutput, value, func() { fmt.Fprintln(deps.Stderr, "context graph not_available; ejecuta lufy-ai context build") })
	}
	fmt.Fprintln(deps.Stderr, err.Error())
	return ExitRuntimeErr
}

func runOpsx(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printOpsxHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "render":
		return runOpsxRender(args[1:], deps)
	case "-h", "--help", "help":
		printOpsxHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando opsx desconocido: %s\n\n", args[0])
		printOpsxHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runOpsxRender(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("opsx render", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	change := fs.String("change", "", "Nombre del change OpenSpec")
	format := fs.String("format", "html", "Formato de salida: html")
	theme := fs.String("theme", "notion-dark", "Tema HTML: notion-dark")
	output := fs.String("output", "", "Ruta de salida opcional")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai opsx render --change <name> [--target <dir>] [--format html] [--theme notion-dark] [--output <path>]")
		fmt.Fprintln(deps.Stderr, "Renderiza artifacts OpenSpec proposal/design/tasks/specs en un HTML offline opcional.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "opsx render no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	res, err := opsx.NewChangeRenderer().Render(opsx.RenderOptions{Target: *target, Change: *change, Format: *format, Theme: *theme, Output: *output})
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	fmt.Fprintf(deps.Stdout, "HTML OpenSpec generado: [%s](%s)\n", res.OutputPath, fileURL(res.OutputPath))
	fmt.Fprintf(deps.Stdout, "Abrir: open %s\n", res.OutputPath)
	return ExitOK
}

func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}

func runMemory(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printMemoryHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "init":
		return runMemoryInit(args[1:], deps)
	case "status":
		return runMemoryStatus(args[1:], deps)
	case "validate":
		return runMemoryValidate(args[1:], deps)
	case "search":
		return runMemorySearch(args[1:], deps)
	case "capture":
		return runMemoryCapture(args[1:], deps)
	case "connect":
		return runMemoryConnect(args[1:], deps)
	case "index":
		return runMemoryIndex(args[1:], deps)
	case "-h", "--help", "help":
		printMemoryHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando memory desconocido: %s\n\n", args[0])
		printMemoryHelp(deps.Stderr)
		return ExitUsageErr
	}
}

type repeatedStringFlag []string

func (f *repeatedStringFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *repeatedStringFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func runMemoryInit(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory init", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai memory init [--target <dir>] [--json]")
		fmt.Fprintln(deps.Stderr, "Crea .lufy/memory con política Git ignored e integra defaults en .lufy/config/project.yaml.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "memory init no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if err := memory.NewService().Init(memory.Options{Target: *target, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemoryStatus(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory status", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "memory status no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Status(memory.Options{Target: *target, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemoryValidate(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory validate", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "memory validate no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Validate(memory.Options{Target: *target, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemorySearch(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory search", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai memory search [--target <dir>] [--json] <query>")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fs.Usage()
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Search(memory.Options{Target: *target, Query: fs.Args()[0], JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemoryCapture(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory capture", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutar")
	title := fs.String("title", "", "Título durable de la nota")
	noteType := fs.String("type", "lesson", "Tipo: decision|rule|flow|lesson|concept")
	var links repeatedStringFlag
	fs.Var(&links, "link", "Slug existente para enlazar; repetible")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai memory capture [--target <dir>] --title <title> [--type <type>] [--link <slug>] [--dry-run] [--json] <texto>")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	body := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if *title == "" || body == "" {
		fs.Usage()
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Capture(memory.CaptureOptions{Target: *target, Title: *title, Type: *noteType, Body: body, Links: links, DryRun: *dryRun, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemoryConnect(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory connect", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutar")
	bidirectional := fs.Bool("bidirectional", false, "Agregar enlace en ambas notas")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai memory connect [--target <dir>] [--bidirectional] [--dry-run] [--json] <from-slug> <to-slug>")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 2 {
		fs.Usage()
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Connect(memory.ConnectOptions{Target: *target, From: fs.Args()[0], To: fs.Args()[1], Bidirectional: *bidirectional, DryRun: *dryRun, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMemoryIndex(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("memory index", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Repositorio target")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutar")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai memory index [--target <dir>] [--dry-run] [--json]")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := memory.NewService().Index(memory.IndexOptions{Target: *target, DryRun: *dryRun, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runUninstall(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	keepState := fs.Bool("keep-state", false, "Preservar .lufy/managed-state/install-state.json")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai uninstall [--target <dir>] [--dry-run] [--yes] [--keep-state]")
		fmt.Fprintln(deps.Stderr, "Remueve assets gestionados por Lufy con backup previo y preserva archivos user-owned.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "uninstall no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, *dryRun, *yes, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if err := uninstaller.NewService().Run(uninstaller.Options{Target: *target, DryRun: *dryRun, Yes: *yes, KeepState: *keepState}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runInit(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	force := fs.Bool("force", false, "Reemplazar .lufy/config/project.yaml existente")
	rescan := fs.Bool("rescan", false, "Refrescar evidencia de stacks, preservar overrides y reportar drift sin cleanup destructivo")
	interactive := fs.Bool("interactive", true, "Abrir selector interactivo para project_profile.surfaces cuando haya TTY")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai init [--target <dir>] [--force] [--rescan] [--interactive=false]")
		fmt.Fprintln(deps.Stderr, "Genera .lufy/config/project.yaml con reglas stack-aware y surface-aware editables.")
		fmt.Fprintln(deps.Stderr, "--rescan compara evidencia actual, preserva overrides de usuario y reporta drift sin borrar stacks ni archivos.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "init no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	if *force && *rescan {
		fmt.Fprintln(deps.Stderr, "init no permite combinar --force y --rescan")
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	opts := projectconfig.Options{Target: *target, Force: *force, Rescan: *rescan}
	if *interactive {
		opts.ProfilePrompt = surfaceProfilePrompt(deps)
	}
	if err := projectconfig.NewService().Run(opts, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMigrateLayout(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("migrate-layout", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar migracion real")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai migrate-layout [--target <dir>] [--dry-run] [--yes] [--json]")
		fmt.Fprintln(deps.Stderr, "Migra rutas legacy .lufy-ai/ y .lufy/project.yaml al layout unificado .lufy/.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "migrate-layout no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	if err := layout.NewService().Run(layout.Options{Target: *target, DryRun: *dryRun, Yes: *yes, JSON: *jsonOutput}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runScan(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	interactive := fs.Bool("interactive", true, "Abrir selector interactivo para project_profile.surfaces cuando haya TTY")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai scan [--target <dir>] [--interactive=false]")
		fmt.Fprintln(deps.Stderr, "Escanea stacks y superficies, crea o actualiza .lufy/config/project.yaml preservando overrides.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "scan no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	opts := projectconfig.Options{Target: *target, Rescan: true}
	if *interactive {
		opts.ProfilePrompt = surfaceProfilePrompt(deps)
	}
	if err := projectconfig.NewService().Run(opts, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runMerge(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	acceptTheirs := fs.Bool("accept-theirs", false, "Resolver aceptando <path>.lufy-new sin LUFY_MERGE_TOOL")
	acceptOurs := fs.Bool("accept-ours", false, "Resolver preservando el target local sin LUFY_MERGE_TOOL")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai merge [--target <dir>] [--accept-theirs|--accept-ours] <path>")
		fmt.Fprintln(deps.Stderr, "Reconcilia target, ancestor y .lufy-new usando LUFY_MERGE_TOOL o una resolución no interactiva.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fs.Usage()
		return ExitUsageErr
	}
	if *acceptTheirs && *acceptOurs {
		fmt.Fprintln(deps.Stderr, "merge no permite combinar --accept-theirs y --accept-ours")
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if err := merger.NewService().Run(merger.Options{Target: *target, Path: fs.Args()[0], AcceptTheirs: *acceptTheirs, AcceptOurs: *acceptOurs}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runUpgrade(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	to := fs.String("to", "", "Versión destino vX.Y.Z")
	baseURL := fs.String("base-url", "", "Base URL de releases")
	dryRun := fs.Bool("dry-run", false, "Mostrar descarga/reemplazo sin mutar")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "upgrade no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if err := upgrade.NewService().Run(upgrade.Options{To: *to, BaseURL: *baseURL, DryRun: *dryRun}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runStatus(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	verbose := fs.Bool("verbose", false, "Mostrar detalle por asset")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "status no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	if err := status.NewService().Run(status.Options{Target: *target, JSON: *jsonOutput, Verbose: *verbose, Scope: scope}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runInfo(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("info", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "info no acepta argumentos posicionales")
		return ExitUsageErr
	}
	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	if err := governance.NewService().Info(governance.Options{Target: *target, JSON: *jsonOutput, Scope: scope}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runDoctor(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "doctor no acepta argumentos posicionales")
		return ExitUsageErr
	}
	if !*jsonOutput {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	if err := governance.NewService().Doctor(governance.Options{Target: *target, JSON: *jsonOutput, Scope: scope}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runPin(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("pin", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	reason := fs.String("reason", "", "Motivo opcional del freeze")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai pin [--target <dir>] [--reason <texto>] <path>")
		fmt.Fprintln(deps.Stderr, "Congela un asset gestionado para que sync lo preserve sin modificar.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fs.Usage()
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if err := governance.NewService().Pin(governance.PinOptions{Target: *target, Path: fs.Args()[0], Reason: *reason}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runUnpin(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("unpin", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai unpin [--target <dir>] <path>")
		fmt.Fprintln(deps.Stderr, "Remueve el freeze de un asset gestionado para permitir sync normal.")
	}
	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) != 1 {
		fs.Usage()
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	if err := governance.NewService().Unpin(governance.PinOptions{Target: *target, Path: fs.Args()[0]}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runVersion(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "version no acepta argumentos posicionales")
		return ExitUsageErr
	}
	fmt.Fprintln(deps.Stdout, version.Current().String())
	return ExitOK
}

func runSync(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)

	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	harnessFlags := addToolFlag(fs)

	fs.Usage = func() {
		fmt.Fprintf(deps.Stderr, "Uso: lufy-ai sync [--target <dir>] [--scope project|global|both] [--tool %s] [--dry-run] [--yes]\n", writableToolUsage())
		fmt.Fprintln(deps.Stderr, "Sincroniza assets gestionados con manifest/hash/backup sin tocar drift local ni archivos no gestionados.")
	}

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "sync no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}

	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	harness, err := parseHarnessFlags(harnessFlags)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	service := syncer.NewService()
	if err := ensureLayoutForMutation(*target, *dryRun, *yes, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	err = service.Run(syncer.Options{Target: *target, DryRun: *dryRun, Yes: *yes, Scope: scope, Harness: harness}, deps.Stdout)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runVerify(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	jsonOutput := fs.Bool("json", false, "Emitir salida JSON")
	quiet := fs.Bool("quiet", false, "Omitir salida humana si no hay errores")
	verbose := fs.Bool("verbose", false, "Mostrar diagnóstico adicional")
	deep := fs.Bool("deep", false, "Ejecutar validaciones profundas opt-in")
	harnessFlags := addToolFlag(fs)
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	svc := verify.NewService()
	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	harness, err := parseHarnessFlags(harnessFlags)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	if !*jsonOutput && !*quiet {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
	}
	if err := svc.Run(verify.Options{Target: *target, JSON: *jsonOutput, Quiet: *quiet, Verbose: *verbose, Deep: *deep, Scope: scope, ExpectedTool: harness.Tool}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runBackup(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, false, true, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	svc := backup.NewService()
	if _, err := svc.Run(backup.Options{Target: *target}, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runRestore(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	backupPath := fs.String("backup", "", "Ruta a backup o manifest.json")
	list := fs.Bool("list", false, "Lista backups disponibles")
	dryRun := fs.Bool("dry-run", false, "Mostrar restore sin mutaciones")
	yes := fs.Bool("yes", false, "Acepta confirmaciones")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	if *list {
		if err := reportLegacyLayoutForReadOnly(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		if err := backup.NewService().List(*target, deps.Stdout); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return ExitRuntimeErr
		}
		return ExitOK
	}
	if *backupPath == "" {
		fmt.Fprintln(deps.Stderr, "restore requiere --backup <ruta>")
		return ExitUsageErr
	}
	if err := ensureLayoutForMutation(*target, *dryRun, *yes, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	svc := backup.NewService()
	if err := svc.Restore(*target, *backupPath, *dryRun, *yes, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	return ExitOK
}

func runInstall(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)

	target := fs.String("target", ".", "Directorio destino")
	scopeValue := fs.String("scope", "project", "Scope efectivo: project, global o both")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	backup := fs.Bool("backup", false, "Forzar backup cuando aplique")
	harnessFlags := addHarnessFlags(fs)

	fs.Usage = func() {
		fmt.Fprintf(deps.Stderr, "Uso: lufy-ai install [--target <dir>] [--scope project|global|both] [--tool %s] [--methodology-tier T3:none] [--dry-run] [--yes] [--backup]\n", writableToolUsage())
	}

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsageErr
	}

	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "install no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}

	scope, err := assets.ParseScope(*scopeValue)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	harness, err := parseHarnessFlags(harnessFlags)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitUsageErr
	}
	service := installer.NewService()
	if err := ensureLayoutForMutation(*target, *dryRun, *yes, deps.Stdout); err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}
	err = service.Run(installer.Options{
		Target:  *target,
		DryRun:  *dryRun,
		Yes:     *yes,
		Backup:  *backup,
		Scope:   scope,
		Harness: harness,
	}, deps.Stdout)
	if err != nil {
		var actionable ActionableError
		if errors.As(err, &actionable) {
			fmt.Fprintln(deps.Stderr, actionable.Error())
			if actionable.Code != 0 {
				return actionable.Code
			}
			return ExitRuntimeErr
		}
		fmt.Fprintln(deps.Stderr, err.Error())
		return ExitRuntimeErr
	}

	return ExitOK
}

func ensureLayoutForMutation(target string, dryRun bool, allowApply bool, stdout io.Writer) error {
	if dryRun {
		resolved, err := platform.ResolveTargetPath(target)
		if err != nil {
			return err
		}
		report, err := layout.BuildPlan(resolved)
		if err != nil {
			return err
		}
		if len(report.Actions) == 0 && len(report.Conflicts) == 0 {
			return nil
		}
		for _, conflict := range report.Conflicts {
			fmt.Fprintf(stdout, "[dry-run] [layout-conflict] %s -> %s: %s\n", conflict.Source, conflict.Target, conflict.Reason)
		}
		for _, action := range report.Actions {
			if action.Source != "" {
				fmt.Fprintf(stdout, "[dry-run] [layout-%s] %s -> %s\n", action.Kind, action.Source, action.Target)
				continue
			}
			fmt.Fprintf(stdout, "[dry-run] [layout-%s] %s\n", action.Kind, action.Target)
		}
		return nil
	}
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return err
	}
	report, err := layout.BuildPlan(resolved)
	if err != nil {
		return err
	}
	if len(report.Conflicts) > 0 {
		return fmt.Errorf("layout .lufy bloqueado por %d conflicto(s); ejecuta lufy-ai migrate-layout --target %s --dry-run", len(report.Conflicts), target)
	}
	if len(report.Actions) == 0 {
		return nil
	}
	if !allowApply {
		for _, action := range report.Actions {
			if action.Kind == "migrate-copy" {
				return fmt.Errorf("layout legacy detectado; vuelve a ejecutar con --yes para migrar primero o revisa con lufy-ai migrate-layout --target %s --dry-run", target)
			}
		}
		return nil
	}
	_, err = layout.Apply(resolved, report.Actions, stdout)
	return err
}

func reportLegacyLayoutForReadOnly(target string, stdout io.Writer) error {
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return err
	}
	report, err := layout.BuildPlan(resolved)
	if err != nil {
		return err
	}
	if len(report.Legacy) == 0 && len(report.Conflicts) == 0 {
		return nil
	}
	if len(report.Legacy) > 0 {
		fmt.Fprintf(stdout, "legacy layout detectado: %s; revisa con lufy-ai migrate-layout --target %s --dry-run\n", joinComma(report.Legacy), target)
	}
	for _, conflict := range report.Conflicts {
		fmt.Fprintf(stdout, "legacy layout conflictivo: %s -> %s: %s\n", conflict.Source, conflict.Target, conflict.Reason)
	}
	return nil
}

func joinComma(values []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for _, value := range values[1:] {
		out += ", " + value
	}
	return out
}

func printGeneralHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai <comando> [flags]")
	fmt.Fprintln(out, "Comandos:")
	fmt.Fprintln(out, "  init      Genera .lufy/config/project.yaml stack-aware")
	fmt.Fprintln(out, "  scan      Escanea stacks/superficies y actualiza project.yaml")
	fmt.Fprintln(out, "  install   Instala/planifica assets (slice inicial)")
	fmt.Fprintln(out, "  uninstall Remueve assets gestionados por Lufy con backup")
	fmt.Fprintln(out, "  verify    Verifica estado mínimo instalado")
	fmt.Fprintln(out, "  backup    Crea backup mínimo")
	fmt.Fprintln(out, "  restore   Restaura desde backup")
	fmt.Fprintln(out, "  merge     Reconcilia .lufy-new con edits locales")
	fmt.Fprintln(out, "  migrate-layout Migra rutas legacy al layout unificado .lufy/")
	fmt.Fprintln(out, "  memory    Inicializa, valida, busca, captura y conecta memoria Obsidian portable")
	fmt.Fprintln(out, "  sync      Sincroniza assets gestionados con manifest/hash/backup")
	fmt.Fprintln(out, "  status    Resume estado instalado y drift local")
	fmt.Fprintln(out, "  info      Muestra catálogo efectivo, manifest, stacks y surfaces")
	fmt.Fprintln(out, "  doctor    Diagnostica project.yaml, manifest y drift sin mutar")
	fmt.Fprintln(out, "  pin       Congela un asset gestionado para preservar edits locales")
	fmt.Fprintln(out, "  unpin     Remueve el freeze de un asset gestionado")
	fmt.Fprintln(out, "  opsx      Utilidades OpenSpec auxiliares")
	fmt.Fprintln(out, "  context   Construye y consulta el grafo de contexto local")
	fmt.Fprintln(out, "  upgrade   Actualiza el binario lufy-ai a una versión fija")
	fmt.Fprintln(out, "  version   Muestra versión, commit, build date y plataforma")
}

func printContextHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai context <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos:")
	fmt.Fprintln(out, "  scan      Inspecciona fuentes soportadas sin persistir grafo")
	fmt.Fprintln(out, "  build     Genera .lufy/context/graph.json, graph-summary.md y manifest.json")
	fmt.Fprintln(out, "  status    Reporta ready, stale o not_available")
	fmt.Fprintln(out, "  query     Busca nodos por término lexical")
	fmt.Fprintln(out, "  path      Calcula un camino explicable entre nodos")
	fmt.Fprintln(out, "  explain   Explica por qué existe un nodo o edge")
	fmt.Fprintln(out, "  diff      Resume impacto desde git diff --base <ref>")
}

func printOpsxHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai opsx <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos:")
	fmt.Fprintln(out, "  render    Renderiza un change OpenSpec a HTML offline")
}

func printMemoryHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai memory <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos:")
	fmt.Fprintln(out, "  init      Crea .lufy/memory y defaults en project.yaml")
	fmt.Fprintln(out, "  status    Resume notas, drafts y backlinks")
	fmt.Fprintln(out, "  validate  Valida schema de notas y backlinks")
	fmt.Fprintln(out, "  search    Busca en knowledge/maps con rg cuando está disponible")
	fmt.Fprintln(out, "  capture   Crea o actualiza una nota durable y sus backlinks")
	fmt.Fprintln(out, "  connect   Conecta dos notas existentes con backlinks seguros")
	fmt.Fprintln(out, "  index     Reconstruye index/backlinks.json desde wikilinks")
}
