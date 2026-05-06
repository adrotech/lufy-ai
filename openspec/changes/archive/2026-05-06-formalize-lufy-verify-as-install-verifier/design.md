# Design

## Decisión principal

`lufy-ai verify` será el verificador canónico porque ya comparte runtime con instalación/sync/restore, puede consumir `.lufy-ai/install-state.json` y recalcular hashes SHA-256 con el mismo paquete `internal/assets`. No se añadirá `scripts/verify-install.sh`.

## Comportamiento de verificación

El servicio `internal/verify` conserva el flujo actual y añade una capa explícita de estructura crítica:

1. Resolver `--target` con `platform.ResolveTargetPath`.
2. Confirmar que `.lufy-ai/install-state.json` existe como archivo regular no symlink y cargarlo con `state.Load`.
3. Validar JSON parseable para archivos JSON relevantes cuando existan (`opencode.json`, `.opencode/package.json`, `.opencode/package-lock.json`).
4. Validar directorios críticos como directorios reales no symlink:
   - `.opencode/agents`
   - `.opencode/commands`
   - `.opencode/skills`
   - `.opencode/plugins`
   - `.opencode/policies`
5. Validar archivos críticos como archivos regulares no symlink:
   - `.opencode/plugins/agent-observatory.tsx`
   - `AGENTS.md`
   - `tui.json`
   - `openspec/config.yaml`
6. Exigir que los archivos críticos gestionados estén presentes en el manifest.
7. Recorrer todos los assets del manifest y comparar SHA-256 actual contra `targetSHA256`.
8. Reportar Engram como warning no bloqueante o saltarlo con `--no-engram`.

## Reporte y fallos

- Los problemas estructurales incrementan el contador crítico y producen error final `verify falló con N problema(s) crítico(s)`.
- Los warnings de Engram o `sourceChangeID` inesperado no bloquean.
- Los mensajes siguen siendo accionables y en español.

## Alternativas consideradas

- **Crear `scripts/verify-install.sh`**: rechazado porque duplicaría reglas ya disponibles en Go y aumentaría divergencia entre CI, docs y CLI.
- **Validar solo hashes del manifest**: insuficiente porque una instalación podría tener manifest coherente pero carecer de categorías críticas necesarias para OpenCode/OpenSpec.
- **Cambiar esquema de estado**: innecesario; los checks estructurales se derivan de paths esperados y el manifest existente.

## Impacto de compatibilidad

La verificación se vuelve más estricta para instalaciones parciales o antiguas. Eso es deseado: `verify` debe representar el estado instalable actual del kit gestionado. `scripts/install.sh` permanece como wrapper estricto de `lufy-ai install`.
