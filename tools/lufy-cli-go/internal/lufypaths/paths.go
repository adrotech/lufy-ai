package lufypaths

import "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"

const (
	Root          = ".lufy"
	Readme        = ".lufy/README.md"
	ProjectConfig = ".lufy/config/project.yaml"
	InstallState  = ".lufy/managed-state/install-state.json"
	Backups       = ".lufy/managed-state/backups"
	Ancestors     = ".lufy/managed-state/ancestors"
	SkillRegistry = ".lufy/skill-registry.json"
	// Estas rutas sólo permiten diagnosticar y rechazar formatos retirados.
	RetiredProjectConfig = ".lufy/project.yaml"
	RetiredInstallState  = ".lufy-ai/install-state.json"
	RetiredAncestors     = ".lufy-ai/ancestors"
)

func WritePath(targetRoot, rel string) (string, error) { return platform.SafeJoin(targetRoot, rel) }
func ResolveBackupReference(targetRoot, backupID string) (string, error) {
	return platform.SafeJoin(targetRoot, Backups+"/"+backupID)
}
