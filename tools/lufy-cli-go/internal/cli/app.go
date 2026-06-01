package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/merger"
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
	case "sync":
		return runSync(args[1:], deps)
	case "status":
		return runStatus(args[1:], deps)
	case "upgrade":
		return runUpgrade(args[1:], deps)
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

func runUninstall(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	target := fs.String("target", ".", "Directorio destino")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	keepState := fs.Bool("keep-state", false, "Preservar .lufy-ai/install-state.json")
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
	force := fs.Bool("force", false, "Reemplazar .lufy/project.yaml existente")
	rescan := fs.Bool("rescan", false, "Refrescar evidencia de stacks, preservar overrides y reportar drift sin cleanup destructivo")
	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai init [--target <dir>] [--force] [--rescan]")
		fmt.Fprintln(deps.Stderr, "Genera .lufy/project.yaml con reglas stack-aware editables.")
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
	if err := projectconfig.NewService().Run(projectconfig.Options{Target: *target, Force: *force, Rescan: *rescan}, deps.Stdout); err != nil {
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
	noEngram := fs.Bool("no-engram", false, "Omitir integración Engram")
	harnessFlags := addToolFlag(fs)

	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai sync [--target <dir>] [--scope project|global|both] [--tool opencode] [--dry-run] [--yes] [--no-engram]")
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
	err = service.Run(syncer.Options{Target: *target, DryRun: *dryRun, Yes: *yes, NoEngram: *noEngram, Scope: scope, Harness: harness}, deps.Stdout)
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
	noEngram := fs.Bool("no-engram", false, "Omitir chequeo Engram")
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
	if err := svc.Run(verify.Options{Target: *target, NoEngram: *noEngram, JSON: *jsonOutput, Quiet: *quiet, Verbose: *verbose, Deep: *deep, Scope: scope, ExpectedTool: harness.Tool}, deps.Stdout); err != nil {
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
	noEngram := fs.Bool("no-engram", false, "Omitir integración Engram")
	backup := fs.Bool("backup", false, "Forzar backup cuando aplique")
	harnessFlags := addHarnessFlags(fs)

	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai install [--target <dir>] [--scope project|global|both] [--tool opencode] [--methodology-tier T3:none] [--dry-run] [--yes] [--no-engram] [--backup]")
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
	err = service.Run(installer.Options{
		Target:   *target,
		DryRun:   *dryRun,
		Yes:      *yes,
		NoEngram: *noEngram,
		Backup:   *backup,
		Scope:    scope,
		Harness:  harness,
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

func printGeneralHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai <comando> [flags]")
	fmt.Fprintln(out, "Comandos:")
	fmt.Fprintln(out, "  init      Genera .lufy/project.yaml stack-aware")
	fmt.Fprintln(out, "  install   Instala/planifica assets (slice inicial)")
	fmt.Fprintln(out, "  uninstall Remueve assets gestionados por Lufy con backup")
	fmt.Fprintln(out, "  verify    Verifica estado mínimo instalado")
	fmt.Fprintln(out, "  backup    Crea backup mínimo")
	fmt.Fprintln(out, "  restore   Restaura desde backup")
	fmt.Fprintln(out, "  merge     Reconcilia .lufy-new con edits locales")
	fmt.Fprintln(out, "  sync      Sincroniza assets gestionados con manifest/hash/backup")
	fmt.Fprintln(out, "  status    Resume estado instalado y drift local")
	fmt.Fprintln(out, "  upgrade   Actualiza el binario lufy-ai a una versión fija")
	fmt.Fprintln(out, "  version   Muestra versión, commit, build date y plataforma")
}
