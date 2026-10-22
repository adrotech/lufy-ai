# Verificacion B2: diagnostico de runtime Codex

## Aceptacion y alcance

- Sin --codex, doctor conserva el formato anterior: codex_runtime es opt-in y omitempty; CLI y compatibilidad cubiertos por pruebas.
- Con ejecutable explicito, se invocan literalmente --version y features list, sin shell/modelos, con deadline total de cinco segundos y stdout limitado a 64 KiB por proceso. Stderr descartado y codigos de causa sin salida/errores crudos.
- Solo version parseable y cuatro flags permitidas se conservan; duplicados conocidos invalidan flags. Version/flags permiten available/partial/unavailable, nunca demuestran delegacion nativa. native_delegation queda unknown.
- Activacion de configuracion/memoria/grafo locales preserva archivos anteriores mediante copia exclusiva de cinco archivos nuevos, rechazo de symlinks/preexistentes y sin migrar el layout legacy del checkout principal. Sin modelos, credenciales o permisos globales modificados; sin budgets inferidos.

Nueve paths versionados: plan, doc, servicio/runtime tests, CLI y governance, spec canonica y mirror exacto. Locales .lufy config/mem/state/SDD y activechange no se incluyen. Plan autorizado cubre alcance/aceptacion/validacion; unidad T2 del programa T1, reporte y comando juntos para no entregar una capacidad incompleta.

## Validacion

RED real de overflow y GREEN tras buffer privado por composicion: evita exponer interfaces de bytes.Buffer que salten el limite. Reporte test-writer /tmp/lufy-seven-b2-tests-result.yaml validado. scripts/validate.sh exit 0; 68 offline, suite Go, coverage global 80.5% y runtime 92%, vet/build correctos. Shellcheck ausente localmente y shell lint omitido, declarado como limite. No se repitio la suite por delivery.

OpenSpec change/canon strict passed, mirror exacto propio y assets post-sync passed 0.454s. Stage9, guard9 y whitespace antes/despues del commit passed segun root relay.

Probes repetidos con binario recien construido de la validacion, evidencias /private/tmp/lufy-seven-b2-app-probe-final.json y /private/tmp/lufy-seven-b2-path-probe-final.json: app 0.163.0-alpha.10.2 y PATH 0.144.5, ambos runtime available/native unknown. Doctor exit 1 por manifest ausente se conserva: un runtime disponible no vuelve saludable la instalacion.

Memoria primaria ready con una nota; grafo ready con 1116 indexados, 20 sensibles omitidos y cero errores. Son observaciones de ese checkout, no evidencia del candidato ni de handoff nativo. En el worktree context review devuelve escalate, graph not_available, traceability unknown y limites ausentes; tres tracked antes stage no representan el scope total9. La revision directa mantiene todos los paths/mirror y no inventa budgets ni inicializa grafo para obtener proceed.

## Fallo multiplataforma y correccion

El primer HEAD c5eb0f0 fallo en Windows: las expectativas de tests comparaban literales POSIX con rutas normalizadas por filepath.Abs de Windows (drive de fixture). El runtime productivo no cambio. Se agrego fixtureAbsolute solo a runtime_test.go y se normalizaron cinco expectativas de default/seleccion explicita/argv, conservando la comprobacion de path literal sin shell. Gofmt y go test ./internal/codexruntime -count=1 pasaron en 1.074s segun root. Ese fallo historico no se borra ni se presenta como check exitoso; se exigieron cinco checks del HEAD corregido antes de aceptar la recuperacion; el snapshot posterior confirma su exito.

## Publicacion y snapshot

PR https://github.com/adrotech/lufy-ai/pull/245 contra develop; HEAD inicial c5eb0f08ea0cd34c1f32498fbdbd3498c74df5c7; fix de portabilidad publicado en a000af0e8e4d88078f8b2bd033195bc5d989977d, stage1 y guard/whitespace del rango9 correctos. Root snapshot gh pr view confirma OPEN/CLEAN y cinco SUCCESS del HEAD corregido a000af0e8e4d88078f8b2bd033195bc5d989977d, run 39862670088. Windows completo 2026-10-22T22:10:51Z y obligatorio installer 2026-10-22T22:11:12Z.

Autorizacion humana global de los siete bloques vigente. Root relay ejecuta Git/GH; delivery conserva gates. Snapshot delivered elegible archive, no closed anticipado. Task3.1 significa publicacion/checks iniciales y elegibilidad archive; el HEAD final con archive y merge conservan gates independientes.

Con tasks completas, archive validado usa skip-specs porque sync ya aplicado. Stage solo archive B2; source/documentacion/config local de B3 en el worktree permanecen fuera del PR. No usar guard global include-worktree para validar B2 cuando hay B3 dirty: despues del commit se comprueba rango comprometido origin/develop...HEAD, cinco checks finales y merge normal con match-head-commit, sin admin/force. La prueba posterior GitHub confirma cierre sin reescribir retrospectivamente este snapshot. Checkout padre mixed y dist preservados.
