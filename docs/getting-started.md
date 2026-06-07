# Primeros pasos con lufy-ai

Esta guía asume que ya tienes el binario instalado. Para instalarlo, ver [`docs/installation.md`](installation.md).

## Modelo mental

`lufy-ai` instala un harness en un repositorio existente. El harness coordina agentes, specs, skills, validación y delivery sin acoplar el valor de Lufy a una única tool.

Hoy el flujo productivo es:

```text
lufy-ai core -> opencode adapter -> openspec/lufy-sdd/none por tier
```

La separación futura es:

```text
lufy-ai core -> opencode | codex | claude-code | otros adapters
```

Por ahora solo `opencode` escribe archivos. `codex` y `claude-code` son previews dry-run.

## 1. Revisar e instalar

```bash
lufy-ai version
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --dry-run --yes --no-engram
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --yes --no-engram
lufy-ai verify --target /ruta/a/tu/proyecto --tool opencode --no-engram
```

Después de instalar:

1. Revisa que `AGENTS.md` conserve tus convenciones locales.
2. Confirma que contiene `@lufy-ia.harness.md`.
3. Reinicia OpenCode para cargar agentes, comandos, skills y plugin.
4. Ejecuta `lufy-ai status --target /ruta/a/tu/proyecto --verbose` si quieres ver assets y drift.

## 2. Inicializar configuración stack-aware y surface-aware

`init` crea `.lufy/project.yaml`, que es configuración del proyecto destino. No es asset gestionado por hash. Además de stacks técnicos, incluye `project_profile.surfaces` para declarar si el proyecto o una raíz se razona como `frontend`, `backend`, `fullstack`, `mobile`, `cli`, `infra` o `library`.

```bash
lufy-ai init --target /ruta/a/tu/proyecto
```

Para revisar o ajustar la mentalidad de los agentes durante el scan, usa el modo interactivo. En una terminal real abre una TUI Bubble Tea/Charm para revisar superficies detectadas, cambiar su tipo y confirmar el `agent_lens`; en CI, pipes o salidas no TTY conserva la detección automática sin bloquear.

```bash
lufy-ai init --target /ruta/a/tu/proyecto --interactive
lufy-ai scan --target /ruta/a/tu/proyecto
```

Si el proyecto cambia de stack:

```bash
lufy-ai init --target /ruta/a/tu/proyecto --rescan
```

`--rescan` preserva overrides manuales como thresholds, anti-patterns, `project_profile.surfaces` y `workflow_limits`.

Ejemplo frontend con `pnpm`: el proyecto puede declarar comandos de validación que `implementer` hereda sin hardcodearlos en el agente.

```yaml
validation:
  allowed_commands:
    implementer:
      - pnpm typecheck*
      - pnpm lint*
      - pnpm test*
      - pnpm build*
```

## 3. Elegir metodología por tier

Default actual: OpenCode + OpenSpec.

Para dejar T3 sin metodología:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --methodology-tier T3:none --yes --no-engram
```

Para usar Lufy SDD Lite en T2:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --methodology-tier T2:lufy-sdd/lite --yes --no-engram
```

Para combinar OpenSpec Lite en T2 y T3 Express sin spec:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --methodology-tier T2:openspec/lite --methodology-tier T3:none --yes --no-engram
```

Los comandos mutantes bloquean combinaciones inseguras como `T1:none`, `T2:none`, `--tool codex` y `--tool claude-code`.

## 4. Usar el harness

### T3 Express

Usa T3 para cambios triviales, mecánicos, documentales o locales. El agente puede implementar directamente con validación proporcional.

Ejemplos:

- corregir typos;
- actualizar una línea de docs;
- ajustar un texto de ayuda;
- cambios de config sin riesgo transversal.

### T2 SDD Lite

Usa T2 cuando hay comportamiento observable pero el cambio es acotado.

Artefactos esperados:

- mini-spec o handoff corto;
- criterios `WHEN`/`THEN`;
- validación agrupada;
- Result Contract compacto.

Template:

```text
.opencode/templates/sdd-lite.md
```

### T1 Full SDD

Usa T1 para arquitectura, contratos públicos, seguridad, delivery policy o alta incertidumbre.

Comandos:

| Paso | Comando |
| --- | --- |
| Explorar | `/opsx-explore` |
| Proponer | `/opsx-propose` |
| Implementar | `/opsx-apply` |
| Verificar | `/opsx-verify` |
| Sincronizar specs | `/opsx-sync` |
| Archivar | `/opsx-archive` |
| Diagnóstico versión | `/opsx-version` |

## 5. Lifecycle de mantenimiento

### Ver estado

```bash
lufy-ai status --target /ruta/a/tu/proyecto
lufy-ai status --target /ruta/a/tu/proyecto --json --verbose
```

### Sincronizar assets de Lufy

```bash
lufy-ai sync --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
lufy-ai sync --target /ruta/a/tu/proyecto --yes --no-engram
lufy-ai verify --target /ruta/a/tu/proyecto --no-engram
```

### Resolver drift

```bash
lufy-ai status --target /ruta/a/tu/proyecto --verbose
LUFY_MERGE_TOOL="tu-merge-tool" lufy-ai merge --target /ruta/a/tu/proyecto <path>
```

### Backup y restore

```bash
lufy-ai backup --target /ruta/a/tu/proyecto
lufy-ai restore --target /ruta/a/tu/proyecto --list
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id> --dry-run
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id> --yes
```

### Desinstalar y reinstalar

```bash
lufy-ai uninstall --target /ruta/a/tu/proyecto --dry-run
lufy-ai uninstall --target /ruta/a/tu/proyecto --yes
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --yes --no-engram
lufy-ai verify --target /ruta/a/tu/proyecto --tool opencode --no-engram --quiet
```

`uninstall` preserva `opencode.json`, preserva `AGENTS.md` y elimina solo la referencia `@lufy-ia.harness.md`. Si encuentra drift en assets gestionados, bloquea antes de mutar.

## 6. Comandos slash disponibles

| Namespace | Comando | Uso |
| --- | --- | --- |
| OpenSpec | `/opsx-explore` | Investigación read-only. |
| OpenSpec | `/opsx-propose` | Crear proposal, design, specs y tasks. |
| OpenSpec | `/opsx-apply` | Implementar tareas de un cambio activo. |
| OpenSpec | `/opsx-verify` | Verificar completitud contra specs/tasks. |
| OpenSpec | `/opsx-sync` | Aplicar deltas validados a specs principales. |
| OpenSpec | `/opsx-archive` | Archivar cambio terminado. |
| OpenSpec | `/opsx-version` | Reportar fuente OpenSpec efectiva. |
| Lufy | `/lufy.timereport` | Reporte local de tiempo/ROI. |
| Lufy | `/lufy.onboard` | Onboarding/demo cuando esté disponible en el target. |

`/opsx-*` se conserva como namespace estable de OpenSpec. `/lufy.*` queda para capacidades propias del kit.

## 7. Roles instalados

| Agente | Rol |
| --- | --- |
| `orchestrator` | Coordina y elige el menor flujo seguro. |
| `sdd-router` | Clasifica T1/T2/T3 en modo read-only/no-shell. |
| `explorer` | Investiga impacto y riesgos sin editar. |
| `implementer` | Implementa cambios acotados sin delivery. |
| `test-writer` | Apoya TDD stack-aware en T1/T2 sustantivos. |
| `validator` | Valida y diagnostica sin editar. |
| `reviewer` | Revisa calidad, cobertura y riesgo. |
| `delivery` | Git/GitHub solo con autorización explícita. |

## 8. Desarrollo local de lufy-ai

```bash
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

Validación del repo:

```bash
cd /tmp/lufy-ai
scripts/validate.sh
```

No hay `npm test`, `npm run typecheck` ni `tsc` global en la raíz.

## 9. Troubleshooting rápido

### Los agentes no cargan

1. Reinicia OpenCode.
2. Verifica `.opencode/agents/`.
3. Verifica que `AGENTS.md` tenga `@lufy-ia.harness.md`.
4. Corre `lufy-ai verify --target <repo> --no-engram`.

### El plugin TUI no aparece

1. Verifica `tui.json`.
2. Verifica `.opencode/plugins/agent-observatory.tsx`.
3. Reinicia OpenCode.
4. Usa `/observatory`.

### Quiero remover Lufy del repo

```bash
lufy-ai uninstall --target <repo> --dry-run
lufy-ai uninstall --target <repo> --yes
```

Si el comando bloquea por drift, revisa `status --verbose` antes de decidir cómo resolver.

## Más documentación

- [README raíz](../README.md)
- [Instalación completa](installation.md)
- [Arquitectura](architecture.md)
- [Estado](status.md)
- [Roadmap](roadmap.md)
- [CLI Go README](../tools/lufy-cli-go/README.md)
- [OpenSpec](../openspec/README.md)
