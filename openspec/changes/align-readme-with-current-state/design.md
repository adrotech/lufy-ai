## Context

El README actual ya preserva el banner y describe piezas reales del repositorio, pero todavía contiene una sección extensa de dirección futura sobre templates, stacks y un subagente `infra-cloud-sre`. Ese contenido pertenece al roadmap o a documentos de diseño porque esos templates/subagentes no existen como assets instalables.

El estado real del repo incluye una CLI Go en `tools/lufy-cli-go/`, wrapper Bash estricto en `scripts/install.sh`, comandos `install`, `verify`, `backup`, `restore` y trabajo de `sync` en curso. También hay trabajo de CI en curso, por lo que la documentación debe distinguir entre capacidades implementadas, propuestas activas y planes futuros.

## Goals / Non-Goals

**Goals:**

- Convertir `README.md` en una entrada breve orientada a estado real, quickstart y navegación.
- Mantener documentación humana en español sin traducir identificadores técnicos, rutas, comandos ni flags.
- Separar el contenido futuro de templates/subagentes hacia `docs/roadmap.md` o docs específicas.
- Sincronizar `docs/getting-started.md` y `tools/lufy-cli-go/README.md` con los comandos y límites actuales.
- Evitar afirmar validaciones, CI o features como listas si solo están propuestas o en curso.

**Non-Goals:**

- No implementar cambios de código, CLI, CI, installer, sync, templates ni agentes.
- No añadir templates por stack ni nuevas capacidades instalables.
- No cambiar contratos públicos de comandos, flags, rutas gestionadas ni política de delivery.
- No resolver o archivar otras proposals OpenSpec en curso.

## Decisions

1. README como landing page operativa.

   El README debe conservar el banner y enlaces relevantes, pero reducir la narrativa especulativa. La alternativa era mantener templates futuros en el README marcados como roadmap; se descarta porque RM-011 pide estado real + quickstart y porque los usuarios consumen el README como contrato inicial del repo.

2. Roadmap como destino del contenido futuro.

   Las ideas de templates, separación React/Next.js/Astro, subagente Infra/SRE y mapas futuros deben vivir en `docs/roadmap.md` o docs específicas enlazadas desde ahí. La alternativa era borrarlas; se descarta porque siguen siendo contexto útil de producto si no se presentan como disponible.

3. Quickstart basado en CLI Go y wrapper estricto.

   `README.md` y `docs/getting-started.md` deben mostrar primero el flujo actual: compilar `tools/lufy-cli-go/bin/lufy-ai` cuando haga falta, ejecutar `lufy-ai install` o `scripts/install.sh`, usar `--dry-run`, `--yes`, `--no-engram` y validar con `verify`. La alternativa era documentar instalación manual de assets; se descarta porque el wrapper ya no debe reintroducir lógica legacy ni sugerir copias manuales como camino principal.

4. Estado explícito para trabajo en curso.

   Sync y CI deben describirse con precisión: capacidades disponibles solo si existen en la rama actual y trabajo en curso/proposal si todavía no está completado. La alternativa era documentar el objetivo final; se descarta porque confunde roadmap con estado instalable.

5. Validación documental sin toolchain inventado.

   La implementación de esta proposal debe validarse con revisión estática, `git diff --check` y comandos OpenSpec disponibles. No se deben añadir ni prometer `npm test`, `tsc` o validación raíz inexistente.

## Risks / Trade-offs

- Pérdida de contexto estratégico al adelgazar el README -> Mitigar moviendo futuro a `docs/roadmap.md` y manteniendo enlaces claros.
- Drift entre README y CLI Go si cambia el slice de comandos durante proposals paralelas -> Mitigar referenciando `tools/lufy-cli-go/README.md` como detalle operativo y usando lenguaje de estado actual/en curso.
- Confusión por proposals simultáneas de CI y sync -> Mitigar evitando afirmar finalización si el trabajo sigue en curso.
- Enlaces rotos al mover secciones -> Mitigar revisando anchors internos y rutas relativas en los archivos modificados.

## Migration Plan

1. Reescribir `README.md` como landing page de estado real + quickstart + enlaces.
2. Consolidar contenido futuro de templates/subagentes en `docs/roadmap.md` o docs específicas sin presentarlo como instalable.
3. Actualizar `docs/getting-started.md` a español y al flujo actual de CLI Go.
4. Ajustar `tools/lufy-cli-go/README.md` para reflejar comandos y validaciones actuales sin sobreprometer CI/sync.
5. Validar estáticamente rutas, enlaces locales, ausencia de features falsas y whitespace.

Rollback: al ser documentación, revertir los archivos documentales tocados por esta change si la revisión detecta pérdida de información o drift nuevo.

## Open Questions

- Ninguna bloqueante. Si durante apply se detecta que una capability de sync o CI ya quedó completamente implementada en la rama, documentarla como estado actual; si no, mantenerla como trabajo en curso.
