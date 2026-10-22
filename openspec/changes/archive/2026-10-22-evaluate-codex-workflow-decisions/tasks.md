# Tasks: bloque 2, decisiones conductuales de Codex

Autorización global recibida: «continua 100% tenes mi autorizacion para finalizar todos los bloques». Incluye delivery y merge condicionado a checks; no omite aceptación ni seguridad.

## 0. Precondiciones

- [x] 0.1 Confirmar alcance aprobado: cinco grupos del bloque 2, sólo Codex, sin cambios de permisos ni auth real.
- [x] 0.2 Inspeccionar CLI local, autenticación disponible, contratos y agentes Codex; definir criterios y aislamiento.
- [x] 0.3 Validar proposal/design/delta en modo estricto.

## 1. P1 — Piloto reproducible

- [x] 1.1 Definir seis casos sintéticos, prompts y expectativas observables; séptima invocación como receptor del handoff.
- [x] 1.2 Delegar regresiones del evaluador a test-writer con RED real y casos adversos; 19 pruebas finales.
- [x] 1.3 Implementar ejecución secuencial, timeout, captura de JSONL, snapshots y evaluación sin falsos positivos por autodeclaración. Añadida parada de infraestructura tras observar 429/timeouts.
- [x] 1.4 Documentar uso, datos privados, comandos, límites y diferencia entre CLI y app.
- [x] 1.5 Recuperación diagnosticada: selección opcional de modelo/esfuerzo por invocación, metadata veraz y pruebas RED/GREEN; preservar defaults y configuración global. Ocho métodos nuevos: RED previo y GREEN 27/27 con los 19 anteriores; no se cambiaron criterios del evaluador.
- [x] 1.6 Corregir el falso positivo observado cuando stderr registra un error de herramientas omitido en JSONL. RED delegado: un falsoverde y ocho errores; GREEN raíz: 35/35. Registrar desactivación solicitada sin afirmar que fue verificada y conservar evidencia original al reevaluar.

## 2. P2 — Ejecución conductual

- [x] 2.1 Ejecutar análisis sin escrituras y comprobar traza/snapshot: passed.
- [x] 2.2 Ejecutar mecánico multifichero y auth de una línea; comprobar tier y gates: mecánico T3 correcto en smoke, auth T1/fast path false en corrida completa; timeout mecánico posterior conservado como inconcluso.
- [x] 2.3 Ejecutar auto-chain sin aprobación y comprobar bloqueo sin mutaciones. Corrida pending-explicit del 2026-10-22: T2, fast path false, blocked, snapshots iguales, sin errores del router; auditoría offline conserva passed.
- [x] 2.4 Última corrida aislada pasó: agents.enabled=false, apps desactivadas, handles host aislados; turno completo, bloqueo visible y sin revisión inline/acciones/routererrors. Metadata no afirma inventario exhaustivo verificado.
- [x] 2.5 Productor real de la app passed; receptor recuperado --handoff-from passed tras revalidar fuentes/instrucciones/snapshots/hash/traza. Referencias, aceptación, ejemplos y restricciones preservados sin mutaciones.
- [x] 2.6 Revisar trazas y registrar resultados reales, inconclusos/fallos y limitaciones en `docs/codex-behavior-pilot-2026-10-21.md`.

## 3. P3 — Verificación y cierre

- [x] 3.1 Revalidar unittest y checks pertinentes: 68/68 pruebas Python, OpenSpec estricto y whitespace de archivos nuevos correctos. Baseline Go anterior: scripts/validate.sh passed, coverage 80.4%, vet/build y paridad correctos; sin nuevos cambios Go/assets en estos ajustes del piloto.
- [x] 3.2 B2-AC1–AC6 verificados con muestras reales identificadas; OpenSpec estricto válido y límites preservados.
- [x] 3.3 Spec codex-behavior-pilot y mirror exacto sincronizados tras aceptación.
- [x] 3.4 Review y publicación autorizada en PR #243; cinco checks iniciales exitosos y elegibilidad de archive comprobada. HEAD con archive y merge final siguen gates explícitos en verification.md/result.yaml.

Configuración de proyecto ausente; `workflow_limits`, memoria y grafo se reportan no disponibles y no se inicializan como efecto lateral. Las tareas de ejecución no se marcan completas sólo por disponer del runner.

## Historial de estado y bloqueos (preservado)

Estado del bloque 2: `blocked` para aceptación conductual completa. P1 implementado/validado; P2 parcial. Auto-chain, delegación y handoff no produjeron evidencia concluyente por 429/timeouts. Una recuperación acotada terminó por timeout; no se lanzaron más llamadas. Recovery exacto en el reporte. La spec no se sincroniza ni archiva mientras la aceptación y esas tareas sigan pendientes.

Actualización posterior: auto-chain obtuvo evidencia concluyente en la corrida con selección explícita. Delegación quedó bloqueada visiblemente, pero un error real `collab spawn failed` en stderr no apareció como acción en JSONL; la reevaluación corregida la clasifica inconclusa. `--disable multi_agent` se solicita, pero no se demuestra desactivación efectiva. El productor de handoff terminó por timeout y el receptor no se ejecutó. Quedan pendientes 2.4/2.5 y aceptación completa; no se relajan sus criterios para cerrar.

Recuperación del 2026-10-22, autorizada por «continua»: auto-chain volvió a terminar por timeout a los 180 s; la parada de infraestructura evitó las otras tres invocaciones. Evidencia nueva en `/private/tmp/lufy-codex-pilot-recovery-20261022`. Se conservan los resultados anteriores y las advertencias del runtime sin atribuir causalidad ni modificar configuración global. Aceptación y tareas pendientes permanecen sin marcar.

Handoff test-writer: `/private/tmp/lufy-block2-tests-result.yaml`, validado con `lufy-ai result validate --role test-writer`; RED inicial y adversarial separados de GREEN ejecutado por root. El piloto ejecuta modelos realmente sólo con `--live`; listar escenarios no cuenta como corrida.

Actualización de esta sesión: CLI incluida en la app0.163.0-alpha.10.2 seleccionable mediante --codex; productor/receptor de handoff completos. Configuración oficial agents.enabled=false solicitada, pero muestra delegation medium timeout180s. Recuperación acotada low conserva configuración, criterios y evidencia separada; sin aceptación automática. Nueva hipótesis de aislamiento: handles de sesión/app-tools heredados del proceso host; se investigan sin cambios globales ni eliminación de restricciones de sandbox. Autorización delivery global vigente.

Estado actual: validated. Todas las líneas de conducta tienen muestras correctas bajo condiciones documentadas, incluida delegación en /private/tmp/lufy-codex-pilot-isolated-apps-disabled-20261022 (JSONL72f31ac0927c4262e5a301f1d31234053f65807a2e3fa7017b4b818c2412f8be). Los inconclusos históricos no se reinterpretan. GREEN68/68; sync aplicado, quedan gates de delivery/archive.
