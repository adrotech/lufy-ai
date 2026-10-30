> Informe histórico del modelo anterior. Sus resultados se conservan; sus comandos y roles no describen la rama inline actual. Ver [estado vigente](status.md).

# Verificación del perfil collaboration

El seguimiento aprobado añade un perfil diagnóstico opt-in para la API observada, preservando el ensayo legacy y su criterio de cierre completo. Una prueba parcial no se publica como verificación integral.

## Evidencia previa y causa reproducida

El chat de la app aceptó un reviewer nativo, recibió `null` y recuperó una vez al mismo agente con el contrato esperado. El validador rechazó null (exit 3) y aceptó el segundo resultado (exit 0, fingerprint `663d89a570f132e376de9c0ac01232f1b376932930f7a149bf421169bd744e54`). No se observaron hooks de ese chat ni una API de cierre.

En el CLI `0.163.0-alpha.10.2` se capturaron esquemas `additional_tools` mediante un proveedor loopback sin Authorization que devolvió HTTP400 deliberado, sin ejecutar modelos. Los cuatro controles mantuvieron el mismo reviewer:

| Configuración temporal de confianza | ignore-user-config | reviewer anunciado |
| --- | --- | --- |
| Ausente | Sí | No |
| Ausente | No | No |
| Cargada | No | Sí |
| Presente pero ignorada | Sí | No |

El override `projects.<fixture>.trust_level=trusted` por sí solo no bastó. Un control adicional registró reviewer explícitamente mediante `agents.reviewer.config_file`/`description` y lo anunció conservando ignore-user-config. Ese registro es la alternativa usada por el perfil, sin copiar/enlazar credenciales ni cargar configuración arbitraria de usuario. Los controles locales no atestiguan el proveedor autenticado real.

La API observada ofrece `spawn_agent`, `followup_task`, `list_agents` y `wait_agent`, sin `close_agent`. No se equipara `completed` con cierre ni `functions.wait` con espera de agentes. Referencias: [Subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents) y [autenticación de proveedores](https://learn.chatgpt.com/docs/auth).

## Ensayo acotado

Se ejecutó una sola corrida de hasta 180 segundos con el modelo/esfuerzo solicitados `gpt-6.1-sol/medium`, sin afirmar identidad servida. El CLI terminó exit 0, sin timeout; el evaluador devolvió exit 1: `inconclusive/tool_unknown`. Los hooks usaron nombres concatenados como `collaborationspawn_agent`, en lugar del separador que esperaba el normalizador. Se preservó ese resultado original.

| Evidencia | Observación |
| --- | --- |
| Lifecycle | SessionStart, SubagentStart reviewer, dos SubagentStop del mismo ID, Stop y SessionEnd; lifecycle Lufy exit 0 y nonce coherente |
| Herramientas | Un spawn, un followup, dos waits y dos inspecciones; callbacks PreToolUse/PostToolUse |
| Historia persistida | Un único child enlazado al parent, `agentRole=reviewer`, secuencia started/completed/interacted/completed |
| Payloads del child | Dos turnos completed: null rechazado (exit 3), luego JSON idéntico al fixture aceptado (exit 0) |
| Fingerprint recuperado | `7a41a92fd6c1cffe6ef403a529ebcad53309c2678bf964a544dd00765226133e` |
| Cierre | not_available; no se afirma native_execution_verified |

La auditoría posterior usó sólo `thread/read` del app-server sobre el parent de esta corrida y su único child enlazado, sin reanudar ni invocar modelos. El enlace entre `agentThreadId` y `agentPath` permitió comprobar la identidad UUID/nombre canónico mediante datos del runtime, sin deducirla del texto del modelo. Validó snapshots iguales, hashes originales de JSONL/hooks, rol, nonce y payloads. Esto acredita delegación y recuperación **en esa muestra CLI** por evidencia persistida; no cambia el veredicto automático original ni prueba cierre o calidad de review productiva.

Artifacts locales: `/private/tmp/lufy-native-collaboration-20261023-a/`, con `result.json` original intacto y `persisted-history-audit.json` separado. Reporte JSONL, prompts e historias completas quedan locales; esta documentación publica sólo campos sintéticos revisados.

Después del ensayo se añadieron aliases exactos para los nombres concatenados observados y controles offline del ack vacío de followup. Un ack vacío sólo puede acreditarse si la inspección posterior demuestra el contrato esperado del mismo agente; errores explícitos nunca se ignoran. No se repitieron modelos ni se presenta esa corrección posterior como validada live. La cobertura final del normalizador corregido es offline.

## Validación

### Seguimiento acotado T2: timeout de espera

Alcance: normalización/evaluación de `wait_agent`, pruebas adversariales, documentación y escenario canónico con su mirror. Programa T1, slice T2 SDD Lite sin fast path; una única unidad revisable porque código, tests y aceptación representan el mismo comportamiento. Seis rutas estimadas, sin paralelismo ni inferencias adicionales. `workflow_limits` y context graph no disponibles en este worktree; se usó evidencia directa de archivos/Git. La autorización vigente cubre implementación, validación y entrega a develop; no promoción a main.

Criterios WHEN/THEN: cuando una espera devuelve `timed_out: true` sin errores, sólo una inspección posterior del mismo agente con el contrato válido permite acreditar transporte. Cuando falta ese estado posterior, el booleano tiene tipo incorrecto, aparece un error explícito o la respuesta es desconocida, transporte permanece inconcluso. El resultado integral y cleanup conservan su limitación, y el perfil legacy conserva su aceptación.

La segunda corrida autorizada, en `lufy-native-collaboration-20261023-b`, terminó con CLI exit 0 y evaluador exit 1 (`collaboration_evidence_missing`): hubo tres esperas; la intermedia agotó su tiempo antes de completar el child. La historia persistida del parent y su único child confirma reviewer y dos turnos completados: null y el envelope esperado. Hooks Lufy, nonce y snapshots pasaron. El resultado original se conserva intacto.

La auditoría posterior sólo leyó esas historias y el rollout propio indicado por `thread/read`, sin reanudar ni invocar modelos. Vinculó siete outputs por `call_id`, comprobó igualdad de todos los campos normalizados originales y recuperó las esperas booleanas `false / true / false`. Una reevaluación offline del normalizador corregido obtuvo `transport_status=passed`, `role_status=requested`, `cleanup_status=not_available` y estado integral inconcluso. La historia acredita reviewer por su enlace UUID/nombre; el normalizador no inventa ese mapeo. No se presenta el replay como una nueva prueba live.

Evidencia local separada: `/private/tmp/lufy-native-collaboration-20261023-b/persisted-history-audit.json`; SHA-256 del resultado original `0e36df8fa2497476f5c7411fe5d952c30171349098f39f6c6ab1bc0e53ed9cc8`, del rollout `07e16f154934cfd79e7cff4a2171489ea485e5110b5f822c1a985cd74a7765c2`. El digest del payload recuperado es `8cdadc1a3c6a3822ab604fe02168ca8b46fe0b30add518bbb0a77eea5756b544` (hash del evaluador; distinto del fingerprint del validador Result Contract). Los cuerpos completos permanecen locales.

RED: dos pruebas nuevas fallaron antes del fix por ausencia de `wait_outcome`. GREEN final: 45 pruebas nativas, 0,066 s. La cobertura adversarial exige inspección posterior incluso si un timeout aparece después de una inspección válida; mantiene errores explícitos, respuestas desconocidas y legacy fuera de la excepción. `scripts/validate.sh` terminó exit 0 con 131 pruebas offline, cobertura Go 80,5%, vet/build y PR guard. Shellcheck no está disponible localmente; CI aplica su gate. OpenSpec strict, paridad del mirror, whitespace contra origin/develop y replay offline pasaron. Los checks remotos se registran en el reporte de entrega.

### Evidencia del PR #251

RED inicial: normalizador sin task_name/target/list_agents y evaluador ausente. RED adversarial: rol contradictorio y otro agente running omitido. RED de integración: un nonce incorrecto podía conservar transporte pasado por no evaluar el gate de una faceta independiente. RED posterior al live: nombres concatenados y ack vacío de followup. Los casos se corrigieron; 43 pruebas nativas focales pasaron. La validación agrupada final `scripts/validate.sh` terminó exit 0: 129 pruebas offline, cobertura Go 80,5%, vet y build. Shellcheck no está disponible localmente y CI aplica su gate. OpenSpec change/spec y paridad canonical/mirror pasaron.
