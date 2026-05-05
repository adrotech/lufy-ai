package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/syncer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

func Run(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printGeneralHelp(deps.Stderr)
		return ExitUsageErr
	}

	switch args[0] {
	case "install":
		return runInstall(args[1:], deps)
	case "verify":
		return runVerify(args[1:], deps)
	case "backup":
		return runBackup(args[1:], deps)
	case "restore":
		return runRestore(args[1:], deps)
	case "sync":
		return runSync(args[1:], deps)
	case "-h", "--help", "help":
		printGeneralHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Comando desconocido: %s\n\n", args[0])
		printGeneralHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runSync(args []string, deps Dependencies) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)

	target := fs.String("target", ".", "Directorio destino")
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	noEngram := fs.Bool("no-engram", false, "Omitir integración Engram")

	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai sync [--target <dir>] [--dry-run] [--yes] [--no-engram]")
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

	service := syncer.NewService()
	err := service.Run(syncer.Options{Target: *target, DryRun: *dryRun, Yes: *yes, NoEngram: *noEngram}, deps.Stdout)
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
	noEngram := fs.Bool("no-engram", false, "Omitir chequeo Engram")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
	}
	svc := verify.NewService()
	if err := svc.Run(verify.Options{Target: *target, NoEngram: *noEngram}, deps.Stdout); err != nil {
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
	dryRun := fs.Bool("dry-run", false, "Mostrar restore sin mutaciones")
	yes := fs.Bool("yes", false, "Acepta confirmaciones")
	if err := fs.Parse(args); err != nil {
		return ExitUsageErr
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
	dryRun := fs.Bool("dry-run", false, "Mostrar plan sin mutaciones")
	yes := fs.Bool("yes", false, "Aceptar confirmaciones")
	noEngram := fs.Bool("no-engram", false, "Omitir integración Engram")
	backup := fs.Bool("backup", false, "Forzar backup cuando aplique")

	fs.Usage = func() {
		fmt.Fprintln(deps.Stderr, "Uso: lufy-ai install [--target <dir>] [--dry-run] [--yes] [--no-engram] [--backup]")
	}

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		return ExitUsageErr
	}

	if len(fs.Args()) > 0 {
		fmt.Fprintln(deps.Stderr, "install no acepta argumentos posicionales")
		fs.Usage()
		return ExitUsageErr
	}

	service := installer.NewService()
	err := service.Run(installer.Options{
		Target:   *target,
		DryRun:   *dryRun,
		Yes:      *yes,
		NoEngram: *noEngram,
		Backup:   *backup,
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
	fmt.Fprintln(out, "  install   Instala/planifica assets (slice inicial)")
	fmt.Fprintln(out, "  verify    Verifica estado mínimo instalado")
	fmt.Fprintln(out, "  backup    Crea backup mínimo")
	fmt.Fprintln(out, "  restore   Restaura desde backup")
	fmt.Fprintln(out, "  sync      Sincroniza assets gestionados con manifest/hash/backup")
}
