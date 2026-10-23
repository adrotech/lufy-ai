# Verificacion B4: eventos explicitos de review

## Aceptacion y contrato

- Start/complete explicitos del host generan review.started y review.completed con TaskRef del intento y timestamps observados; metricas forman el par sin duplicar retries.
- Complete/rework/reopen requieren inicio y Result Contract estricto con claim reviewer permitido. Complete exige outcome passed|failed y no admite passed si contrato blocked/escalated. Falta de evidencia, timestamp invalido o conflicto no produce conclusion falsa.
- Lifecycle SubagentStart/SubagentStop/SessionEnd sigue generico, sin inferencia automatica de review. Las operaciones no ejecutan modelos ni autentican identidad de agente; contrato estructural valido tampoco prueba calidad de review.
- Eventos KindEvidence sin checkpoint conservan gates y estado terminal independientes. No conceden aprobacion ni delivery y no ligan todavia evidencia a candidato Git.

IDs opacos validados; review/operation hasheados y RunID selecciona el stream. La clave separa hashes de componentes para evitar colisiones con ':'. Retry conserva timestamp, devuelve duplicate_noop y conflicto con contenido distinto; CAS/reintentos acotados sobre FileStore real. Persisten metadata/outcome/fingerprint canonical, no cuerpos/findings, rutas de evidencia, prompts ni entorno. No inicializa instalacion ni migra layout; el ledger debe estar habilitado.

## Evidencia y triangulacion

RED de tuplecollision y help antes de corregir; GREEN CLI/servicio completo y triangulacion adversa sobre FileStore/proyector reales: retries concurrentes, conflictos, privacidad, reloj y checkpoints. Reporte test-writer /private/tmp/lufy-seven-b4-tests-result.yaml validado, fingerprint8fd198d4f60fb87bec1105e29d65250c5b15f2ad820f3dee62cfae8bf4f8c814.

scripts/validate.sh exit0, log /private/tmp/lufy-seven-b4-validation.log: 98 offline, suite Go coverage global80.5% y reviewevents93%, vet/build. Shellcheck local ausente y lint omitido, limitacion declarada. OpenSpec change/canon strict passed, espejo exacto propio y assets post-sync0.434s passed. No se repitieron tests por delivery.

Review manual real del productor registrada por host: start22:36:27.859229, complete22:37:04.000799, duracion36141ms; metrics available con1completed y2recognized, retry duplicate_noop. Contrato local /private/tmp/lufy-seven-b4-review.json; cuerpo no persistido. Es la observacion de esa review declarada, no prueba de actividad ausente, identidad nativa, costes/tokens o candidato Git exacto. B5 cubre binding por separado.

## Scope y review workload

Diez paths: dos docs, tres CLI, tres service y spec/mirror. Manifest propio /private/tmp/lufy-seven-b4-paths.json, source CLI/servicio/docs/proposal revisados, stage vacio antes relay y espejo exacto. Root confirmo stage10/guard worktree+committed/whitespace/push correctos.

T1 por nueva CLI y contrato operacional, una unidad coherente cubierta por plan/aceptacion/validacion; separar productor y consumo observable dejaria contrato incompleto. context review conserva escalate, graph not_available, traceability unknown y limites ausentes; dos tracked iniciales no representan scope10. Fallback dirigido sin budgets inventados. Locales .lufy/stores y artifacts activos excluidos; padre mixed/dist preservados. Sin OpenCode/Pi ni permisos de reviewer/orchestrator ampliados.

## Snapshot y gates

PR https://github.com/adrotech/lufy-ai/pull/247 contra develop, HEAD inicial21e051cc7807e903e058f6e8d20311684a8a88d2. Root snapshot confirma OPEN/CLEAN y cinco SUCCESS del HEAD21e051cc7807e903e058f6e8d20311684a8a88d2, run39866564899: Quality2026-10-22T22:41:17Z, Ubuntu22:40:24Z, macOS22:40:30Z, Windows22:42:23Z e installer22:42:47Z.

Autorizacion humana global del programa vigente, root relay ejecuta Git/GH y delivery conserva gates. Snapshot delivered elegible archive, sin closed anticipado. Task3.1 significa publicacion/checks iniciales/elegibilidad archive; con tasks completas, archive validado usa skip-specs porque sync aplicado.

Stage solo seis artifacts archive B4. Excluir B5 source/docs/locales; no include-worktree global con trabajo ajeno dirty. Tras commit comprobar rango origin/develop...HEAD, cinco checks del nuevo HEAD y merge normal con match-head-commit exacto, sin admin/force. Prueba posterior GitHub cierra el bloque sin reescribir snapshot. No polling/escalaciones propias ni cambios del checkout primario.
