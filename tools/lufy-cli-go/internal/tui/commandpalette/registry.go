package commandpalette

type ParamKind string

const (
	ParamText   ParamKind = "text"
	ParamBool   ParamKind = "bool"
	ParamChoice ParamKind = "choice"
	ParamArg    ParamKind = "arg"
)

type CommandSpec struct {
	ID          string
	Title       string
	Description string
	Args        []string
	Params      []ParamSpec
}

type ParamSpec struct {
	Name        string
	Flag        string
	Kind        ParamKind
	Description string
	Default     string
	Choices     []string
	Required    bool
}

type ParamValue struct {
	Spec  ParamSpec
	Value string
}

func Registry() []CommandSpec {
	commonTarget := ParamSpec{Name: "target", Flag: "--target", Kind: ParamText, Default: ".", Description: "Directorio del proyecto"}
	dryRun := ParamSpec{Name: "dry-run", Flag: "--dry-run", Kind: ParamBool, Description: "Mostrar plan sin escribir"}
	yes := ParamSpec{Name: "yes", Flag: "--yes", Kind: ParamBool, Description: "Confirmar mutaciones"}
	jsonOut := ParamSpec{Name: "json", Flag: "--json", Kind: ParamBool, Description: "Emitir JSON"}
	scope := ParamSpec{Name: "scope", Flag: "--scope", Kind: ParamChoice, Default: "project", Choices: []string{"project", "global", "both"}, Description: "Scope de assets"}
	tool := ParamSpec{Name: "tool", Flag: "--tool", Kind: ParamChoice, Default: "opencode", Choices: []string{"opencode", "codex", "claude-code"}, Description: "Tool adapter"}
	return []CommandSpec{
		{ID: "setup", Title: "Setup", Description: "Configura LUFY end-to-end", Args: []string{"setup"}, Params: []ParamSpec{commonTarget, dryRun, yes, jsonOut, {Name: "skip-version-check", Flag: "--skip-version-check", Kind: ParamBool, Description: "Omitir chequeo de version"}, {Name: "require-latest", Flag: "--require-latest", Kind: ParamBool, Description: "Bloquear si no esta actualizado"}, {Name: "check-new-features", Flag: "--check-new-features", Kind: ParamBool, Description: "Revisar nuevas capacidades"}}},
		{ID: "init", Title: "Init", Description: "Genera project.yaml stack-aware", Args: []string{"init"}, Params: []ParamSpec{commonTarget, {Name: "force", Flag: "--force", Kind: ParamBool, Description: "Reemplazar config existente"}, {Name: "rescan", Flag: "--rescan", Kind: ParamBool, Description: "Preservar overrides y refrescar evidencia"}, {Name: "interactive", Flag: "--interactive", Kind: ParamChoice, Default: "true", Choices: []string{"true", "false"}, Description: "Selector interactivo de superficies"}}},
		{ID: "scan", Title: "Scan", Description: "Escanea stacks y superficies", Args: []string{"scan"}, Params: []ParamSpec{commonTarget, {Name: "interactive", Flag: "--interactive", Kind: ParamChoice, Default: "true", Choices: []string{"true", "false"}, Description: "Selector interactivo de superficies"}}},
		{ID: "install", Title: "Install", Description: "Instala assets gestionados", Args: []string{"install"}, Params: []ParamSpec{commonTarget, scope, tool, dryRun, yes, {Name: "backup", Flag: "--backup", Kind: ParamBool, Description: "Forzar backup"}}},
		{ID: "sync", Title: "Sync", Description: "Sincroniza assets gestionados", Args: []string{"sync"}, Params: []ParamSpec{commonTarget, scope, tool, dryRun, yes}},
		{ID: "verify", Title: "Verify", Description: "Verifica instalacion y drift", Args: []string{"verify"}, Params: []ParamSpec{commonTarget, scope, tool, jsonOut, {Name: "quiet", Flag: "--quiet", Kind: ParamBool, Description: "Omitir salida si no hay errores"}, {Name: "verbose", Flag: "--verbose", Kind: ParamBool, Description: "Detalle adicional"}, {Name: "deep", Flag: "--deep", Kind: ParamBool, Description: "Checks profundos opt-in"}}},
		{ID: "doctor", Title: "Doctor", Description: "Diagnostica config, manifest y drift", Args: []string{"doctor"}, Params: []ParamSpec{commonTarget, scope, jsonOut}},
		{ID: "status", Title: "Status", Description: "Resume estado instalado", Args: []string{"status"}, Params: []ParamSpec{commonTarget, scope, jsonOut, {Name: "verbose", Flag: "--verbose", Kind: ParamBool, Description: "Detalle por asset"}}},
		{ID: "info", Title: "Info", Description: "Muestra catalogo efectivo y metadata", Args: []string{"info"}, Params: []ParamSpec{commonTarget, scope, jsonOut}},
		{ID: "upgrade", Title: "Upgrade", Description: "Actualiza el binario a una version fija", Args: []string{"upgrade"}, Params: []ParamSpec{{Name: "to", Flag: "--to", Kind: ParamText, Required: true, Description: "Version destino vX.Y.Z"}, {Name: "base-url", Flag: "--base-url", Kind: ParamText, Description: "Base URL alternativa"}, dryRun}},
		{ID: "memory-init", Title: "Memory Init", Description: "Inicializa memoria Obsidian", Args: []string{"memory", "init"}, Params: []ParamSpec{commonTarget, jsonOut}},
		{ID: "memory-status", Title: "Memory Status", Description: "Estado de memoria", Args: []string{"memory", "status"}, Params: []ParamSpec{commonTarget, jsonOut}},
		{ID: "memory-validate", Title: "Memory Validate", Description: "Valida memoria y backlinks", Args: []string{"memory", "validate"}, Params: []ParamSpec{commonTarget, jsonOut}},
		{ID: "memory-search", Title: "Memory Search", Description: "Busca notas durables", Args: []string{"memory", "search"}, Params: []ParamSpec{commonTarget, jsonOut, {Name: "query", Kind: ParamArg, Required: true, Description: "Texto a buscar"}}},
		{ID: "context-build", Title: "Context Build", Description: "Construye grafo de contexto", Args: []string{"context", "build"}, Params: []ParamSpec{commonTarget, jsonOut}},
		{ID: "context-status", Title: "Context Status", Description: "Estado del grafo", Args: []string{"context", "status"}, Params: []ParamSpec{commonTarget, jsonOut}},
		{ID: "context-query", Title: "Context Query", Description: "Consulta grafo por termino", Args: []string{"context", "query"}, Params: []ParamSpec{commonTarget, jsonOut, {Name: "term", Kind: ParamArg, Required: true, Description: "Termino de busqueda"}}},
		{ID: "context-diff", Title: "Context Diff", Description: "Impacto desde un ref Git", Args: []string{"context", "diff"}, Params: []ParamSpec{commonTarget, jsonOut, {Name: "base", Flag: "--base", Kind: ParamText, Required: true, Description: "Referencia Git base"}}},
		{ID: "conflicts-plan", Title: "Conflicts Plan", Description: "Planifica conflictos sin mutar", Args: []string{"conflicts", "plan"}, Params: []ParamSpec{commonTarget, scope, tool, jsonOut}},
		{ID: "pr-guard", Title: "PR Guard", Description: "Detecta paths ignorados o internos antes de push/PR", Args: []string{"pr", "guard"}, Params: []ParamSpec{commonTarget, jsonOut, {Name: "base", Flag: "--base", Kind: ParamText, Default: "origin/develop", Description: "Referencia Git base"}, {Name: "include-worktree", Flag: "--include-worktree", Kind: ParamBool, Description: "Incluir cambios locales pendientes"}}},
		{ID: "backup", Title: "Backup", Description: "Crea backup minimo", Args: []string{"backup"}, Params: []ParamSpec{commonTarget}},
		{ID: "restore", Title: "Restore", Description: "Restaura desde backup", Args: []string{"restore"}, Params: []ParamSpec{commonTarget, {Name: "backup", Flag: "--backup", Kind: ParamText, Required: true, Description: "Manifest o directorio de backup"}, dryRun, yes}},
		{ID: "merge", Title: "Merge", Description: "Reconcilia .lufy-new", Args: []string{"merge"}, Params: []ParamSpec{commonTarget, {Name: "accept-theirs", Flag: "--accept-theirs", Kind: ParamBool, Description: "Aceptar version nueva"}, {Name: "accept-ours", Flag: "--accept-ours", Kind: ParamBool, Description: "Preservar local"}, {Name: "path", Kind: ParamArg, Required: true, Description: "Path a reconciliar"}}},
		{ID: "pin", Title: "Pin", Description: "Congela un asset gestionado", Args: []string{"pin"}, Params: []ParamSpec{commonTarget, {Name: "reason", Flag: "--reason", Kind: ParamText, Description: "Motivo"}, {Name: "path", Kind: ParamArg, Required: true, Description: "Asset a congelar"}}},
		{ID: "unpin", Title: "Unpin", Description: "Remueve freeze de asset", Args: []string{"unpin"}, Params: []ParamSpec{commonTarget, {Name: "path", Kind: ParamArg, Required: true, Description: "Asset a descongelar"}}},
		{ID: "uninstall", Title: "Uninstall", Description: "Remueve assets gestionados", Args: []string{"uninstall"}, Params: []ParamSpec{commonTarget, dryRun, yes, {Name: "keep-state", Flag: "--keep-state", Kind: ParamBool, Description: "Preservar estado"}}},
		{ID: "opsx-render", Title: "Opsx Render", Description: "Renderiza change OpenSpec", Args: []string{"opsx", "render"}, Params: []ParamSpec{commonTarget, {Name: "change", Kind: ParamArg, Required: true, Description: "Change ID"}, {Name: "output", Flag: "--output", Kind: ParamText, Description: "Archivo HTML destino"}}},
		{ID: "version", Title: "Version", Description: "Muestra version del binario", Args: []string{"version"}},
	}
}

func InitialValues(spec CommandSpec) []ParamValue {
	values := make([]ParamValue, len(spec.Params))
	for i, param := range spec.Params {
		values[i] = ParamValue{Spec: param, Value: param.Default}
	}
	return values
}

func BuildArgs(spec CommandSpec, values []ParamValue) []string {
	args := append([]string{}, spec.Args...)
	for _, value := range values {
		param := value.Spec
		switch param.Kind {
		case ParamBool:
			if value.Value == "true" {
				args = append(args, param.Flag)
			}
		case ParamChoice:
			if value.Value != "" && value.Value != param.Default {
				if isBoolChoice(param) {
					args = append(args, param.Flag+"="+value.Value)
					continue
				}
				args = append(args, param.Flag, value.Value)
			}
		case ParamText:
			if value.Value != "" && value.Value != param.Default {
				args = append(args, param.Flag, value.Value)
			}
		case ParamArg:
			if value.Value != "" {
				args = append(args, value.Value)
			}
		}
	}
	return args
}

func MissingRequired(values []ParamValue) []string {
	missing := []string{}
	for _, value := range values {
		if value.Spec.Required && value.Value == "" {
			missing = append(missing, value.Spec.Name)
		}
	}
	return missing
}

func isBoolChoice(param ParamSpec) bool {
	if param.Kind != ParamChoice || len(param.Choices) != 2 {
		return false
	}
	return param.Choices[0] == "true" && param.Choices[1] == "false" || param.Choices[0] == "false" && param.Choices[1] == "true"
}
