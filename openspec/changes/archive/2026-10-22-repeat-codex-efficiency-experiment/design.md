# Diseño

Mantener `pilot.execute_case` y el caso auto-chain: el usuario sintético no aprobó post-plan, por lo tanto el resultado válido es blocked sin editar. El ensayo sigue siendo emulated_cli y no verifica delegación nativa ni identidad del modelo servido. Pedir gpt-6.1-sol/medium por invocación no cambia defaults globales.

El experimento captura procedencia antes y después del proceso: path canónico, hash binario y versión observada, más hashes de experiment.py/pilot.py/scenarios.json, rol implementer, delivery.md, mapping, fuente común y renderer. El registro JSON mantiene esa metadata junto con su evidencia existente; ausencia o diferencia hace la muestra no elegible. Una comprobación global antes de nuevas llamadas evita continuar tras drift. No se firma autoría ni se toma un lock contra cambios hostiles concurrentes.

El perfil argv normalizado elimina únicamente valores de -C/--output-schema/-o, preservando orden y demás flags. Un sandbox/flag/modelo distinto invalida el par. El CLI de versión se ejecuta sin shell con timeout; no inicia modelos. Hashes/procedencia quedan locales y se resumen para el reporte compartible. El runner no modifica Lufy productivo ni reduce sus instrucciones por un resultado positivo.

Hasta dos pares, orden baseline/candidato y luego candidato/baseline. Cualquier timeout/error/resultado inconcluso o deriva detiene nuevas llamadas. El primer fallo no se repite para obtener pass. La muestra pequeña no prueba coste ni rendimiento global; tokens de entrada y duración observados no son coste facturado ni latencia aislada del modelo. Evidencia previa incompleta se conserva separada.
