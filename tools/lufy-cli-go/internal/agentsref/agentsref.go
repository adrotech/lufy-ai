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
	Reference         = "@lufy-ia.harness.md"
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
	return bytes.Contains(body, []byte(Reference)) || ContainsManagedBlock(body)
}

func ContainsManagedBlock(body []byte) bool {
	return bytes.Contains(body, []byte(BeginMarker)) && bytes.Contains(body, []byte(EndMarker))
}

func ManagedBlock() string {
	return BeginMarker + `
# Lufy AI Harness

- Responde en español para comunicación humana.
- Trata este repo como un proyecto gobernado por Lufy.
- Antes de cambios no triviales, clasifica el trabajo como T1/T2/T3.
- T1: arquitectura, contratos públicos, seguridad, cambios transversales o alta incertidumbre; usar workflow SDD completo.
- T2: cambio funcional acotado, bug relevante, agente/skill o refactor controlado; usar SDD Lite o handoff estructurado.
- T3: cambio trivial, mecánico, documental o local; permite ejecución directa con validación proporcional.
- Usa skills repo-locales en .agents/skills cuando apliquen.
- Si existe '.codex/lufy-agent-mapping.md', declara 'agent_execution_mode' como 'native', 'emulated' o 'inline' antes de delegar roles Lufy.
- En Codex, '@orchestrator' o '@<rol-lufy>' es una solicitud de delegacion: usa subagent tooling con spawn/wait/close cuando este disponible; no respondas como ese rol en el mismo hilo para luego seguir ejecutando.
- Si Codex no expone subagent tooling para una delegacion solicitada o requerida, reporta el bloqueo y no conviertas silenciosamente la solicitud en ejecucion inline.
- No afirmes haber usado subagentes Lufy nativos si el runtime solo expone roles genericos como 'default', 'explorer' o 'worker'.
- No hagas commit, push, PR ni delivery sin autorización explícita.
- Reporta siempre comandos de validación reales y resultados reales.
- No asumas tooling no detectado en el repo.
- Preserva trabajo local no relacionado.
` + EndMarker + "\n"
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
	return true, ContainsReference(body), nil
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
	if ContainsReference(body) {
		return nil
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
	lines := strings.Split(text, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == Reference {
			continue
		}
		filtered = append(filtered, line)
	}
	text = strings.Join(filtered, "\n")
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
