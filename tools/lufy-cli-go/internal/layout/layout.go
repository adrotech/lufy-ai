package layout

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target string
	DryRun bool
	Yes    bool
	JSON   bool
}

type Action struct {
	Kind   string `json:"kind"`
	Source string `json:"source,omitempty"`
	Target string `json:"target"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type Conflict struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Reason string `json:"reason"`
}

type Report struct {
	TargetRoot string     `json:"targetRoot"`
	Actions    []Action   `json:"actions"`
	Conflicts  []Conflict `json:"conflicts"`
	Legacy     []string   `json:"legacyDetected,omitempty"`
	Applied    bool       `json:"applied"`
}

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) Run(opts Options, stdout io.Writer) error {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return err
	}
	if !opts.DryRun {
		lock, err := platform.AcquireLock(target)
		if err != nil {
			return err
		}
		defer lock.Release()
	}
	report, err := BuildPlan(target)
	if err != nil {
		return err
	}
	if !opts.JSON {
		printReport(report, opts.DryRun, stdout)
	}
	if opts.DryRun {
		if opts.JSON {
			body, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			_, _ = stdout.Write(append(body, '\n'))
		}
		return nil
	}
	if len(report.Conflicts) > 0 {
		if opts.JSON {
			body, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			_, _ = stdout.Write(append(body, '\n'))
		}
		return fmt.Errorf("migrate-layout bloqueado por %d conflicto(s); resuelve rutas nuevas/legacy y reintenta", len(report.Conflicts))
	}
	if mutating(report.Actions) && !opts.Yes {
		if opts.JSON {
			body, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			_, _ = stdout.Write(append(body, '\n'))
		}
		return fmt.Errorf("migrate-layout requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan")
	}
	applyOut := stdout
	if opts.JSON {
		applyOut = io.Discard
	}
	applied, err := Apply(target, report.Actions, applyOut)
	if err != nil {
		return err
	}
	report.Applied = applied
	if opts.JSON {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, _ = stdout.Write(append(body, '\n'))
		return nil
	}
	if applied {
		fmt.Fprintln(stdout, "legacy layout migrated; .lufy-ai preserved for rollback")
	}
	return nil
}

func BuildPlan(target string) (Report, error) {
	var report Report
	report.TargetRoot = target
	pairs := []struct {
		legacy                string
		canonical             string
		preferCanonicalIfBoth bool
	}{
		{lufypaths.LegacyProjectConfig, lufypaths.ProjectConfig, true},
		{lufypaths.LegacyInstallState, lufypaths.InstallState, false},
		{lufypaths.LegacyBackups, lufypaths.Backups, false},
		{lufypaths.LegacyAncestors, lufypaths.Ancestors, false},
		{lufypaths.LegacyOpenSpecCache, lufypaths.OpenSpecCache, false},
		{lufypaths.LegacyLufySDD, lufypaths.LufySDD, false},
	}
	for _, pair := range pairs {
		legacyPath, err := platform.SafeJoin(target, pair.legacy)
		if err != nil {
			return report, err
		}
		if !pathExists(legacyPath) {
			continue
		}
		report.Legacy = append(report.Legacy, pair.legacy)
		canonicalPath, err := platform.SafeJoin(target, pair.canonical)
		if err != nil {
			return report, err
		}
		if pathExists(canonicalPath) {
			action, conflict, err := existingPairPlan(legacyPath, canonicalPath)
			if err != nil {
				return report, err
			}
			if action == "stale" {
				report.Actions = append(report.Actions, Action{Kind: "legacy-stale", Source: pair.legacy, Target: pair.canonical, Status: "already-migrated", Reason: "ruta nueva ya existe con el mismo contenido"})
				continue
			}
			if action == "merge" {
				report.Actions = append(report.Actions, Action{Kind: "migrate-copy", Source: pair.legacy, Target: pair.canonical, Status: "planned", Reason: "ruta nueva existe; se copiaran solo entradas legacy no conflictivas"})
				continue
			}
			if pair.preferCanonicalIfBoth {
				report.Actions = append(report.Actions, Action{Kind: "legacy-stale", Source: pair.legacy, Target: pair.canonical, Status: "stale", Reason: "ruta nueva preferida; legacy obsoleto preservado"})
				continue
			}
			report.Conflicts = append(report.Conflicts, Conflict{Source: pair.legacy, Target: pair.canonical, Reason: conflict})
			continue
		}
		report.Actions = append(report.Actions, Action{Kind: "migrate-copy", Source: pair.legacy, Target: pair.canonical, Status: "planned"})
	}
	readmePath, err := platform.SafeJoin(target, lufypaths.Readme)
	if err != nil {
		return report, err
	}
	if !pathExists(readmePath) {
		report.Actions = append(report.Actions, Action{Kind: "write-readme", Target: lufypaths.Readme, Status: "planned", Reason: "documentar layout .lufy"})
	}
	sort.Slice(report.Actions, func(i, j int) bool {
		if report.Actions[i].Target == report.Actions[j].Target {
			return report.Actions[i].Kind < report.Actions[j].Kind
		}
		return report.Actions[i].Target < report.Actions[j].Target
	})
	return report, nil
}

func Apply(target string, actions []Action, stdout io.Writer) (bool, error) {
	applied := false
	var migrated []Action
	for _, action := range actions {
		switch action.Kind {
		case "migrate-copy":
			migrated = append(migrated, action)
		}
	}
	if len(migrated) > 0 {
		if err := writeMigrationBackup(target, migrated); err != nil {
			return false, err
		}
	}
	for _, action := range actions {
		switch action.Kind {
		case "migrate-copy":
			src, err := platform.SafeJoin(target, action.Source)
			if err != nil {
				return applied, err
			}
			dst, err := platform.SafeJoin(target, action.Target)
			if err != nil {
				return applied, err
			}
			if err := copyPath(src, dst); err != nil {
				return applied, err
			}
			fmt.Fprintf(stdout, "- [migrate-layout] %s -> %s\n", action.Source, action.Target)
			applied = true
		case "write-readme":
			dst, err := platform.SafeJoin(target, action.Target)
			if err != nil {
				return applied, err
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return applied, err
			}
			if err := platform.WriteFileAtomic(dst, []byte(ReadmeContent()), 0o644); err != nil {
				return applied, err
			}
			fmt.Fprintf(stdout, "- [migrate-layout] %s\n", action.Target)
			applied = true
		}
	}
	return applied, nil
}

func Ensure(target string, stdout io.Writer) error {
	report, err := BuildPlan(target)
	if err != nil {
		return err
	}
	if len(report.Conflicts) > 0 {
		return fmt.Errorf("layout .lufy bloqueado por %d conflicto(s); ejecuta lufy-ai migrate-layout --target %s --dry-run", len(report.Conflicts), target)
	}
	_, err = Apply(target, report.Actions, stdout)
	return err
}

func ReadmeContent() string {
	return `# .lufy

Workspace local de Lufy para este repositorio.

- ` + "`config/`" + `: configuracion editable del proyecto, stacks, superficies y limites de workflow.
- ` + "`memory/`" + `: memoria Obsidian portable y notas locales del proyecto.
- ` + "`workflows/`" + `: artefactos de metodologias Lufy, como SDD.
- ` + "`managed-state/`" + `: estado interno de instalacion, backups y ancestors para sync/restore.
- ` + "`cache/`" + `: caches locales reutilizables, como OpenSpec.

No borres ` + "`managed-state/`" + ` manualmente salvo que vayas a reinstalar Lufy en este repo.
`
}

func writeMigrationBackup(target string, actions []Action) error {
	root, err := platform.SafeJoin(target, lufypaths.Backups)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, time.Now().UTC().Format("20060102T150405.000000000Z")+"-layout-migration")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	type item struct {
		Source string `json:"source"`
		Backup string `json:"backup"`
		SHA256 string `json:"sha256,omitempty"`
	}
	var items []item
	for _, action := range actions {
		src, err := platform.SafeJoin(target, action.Source)
		if err != nil {
			return err
		}
		backupRel := filepath.Join("legacy", action.Source)
		dst := filepath.Join(dir, backupRel)
		if err := copyPath(src, dst); err != nil {
			return err
		}
		hash := ""
		if info, err := os.Lstat(src); err == nil && info.Mode().IsRegular() {
			hash, _ = assets.FileSHA256(src)
		}
		items = append(items, item{Source: action.Source, Backup: filepath.ToSlash(backupRel), SHA256: hash})
	}
	body, err := json.MarshalIndent(map[string]any{
		"schemaVersion": 1,
		"createdAt":     time.Now().UTC().Format(time.RFC3339),
		"cause":         "layout-migration",
		"items":         items,
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return platform.WriteFileAtomic(filepath.Join(dir, "manifest.json"), body, 0o644)
}

func printReport(report Report, dryRun bool, stdout io.Writer) {
	if len(report.Actions) == 0 && len(report.Conflicts) == 0 {
		fmt.Fprintln(stdout, "Layout .lufy ya está actualizado")
		return
	}
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}
	for _, conflict := range report.Conflicts {
		fmt.Fprintf(stdout, "%s[conflict] %s -> %s: %s\n", prefix, conflict.Source, conflict.Target, conflict.Reason)
	}
	for _, action := range report.Actions {
		if action.Source != "" {
			fmt.Fprintf(stdout, "%s[%s] %s -> %s\n", prefix, action.Kind, action.Source, action.Target)
			continue
		}
		fmt.Fprintf(stdout, "%s[%s] %s\n", prefix, action.Kind, action.Target)
	}
}

func mutating(actions []Action) bool {
	for _, action := range actions {
		if action.Kind == "migrate-copy" || action.Kind == "write-readme" {
			return true
		}
	}
	return false
}

func existingPairPlan(legacyPath, canonicalPath string) (string, string, error) {
	li, err := os.Lstat(legacyPath)
	if err != nil {
		return "", "", err
	}
	ci, err := os.Lstat(canonicalPath)
	if err != nil {
		return "", "", err
	}
	if li.IsDir() || ci.IsDir() {
		if !li.IsDir() || !ci.IsDir() {
			return "", "ruta nueva y legacy tienen tipos distintos", nil
		}
		relation, err := dirOverlayRelation(legacyPath, canonicalPath)
		if err != nil {
			return "", "", err
		}
		switch relation {
		case "same", "canonical-superset":
			return "stale", "", nil
		case "merge-safe":
			return "merge", "", nil
		default:
			return "", "ruta nueva y legacy existen con contenido distinto", nil
		}
	}
	same, err := sameContent(legacyPath, canonicalPath)
	if err != nil {
		return "", "", err
	}
	if same {
		return "stale", "", nil
	}
	return "", "ruta nueva y legacy existen con contenido distinto", nil
}

func sameContent(a, b string) (bool, error) {
	ai, err := os.Lstat(a)
	if err != nil {
		return false, err
	}
	bi, err := os.Lstat(b)
	if err != nil {
		return false, err
	}
	if ai.IsDir() || bi.IsDir() {
		if !ai.IsDir() || !bi.IsDir() {
			return false, nil
		}
		return sameDir(a, b)
	}
	if !ai.Mode().IsRegular() || !bi.Mode().IsRegular() {
		return false, nil
	}
	ah, err := assets.FileSHA256(a)
	if err != nil {
		return false, err
	}
	bh, err := assets.FileSHA256(b)
	if err != nil {
		return false, err
	}
	return ah == bh, nil
}

func dirOverlayRelation(legacyDir, canonicalDir string) (string, error) {
	legacy, err := dirHashes(legacyDir)
	if err != nil {
		return "", err
	}
	canonical, err := dirHashes(canonicalDir)
	if err != nil {
		return "", err
	}
	missing := false
	for rel, legacyHash := range legacy {
		canonicalHash, ok := canonical[rel]
		if !ok {
			missing = true
			continue
		}
		if canonicalHash != legacyHash {
			return "conflict", nil
		}
	}
	if missing {
		return "merge-safe", nil
	}
	if len(canonical) == len(legacy) {
		return "same", nil
	}
	return "canonical-superset", nil
}

func sameDir(a, b string) (bool, error) {
	am, err := dirHashes(a)
	if err != nil {
		return false, err
	}
	bm, err := dirHashes(b)
	if err != nil {
		return false, err
	}
	if len(am) != len(bm) {
		return false, nil
	}
	for rel, ah := range am {
		if bm[rel] != ah {
			return false, nil
		}
	}
	return true, nil
}

func dirHashes(root string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink no soportado: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		hash, err := assets.FileSHA256(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = hash
		return nil
	})
	return out, err
}

func copyPath(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlink no soportado: %s", src)
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("path no es archivo regular ni directorio: %s", src)
	}
	body, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return platform.WriteFileAtomic(dst, body, info.Mode().Perm())
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if strings.Contains(rel, "..") {
			return fmt.Errorf("rel inseguro: %s", rel)
		}
		target := filepath.Join(dst, rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink no soportado: %s", path)
		}
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return platform.WriteFileAtomic(target, body, info.Mode().Perm())
	})
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
