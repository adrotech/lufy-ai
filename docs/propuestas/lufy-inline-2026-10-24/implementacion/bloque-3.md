# Bloque 3: evidencia, revisión y capacidades

Estado: implementado y validado en `codex/lufy-inline`, checkout `/private/tmp/lufy-inline-20261026`, sobre `origin/develop` en `1cc18d4647afc61ea1f6b1cb726ba74128fded5f` (`v0.6.25`). La instalación personal no se sustituyó. Trabajo inline, sin subagentes, commit, push o PR.

## Resultado

El producto deja de interpretar `result-contract/v1` como condición para aceptar trabajo, roles o transiciones. El principal puede responder libremente y escoger comprobaciones según el resultado: código, texto, investigación, datos, operación o propuesta.

Los comandos que necesitan evidencia estructurada reciben únicamente comprobaciones y hallazgos declarados. Ese registro es una integración opcional; no es un nuevo envelope general de respuesta ni una máquina de estados. No valida por sí mismo que una afirmación sea verdadera o que una prueba se haya ejecutado.

## Cambios

- Eliminados los paquetes Go `resultcontract` y `resultcontractledger`, el store de receipts, los comandos `result validate/normalize/transition`, su menú y las políticas de salida por rol. No hay normalizador, alias o bridge para el formato retirado.
- Eliminados el registro de instrucciones por contratos de rol y su modelo de `OutputContract`, que dependían del protocolo retirado y no tenían otros consumidores vigentes.
- Retirados los contratos fuente/embedded `result-contract.md` y `result-transition.md`; actualizado el índice de políticas y la guía de registro local.
- Nuevo `internal/evidence`: entrada JSON acotada con `checks` y `findings`, rechazo de duplicados/case aliases/keys desconocidas, observaciones requeridas y fingerprint canónico. Rechaza scores, estados de workflow y envelopes anteriores.
- `evidence inspect` deriva `supported`, `failed`, `inconclusive` o `not_applicable`. Su exit 0 significa que interpretó el archivo; no ejecuta ni acredita una comprobación. Los no aplicables se cuentan aparte.
- `candidate capture/verify` conserva identidad de repositorio, base, commit/tree, paths/modos y políticas; vincula las observaciones sin un claim de rol. El binding actual es `candidate-evidence/v2` y rechaza el anterior. Las políticas vinculadas son trabajo, evidencia, entrega y configuración aplicable. Se conservan las defensas de lectura Git sin efectos laterales.
- `review` registra sólo fingerprints/metadata mediante una operación explícita e idempotente. No admite declarar `passed` con fallos, incertidumbre, sólo no aplicables o hallazgos críticos/altos. Conserva los outcomes inconcluso y no aplicable sin convertirlos en aprobación. La vigencia de un binding no equivale a calidad ni autorización de entrega.
- Lifecycle Codex: sólo orientación `SessionStart`. Eliminados parser de reportes, diagnóstico de gates SDD y registro automático de sesiones/agentes. Los demás eventos se ignoran sin procesar mensajes ni crear registros.
- Retirados `min_evidence_items` y `--evidence-items` de configuración, rescan y revisión contextual. El campo retirado se rechaza, incluso al intentar reemitirlo por `Extra`. No se conserva por compatibilidad.
- TDD y triangulación dejan de ser defaults obligatorios. Se preservan preferencias explícitas del proyecto y el ejecutor puede escribir las pruebas. Las listas de edge cases son orientación, no mínimos universales.
- La revisión contextual conserva observaciones de diff y límites de workload; falta de grafo o trazabilidad se informa sin exigir aprobación humana. Puede recomendar dividir o ajustar concurrencia, sin aprobar calidad ni delivery.
- Añadidos `contracts/evidence.md` y la capacidad `lufy-check`, seleccionados por el catálogo. Guían comprobaciones pertinentes y revisión independiente por utilidad/pedido, sin exigir otro rol o un archivo estructurado. Paridad fuente/embedded comprobada.

## Evidencia

- RED inicial de los tests de evidencia: las funciones no existían. Después de implementar, los casos pasan: defecto crítico, fallo, incertidumbre, no aplicable, identidad de observaciones, ambigüedad y ausencia de filtraciones en errores.
- `GOCACHE=/private/tmp/lufy-inline-go-cache scripts/validate.sh`: PASS, exit 0. Incluye 169 tests Python offline, toda la suite Go, cobertura **80,3%** con mínimo intacto de 80,0%, `go vet`, build, whitespace contra `origin/develop`, guard, pinning, versión/release, YAML, shell, acoplamiento y smoke de hooks.
- Smoke con binario compilado actual: PASS en 14 comprobaciones. Instalación inline y capacidad seleccionada; verify profundo; orientación sin ledger automático; eventos ignorados sin escrituras; observaciones editoriales y factuales; revisión idempotente; crítico impide éxito sin mutaciones; inconcluso/no aplicable no son éxito; comando anterior rechazado; binding sin rol; commit nuevo invalida; observación Git read-only; ledger sin contenido privado.
- Las pruebas de candidatos conservan cobertura de filtros, promisors, submódulos, índices especiales, identidad de repositorio, políticas, modos de archivo y captura estable. No se relajaron para retirar el claim por rol.
- `git diff --check origin/develop`: PASS. Archivos nuevos y paridad de cinco assets del bloque revisados por separado del diff de Git.
- Memoria Obsidian: decisión durable registrada; 14 notas, cero backlinks rotos.

La suite Go con `httptest` se ejecutó con permiso para bind local en loopback. No se ejecutaron modelos nuevos, casos nativos pendientes, publicaciones ni cambios de instalación personal.

## Límites

Los registros estructurados son declaraciones comprobables por el agente/host, no un oráculo de verdad. Las pruebas deterministas demuestran las reglas e integraciones descritas; no establecen todavía mejora empírica del agente entre familias de tareas. Esa evaluación continúa en el bloque 5 y no se reintentaron los casos nativos previamente rechazados.

Los motores/metodologías, instrucciones y comandos históricos ajenos al protocolo retirado continúan con el retiro general del bloque 6. No participan en el catálogo inline nuevo ni justifican conservar el parser eliminado. Los targets temporales generados por el bloque 2 con el mínimo retirado deben recibir configuración actual durante su sustitución puntual; no se migran silenciosamente.

El próximo bloque es la sustitución controlada de la instalación propia, con respaldo/restauración y preservación del trabajo local ajeno. Este informe no declara terminada toda la refactorización.

## Respaldo

- `bloque-3-validation.log`: salida de validación final.
- `bloque-3-smoke.py` y `bloque-3-smoke-result.json`: procedimiento local y resultado integrado.
- `bloques-1-a-3.patch`: código e instrucciones sobre la base develop, con los archivos nuevos de los tres bloques. No se aplicó al checkout original. El reverse-check contra el checkout de implementación confirma que el patch representa ese contenido.
