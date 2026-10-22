# Verificación B5: evidencia ligada al candidato Git

Estado del snapshot: `delivered` para publicación inicial y elegibilidad de archive; merge final pendiente. PR https://github.com/adrotech/lufy-ai/pull/248, rama `codex/bind-candidate-evidence`, HEAD `3c5a8b8e79fe97dbc975415f31bc9bd5ae868e65`.

## Alcance y aceptación

Doce rutas source enumeradas en `/private/tmp/lufy-seven-b5-paths.json`, con spec canónica y mirror sincronizados. Capture/verify compara repositorio, base/HEAD/trees, paths/modos, políticas, evidencia y rol; sidecar JSON estricto. Declara identidad del snapshot, sin autenticar autoría ni atestar ejecución de tests sobre ese candidato. B6, stores locales y checkout principal mixed excluidos.

## Evidencia proporcional

Root relay confirmó `scripts/validate.sh` exit 0 en `/private/tmp/lufy-seven-b5-validation-promisor-final.log`: 98 pruebas offline en 0.163s, Go global 80.5%, candidateevidence 86.6% (9.912s), vet/build exitosos. Shellcheck no disponible localmente. OpenSpec change/canon strict y mirror passed. Reporte test-writer válido, fingerprint `210d4fcec2274282eb7434f5c8d76977bd147546536e077d9ce0ee1588166f5e`.

RED/GREEN conservado: claves case-sensitive y políticas tracked; cuatro callbacks reales de filtros directos/includes; gitlink no inicializado; promisor local con objeto base ausente que se materializaba durante Capture. Correcciones rechazan filtros efectivos, gitlinks y configuración promisor/partial-clone antes de status/resolución de commits/trees; `GIT_NO_LAZY_FETCH=1` y prueba de ODB intacta tras rechazo. Control promisor reproduce fetch local y prueba corregida GREEN en 0.770s. No dependió de red ni nuevos modelos.

Demo con binario fresco: clon limpio del candidato B4 `146b1a6` current, README dirty rechazado exit 4 y restauración current, clave ajena `Role` rechazada exit 4. Esto prueba el mecanismo; no atribuye validaciones históricas a código ausente de ese HEAD.

## Límites y review

Dos snapshots sin lock frente a un mutador hostil; configuración debe permanecer estable y verify ejecutarse inmediatamente antes de consumir. Ignorados/dependencias externas no forman parte del árbol comprometido. Config local conocida es excepción exacta y ligada por hash. Deadline total 10s más cleanup de IO 1s, 2MiB por salida, sidecar 64KiB. No modifica Result Contract, CAS ni gates; no ofrece identidad nativa. Repositorios con filtros, gitlinks o promisor/partial-clone se rechazan conservadoramente.

Context review: escalate por graph `not_available`, traceability `unknown` y límites ausentes; se preserva fallback dirigido, sin inventar budgets. Unidad coherente T1 cubierta por plan aprobado: slices lógicas binding/CLI y lectura Git endurecida; no nueva autorización por conteo.

## Delivery y cierre

Root relay confirma stage/commit/push del follow-up de cinco rutas y guard committed del rango completo de doce rutas, whitespace PASS. Ninguna ruta B6 incluida.

INITIAL_CHECKS_SNAPSHOT: root relay confirmó snapshot fresco de PR248 OPEN/CLEAN, HEAD `3c5a8b8e79fe97dbc975415f31bc9bd5ae868e65`, run `39869413305`, cinco SUCCESS del 2026-10-22: Quality 23:06:33Z, Ubuntu 23:05:35Z, macOS 23:05:52Z, Windows 23:07:57Z y Smoke 23:08:19Z. La corrección de fixture tras el fallo macOS anterior queda confirmada en CI. La publicación inicial está delivered y es elegible para archive; no afirma merge todavía.

Tras ese gate, la tarea de publicación inicial permite archive con seis artifacts B5 únicamente, validación de archive habilitada y `--skip-specs` porque sync ya está aplicado. Commit/push del archive requiere guard committed de dieciocho rutas y whitespace del rango; el nuevo HEAD debe tener cinco SUCCESS y CLEAN antes de merge autorizado con `--match-head-commit`, sin admin/force. La prueba GitHub MERGED y el merge SHA cierran delivery después; este snapshot previo al merge no afirma cierre.
