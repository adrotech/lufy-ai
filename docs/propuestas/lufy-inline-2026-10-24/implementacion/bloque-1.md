# Primer bloque: política inline y piloto

Fecha: 27/10/2026. Estado: implementación del piloto y validación offline completadas; aceptación conductual real parcial. El modelo principal y la instalación personal todavía no se sustituyeron.

Base: `origin/develop` en `1cc18d4647afc61ea1f6b1cb726ba74128fded5f`, `v0.6.25`. Checkout de validación limpio en `/private/tmp/lufy-inline-20261026`, rama local `codex/lufy-inline`. Los seis archivos nuevos también están guardados en el workspace original, sin reemplazar archivos existentes ni mezclar sus cambios previos.

## Implementación

| Archivo | Función |
| --- | --- |
| `.lufy/contracts/work.md` | Política neutral: resultado, alcance, autonomía, consultas humanas, delegación y evidencia por trabajo. |
| `tools/lufy-cli-go/internal/assets/embedded/.lufy/contracts/work.md` | Mirror de distribución idéntico a la fuente. No es una segunda política editable. |
| `tools/codex-behavior-pilot/work-scenarios.json` | Seis casos de análisis, propuesta, escritura, traducción, corrección y datos. |
| `tools/codex-behavior-pilot/work.py` | Preparación aislada, invocación opt-in, observación de acciones y aceptación según resultado. |
| `tools/codex-behavior-pilot/test_work.py` | 38 pruebas nuevas del evaluador, fixtures, runner y oráculo; sin modelos ni autenticación. |
| `tools/codex-behavior-pilot/WORK.md` | Uso, aceptación, aislamiento y límites del piloto. |

El fixture recibe sólo la política nueva y sus restricciones específicas en `AGENTS.md`; no carga la cadena de roles ni los contratos metodológicos anteriores. El agente responde en lenguaje natural, sin tier ni output schema obligatorio. La fuente nueva no está referenciada como reemplazo de las instrucciones principales todavía: esa selección requiere el catálogo/configuración del bloque siguiente.

Para textos y análisis, los controles objetivos de alcance y ejemplos no se confunden con calidad semántica: el experimento conserva `needs_review` y criterios de contenido. Para la corrección del filtro se agregó un oráculo independiente que comprueba comportamiento, conservación de datos y casos distintos de las pruebas visibles.

El código producido por un modelo sólo puede ejecutarse a través de `codex sandbox` con configuración read-only. Antes de importar el candidato, una sonda comprueba que no pueda escribir en el fixture. Si el sandbox falta, falla o permite esa escritura, se conserva `inconclusive`, sin fallback al host y sin invocar el modelo de ese caso. Las pruebas unitarias ejecutan únicamente fuentes sintéticas conocidas escritas por los propios tests.

## Evidencia offline

- `python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_*.py'`: **169/169 passed**, incluidas las 38 nuevas.
- `scripts/validate.sh`: **passed**, incluida la última pasada posterior al ajuste del sandbox del oráculo. Comprobó las 169 pruebas Python, assets, shells, workflows, build, tests/vet Go y cobertura **80,5 %**.
- Fuente y mirror de `work.md` comprobados byte a byte.
- Whitespace de archivos nuevos comprobado explícitamente: al estar untracked, `git diff --check` y el guard de PR no los cubren por sí solos. No se declara PR-ready ni se hizo staging.

Las pruebas adversariales expusieron verdes falsos al terminar el proceso antes de las assertions y al usar symlinks, y una omisión al conservar datos en el caso `None`. Las correcciones se verifican en los tests. La ausencia inicial del módulo fue un error de importación del esqueleto; no se presenta como evidencia RED de comportamiento. El RED/GREEN funcional del filtro corresponde a fixtures sintéticos, no a un caso real de modelo completado.

## Ensayos reales y límites

Las trazas permanecen locales. Se preservaron las ejecuciones originales sin reescribir sus veredictos cuando cambió el evaluador:

| Registro | Resultado observado |
| --- | --- |
| `/private/tmp/lufy-inline-pilot-20261026-a` | La CLI global `0.144.5` rechazó `agents.enabled=false` antes de invocar un modelo. |
| `/private/tmp/lufy-inline-pilot-20261026-b` | La CLI de la app no pudo iniciar su app-server dentro del sandbox externo. |
| `/private/tmp/lufy-inline-pilot-20261026-c` | Con ejecución autorizada, `explain-code` terminó: explicó factor, redondeo y resultado 121, sin cambios en el fixture. El evaluador de esa revisión quedó inconcluso porque aún no reconocía `nl`; se añadió cobertura después, sin cambiar el registro. |
| `/private/tmp/lufy-inline-pilot-20261027-d` | `propose-export` terminó con alternativas y recomendación contextual, sin implementar ni escribir archivos. Quedó inconcluso automáticamente por una secuencia de lecturas no reconocida; después se agregó soporte acotado para secuencias simples y tests negativos de shell. |
| Cuatro casos restantes | No ejecutados: se rechazó la solicitud de ejecución real. No se reintentaron ni se contaron como aprobados. |

El análisis manual de las dos respuestas comprobó el resultado aritmético y la pertinencia de la propuesta frente a sus fuentes. Es evidencia cualitativa de esas muestras, no una aprobación automática de la revisión final del runner. La CLI de app observada fue `0.164.0-alpha.11.1`; modelo default, identidad servida no verificada.

No se completó el caso real de corrección, escritura/traducción o datos, ni la aceptación real de la nueva sonda del sandbox. Tampoco se ensayó paralelismo nativo ni una comparación causal de costo/tiempo contra el baseline. La documentación y los tests no sustituyen esas observaciones pendientes.

## Próxima unidad

El bloque 2 debe reemplazar el modelo de configuración y catálogo: retirar `MethodologyByTier`, sus defaults y flags, seleccionar la política nueva desde el agente principal y actualizar todas sus referencias internas. No hay requisito de lectores antiguos, bridges ni soporte a usuarios legacy.

Este avance no cambia configuración global, no ejecuta instalación/sync, no elimina trabajo local ajeno y no realiza commit, push ni PR. La adopción del modelo único y el retiro de SDD/OpenSpec siguen pendientes.
