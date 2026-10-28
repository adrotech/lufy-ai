> Informe histórico del modelo anterior. Sus resultados se conservan; sus comandos y roles no describen la rama inline actual. Ver [estado vigente](status.md).

# Eficiencia de Lufy en Codex — repetición del bloque 7

## Decisión

Se conserva `keep_baseline`, `adopted=false`. El baseline inline pasó; la variante por referencia agotó 180 segundos sin completar. El par no es comparable y el protocolo detuvo nuevas llamadas: dos intentos, un par incompleto, sin segundo par ni retries para obtener verde. No hay evidencia de ahorro de tokens, coste o latencia para adoptar referencias.

## Perfil y observaciones

Se usó la CLI incluida con ChatGPT, `codex-cli 0.163.0-alpha.10.2`, en `/Applications/ChatGPT.app/Contents/Resources/codex-cli/bin/codex`. SHA-256 observado: `50ab38ba21d0d9f8346f32f41848382f15b556190f3c7a07e885a4fb73e379c8`. Se solicitó `gpt-6.1-sol` / `medium`; identidad del modelo servido no verificada. Las instrucciones, runner y binario permanecieron estables antes/después (`provenance_stable=true`).

| Muestra | Prompt bytes | Tokens de entrada observados | Tiempo observado | Aceptación |
| --- | ---: | ---: | ---: | --- |
| Inline | 700 | 18402 | 12856 ms | passed |
| Referencia | 626 | unavailable | 180038 ms hasta timeout | inconclusive |

La primera ejecución fue del 2026-10-22 23:26:30 al 23:26:43 UTC; la segunda, del 23:26:53 al 23:29:53 UTC. El candidato sólo registró inicio de thread/turn, sin `turn.completed` ni decisión estructurada. El timeout terminó su grupo de procesos y conservó evidencia. Tener menos bytes en el prompt no permite inferir reducción de tokens o coste; el tiempo hasta timeout tampoco es una latencia de respuesta completa comparable.

El caso sintético auto-chain exige bloquear implementación sin aprobación post-plan. Las dos variantes usan las mismas instrucciones y fuentes, con entrega inline o por referencia; sus workspaces temporales son distintos. El perfil solicita sandbox workspace-write y desactivar agentes/hooks/plugins/apps sólo en ese proceso. Se evalúa `emulated_cli`, no delegación nativa; no se alteraron credenciales, defaults ni configuración global. Los grupos POSIX y el timeout son guardas del protocolo, no workflow_limits configurados.

## Reproducción y evidencia

```sh
python3 scripts/render-codex-instructions.py --check
python3 tools/codex-behavior-pilot/experiment.py --live --output <directorio-externo-nuevo> --codex /Applications/ChatGPT.app/Contents/Resources/codex-cli/bin/codex --model gpt-6.1-sol --reasoning-effort medium --pairs 2 --timeout 180 --interval 10
```

Evidencia local: `/private/tmp/lufy-seven-b7-efficiency-20261022/summary.json`, con result.json, JSONL, stderr, prompts y fixtures por muestra. Las trazas quedan locales; este reporte incluye únicamente campos sintéticos revisados. La corrida previa descrita en `codex-execution-experiments-2026-10-22.md` falló en su baseline antes del ajuste de aislamiento: no es el primer par de esta repetición. Una nueva corrida requiere una condición diagnóstica distinta y conserva la autorización/protocolo; no se relanza por deseo de un resultado positivo.

El comparador ahora liga cada muestra a hashes/versión del ejecutable y ocho fuentes, conserva perfil argv y reevalúa trazas, acciones, filesystem, finalización y usage. Las pruebas offline reproducen drift, cambio de sandbox/modelo, fallo al iniciar proceso y un segundo par incompleto; ninguna inicia modelos. El resultado negativo completa la evaluación del bloque, sin afirmar una mejora de rendimiento ni convertir una medición en autoridad de workflow.
