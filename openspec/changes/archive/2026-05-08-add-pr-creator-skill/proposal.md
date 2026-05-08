## Why

El flujo de delivery necesita una forma consistente de producir contenido de Pull Request con trazabilidad, evidencia y secciones operativas sin depender de redacción manual ad hoc. Crear un skill `pr.creator` permite estandarizar el template de PR de GitHub, aprovechar contexto OpenSpec —especialmente `proposal.md`— y ofrecer el mismo formato tanto cuando alguien lo invoque manualmente como cuando `delivery` prepare un PR.

## What Changes

- Añadir una nueva capacidad OpenSpec para un skill OpenCode llamado `pr.creator` bajo `.opencode/skills/`.
- Definir dos flujos de uso:
  - **Modo manual**: una persona o agente invoca `pr.creator` para generar o refinar el contenido del PR sin ejecutar delivery.
  - **Modo integrado desde `delivery`**: el subagente `delivery` debe llamar o aplicar `pr.creator` antes/durante la creación de PR para producir el cuerpo del PR con el template estándar.
- Definir que `pr.creator` genere contenido/plantilla de Pull Request para GitHub, no que cree el PR ni ejecute operaciones Git/GH por sí mismo.
- Mantener `delivery` como responsable de inspección Git, validación final, commit, push, `gh pr create` y sync autorizado; `pr.creator` solo estructura el título/cuerpo/evidencia del PR.
- Incluir en la plantilla: resumen de nueva funcionalidad, `why`, link a tarea asociada cuando exista herramienta de tracking configurada, evidencia de pruebas, sección `Monitors` y sección `Migraciones`.
- Requerir detección automática de migraciones o cambios de tablas/schemas de DB a partir del diff/rutas/patrones disponibles, reflejando el resultado en el template.
- Mantener el contenido humano en español y preservar identificadores técnicos.

## Capabilities

### New Capabilities
- `pr-creator-skill`: Skill OpenCode `pr.creator` con estructura estándar tipo Anthropic para generar contenido de Pull Request en GitHub con trazabilidad, evidencia, monitores y migraciones.

### Modified Capabilities
- Ninguna.

## Impact

- Archivos nuevos esperados bajo `.opencode/skills/pr.creator/`, incluyendo `SKILL.md` y recursos/templates auxiliares según el diseño.
- Actualización esperada de `.opencode/agents/delivery.md` o documentación/configuración equivalente para indicar que `delivery` debe usar `pr.creator` al crear PRs cuando el skill esté disponible.
- Posible documentación interna del skill si se requiere para explicar uso, modo manual, modo integrado y límites.
- No modifica código del producto, instalador, contratos públicos, puertos, auth defaults ni esquema de base de datos.
- No introduce operaciones de delivery automáticas: el skill prepara contenido de PR; creación de PR, commit, push y sync siguen siendo responsabilidad del rol `delivery` con autorización explícita.
