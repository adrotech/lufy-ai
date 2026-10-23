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

RED inicial: normalizador sin task_name/target/list_agents y evaluador ausente. RED adversarial: rol contradictorio y otro agente running omitido. RED de integración: un nonce incorrecto podía conservar transporte pasado por no evaluar el gate de una faceta independiente. RED posterior al live: nombres concatenados y ack vacío de followup. Los casos se corrigieron; 43 pruebas nativas focales pasaron. La validación agrupada final `scripts/validate.sh` terminó exit 0: 129 pruebas offline, cobertura Go 80,5%, vet y build. Shellcheck no está disponible localmente y CI aplica su gate. OpenSpec change/spec y paridad canonical/mirror pasaron.
