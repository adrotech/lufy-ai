## Context

El repositorio usa OpenCode skills bajo `.opencode/skills/` y flujo OpenSpec para cambios planificados. La política de delivery exige evidencia real, trazabilidad y PRs hacia `develop`, pero hoy no hay un skill dedicado a preparar cuerpos de Pull Request con un formato uniforme.

El cambio propuesto debe crear un skill `pr.creator` con estructura estándar tipo Anthropic para skills. El skill debe ayudar a redactar contenido de PR para GitHub a partir de artefactos disponibles del cambio: propuesta OpenSpec, tareas, diffs locales, evidencia de validación y referencias a tracking/observabilidad configuradas.

Hay dos flujos explícitos:

- **Modo manual**: una persona o agente invoca `pr.creator` directamente para preparar el título/cuerpo de un PR o para revisar que la evidencia esté completa. Este modo no hace delivery ni requiere autorización Git/GH.
- **Modo integrado desde `delivery`**: cuando `delivery` tenga autorización para crear un PR, debe usar `pr.creator` como generador estándar del título/cuerpo antes de ejecutar `gh pr create`. `delivery` conserva la responsabilidad de branch safety, validación final, commit, push, creación del PR y sync remoto.

`pr.creator` no debe ejecutar delivery por sí mismo ni reemplazar el rol `delivery`; actúa como helper de contenido y estructura.

## Goals / Non-Goals

**Goals:**
- Definir un skill OpenCode `pr.creator` instalable en `.opencode/skills/pr.creator/` con `SKILL.md` como entrada principal y recursos/templates versionados.
- Generar una plantilla de Pull Request en español que incluya resumen funcional, `why`, trazabilidad a ticket, evidencia de pruebas, `Monitors` y `Migraciones`.
- Soportar invocación manual de `pr.creator` con inputs explícitos del usuario/agente.
- Integrar el uso de `pr.creator` en el flujo de `delivery` para creación de PR, dejando claro que el skill produce contenido y `delivery` ejecuta Git/GH.
- Preferir información derivada de `openspec/changes/<change>/proposal.md` cuando el cambio esté asociado a OpenSpec.
- Detectar migraciones o cambios de tablas/schemas de DB mediante patrones de rutas y diff, y reflejar explícitamente si existen, no existen o requieren confirmación manual.
- Mantener límites de seguridad: no hacer commit, push, PR, sync de proyectos ni mutaciones remotas.

**Non-Goals:**
- No implementar todavía el skill en esta propuesta.
- No crear ni modificar PRs reales en GitHub.
- No integrar APIs externas de Jira, Notion, Grafana, New Relic o Datadog en esta fase; el skill solo debe usar configuración/contexto disponible y pedir confirmación cuando falte información.
- No cambiar el workflow de delivery ni reemplazar `.opencode/policies/delivery.md`.

## Decisions

1. **Estructura estándar de skill con recursos locales**
   - Decisión: crear `.opencode/skills/pr.creator/SKILL.md` y recursos auxiliares bajo la misma carpeta, por ejemplo `templates/` y/o `references/`.
   - Rationale: mantiene el patrón tipo Anthropic de skills autocontenidos y facilita versionar el template sin mezclarlo con código del producto.
   - Alternativa considerada: ubicar solo un template en `.opencode/templates/`. Se descarta porque el requisito pide un skill y porque las instrucciones de uso/detección necesitan lógica operativa documentada.

2. **El skill genera contenido, no ejecuta delivery**
   - Decisión: `pr.creator` debe producir Markdown para el cuerpo del PR y, opcionalmente, sugerir un título; no debe llamar `gh pr create`, `git push`, Project sync ni herramientas remotas.
   - Rationale: respeta la separación de roles: `delivery` conserva las operaciones Git/GH autorizadas.
   - Alternativa considerada: automatizar `gh pr create`. Se descarta por los límites explícitos de permisos y trazabilidad.

3. **Dos modos de invocación con el mismo contrato de salida**
   - Decisión: documentar en `SKILL.md` un modo manual y un modo integrado desde `delivery`, ambos con el mismo output principal: título sugerido y cuerpo Markdown del PR.
   - Rationale: evita divergencia entre PRs preparados manualmente y PRs creados por `delivery`; el mismo template gobierna ambos caminos.
   - Alternativa considerada: crear un template separado para `delivery`. Se descarta porque duplicaría reglas y aumentaría el riesgo de inconsistencias.

4. **`delivery` debe delegar la generación del PR body a `pr.creator`**
   - Decisión: actualizar `.opencode/agents/delivery.md` o documentación/configuración equivalente para que, al crear un PR y cuando `.opencode/skills/pr.creator/` exista, `delivery` cargue/use el skill antes de `gh pr create`.
   - Rationale: el nuevo requisito convierte el uso integrado en parte del flujo esperado de delivery, sin transferir permisos Git/GH al skill.
   - Alternativa considerada: dejar el uso del skill como opcional para `delivery`. Se descarta porque el usuario pidió que `delivery` puede y debe llamarlo al crear PR.

5. **Fuentes de información en orden de prioridad**
   - Decisión: para resumen y `why`, usar `proposal.md` cuando exista; si no, usar diff/commits/contexto proporcionado por el usuario y marcar supuestos.
   - Rationale: OpenSpec ya captura motivación y alcance; reutilizarlo reduce duplicación y deriva manual.
   - Alternativa considerada: depender solo del diff. Se descarta porque el diff suele explicar el qué, pero no siempre el por qué.

6. **Tracking y monitores como secciones condicionales explícitas**
   - Decisión: incluir siempre secciones de ticket y `Monitors`, pero permitir valores `No configurado`, `No aplica` o `Pendiente de confirmar` según la evidencia disponible.
   - Rationale: evita ocultar trazabilidad ausente y obliga a declarar límites reales sin inventar integraciones.
   - Alternativa considerada: omitir secciones cuando no haya datos. Se descarta porque reduce consistencia y puede esconder huecos de delivery.

7. **Migraciones detectadas por heurísticas verificables**
   - Decisión: detectar migraciones y cambios de schema revisando rutas/patrones comunes (`migrations/`, `db/migrate/`, `schema.sql`, `prisma/schema.prisma`, `*.migration.*`, `CREATE TABLE`, `ALTER TABLE`, etc.) en archivos modificados/diff.
   - Rationale: cubre la mayoría de repos sin añadir dependencias externas y permite reportar evidencia concreta.
   - Alternativa considerada: requerir integración específica con cada ORM. Se descarta como demasiado acoplado para un skill genérico.

## Risks / Trade-offs

- **Riesgo: falsos negativos en detección de migraciones** → Mitigación: documentar patrones usados, incluir estado `Pendiente de confirmar` cuando el diff sugiera cambios de persistencia no reconocidos y permitir override manual.
- **Riesgo: enlaces de tickets/monitores inventados por falta de configuración** → Mitigación: el skill MUST usar solo evidencia disponible o pedir datos; si no existen, debe declarar `No configurado`/`No aplica`.
- **Riesgo: duplicación con templates existentes de delivery** → Mitigación: referenciar `.opencode/policies/delivery.md` y reutilizar o alinear con templates existentes en vez de crear reglas contradictorias.
- **Riesgo: confusión entre generar contenido y crear PR** → Mitigación: `SKILL.md` debe declarar explícitamente que no ejecuta `gh`, no muta remoto y no reemplaza autorización de `delivery`.
- **Riesgo: `delivery` omite el skill por flujo legacy** → Mitigación: actualizar las instrucciones de `delivery` para requerir `pr.creator` al crear PR cuando esté instalado, y dejar fallback explícito solo si el skill no existe.

## Migration Plan

- Implementar el skill como archivos nuevos bajo `.opencode/skills/pr.creator/`.
- Actualizar `.opencode/agents/delivery.md` o la configuración equivalente para invocar `pr.creator` durante creación de PR cuando el skill exista.
- Validar estáticamente que el template cubre todos los campos requeridos y que las instrucciones respetan delivery policy.
- No requiere rollback de datos ni migración runtime; revertir los archivos del skill elimina la capacidad.

## Open Questions

- ¿Existe una convención deseada para configurar herramientas de ticket/monitoring en este repo (variables de entorno, archivos de config o texto de contexto)? Si no existe, el skill deberá operar por detección/contexto y confirmación manual.
- ¿El template final debe vivir solo dentro del skill o también sincronizarse con `.opencode/templates/pr-evidence.md` si existe?
