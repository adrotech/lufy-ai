---
name: lufy.close
description: Cierra o finaliza un cambio activo del workflow LUFY. Usar cuando el usuario pide finalizar la spec/cambio activo, actualizar información, sincronizar/archivar, chequear PR cerrado o mergeado y limpiar la rama de forma segura.
license: MIT
compatibility: Requiere OpenSpec CLI; los chequeos Git/GH son opcionales solo cuando delivery o limpieza de PR/rama no aplican.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Cierre De Workflow LUFY

Finaliza un cambio activo del workflow LUFY o respaldado por OpenSpec después de implementación, validación, sync de specs, delivery, PR cerrado/mergeado y limpieza segura de rama.

**Entrada**: Opcionalmente especificar un nombre de cambio. Si se omite, inferirlo solo cuando exista un único cambio activo o la conversación lo nombre claramente; si no, pedir al usuario que elija.

## Pasos

1. **Seleccionar el cambio**

   - Si se proporciona un nombre, usarlo.
   - Si no, ejecutar `openspec list --json` y seleccionar solo cuando no sea ambiguo.
   - Anunciar: `Usando cambio: <name>`. Para sobrescribirlo, ejecutar `/lufy.close <other>`.

   Contexto específico del repo:
   - La spec activa/foco actual es `install-managed-assets-with-hash-idempotency`.
   - `migrate-installer-to-go-cli` no debe archivarse mientras tenga tasks incompletas.
   - Arquitectura del instalador: CLI Go en `tools/lufy-cli-go`; `scripts/install.sh` es un wrapper estricto sin fallback legacy.

2. **Actualizar evidencia de cierre**

   Recolectar evidencia compacta antes de mutar cualquier cosa:

   ```bash
   openspec status --change "<name>" --json
   openspec validate "<name>"
   git status --short
   git status --branch --short
   git branch --show-current
   git log --oneline -10
   ```

   Si existe upstream, revisar commits locales sin push con:

   ```bash
   git log --oneline @{u}..HEAD
   ```


3. **Verificar tasks y artifacts**

   - Leer `openspec/changes/<name>/tasks.md`.
   - Bloquear si queda cualquier task `- [ ]`.
   - Bloquear si faltan artifacts requeridos o están incompletos.
   - Tratar los checkboxes como necesarios pero no suficientes para `closed` o archive readiness.

4. **Verificar implementación y validación**

   - Usar la semántica de `/opsx-verify <name>` o el skill instalado `openspec-verify-change`.
   - Ejecutar o reutilizar evidencia de validación agrupada vigente del bloque/proposal.
   - En este repositorio, preferir `scripts/validate.sh` cuando el cambio toque assets gestionados, CLI Go o instrucciones de workflow.
   - Si la validación no puede ejecutarse, reportar la limitación exacta y evidencia estática/manual.

5. **Sincronizar specs principales cuando corresponda**

   - Revisar `openspec/changes/<name>/specs/` para detectar delta specs.
   - Si existen deltas y no están reflejados en `openspec/specs/`, ejecutar o solicitar `/opsx-sync <name>` antes de archivar.
   - Después del sync, ejecutar `openspec validate "<name>"` y `openspec validate --all` cuando estén disponibles.
   - Nunca archivar con delta specs sin sincronizar.

6. **Verificar docs/specs sin commit y commits sin push**

   Este gate aplica especialmente a cambios docs-only, OpenSpec-only o workflow-only.

   - Revisar cambios sin commit/stage con:

   ```bash
   git status --short
   git diff --name-only
   git diff --cached --name-only
   ```

   - Si aparecen archivos de documentación, specs o workflow sin commit, retornar `delivery_pending` o `blocked` con la acción exacta. Rutas típicas:
     - `docs/**`
     - `openspec/**`
     - `.lufy/**`
     - `.opencode/**`
     - `AGENTS.md`, `AGENTS.md.template`, `lufy-ia.harness.md`
     - `README.md`, `CHANGELOG.md`, `SECURITY.md`
   - Si existen commits locales sin push en la rama actual y el cierre requiere PR/delivery, retornar `delivery_pending` con instrucción de push/autorización.
   - No marcar `closed` ni ejecutar archive si hay docs/specs/workflow artifacts pendientes de commit o push, salvo que el usuario declare explícitamente que esos cambios no pertenecen al cierre y no afectan el archive.

7. **Verificar delivery y estado de PR**

   Determinar si Git/GH delivery es requerido para este cambio:

   - Si el cambio no requiere commit/PR/delivery y no hay cambios pendientes ni commits locales sin push, marcar delivery como `not_applicable` con evidencia.
   - Si delivery es requerido pero no está autorizado o está incompleto, retornar `delivery_pending` con la autorización/acción exacta.
   - Si existe URL/número de PR en contexto, metadata Git, commits recientes o input del usuario, verificar con:

   ```bash
   gh pr view <PR> --json url,state,mergedAt,headRefName,baseRefName,mergeStateStatus,statusCheckRollup
   gh pr checks <PR>
   ```

   - Tratar `MERGED` o cierre no mergeado como evidencia de cierre solo cuando el usuario/workflow acepte explícitamente cierre sin merge.
   - Si el PR está abierto, los checks están pendientes/fallidos o falta evidencia de un PR requerido, retornar `delivery_pending` o `blocked`.

8. **Evaluar limpieza de rama**

   La limpieza de rama es opcional y destructiva. Debe ser segura y explícitamente autorizada.

   - Inspeccionar estado con `git status --short`, `git branch --show-current`, `git branch --merged`, `git remote -v` y upstream cuando aplique.
   - Nunca borrar `main`, `develop`, `master`, `development` ni la rama actual.
   - Nunca borrar una rama con commits no mergeados o sin push salvo pedido explícito y riesgo reportado.
   - Si la limpieza es segura pero no está autorizada, retornar `delivery_pending` con comandos exactos a autorizar, por ejemplo:

   ```bash
   git branch -d <branch>
   git push origin --delete <branch>
   ```

   - Si la limpieza está autorizada, enrutar por `delivery`; los comandos de borrado deben pedir permiso al usuario.

9. **Archivar el cambio**

   Usar `/opsx-archive <name>` solo después de resolver tasks, artifacts, validación, sync, docs/specs sin commit, commits sin push, delivery/PR y decisiones de cleanup.

   - Si el archive termina correctamente, reportar la ruta.
   - Si el archive queda bloqueado, retornar el menor paso exacto de recuperación.

10. **Devolver Result Contract**

   La salida final DEBE usar Result Contract envelope v1 e incluir:

   - `status`: `closed`, `delivery_pending`, `sync_pending` o `blocked`.
   - `artifacts.changed`: ruta de archive, specs sincronizadas, limpieza de rama o `none`.
   - `evidence.commands`: comandos OpenSpec, validación, Git y GH con resultados.
   - `risks`: riesgos pendientes de PR/rama/delivery/sync/docs o `none`.
   - `next_recommended`: owner y acción exacta.

## Guardrails

- No adivinar el cambio cuando existan múltiples cambios activos.
- No archivar cambios con tasks incompletas o delta specs sin sync.
- No tratar checkboxes, validación o aprobación del usuario como autorización Git/GH.
- No marcar `closed` si quedan docs/specs/workflow artifacts sin commit o commits locales sin push que pertenezcan al cierre.
- No crear commits, push, PRs, comentarios, sync de Projects ni borrar ramas sin autorización explícita y routing por `delivery`.
- No borrar ramas protegidas ni la rama actual.
- No afirmar PR cerrado/mergeado sin evidencia de `gh` o evidencia explícita provista por el usuario.
- Mantener outputs compactos y en español, preservando identificadores técnicos.
