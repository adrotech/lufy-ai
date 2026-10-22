# Regresión nativa de Lufy en Codex

El bloque 3 añade un ensayo opt-in de transporte nativo y recuperación. Exige reviewer observado, handoff vacío, recuperación sobre el mismo ID, segundo envelope válido y cierre reconocido. Instrucciones, flags y descripciones del modelo no sustituyen esas observaciones.

## Ejecución y aislamiento

```sh
python3 tools/codex-behavior-pilot/native.py --live \
  --output /tmp/lufy-native-evidence \
  --codex /ruta/al/codex --lufy /ruta/al/lufy-ai \
  --model gpt-6.1-sol --reasoning-effort medium --timeout 240
```

El directorio debe ser nuevo y externo al checkout. El runner live requiere POSIX; evaluador/normalizador tienen pruebas portables y los controles de procesos simulados declaran skip fuera de POSIX. CI ejecuta únicamente las pruebas offline.

El perfil predeterminado usa hooks de proyecto. `--hook-delivery inline` permite diagnosticar una capa de proyecto no observada mediante un override TOML de la invocación. Los perfiles se informan por separado; no son un par de rendimiento. Se conserva copia exacta de los hooks del producto y se instrumenta la definición del fixture para observar callbacks y llamar al lifecycle Lufy real. Esto no prueba el wrapper instalado sin instrumentar.

Se solicita sandbox read-only, aprobación `never`, user config ignorada y apps/plugins desactivados, sin afirmar que el inventario haya verificado esas flags. Se conservan autenticación y restricciones heredadas; sólo se eliminan los cuatro handles host del piloto previo. No se usa modo ephemeral.

Los hooks del fixture se revisan y hashean. Su confianza queda limitada a la invocación mediante el flag de automatización documentado, sin persistir confianza ni ampliar sandbox. Hooks normales nuevos o modificados requieren revisión propia. Referencias consultadas el 2026-10-22: [Hooks](https://learn.chatgpt.com/docs/hooks) y [Subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents).

## Evidencia

`events.jsonl` conserva el stream del escenario sintético. `hooks.jsonl` conserva eventos, IDs seudonimizados, clasificación de payload y fingerprint, sin cuerpos de mensajes ni valores de entorno. El envelope sintético se valida con `lufy-ai result validate` antes de consumir cuota; sólo su JSON idéntico se clasifica válido. Esto prueba transporte, no calidad de review productiva.

`result.json` conserva hashes de stream, callbacks, fuentes, observer y envelope, snapshots y procedencia. Los callbacks deben tener el nonce de su corrida. Un error desconocido, un callback ausente o un ID incoherente mantiene `inconclusive`. Una acción prohibida observada o mutación produce `failed`, incluso con gaps de trazabilidad. Ningún resultado avanza gates.

## Observaciones locales

CLI de la aplicación `0.163.0-alpha.10.2`; modelo/esfuerzo solicitados `gpt-6.1-sol/medium`, identidad efectiva del modelo no atestada.

| Intento | Perfil | Observación | Resultado preservado |
| --- | --- | --- | --- |
| A | project | Sin callbacks ni reviewer observado; el modelo declaró bloqueo | inconclusive — native_reviewer_missing |
| B | inline | SessionStart, Stop y SessionEnd observados, los tres con lifecycle Lufy exit 0; sin reviewer ni recuperación observados | inconclusive — error_item_unknown |

B detectó además un warning de ajuste de timeout SessionEnd. El fixture se corrigió a tres segundos y se cubrió offline; esa corrección posterior no se presenta como otro ensayo exitoso. El warning permanece en el artifact original. No se repitió buscando verde sin nueva evidencia.

Artifacts locales: `/private/tmp/lufy-native-regression-20261022-a` y `/private/tmp/lufy-native-regression-20261022-b`. El CLI terminó exit 0 en ambos; el evaluador devolvió exit 1 porque faltaba evidencia. La descripción del modelo no prueba inventario. La cadena nativa completa queda **no verificada en estos perfiles CLI**.

Se recibieron handoffs de roles nativos de la aplicación para pruebas y delivery, con Result Contracts validados. Esa evidencia específica no reemplaza la cadena completa del ensayo CLI ni se usa para marcarla exitosa.
