# Verificacion B3: regresion nativa observable

## Aceptacion del protocolo

El protocolo solo permite passed con reviewer nativo observado, lifecycle y llamadas de herramientas reconocidas, primer wait con handoff vacio, recuperacion en el mismo ID, segundo wait con envelope exacto validado, cierre y fuentes intactas. Instrucciones, flags y descripciones del modelo no sustituyen esas observaciones.

Falta de capacidad, inventario, hooks o trazabilidad conserva unavailable/inconclusive; nunca emula al reviewer silenciosamente. Mutacion de fuentes o accion prohibida observada produce failed incluso si tambien faltan otros datos. Tipos de item/evento desconocidos y errores no reconocidos degradan a inconclusive. El unico warning permitido es el literal conocido del flag de confianza del fixture revisado; no se permite cualquier error por heuristica.

Observer conserva metadata content-free, IDs seudonimizados, clasificacion/fingerprint y nonce de corrida; envelope sintetico validado antes de live. Reporte conserva hashes de fuentes, observer, fixture hooks, envelope, stream y callbacks, snapshots y correlacion. No se parsean transcripts privados ni se publican streams/outputs completos. Esto prueba transporte cuando hay evidencia completa, no calidad de review productiva ni gates reales.

## Aislamiento y riesgo revisado

Live POSIX opt-in con output nuevo externo, sandbox read-only y approval never. User config ignorada, apps/plugins desactivados, cuatro handles host omitidos del entorno hijo; auth y restricciones heredadas conservadas, sin cambios globales ni modo ephemeral.

Fixture hooks revisados y hasheados; confianza limitada a esta invocacion por --dangerously-bypass-hook-trust y override del fixture trusted, sin persistir confianza ni ampliar sandbox. Se copia la definicion de hooks del producto y se instrumenta el fixture para registrar callbacks y ejecutar lifecycle Lufy real. No prueba el wrapper instalado sin instrumentar. Hooks nuevos/modificados normales requieren revision propia.

## Resultados reales preservados

| Intento | Observacion | Resultado |
| --- | --- | --- |
| A/project | Sin callbacks ni reviewer observado; texto del modelo declara bloqueo | CLI exit0, evaluador exit1/inconclusive native_reviewer_missing |
| B/inline | SessionStart, Stop y SessionEnd observados, cada lifecycle Lufy exit0; sin reviewer ni recovery | CLI exit0, evaluador exit1/inconclusive error_item_unknown |

Artifacts privados locales /private/tmp/lufy-native-regression-20261022-a y -b. CLI app 0.163.0-alpha.10.2; gpt-6.1-sol/medium solicitados, identidad servida no verificada. native_execution_verified no se afirma; cadena completa NO verificada en esos perfiles. Capacidad no expuesta no bloquea entrega del protocolo fail-closed aceptado y no se convierte en pass nativo.

El warning de clamp SessionEnd se conserva en el artifact B. Fix del timeout3s cubierto offline despues, sin atribuir otro ensayo real exitoso ni causalidad. Dos intentos bounded, sin reintentos ciegos. Handoffs nativos de test-writer/delivery de la app son evidencia distinta y no reemplazan la cadena CLI.

## Validacion y scope

scripts/validate.sh exit0 con97 offline, suite Go coverage80.5, vet/build; shell lint local omitido por shellcheck ausente. Hardening posterior Python-only: RED real de unknownitems y GREEN final98 passed0.167s, sin cambios Go posteriores. TDD report inicial validado por root fingerprint98ad857bad0e08f0fad6a6aec4ec8d3ce9ed7b1598d0808d6ce36147a6682413; RED adicionales registrados separadamente.

OpenSpec strict change/canon passed, mirror exacto propio, assets post-sync0.476s passed. Nueve rutas de runner/evaluador/observer/tests/docs/spec/mirror, cubiertas por plan T1 de protocolo transversal/confianza de hooks; una unidad coherente evita evaluar datos sin procedencia. Stage9 exacto, guard9 y whitespace antes/despues commit passed segun root.

context review previo stage: escalate, graph not_available, traceability unknown y limites ausentes; observaba dos tracked, no scope total9. No se inventan budgets; revision directa conserva mirror y workload real. Locales .lufy y activechange fuera del PR; B4 source independiente no entra al archive B3.

## Snapshot de publicacion y gates

PR https://github.com/adrotech/lufy-ai/pull/246 contra develop, HEAD inicial13e15d11ef81c07e11b5afe093d6f7b54d6dcaf6. Root snapshot gh pr view confirma OPEN/CLEAN y cinco SUCCESS en HEAD13e15d11ef81c07e11b5afe093d6f7b54d6dcaf6, run39864770558; Windows2026-10-22T22:26:58Z y obligatorio installer2026-10-22T22:27:18Z.

Autorizacion global humana vigente; root relay ejecuta Git/GH y delivery conserva gates. Snapshot delivered del protocolo, no native passed ni closed anticipado. Task3.1 significa publicacion/checks iniciales y elegibilidad archive. Archive validado con tasks completas y skip-specs por sync ya aplicado; stage solo seis artifacts archive B3.

No usar include-worktree global cuando B4 esta dirty; tras commit se verifica rango origin/develop...HEAD, cinco checks del nuevo HEAD y merge normal con match-head-commit, sin admin/force. La prueba posterior GitHub cierra entrega del protocolo sin reescribir este snapshot ni cambiar los inconclusos live. Padre mixed y dist preservados.
