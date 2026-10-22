# Verificacion del bloque 1: regresiones offline en CI

## Aceptacion

| Criterio | Evidencia |
| --- | --- |
| B1-AC1 | Filtros pull_request y push incluyen tools/codex-behavior-pilot/**; workflow YAML validado y PR 244 ejecuto Quality y matriz de plataformas |
| B1-AC2 | scripts/validate.sh invoca unittest obligatorio bajo set -euo pipefail; suite agrupada local ejecuto las 68 regresiones y fallo se propaga. Suite offline usa procesos simulados, sin --live/modelos/auth |
| B1-AC3 | Revision directa verifica command -v python3, diagnostico de instalacion y return 1 ante dependencia ausente; Ubuntu instala python3 en Quality, sin convertir ausencia en skip exitoso |

Scope inicial de seis paths: workflow, script, README, plan de siete bloques y codex-behavior-pilot spec/mirror. Plan autorizado cubre alcance/aceptacion/validacion; slice T2 del programa T1, una unidad coherente que une disparador y comando. Mirror exacto comprobado por delivery. No se modificaron permisos, modelos globales ni protocolo live.

## Validacion y sync

Root ejecuto scripts/validate.sh: exit 0, 68 unittest offline, suite Go, coverage 80.4%, vet/build correctos. Shellcheck ausente en entorno local: shell lint omitido y limitacion registrada; CI instala shellcheck/ripgrep/python3. Revision directa del script y workflow conserva diagnosticos y propagacion de error. No se agregan pruebas que solo dupliquen YAML.

OpenSpec strict del change y spec canonica passed tras sync. Assets post-sync passed 0.530s. Guard y whitespace de seis paths iniciales passed antes y despues del commit.

context review anterior al stage reporto escalate, graph not_available, traceability unknown y limites ausentes; observaba tres tracked, no todo el scope de seis. La revision directa conserva volumen real, plan y mirror. No se inventan budgets ni se inicializa grafo para obtener proceed.

## Publicacion inicial y snapshot

Autorizacion global humana vigente para planificar/implementar/validar/entregar los siete bloques en orden. Root actua como relay mutante; delivery mantiene gates. PR https://github.com/adrotech/lufy-ai/pull/244 contra develop, HEAD inicial 961d378a1be2d6ab7cf06b2d68204ed34ee6abe3.

Snapshot propio y root de gh pr view: OPEN/CLEAN y cinco SUCCESS; ultimo check obligatorio Test, build and smoke Go installer completo 2026-10-22T21:50:01Z. Run 39860052104, Quality 1m5s, Ubuntu 33s, macOS 47s, Windows 2m9s e installer 23s. Watch local termino exit 0. El intento de consultar el log detallado encontro error de conexion y la escalacion no devolvio respuesta antes de interrupcion; no se presenta ese log como evidencia obtenida ni se repite la escalacion. El estado completo de los checks si fue confirmado por consulta directa.

Snapshot delivered, elegible para archive; no closed anticipado. La tarea 3.1 significa publicacion/checks iniciales y elegibilidad archive. Con tareas completas, archive se valida y usa skip-specs porque sync ya aplicado. El nuevo HEAD con archive debe pasar cinco checks antes del merge autorizado con match-head-commit, sin admin/force. La prueba GitHub posterior cerrara el bloque sin reescribir este snapshot.

Source B2 esta separado y sin stage; no entra en commit de archive B1 ni en la revision del rango comprometido de este PR. Se conserva checkout padre mixed y dist ajeno.
