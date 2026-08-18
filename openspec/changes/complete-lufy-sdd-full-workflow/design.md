# Design: Lufy SDD Full nativo

## Context

El adapter actual solo renderiza `README.md`, `changes/`, `decisions/`, `verification/` y, en mode full, `specs/`. La CLI filtra correctamente esos assets por metodología, pero el lifecycle sigue dependiendo de instrucciones OpenSpec. El diseño debe mantener la arquitectura hexagonal: el dominio de metodología define capacidades, el motor Lufy SDD implementa filesystem y parsing, la CLI adapta argumentos/salidas y los tool adapters solo exponen instrucciones.

## Goals

- Ofrecer un lifecycle T1 completo sin binario o cache OpenSpec.
- Mantener artifacts Markdown legibles y editables a mano.
- Hacer que validación, sync y archive fallen antes de mutar cuando el input es ambiguo.
- Conservar OpenSpec como preset default y permitir convivencia explícita.
- Compartir el mismo contrato semántico entre CLI, comandos y skills.

## Decisions

### 1. Layout canónico y versionado

```text
.lufy/workflows/sdd/
  config.yaml
  README.md
  changes/
    <change>/
      change.yaml
      proposal.md
      design.md
      tasks.md
      specs/<capability>/spec.md
      change-overview.html
  specs/<capability>/spec.md
  decisions/
  verification/<change>/
  archive/YYYY-MM-DD-<change>/
```

`config.yaml` es gestionado. El contenido de cada change, las specs activas, decisiones, evidencia y la vista HTML derivada son user-owned aunque vivan debajo de una raíz instalada por Lufy. El catálogo solo gestiona archivos bootstrap y `.gitkeep`; nunca registra artifacts creados por el usuario como assets reemplazables.

### 2. Contrato Markdown compatible por intención

Los specs delta usan exactamente:

- `## ADDED Requirements`
- `## MODIFIED Requirements`
- `## REMOVED Requirements`
- `### Requirement: <título>`
- `#### Scenario: <título>` con cláusulas `WHEN` y `THEN`; `GIVEN` es opcional.

La compatibilidad sintáctica reduce fricción de migración, pero el motor es propio. No se leen schemas internos ni metadata privada de OpenSpec.

### 3. Namespace CLI nativo

```text
lufy-ai sdd new --change <name> [--mode full|lite] [--capability <name>] [--target <dir>] [--json]
lufy-ai sdd status [--change <name>] [--target <dir>] [--json]
lufy-ai sdd validate --change <name> [--strict] [--target <dir>] [--json]
lufy-ai sdd sync --change <name> [--target <dir>] [--json]
lufy-ai sdd archive --change <name> [--target <dir>] [--json]
```

`new` crea scaffolding y falla si el change ya existe. Full conserva proposal, design, tasks y specs delta; Lite usa proposal y tasks como change acotado y no exige specs. `status` deriva el avance desde artifacts y checkboxes. `validate` no muta fuentes semánticas ni specs, pero materializa la vista HTML derivada. `sync` y `archive` son mutantes y no aceptan paths arbitrarios: solo IDs kebab-case resueltos mediante safe join.

### 4. Modelo de estado derivado con metadata mínima

`change.yaml` contiene schema version, ID, mode y estado declarado. Cuando metadata antigua no declara mode se interpreta como Full. El estado efectivo también considera artifacts y tareas:

```text
proposed -> in_progress -> verification -> completed -> archived
```

Una lista de tareas completa no equivale por sí sola a cierre. `sync` registra en `change.yaml` el digest de los deltas aplicados. Si el delta cambia después, archive vuelve a bloquear hasta un nuevo sync. La evidencia externa de delivery sigue siendo responsabilidad del harness/delivery; el CLI no inventa estado GitHub.

### 5. Validación semántica determinística

El parser trabaja por headings Markdown y produce diagnósticos ordenados por path, línea y código. En strict mode son errores:

- artifact requerido faltante o vacío;
- marker delta desconocido o ausente;
- requirement sin scenario;
- scenario sin `WHEN` o `THEN`;
- requirement duplicado dentro del change;
- capability o change ID inseguro;
- archivo, directorio intermedio o destino relevante que sea symlink/no regular.

La salida JSON usa un schema versionado `lufy-sdd-report/v1` común a todas las acciones.

### 6. Sync por plan y preflight

El motor carga todos los deltas y todas las specs destino antes de escribir. `ADDED` exige que el título no exista; `MODIFIED` y `REMOVED` exigen una coincidencia única. Cualquier ambigüedad aborta el plan completo.

Las escrituras usan temporales en el mismo directorio, backup de destinos existentes y rename. Si una escritura falla, se restauran los destinos ya tocados. El resultado registra acciones `create`, `modify` y `remove` de forma estable.

### 7. Archive con gates locales

Archive requiere:

- artifacts válidos en strict mode;
- todas las tareas marcadas;
- digest sincronizado igual al delta actual;
- directorio de evidencia `verification/<change>/` con al menos un archivo regular.

El move a `archive/YYYY-MM-DD-<change>` falla si el destino existe. No sobrescribe archives ni sigue symlinks.

### 8. Superficies de agentes desacopladas

Las acciones canónicas se documentan bajo `.lufy/workflows/sdd/actions/`. OpenCode y Codex reciben wrappers mínimos que invocan el mismo contrato y reportan `methodology_id: lufy-sdd`. El catálogo marca estos wrappers como `methodology-skill` o `methodology-command` de Lufy SDD para que no aparezcan cuando solo se selecciona OpenSpec.

### 9. Vista HTML derivada e integrada

Cada change contiene `change-overview.html`, una vista autocontenida y determinística de sus artifacts Markdown. Full reúne proposal, design, tasks y specs delta ordenadas; Lite reúne proposal y tasks. El HTML no contiene recursos remotos, scripts ni estado autoritativo: los Markdown y `change.yaml` siguen siendo la fuente de verdad.

`new` materializa la vista inicial. `validate` la regenera después de que el agente edita los Markdown; `sync` y `archive` la refrescan antes de completar su transición. `status` permanece sin escrituras. La generación es atómica, rechaza symlinks y no expone un comando o skill público separado.

Lite no sincroniza specs: `sync` responde `not_applicable` sin mutaciones. Archive Lite exige validación strict, tasks completas y evidencia, pero no un digest de deltas inexistentes.

### 10. Routing del harness consciente de metodología

T1/T2/T3 continúa definiendo la profundidad del trabajo; la selección instalada por tier define el adapter concreto. T1 + `lufy-sdd/full` usa el lifecycle Lufy Full, T2 + `lufy-sdd/lite` usa Lufy Lite y las selecciones OpenSpec conservan su propio lifecycle. El router no sustituye una metodología configurada por OpenSpec ni interpreta `execution_mode` como identidad del adapter.

El orchestrator carga skills concretos según `methodology_id`, verifica artifacts por mode y propaga un contrato explícito del overview. Para Lufy SDD el overview es `automatic`: nunca pregunta si debe generarlo y reporta el path materializado/refrescado. Implementer trabaja desde Markdown fuente; validator ejecuta strict validation y comprueba el overview; reviewer lo trata como vista derivada; delivery preserva los artifacts user-owned.

El PR guard mantiene `.lufy/` como superficie interna por defecto, pero excluye del falso positivo las raíces declaradas user-owned de Lufy SDD: `changes/`, `specs/`, `decisions/`, `verification/` y `archive/`. Un `.gitignore` explícito todavía bloquea y requiere decisión humana.

## Alternatives Considered

### Envolver el CLI OpenSpec

Rechazado: mantendría la dependencia que esta iniciativa busca eliminar y duplicaría el resolver stay-updated.

### Inventar un formato YAML completo para specs

Rechazado: empeora la edición humana y la migración. YAML queda limitado a metadata/config; requirements y scenarios continúan en Markdown.

### Implementar solo skills sin motor CLI

Rechazado: repetiría el problema actual. Sin comandos determinísticos no hay validación, sync o archive verificables fuera de una sesión de agente.

## Risks and Mitigations

- **Parser Markdown demasiado permisivo:** gramática deliberadamente pequeña, diagnósticos con línea y fixtures negativos.
- **Sync destructivo:** preflight global, backup, rename y rollback; nunca aproximar títulos.
- **Explosión de assets duplicados:** acciones canónicas tool-neutral y wrappers delgados.
- **Confusión OpenSpec/Lufy SDD:** ownership explícito por metodología y docs de coexistencia.
- **Cambio amplio con toolchain ausente:** tareas separadas por bloques y validación Go pendiente explícita hasta disponer de Go 1.24.2.

## Migration Plan

1. Instalar o sincronizar con `T1:lufy-sdd/full`.
2. Crear nuevos changes con `lufy-ai sdd new`; los changes OpenSpec existentes continúan su lifecycle original.
3. Para migración manual, copiar contenido semántico a un nuevo change Lufy SDD y ejecutar `validate --strict` antes de sync.
4. Mantener `openspec/` intacto hasta que el usuario decida retirarlo fuera del catálogo gestionado.

## Open Questions Deferred

- Importador automatizado OpenSpec → Lufy SDD.
- Baseline remoto/stay-updated propio.
- Personalización visual o exportación a otros formatos del overview derivado.
