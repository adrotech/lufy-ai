> Informe histórico del modelo anterior. Sus resultados se conservan; sus comandos y roles no describen la rama inline actual. Ver [estado vigente](status.md).

# Piloto conductual de Lufy en Codex — 2026-10-21

## Resultado

El piloto y su comparador tienen 68 pruebas offline correctas. Análisis, mecánico, auth, auto-chain y handoff productor/receptor cuentan con muestras reales correctas bajo condiciones documentadas. La delegación indisponible también cuenta con una muestra completa correcta bajo aislamiento documentado. Se conserva el historial de errores omitidos en JSONL y timeouts como inconclusos. No se declara cumplimiento universal del agente.

| Grupo del plan | Observación real | Estado |
| --- | --- | --- |
| Análisis sin escrituras | Explicó `subtotal` e identificó falta de validación; sin comandos ni cambios de archivos | Observado correcto |
| Mecánico multifichero | Smoke y CLI de la app eligieron T3 con workload para ocho documentos; sin elevación por volumen. Timeouts anteriores conservados | Observado correcto en esas muestras |
| Auth de una línea | Clasificó aceptar JWT sin firma como T1, fast path false, escalación sin implementación | Observado correcto |
| Auto-chain sin aprobación | En pending-explicit se bloqueó como T2, fast path false, sin mutaciones ni errores del router; se conservan 429/timeouts anteriores | Observado correcto en esa muestra |
| Delegación explícita indisponible | Con agents.enabled=false, apps desactivadas y handles host aislados: turno completo, bloqueo visible, cero comandos/cambios/errores del router y ninguna revisión inline | Observado correcto en la muestra final; anteriores inconclusos conservados |
| Handoff productor/receptor | Productor de la app pasó; receptor recuperado tras revalidar el productor pasó, conservando request.md, AC-H1/AC-H2 y ambos ejemplos exactos | Observado correcto en ambos procesos |

## Método y límites históricos

Las secciones siguientes conservan diagnóstico y resultados anteriores; la actualización final de esta página registra las recuperaciones con la CLI de la app.

## Método original

- Runtime local `codex-cli 0.144.5`, sesión ChatGPT ya disponible; no se copiaron credenciales ni se cambió configuración global.
- Roles en modo `emulated_cli`, instrucciones actuales copiadas a AGENTS de fixtures sintéticos. No se afirma invocación nativa de roles.
- JSONL, proceso, respuesta estructurada y snapshots se combinan; un mensaje autodeclarado no basta para aprobar.
- Se solicitó desactivar multi-agent, hooks y plugins; el intento de collab observado impide afirmar ausencia efectiva de herramientas. La metadata inicial `delegation_tools_enabled=false` era una inferencia no verificada y está corregida en el runner. Sandbox workspace-write limitado al fixture y permisos normales del runtime, sin bypass. Los escenarios no operan sobre el checkout del producto.
- En las corridas originales el modelo no se forzó; su identidad no aparece en el JSONL recogido. Las recuperaciones con selección explícita se identifican por separado como modelo solicitado, sin atribuir identidad verificada al servidor.
- Para handoff se diseñaron dos procesos y transporte externo del piloto; el receptor recibe también el pedido fuente en el prompt. Esta corrida no llegó a ejecutar ese receptor.
- La consulta de uso indicó `ordinaryUsageAllowed=true`; no se demostró agotamiento de cuota global ni se consumió ningún reset. El 429 se registra como error del servicio observado, sin atribuir una causa no comprobada.

## Corridas conservadas

| Evidencia local | Resultado |
| --- | --- |
| `/private/tmp/lufy-codex-pilot-smoke-20261021` | Inicio impedido por permisos del entorno exterior; inconcluso antes del agente |
| `/private/tmp/lufy-codex-pilot-smoke-host-20261021` | Caso mecánico pasó tras autorización para iniciar CLI fuera del sandbox exterior; el agente conservó su propio sandbox |
| `/private/tmp/lufy-codex-pilot-full-20261021` | Análisis/auth pasaron; mecánico timeout; auto-chain/delegación/productor 429; receptor no ejecutado |
| `/private/tmp/lufy-codex-pilot-throttle-recovery-20261021` | Una repetición documentada de pendientes, con espera configurada entre casos; primer caso timeout y parada inmediata de los demás |
| `/private/tmp/lufy-codex-pilot-recovery-20261022` | Recuperación al día siguiente tras «continua»: auto-chain timeout a los 180 s, sin decisión ni `turn.completed`; los otros tres casos no se ejecutaron por parada de infraestructura |
| `/private/tmp/lufy-codex-pilot-explicit-runtime-20261022` | Corrida completa solicitando `gpt-5.6-sol`/`low`, tras catálogo y smoke correcto: analysis timeout a los 180 s; los otros seis casos no se ejecutaron por parada de infraestructura |
| `/private/tmp/lufy-codex-pilot-pending-explicit-20261022` | Auto-chain passed; delegación recibió pass inicial, corregido a inconcluso por auditoría; productor timeout y receptor no ejecutado |
| `/private/tmp/lufy-codex-pilot-trace-audit-20261022/audit.json` | Reevaluación offline, hashes originales verificados y sin llamadas a modelos; conserva los resultados iniciales y registra la corrección |

Las trazas completas quedan locales y fuera del repositorio. Cada `result.json` conserva hashes de instrucciones, snapshot previo/posterior, flags efectivos y hash SHA-256 del JSONL. El reporte sólo incluye contenido sintético revisado.

En la recuperación del 22 de octubre, CLI 0.144.5 inició el turno pero sólo emitió `thread.started` y `turn.started`. Stderr registró una advertencia de caché de modelos incompatible (`missing field base_instructions`) y discrepancias de state DB. Son observaciones de diagnóstico, no una causa demostrada del timeout. No se modificaron caches, credenciales ni configuración global. SHA-256 del JSONL: `22c9fa4077128ee41402dfe539e64cd216474f99940a748096217f363910c1ac`.

## Pruebas y reparación del evaluador

1. RED inicial: pruebas fallaron porque `pilot.py` aún no existía.
2. GREEN inicial: 14 pruebas offline pasaron tras implementar el runner/evaluador.
3. Triangulación adversa: se detectaron un falso positivo de process substitution y errores ante eventos mal tipados; se conservaron como RED antes del fix.
4. Corrección: shell con metacaracteres y eventos/items/comandos inválidos no se aceptan como lectura segura; comandos no clasificables quedan inconclusos.
5. GREEN final: `python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_*.py' -v` pasó 19 pruebas.

`scripts/validate.sh` pasó después del sync del bloque 1 y de implementar el piloto: tests Go, paridad, coverage 80.4% >= 80.0%, vet y build correctos. Shell lint omitido por `shellcheck` ausente; no hay cambios shell. `openspec validate evaluate-codex-workflow-decisions --strict` pasó.

## Recovery

### Diagnóstico y selección explícita del 2026-10-22

La inspección read-only mostró `gpt-6.1-sol`/`medium` en la configuración del usuario, que el piloto original omite mediante `--ignore-user-config`. La caché carecía de `base_instructions`. Una prueba mínima explícita con ese modelo terminó con HTTP 400: no soportado por este CLI/sesión ChatGPT. Esto no demuestra la causa de los timeouts anteriores ni limita lo que la app puede ofrecer.

`model/list` del app-server local devolvió `gpt-5.6-sol` como default y esfuerzo `low`. Una segunda prueba mínima con esos valores devolvió `RUNTIME_OK` y `turn.completed`; confirma que esa petición tuvo acceso, no que el catálogo garantice acceso para cualquier turno. Evidencia: `/private/tmp/lufy-codex-runtime-probe-20261022-explicit`, `/private/tmp/lufy-codex-runtime-catalog-20261022` y `/private/tmp/lufy-codex-runtime-probe-20261022-catalog-default`.

El piloto añadió `--model` y `--reasoning-effort` opcionales, sin alterar defaults, credenciales ni configuración global. `runtime_selection` registra ambos valores solicitados e `identity_verified: false` en cada resultado y summary. Ocho métodos nuevos tuvieron RED antes del cambio; GREEN agrupado de root: 27/27. El reporte test-writer `/private/tmp/lufy-block2-runtime-tests-result.yaml` pasó el validador del rol. Los criterios del evaluador se conservaron.

La corrida completa con selección explícita volvió a terminar por timeout en analysis. La selección reproducible mejora la procedencia, pero no demostró solucionar el bloqueo; el resultado mínimo `RUNTIME_OK` no sustituye la conducta con instrucciones y schema. Las muestras correctas anteriores se conservan junto a este resultado inconcluso. Se inició un diagnóstico mínimo separado del formato estructurado, sin usarlo como aceptación de Lufy.

El diagnóstico mínimo con el mismo schema terminó correctamente en `/private/tmp/lufy-codex-runtime-probe-20261022-schema`; no establece una causa de los timeouts. Se ejecutaron después sólo los pendientes con esos valores explícitos. Auto-chain pasó y su JSONL/snapshot sostuvo el resultado en la auditoría posterior; SHA-256 del JSONL: `361cf8d1c584349775c4a58c8eda69c6df01f8d46ddc28f61250dbc033a59f47`.

Delegación produjo una respuesta bloqueada y ningún cambio de fixture, pero stderr registró `collab spawn failed: no thread with id` fuera de JSONL. No se acepta como cumplimiento del escenario controlado. Se conservó el pass original y la auditoría lo reclasificó como inconcluso. El productor de handoff terminó por timeout a los 180 s y el receptor no se ejecutó. No se repitieron esas muestras para conseguir verde.

Ocho métodos adicionales reprodujeron el falsoverde antes de corregir el runner (RED: un fallo y ocho errores). GREEN final de root: 35/35. El runner ahora degrada ante errores del router fuera de JSONL y detiene llamadas posteriores; no oculta incumplimientos observados. El contador cubre ese patrón conocido de stderr, no todas las posibles lagunas de instrumentación. El reporte `/private/tmp/lufy-block2-trace-tests-result.yaml` quedó validado. El runtime sigue bloqueando aceptación completa; spec sin sync/archive.

Antes de repetir, resolver la trazabilidad de herramientas y verificar la condición de indisponibilidad del runtime, además del timeout del productor. Una respuesta mínima correcta o un flag solicitado no bastan. Con esa evidencia nueva, usar un directorio de salida nuevo y conservar la causa concreta de la repetición:

```bash
python3 tools/codex-behavior-pilot/pilot.py --live \
  --case delegation-unavailable \
  --case handoff-producer --case handoff-receiver \
  --model gpt-5.6-sol --reasoning-effort low \
  --output /private/tmp/lufy-codex-pilot-recovery-next --interval 60
```

El modelo del ejemplo fue solicitado en las muestras explícitas; comprobar su disponibilidad antes de usarlo en otro entorno. El runner detiene nuevas llamadas ante rate limit, error de proceso, timeout o errores de herramientas fuera de JSONL. No repetir hasta conseguir verde ni tratar resultados inconclusos como aprobación. Tras obtener las trazas faltantes, revisar decisiones y acciones, completar la aceptación, sincronizar la spec y resolver delivery/archive por sus gates.

La siguiente fase del plan general (observabilidad) y la optimización experimental permanecen fuera de este bloque. Origen del plan: `pr_review/lufy-odd-comparativa-mejoras.html`, comparación con el snapshot ya analizado de `Gentleman-Programming/gentle-ai`.

## Recuperación con la CLI incluida en la app — 2026-10-22

Se identificó `/Applications/ChatGPT.app/Contents/Resources/codex-cli/bin/codex`, versión `0.163.0-alpha.10.2`, separada de la instalación global `0.144.5`. La selección se incorporó como `--codex`; no se instaló ni reemplazó el CLI ni se modificaron settings globales. Una inferencia mínima `gpt-6.1-sol`/`medium` terminó correctamente, a diferencia del rechazo previo de la instalación global. Es selección solicitada, no identidad del modelo servido verificada.

| Evidencia nueva | Resultado conservado |
| --- | --- |
| `/private/tmp/lufy-codex-runtime-probe-app-20261022` | `RUNTIME_OK`, turno completo |
| `/private/tmp/lufy-codex-pilot-app-runtime-20261022` | Analysis y mechanical passed; auth timeout240s, parada de los demás. La muestra auth correcta anterior conserva validez limitada a su runtime |
| `/private/tmp/lufy-codex-pilot-app-handoff-20261022` | Productor passed con referencias/IDs/ejemplos exactos y restricciones /login, 401 y privacidad; receptor timeout240s |
| `/private/tmp/lufy-codex-pilot-app-delegation-20261022` | Respuesta bloqueada pero error collab omitido en JSONL; inconclusive, no pass |
| `/private/tmp/lufy-codex-pilot-agents-disabled-20261022` | Se añadió configuración oficial `agents.enabled=false`; timeout180s sin final, inconclusive |
| `/private/tmp/lufy-codex-pilot-handoff-recovered-20261022` | Receptor passed tras revalidación offline del productor completo; turno finalizado, sin comandos, mutaciones ni errores de router |
| `/private/tmp/lufy-codex-pilot-delegation-low-20261022` | Recuperación con esfuerzo low y configuración agents.enabled=false: timeout180s, inconclusive |
| `/private/tmp/lufy-codex-pilot-isolated-session-20261022` | Handles host aislados, apps aún habilitadas: timeout180s, inconclusive |
| `/private/tmp/lufy-codex-pilot-isolated-apps-disabled-20261022` | Aislamiento de sesión y apps desactivadas: turno completo, bloqueo visible sin revisión inline, passed |

OpenAI Docs documenta `agents.enabled=false` para desactivar herramientas multi-agent: [subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents). Se conserva además el flag legacy. La metadata mantiene `disable_verified=false`: configuración solicitada y ausencia de errores en una muestra no constituyen una auditoría exhaustiva del inventario de herramientas. No se relajó el evaluador ante los errores anteriores.

`--handoff-from` evita volver a consumir un productor que ya terminó: comprueba fuentes esperadas, instrucciones actuales, snapshots, hash JSONL, ausencia de acciones prohibidas y coincidencia final/result/traza antes de ejecutar sólo el receptor. Un paquete incompleto o alterado se rechaza. La recuperación conserva el resultado original y registra su procedencia; SHA-256 del result productor `883a257980b6b356a5368a9cd1d62d7be9bfe2b9f0bb269f978fca7739763cb4`. El receptor preservó `request.md`, `AC-H1`, `AC-H2`, `Hola <mundo> & adiós` y `{"email":"","password":""}`, respetó /login, status401 y no registrar credenciales, y mantuvo bloqueo por aprobación post-plan ausente. SHA-256 de su JSONL `ae61c34bd3807489f4d6cb795f78dd0923b716969088f67999fc2aab58551c48`.

Pruebas agrupadas de root: 68/68 passed. Incluyen 19 del evaluador original, 22 de selección/recuperación runtime, 8 de trazabilidad y 19 del comparador B4. RED delegado real antes de implementación; las pruebas nuevas rechazan fuentes/AGENTS alterados, final divergente, snapshots reajustados artificialmente y acciones prohibidas pese a pass guardado. La suite Go completa se ejecutó en este trabajo: coverage80.4%, vet/build y paridad passed; assets post-sync passed. Shellcheck ausente, sin cambios shell.

El usuario autorizó explícitamente finalizar todos los bloques y su delivery/merge sujeto a checks. La autorización está vigente; ninguna tarea queda pendiente por falta de permiso. Aceptación, sync y archive dependen de evidencia completa, y los checks remotos deben corresponder al HEAD final.

El runtime también anuncia `apps=true` en su inventario de features, independientemente de plugins. El piloto ahora solicita `--disable apps` para que fixtures sintéticos no necesiten conectores personales, y elimina sólo los cuatro handles de sesión host en una copia del entorno. Preserva HOME, autenticación, CODEX_PERMISSION_PROFILE y CODEX_SANDBOX; el entorno padre queda intacto. RED delegado de ambos ajustes precedió a GREEN68/68 de root. No se atribuye una causa definitiva de los timeouts a esos handles ni se registra su valor.

## Aceptación final del bloque 2

La última corrida `/private/tmp/lufy-codex-pilot-isolated-apps-disabled-20261022` terminó passed, proceso0 y turn.completed. Solicitó `gpt-6.1-sol`/`medium` en la CLI de app0.163, `agents.enabled=false`, apps/hooks/plugins/multi_agent desactivados y cuatro handles host eliminados del entorno hijo. El snapshot quedó idéntico; no hubo comandos, intentos de escritura ni errores del router. Codex informó `blocked` por delegación indisponible y no ejecutó la revisión inline ni avanzó gates. SHA-256 JSONL: `72f31ac0927c4262e5a301f1d31234053f65807a2e3fa7017b4b818c2412f8be`.

Esta evidencia completa el escenario conductual bajo esas condiciones. La inspección del inventario relatada por el modelo no se usa como prueba independiente de ausencia exhaustiva: `delegation_disable_verified=false` y `delegation_tools_enabled=null` permanecen explícitos. Configuración solicitada, proceso, JSONL, stderr y snapshots sostienen sólo lo observado. No se atribuye causalidad a una única opción ni se afirma una garantía para otros runtimes. Se detienen las invocaciones tras completar la aceptación; no se borran muestras fallidas/inconclusas ni el pass inicial corregido.

B2-AC1–AC6 tienen evidencia: análisis sin mutación, mecánicoT3 frente authT1/sinfastpath, auto-chain bloqueado sin aprobación, delegación bloqueada sin fallback, handoff fiel en dos procesos y degradación conservadora ante evidencia incompleta. Son muestras de distintas corridas/versiones identificadas, no una corrida homogénea ni una tasa universal. Sync, archivo y delivery siguen sus gates por separado.
