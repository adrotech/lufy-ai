package lufysdd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

var safeID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	now   func() time.Time
	write func(string, []byte, os.FileMode) error
}

func NewService() Service {
	return Service{now: time.Now, write: atomicWrite}
}

func (s Service) New(target, change, capability string) (Report, error) {
	return s.NewWithMode(target, change, capability, ModeFull)
}

func (s Service) NewWithMode(target, change, capability string, mode Mode) (Report, error) {
	root, err := workflowRoot(target)
	if err != nil {
		return Report{}, err
	}
	if err := validateID("change", change); err != nil {
		return Report{}, err
	}
	if mode == "" {
		mode = ModeFull
	}
	if mode != ModeFull && mode != ModeLite {
		return Report{}, fmt.Errorf("mode inválido %q; usa full o lite", mode)
	}
	if capability == "" {
		capability = change
	}
	if err := validateID("capability", capability); err != nil {
		return Report{}, err
	}
	if err := requireDirectory(root); err != nil {
		return Report{}, fmt.Errorf("workflow Lufy SDD no disponible: %w", err)
	}
	changesRoot, err := platform.SafeJoin(root, "changes")
	if err != nil {
		return Report{}, err
	}
	if err := requireDirectory(changesRoot); err != nil {
		return Report{}, fmt.Errorf("directorio changes inválido: %w", err)
	}
	changeRoot, err := platform.SafeJoin(changesRoot, change)
	if err != nil {
		return Report{}, err
	}
	if err := os.Mkdir(changeRoot, 0o755); err != nil {
		if os.IsExist(err) {
			return Report{}, fmt.Errorf("change %q ya existe; no se sobrescribió ningún artifact", change)
		}
		return Report{}, err
	}
	created := true
	defer func() {
		if !created {
			_ = os.RemoveAll(changeRoot)
		}
	}()

	type scaffoldFile struct {
		rel  string
		body string
	}
	files := []scaffoldFile{
		{rel: "change.yaml", body: metadataScaffold(change, mode)},
		{rel: "proposal.md", body: proposalScaffold(change, mode)},
	}
	if mode == ModeFull {
		files = append(files, scaffoldFile{rel: "design.md", body: designScaffold(change)})
	}
	files = append(files, scaffoldFile{rel: "tasks.md", body: tasksScaffold(change)})
	if mode == ModeFull {
		files = append(files, scaffoldFile{rel: filepath.Join("specs", capability, "spec.md"), body: specScaffold(capability)})
	}
	for _, file := range files {
		rel, body := file.rel, file.body
		path, joinErr := platform.SafeJoin(changeRoot, rel)
		if joinErr != nil {
			created = false
			return Report{}, joinErr
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
			created = false
			return Report{}, mkdirErr
		}
		if writeErr := writeExclusive(path, []byte(body)); writeErr != nil {
			created = false
			return Report{}, writeErr
		}
	}
	if err := s.renderOverview(root, change, mode); err != nil {
		created = false
		return Report{}, err
	}
	return Report{Schema: ReportSchema, Action: "new", Change: change, Mode: mode, Status: "proposed", Root: root, Progress: Progress{}, Diagnostics: []Diagnostic{}}, nil
}

func metadataScaffold(change string, mode Mode) string {
	return fmt.Sprintf("schemaVersion: 1\nid: %s\nmode: %s\nstatus: proposed\nsyncedDigest: \"\"\n", change, mode)
}

func proposalScaffold(change string, mode Mode) string {
	diagramGuidance := "Usa esta sección solo si reduce ambigüedad. En `full`, los diagramas completos viven en `design.md`."
	if mode == ModeLite {
		diagramGuidance = "Opcional para Lite: incluye un diagrama compacto solo si aclara flujo, componentes o datos."
	}
	return fmt.Sprintf(`# Proposal: %[1]s

## LLM Objective

Describe el resultado observable que debe alcanzar el LLM.

- Outcome principal:
- Usuario, sistema o rol beneficiado:
- Señal objetiva de éxito:
- Tradeoff que no debe optimizarse accidentalmente:

## Problem

Describe el problema actual, su impacto y por qué debe resolverse ahora.

## Current Behavior

- Comportamiento actual:
- Archivos, módulos o flujos afectados:
- Evidencia disponible:

## Target Behavior

- Comportamiento esperado:
- Contratos públicos o internos que cambian:
- Estados, errores o bordes esperados:

## Scope

### In Scope

-

### Out of Scope

-

## Constraints

- No cambiar contratos públicos, seguridad, esquema de datos, puertos o defaults salvo autorización explícita en este proposal.
- Preservar trabajo local no relacionado y assets user-owned.
- Mantener cambios mínimos y alineados con patrones existentes.
- Registrar cualquier desviación como riesgo antes de implementar.

## Environment

- Runtime/toolchain esperado:
- Comandos reales de validación:
- Dependencias o servicios externos:
- Variables/configuración necesarias:

## Acceptance Criteria

- **WHEN** ocurre una condición observable
- **THEN** el sistema produce un resultado verificable

## Diagrams

%[2]s

`+"```mermaid"+`
flowchart TD
  A["Input / trigger"] --> B["Change behavior"]
  B --> C["Validated outcome"]
`+"```"+`

## Risks

-

## Open Questions

-
`, change, diagramGuidance)
}

func designScaffold(change string) string {
	return fmt.Sprintf(`# Design: %[1]s

## Context

Resume el sistema existente, dependencias relevantes, ownership de módulos y límites que condicionan el diseño.

## Goals And Non-Goals

### Goals

-

### Non-Goals

-

## Architecture Overview

Describe la arquitectura objetivo y cómo se integra con la estructura actual.

`+"```mermaid"+`
flowchart LR
  User["Actor / caller"] --> Surface["Public surface"]
  Surface --> Service["Application service"]
  Service --> Domain["Domain rules"]
  Service --> Port["Port / adapter"]
`+"```"+`

## Component Model

| Component | Responsibility | Inputs | Outputs | Owner/Boundary |
| --- | --- | --- | --- | --- |
| <component> | <responsibility> | <inputs> | <outputs> | <boundary> |

`+"```mermaid"+`
graph TD
  A["Component A"] --> B["Component B"]
  B --> C["Component C"]
`+"```"+`

## Data And Persistence

Completar si el cambio toca persistencia, archivos, cache, eventos o estructuras durables. Si no aplica, indicar `+"`not_applicable`"+` y justificar.

`+"```mermaid"+`
erDiagram
  ENTITY_A {
    string id
  }
  ENTITY_B {
    string id
  }
  ENTITY_A ||--o{ ENTITY_B : relates_to
`+"```"+`

## Workflow

`+"```mermaid"+`
sequenceDiagram
  participant Caller
  participant Service
  participant Adapter
  Caller->>Service: request
  Service->>Adapter: side effect
  Adapter-->>Service: result
  Service-->>Caller: response
`+"```"+`

## Design Patterns

- Pattern(s) selected:
- Why these patterns fit:
- Alternatives rejected:
- Existing local patterns reused:

## Security, Privacy And Safety

- Auth/authz impact:
- Secret handling:
- User data impact:
- File/system boundary impact:
- Abuse or failure mode:

## Operational Concerns

- Performance:
- Observability:
- Migration/backfill:
- Rollback:
- Compatibility:

## Decisions

### Decision: <title>

- Context:
- Decision:
- Consequences:

## Validation Strategy

- Unit/integration tests:
- Static checks:
- Manual/static review:
- Cross-platform concerns:

## Risks

-
`, change)
}

func tasksScaffold(change string) string {
	return fmt.Sprintf(`# Tasks: %[1]s

- [ ] 1. Analysis and constraints
  - [ ] 1.1 Revisar archivos existentes, dependencias, ownership y patrones locales.
  - [ ] 1.2 Confirmar alcance, non-goals, restricciones y entorno real.
  - [ ] 1.3 Identificar riesgos, decisiones pendientes y necesidad de escalar Lite -> Full.

- [ ] 2. Design alignment
  - [ ] 2.1 Para Full, completar arquitectura, componentes, datos, seguridad, operaciones y diagramas.
  - [ ] 2.2 Para Lite, mantener contexto compacto y agregar diagramas solo si reducen ambigüedad.
  - [ ] 2.3 Confirmar criterios WHEN/THEN observables antes de implementar.

- [ ] 3. Implementation
  - [ ] 3.1 Implementar por bloques coherentes, sin mezclar cambios no relacionados.
  - [ ] 3.2 Preservar contratos públicos, seguridad, datos y defaults salvo autorización explícita.
  - [ ] 3.3 Actualizar tests/docs ligados al cambio.

- [ ] 4. Validation
  - [ ] 4.1 Ejecutar validación proporcional real.
  - [ ] 4.2 Registrar comandos, resultados y limitaciones bajo `+"`verification/%[1]s/`"+`.
  - [ ] 4.3 Verificar que `+"`change-overview.html`"+` fue refrescado por el lifecycle.

- [ ] 5. Sync and delivery readiness
  - [ ] 5.1 Para Full, ejecutar sync cuando los deltas estén validados.
  - [ ] 5.2 Revisar estado de gates: implemented, validated, delivery_pending, delivered o closed.
  - [ ] 5.3 No archivar ni reportar cierre sin evidencia de validación, sync y delivery cuando aplique.
`, change)
}

func specScaffold(capability string) string {
	return fmt.Sprintf(`# %[1]s Specification Delta

## Intent

Define los cambios observables del capability. Cada requirement debe ser testeable o verificable por revisión estática concreta.

Usa `+"`## ADDED Requirements`"+`, `+"`## MODIFIED Requirements`"+` o `+"`## REMOVED Requirements`"+` según corresponda. No mantengas secciones vacías.

## ADDED Requirements

### Requirement: Describe the required behavior

El sistema SHALL describir un comportamiento verificable, con inputs, outputs, límites y errores esperados cuando aplique.

#### Scenario: Describe the expected outcome

- **GIVEN** un estado inicial relevante
- **WHEN** ocurre una condición concreta
- **THEN** ocurre un resultado observable
`, capability)
}

func (s Service) Validate(target, change string, strict bool) (Report, error) {
	report, err := s.Check(target, change, strict)
	if err != nil {
		return Report{}, err
	}
	if err := s.renderOverview(report.Root, change, report.Mode); err != nil {
		return Report{}, err
	}
	return report, nil
}

// Check validates a change without materializing derived artifacts. It is used
// by read-only installation verification; the sdd validate command uses
// Validate so normal authoring still refreshes the integrated overview.
func (Service) Check(target, change string, strict bool) (Report, error) {
	report, err := inspect(target, change, "validate", strict)
	if err != nil {
		return Report{}, err
	}
	if report.Valid() {
		report.Status = "valid"
	} else {
		report.Status = "invalid"
	}
	return report, nil
}

func (Service) Status(target, change string) (Report, error) {
	report, err := inspect(target, change, "status", false)
	if err != nil {
		return Report{}, err
	}
	switch {
	case !report.Valid():
		report.Status = "blocked"
	case report.Progress.Total == 0:
		report.Status = "proposed"
	case report.Progress.Complete < report.Progress.Total:
		report.Status = "in_progress"
	default:
		changeRoot, joinErr := platform.SafeJoin(report.Root, filepath.Join("changes", change))
		if joinErr != nil {
			return Report{}, joinErr
		}
		if report.Mode == ModeFull {
			metadataPath, joinErr := platform.SafeJoin(changeRoot, "change.yaml")
			if joinErr != nil {
				return Report{}, joinErr
			}
			metadata, metadataErr := readMetadata(metadataPath)
			if metadataErr != nil || metadata.SyncedDigest == "" || metadata.SyncedDigest != report.DeltaDigest {
				report.Status = "sync_pending"
				break
			}
		}
		evidence, diagnostics, evidenceErr := verificationEvidence(report.Root, change)
		if evidenceErr != nil {
			return Report{}, evidenceErr
		}
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
		if evidence && report.Valid() {
			report.Status = "completed"
		} else {
			report.Status = "verification"
		}
	}
	return report, nil
}

func inspect(target, change, action string, strict bool) (Report, error) {
	root, err := workflowRoot(target)
	if err != nil {
		return Report{}, err
	}
	if err := validateID("change", change); err != nil {
		return Report{}, err
	}
	changeRoot, err := platform.SafeJoin(root, filepath.Join("changes", change))
	if err != nil {
		return Report{}, err
	}
	if err := requireDirectory(changeRoot); err != nil {
		return Report{}, fmt.Errorf("change %q no disponible: %w", change, err)
	}
	report := Report{Schema: ReportSchema, Action: action, Change: change, Mode: ModeFull, Status: "unknown", Root: root, Diagnostics: []Diagnostic{}}
	metadataBody, metadataReadErr := readRegular(changeRoot, "change.yaml")
	if metadataReadErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: metadataReadErr.Error()})
	} else if strings.TrimSpace(string(metadataBody)) == "" {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "empty_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "artifact vacío"})
	} else {
		metadataPath, metadataPathErr := platform.SafeJoin(changeRoot, "change.yaml")
		if metadataPathErr != nil {
			return Report{}, metadataPathErr
		}
		metadata, metadataErr := readMetadata(metadataPath)
		if metadataErr != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "invalid_metadata", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: metadataErr.Error()})
		} else {
			report.Mode = metadata.Mode
			if metadata.ID != change {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "metadata_id_mismatch", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "metadata.id no coincide con el directorio del change"})
			}
		}
	}
	required := []string{"proposal.md", "tasks.md"}
	if report.Mode == ModeFull {
		required = []string{"proposal.md", "design.md", "tasks.md"}
	}
	for _, rel := range required {
		body, readErr := readRegular(changeRoot, rel)
		if readErr != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: readErr.Error()})
			continue
		}
		if strings.TrimSpace(string(body)) == "" {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "empty_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: "artifact vacío"})
		}
		if rel == "tasks.md" {
			report.Progress = TaskProgress(string(body))
			if strict && report.Progress.Total == 0 {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_tasks", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: "tasks.md no contiene checkboxes"})
			}
		}
	}
	if report.Mode == ModeLite {
		sortDiagnostics(report.Diagnostics)
		return report, nil
	}

	specRoot, joinErr := platform.SafeJoin(changeRoot, "specs")
	if joinErr != nil {
		return Report{}, joinErr
	}
	if directoryErr := requireDirectory(specRoot); directoryErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: directoryErr.Error()})
		sortDiagnostics(report.Diagnostics)
		return report, nil
	}
	entries, readErr := os.ReadDir(specRoot)
	if readErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: readErr.Error()})
	} else {
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		found := 0
		digest := sha256.New()
		for _, entry := range entries {
			capability := entry.Name()
			if !entry.IsDir() {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "invalid_spec_entry", Path: filepath.ToSlash(filepath.Join("changes", change, "specs", capability)), Message: "se esperaba un directorio de capability"})
				continue
			}
			if err := validateID("capability", capability); err != nil {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_capability", Path: filepath.ToSlash(filepath.Join("changes", change, "specs", capability)), Message: err.Error()})
				continue
			}
			rel := filepath.Join("specs", capability, "spec.md")
			body, specErr := readRegular(changeRoot, rel)
			path := filepath.ToSlash(filepath.Join("changes", change, rel))
			if specErr != nil {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_spec", Path: path, Message: specErr.Error()})
				continue
			}
			found++
			_, _ = digest.Write([]byte(capability))
			_, _ = digest.Write([]byte{0})
			_, _ = digest.Write(body)
			_, _ = digest.Write([]byte{0})
			_, diagnostics := ParseDelta(path, capability, string(body))
			report.Diagnostics = append(report.Diagnostics, diagnostics...)
		}
		if found > 0 {
			report.DeltaDigest = hex.EncodeToString(digest.Sum(nil))
		}
		if strict && found == 0 {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: "el mode full requiere al menos una spec delta"})
		}
	}
	sortDiagnostics(report.Diagnostics)
	return report, nil
}

func workflowRoot(target string) (string, error) {
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return "", err
	}
	return platform.SafeJoin(resolved, lufypaths.LufySDD)
}

func validateID(kind, value string) error {
	if !safeID.MatchString(value) {
		return fmt.Errorf("%s ID inválido %q; usa kebab-case seguro", kind, value)
	}
	return nil
}

func requireDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("%s no es un directorio regular", path)
	}
	return nil
}

func readRegular(root, rel string) ([]byte, error) {
	path, err := platform.SafeJoin(root, rel)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s no es un archivo regular", path)
	}
	return os.ReadFile(path)
}

func writeExclusive(path string, body []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(body); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
