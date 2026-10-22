# Verificación B6: instrucciones Codex coherentes

Estado del snapshot `delivered` para publicación inicial y elegibilidad de archive; merge final pendiente. PR https://github.com/adrotech/lufy-ai/pull/249, HEAD `1883e0e3e093e00bf76364a6dfaaa150d6c5fdd7`, base fresca `73c6c90bd74b350156061f350d06874c3c016653`.

Diecinueve rutas source/mirror/CI cubiertas por plan y manifest `/private/tmp/lufy-seven-b6-paths.json`. Autoridad en delivery.md; wording compartido canónico y tres roles inline conservados, permisos particulares intactos. Renderer check por defecto, write explícito con validación de todas las rutas y markers, drift/paridad observables. Mapping distingue delegación explícita de identidad inline, hooks configurados de capacidades observadas y cleanup no disponible.

Root confirma scripts/validate.sh exit0 (`/private/tmp/lufy-seven-b6-validation-ci-final.log`), 105 offline0.185s, Go80.5%, vet/build; shellcheck ausente localmente. Siete fixtures GREEN0.024s tras RED real de script faltante, preservación CRLF y filtros CI de source/renderer; assets post-sync0.601s y OpenSpec change/canonical strict passed. Handoff validator válido fingerprint75b2e0a84886a7f95a8aa89f59fd5493a3d6c6d0258f4af495f1576d16ce2f92, fase principal sin claim de delegación nativa.

Review dirigido de source y rerelecturas de dos archivos afectados en cada fix; seis pares Codex y spec/mirror idénticos. Sin findings pendientes. La distribución textual no prueba conducta, aislamiento, tokens ni latencia; baseline inline se conserva hasta B7. Graph no disponible, memory ready sin matches; fallback dirigido y sin presupuestos inventados. Write no promete transacción ante IO fallido o mutadores concurrentes.

Root relay confirmó stage19 exacto, guards/whitespace y commit/push; PR249 adjuntado. Source B7, activechange y stores locales quedan fuera.

INITIAL_CHECKS_SNAPSHOT: root confirmó snapshot fresco del 2026-10-22, HEAD `1883e0e3e093e00bf76364a6dfaaa150d6c5fdd7` OPEN/CLEAN, run `39870649647`, cinco SUCCESS: Quality23:17:40Z, Ubuntu23:16:48Z, macOS23:17:17Z, Windows23:19:04Z y Smoke23:19:23Z. Publicación inicial elegible para archive; sin afirmar merge todavía.

Cinco SUCCESS/CLEAN del HEAD inicial permiten tarea de publicación inicial y archive6 específico con skip-specs por sync previo, validación habilitada. El nuevo HEAD archive requiere guard committed25, whitespace rango y otros cinco SUCCESS/CLEAN antes de merge exacto autorizado. Snapshot previo al merge no afirma cierre; GitHub MERGED/merge SHA deben acreditarlo después.
