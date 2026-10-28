package agentsref

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const (
	AgentsFile        = "AGENTS.md"
	HarnessFile       = "lufy-ia.harness.md"
	RetiredReference  = "@lufy-ia.harness.md"
	BeginMarker       = "<!-- LUFY:BEGIN codex-harness -->"
	EndMarker         = "<!-- LUFY:END codex-harness -->"
	IntegrationPolicy = "bloque gestionado LUFY en AGENTS.md"
)

func MinimalContent() []byte {
	return []byte("# AGENTS.md\n\n" + ManagedBlock())
}

func RecommendedInstallAction() string {
	return "ejecuta `lufy-ai install --target <target> --yes` para crear/agregar el bloque gestionado LUFY en AGENTS.md, o edita AGENTS.md manualmente"
}

func ContainsReference(body []byte) bool {
	return ContainsManagedBlock(body)
}

func ContainsManagedBlock(body []byte) bool {
	return bytes.Contains(body, []byte(BeginMarker)) && bytes.Contains(body, []byte(EndMarker))
}

func ManagedBlock() string {
	return `<!-- LUFY:BEGIN codex-harness -->
# Lufy AI Harness

- Responde en español y trabaja como agente principal inline.
- Lee execution en .lufy/config/project.yaml y aplica .lufy/contracts/work.md.
- Entiende el resultado pedido y el contexto existente; actúa dentro del alcance autorizado hasta comprobarlo.
- Adapta la evidencia a análisis, propuestas, escritura, código, datos u operaciones. No exijas tiers, metodología ni handoffs para avanzar.
- Delega según execution.delegation: off, on_request o auto. Una instrucción inline del usuario prevalece; delegar nunca amplía permisos.
- Usa capacidades y skills locales pertinentes, con contexto mínimo, sin imponer una cadena de roles.
- Consulta sólo decisiones materiales que no puedas resolver; continúa el trabajo independiente mientras esperas.
- Preserva trabajo ajeno, comprueba el resultado y reporta evidencia real y límites pendientes.
- Commit, push, PR, publicación y operaciones externas requieren autorización explícita; aplica .lufy/contracts/delivery.md cuando corresponda.
<!-- LUFY:END codex-harness -->
`
}

func Status(targetRoot string) (exists bool, hasReference bool, err error) {
	path, err := platform.SafeJoin(targetRoot, AgentsFile)
	if err != nil {
		return false, false, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return true, false, fmt.Errorf("%s no es archivo regular seguro", AgentsFile)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return true, false, err
	}
	return true, bytes.Contains(body, []byte(ManagedBlock())), nil
}

func InsertReference(targetRoot string) error {
	path, err := platform.SafeJoin(targetRoot, AgentsFile)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return writeAgentsFile(path, MinimalContent())
	}
	if err != nil {
		return err
	}
	if bytes.Contains(body, []byte(ManagedBlock())) {
		return nil
	}
	if ContainsManagedBlock(body) || bytes.Contains(body, []byte(RetiredReference)) {
		return fmt.Errorf("bloque LUFY no vigente en AGENTS.md; reemplázalo por la política inline antes de instalar")
	}
	updated := appendReference(body)
	return writeAgentsFile(path, updated)
}

func RemoveReference(targetRoot string) (bool, error) {
	path, err := platform.SafeJoin(targetRoot, AgentsFile)
	if err != nil {
		return false, err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !ContainsReference(body) {
		return false, nil
	}
	updated := removeReferenceLines(body)
	if bytes.Equal(body, updated) {
		return false, nil
	}
	return true, writeAgentsFile(path, updated)
}

func appendReference(body []byte) []byte {
	text := string(body)
	if text == "" {
		return MinimalContent()
	}
	var b strings.Builder
	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") {
		b.WriteString("\n")
	}
	if !strings.HasSuffix(b.String(), "\n\n") {
		b.WriteString("\n")
	}
	b.WriteString(ManagedBlock())
	return []byte(b.String())
}

func removeReferenceLines(body []byte) []byte {
	text := removeManagedBlocks(string(body))

	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return nil
	}
	return []byte(text + "\n")
}

func removeManagedBlocks(text string) string {
	for {
		start := strings.Index(text, BeginMarker)
		if start < 0 {
			return text
		}
		end := strings.Index(text[start:], EndMarker)
		if end < 0 {
			return text
		}
		end += start + len(EndMarker)
		for end < len(text) && (text[end] == '\r' || text[end] == '\n') {
			end++
		}
		text = text[:start] + text[end:]
	}
}

func writeAgentsFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destino symlink no permitido: %s", path)
	}
	return platform.WriteFileAtomic(path, body, 0o644)
}
