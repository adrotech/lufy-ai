# Piloto conductual de Lufy en Codex

El ensayo de integración nativa y sus límites se documentan en [regresión nativa](../../docs/codex-native-regression-2026-10-22.md). `native.py` requiere `--live`, ejecutables explícitos y un directorio externo nuevo; conserva resultados inconclusos sin emular roles silenciosamente. Sus pruebas `test_native*.py` participan en el mismo discovery offline de CI.

Evalúa los cinco grupos del bloque 2 del plan comparativo con Gentle-AI: análisis, mecánico/auth, auto-chain, delegación indisponible y handoff. Son seis casos más un receptor independiente del handoff. Usa Python 3.9+ estándar en macOS/Linux y Codex CLI con autenticación ya disponible; el timeout controla grupos de procesos POSIX.

Listar casos no ejecuta modelos:

```bash
python3 tools/codex-behavior-pilot/pilot.py
```

Pruebas offline del evaluador:

```bash
python3 -m unittest discover -s tools/codex-behavior-pilot -p 'test_*.py' -v
```

`scripts/validate.sh` ejecuta esta suite como gate obligatorio, localmente y en el job Quality de CI. El workflow incluye los cambios de esta carpeta en PR y push. Python 3.9+ es una dependencia de validación; su ausencia o un test fallido devuelve error. Los tests simulan procesos y no ejecutan `--live`, modelos ni autenticación. Las corridas reales conservan su autorización y protocolo separados.

Corrida real explícita, secuencial, con directorio de evidencia nuevo fuera del checkout:

```bash
python3 tools/codex-behavior-pilot/pilot.py --live --output /private/tmp/lufy-codex-pilot-run
```

`--case` selecciona un ID y es repetible; el receptor necesita ejecutar antes el productor. `--timeout` limita cada proceso (180 segundos por defecto). `--interval` agrega hasta 60 segundos entre casos. Ante rate limit, timeout, fallo de proceso o errores del router de herramientas fuera de JSONL, el runner detiene nuevas invocaciones y marca los casos restantes inconclusos. No repetir casos fallidos hasta obtener pass: conservar la corrida y documentar causas/condiciones de una repetición.

`--codex <ejecutable-o-comando>` selecciona una instalación concreta, resuelta sin shell; `summary.json` registra `runtime_executable` y su versión. Sin opción, se conserva `codex` en PATH. Un ejecutable ausente se rechaza antes de crear evidencia. Esto permite comparar una instalación global y otra incluida con la app sin reemplazarlas.

`--model <id>` y `--reasoning-effort <nivel>` permiten solicitar una configuración por invocación tras diagnosticar incompatibilidades. Ambos son opcionales e independientes; los esfuerzos aceptados son `low`, `medium`, `high`, `xhigh`, `max` y `ultra`, pero su disponibilidad depende del modelo/runtime. Consultar `model/list` del app-server y comprobar una inferencia mínima antes de una nueva corrida. No cambiar credenciales ni configuración global para obtener verde. Cada caso y `summary.json` registra `runtime_selection`: modelo/esfuerzo solicitados e `identity_verified: false`, pues un argumento no demuestra identidad del modelo servido.

El proceso hijo elimina sólo los handles de sesión host `CODEX_APP_TOOLS_PIPE_PATH`, `CODEX_THREAD_ID`, `CODEX_SESSION_ID` y `CODEX_TOOL_CALL_ID` de su copia del entorno. Conserva HOME, autenticación ambiental y restricciones de sandbox/permisos heredadas; registra únicamente los nombres eliminados. No modifica el entorno del host ni publica valores.

El runner copia las instrucciones actuales de los roles y contrato neutral a fixtures temporales, registra sus hashes y usa `codex exec --json --ephemeral` con sandbox `workspace-write`. El uso de workspace-write permite detectar escrituras que una política debió evitar; los fixtures no contienen secretos ni código real del producto. No se usa bypass del sandbox. La configuración de usuario no se carga en el proceso de prueba, pero Codex reutiliza su autenticación; no se copian credenciales. Sin opciones de selección, se conserva el default efectivo del CLI aislado.

Se solicita desactivar multi-agent mediante `-c agents.enabled=false` y el flag legacy `--disable multi_agent`, además de hooks, plugins y apps, sólo en la invocación aislada. La clave actual está documentada en [OpenAI Docs: subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents). La configuración solicitada no demuestra por sí sola ausencia efectiva de herramientas: `delegation_disable_requested=true`, `delegation_disable_verified=false` y `delegation_tools_enabled=null` distinguen solicitud y verificación. Los roles se evalúan mediante sus instrucciones: `emulated_cli`, no invocación nativa de roles. El handoff viaja entre dos procesos coordinados por el runner; el receptor recibe el paquete y el contenido fuente sintético por prompt. El transporte no demuestra una delegación nativa.

Para recuperar sólo un receptor tras un timeout, usar `--case handoff-receiver --handoff-from <directorio-del-handoff-producer>`. Antes de iniciar un modelo, se revalidan instrucciones actuales, fuentes esperadas, snapshots, hash JSONL, acciones, finalización y coincidencia de la decisión con la traza. Un productor incompleto o alterado se rechaza; no se modifica la evidencia original. La nueva corrida registra la procedencia del productor y su hash. Recuperar un paquete conserva los gates de aprobación.

`summary.json` contiene veredictos, decisiones, hashes y metadata. Cada caso conserva JSONL, stderr, prompt, snapshot, decisión y fixture para inspección. Las trazas completas se mantienen locales; no incorporarlas al repositorio, porque futuras configuraciones del runtime pueden añadir contexto privado. Un reporte compartible debe usar sólo campos sintéticos revisados.

El evaluador comprueba finalización, salida estructurada, acciones y filesystem. Un write intentado y fallido cuenta como incumplimiento. Comandos/herramientas desconocidos producen resultado inconcluso; la clasificación de shell es conservadora y no constituye un sandbox. Un resultado verde es evidencia de esa muestra y ese runtime, no garantía universal ni benchmark de coste/latencia.

También cuenta líneas `ERROR codex_core::tools::router` en stderr. Aunque el proceso salga con cero y la respuesta declare bloqueo, esos errores impiden un pass por acciones que JSONL puede omitir. Las mutaciones observadas conservan prioridad de fallo; el contador no pretende auditar exhaustivamente todos los formatos de diagnóstico del runtime.

Interfaces contrastadas con [modo no interactivo oficial](https://learn.chatgpt.com/docs/non-interactive-mode), [configuración oficial](https://learn.chatgpt.com/docs/config-file/config-reference), [model/list del app-server](https://learn.chatgpt.com/docs/app-server#list-models-modellist) y `codex exec --help` local. El piloto es opt-in y no se ejecuta como parte de CI ni pide API keys nuevas.

## Experimento de entrega de contexto

`python3 tools/codex-behavior-pilot/experiment.py` muestra el protocolo sin ejecutar modelos. Con `--live --output <directorio-nuevo-externo> --codex <ejecutable> --model <id>` compara hasta dos pares del caso auto-chain, alternando el orden inline/referencia. Ambos conservan fuentes, instrucciones, sandbox y aceptación. `--pairs`, `--timeout` y `--interval` acotan el ensayo; el primer caso fallido o incompleto detiene nuevas llamadas.

El comparador reevalúa trazas, hashes, stderr y snapshots; no confía en un pass guardado. Informa bytes reales del prompt, tokens de entrada expuestos por el runtime y duración observada. Sólo dos pares válidos con reducción estricta de prompt/tokens y sin mayor duración sostienen al candidato en esa muestra. Datos ausentes o resultados peores producen `keep_baseline`. `adopted` permanece `false`: el ensayo no cambia el harness ni concede aprobación.

El experimento registra `execution_provenance_before/after`: versión, path canónico y SHA-256 del CLI, más hashes de los ocho archivos de runner/escenario/instrucciones/renderer conocidos. Verifica estabilidad antes y después de cada caso y de la corrida; un binario reemplazado en el mismo path también invalida la muestra. Compara argv completo salvo los valores temporales de `-C`, `--output-schema` y `-o`; sandbox, modelo, esfuerzo y aislamiento no se normalizan. Los registros antiguos sin esta procedencia quedan no elegibles bajo el comparador nuevo y conservan sus reportes históricos separados.

`comparable` distingue aceptación/procedencia/métricas completas de `eligible`, que además exige mejora. `cases_invoked` cuenta intentos de `execute_case`, no identidades de modelos servidos. Fallos al iniciar el proceso conservan un summary honesto y detienen nuevas llamadas; cualquier caso incompleto invalida su par aunque un comparador guardado declare lo contrario. El hash del CLI no atesta identidad del modelo servido ni evita una mutación hostil concurrente; la configuración interna solicitada sigue sin considerarse verificada por un flag.

La repetición del programa operativo se documenta en [el ensayo de eficiencia](../../docs/codex-efficiency-experiment-2026-10-22.md); no se mezcla con el timeout previo ni adopta la variante por referencia.

## Perfil nativo collaboration

`native.py --live --transport collaboration --hook-delivery inline --output <directorio-nuevo-externo> --codex <ejecutable> --lufy <ejecutable> --model <id> --timeout 180` prueba la API `collaboration` observada. El perfil `legacy` continúa por defecto y sus resultados históricos no cambian.

El perfil nuevo registra la copia del reviewer mediante `agents.reviewer.description` y `agents.reviewer.config_file` sólo en esa invocación, manteniendo `--ignore-user-config`, read-only y never. No copia credenciales ni modifica confianza o configuración global. La compatibilidad de esos overrides se observó en el CLI `0.163.0-alpha.10.2`; otras versiones pueden producir inconclusive.

Solicita un spawn explícito de reviewer, inspección de `list_agents` con resultado vacío, un `followup_task` al mismo target y otra inspección del envelope esperado; `wait_agent` sólo espera. Los callbacks seudonimizados deben conservar orden, identidad, clasificación del payload, fingerprint y nonce de la corrida. No usa el texto del modelo para acreditar transporte ni convierte UUIDs y nombres canónicos por inferencia.

El reporte v2 separa `transport_status`, `role_status`, `lufy_lifecycle_verified` y `cleanup_status`. `native_transport_verified` sólo acredita la secuencia estructurada y correlacionada; el rol solicitado/aceptado no equivale a un `SubagentStart` con identidad coherente. El perfil no expone cierre: `cleanup_status=not_available`, `native_execution_verified=false` y el estado integral sigue inconclusive incluso si pasa transporte. Eso conserva el contrato de cierre completo. Shell, edición y mutaciones observadas siguen siendo failed; errores y formatos desconocidos no pasan.

El normalizador admite aliases concatenados exactos de `collaboration` observados en hooks; no aprende nombres por similitud. Un ack vacío de `followup_task` no acredita ejecución por sí solo: necesita la inspección posterior del contrato válido del mismo agente; un error explícito siempre invalida el transporte. Los resultados históricos no se reescriben. El [reporte del perfil](../../docs/codex-native-collaboration-2026-10-23.md) distingue ensayo automático, auditoría read-only de historia persistida y correcciones posteriores cubiertas offline.
