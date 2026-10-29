# Comprobaciones, candidato y review

El agente comprueba el resultado con fuentes y herramientas pertinentes. La [política distribuida](../.lufy/contracts/evidence.md) define esa responsabilidad. Los comandos de esta guía son integraciones opcionales; una respuesta ordinaria no necesita un archivo JSON.

## Comprobaciones declaradas

Guardá un archivo externo al checkout, por ejemplo `/tmp/comprobaciones.json`:

```json
{
  "checks": [
    {"name": "Comportamiento solicitado", "result": "passed", "evidence": "Prueba observada: exit 0"},
    {"name": "Revisión visual", "result": "not_applicable", "reason": "No cambió una interfaz visual"}
  ],
  "findings": []
}
```

Reemplazá el texto de ejemplo por observaciones reales. Los resultados son `passed`, `failed`, `inconclusive` y `not_applicable`; este último requiere un motivo. Los hallazgos usan severidad `critical`, `high`, `medium` o `low`. No hay score, rol ni estado de workflow.

```sh
lufy-ai evidence inspect --file /tmp/comprobaciones.json
```

Deriva `supported`, `failed`, `inconclusive` o `not_applicable` de lo declarado. Exit 0 significa que interpretó el archivo, incluso si la conclusión es fallida. No ejecuta pruebas ni autentica las afirmaciones. Vigencia del candidato, calidad y autorización se comprueban por separado.

## Vinculación al candidato Git

```sh
lufy-ai candidate capture --target /ruta/al/checkout-limpio --base origin/develop --evidence /tmp/comprobaciones.json > /tmp/candidato.json
lufy-ai candidate verify --target /ruta/al/checkout-limpio --base origin/develop --evidence /tmp/comprobaciones.json --binding /tmp/candidato.json --json
```

Se exige un candidato comprometido limpio, incluyendo untracked no ignorados. La excepción exacta es `.lufy/config/project.yaml` untracked, cuyo contenido se liga por hash; si está tracked y modificado, se rechaza. Guardá el sidecar fuera del checkout. Las políticas `work.md`, `evidence.md` y `delivery.md` deben estar tracked bajo `.lufy/contracts` y ser archivos regulares sin symlinks.

El schema `candidate-evidence/v2` liga identidad local del repositorio, commits/árboles de base y candidato, paths/modos, evidencia y políticas/configuración. Otro commit con el mismo árbol, cambios de base o evidencia y ciertas reubicaciones invalidan la vinculación. No contiene roles ni el contrato retirado.

Las lecturas Git usan argv literal y límites de tiempo/salida. Rechazan flags de índice que ocultan cambios, filtros Git efectivos, gitlinks/submódulos y partial clones/promisors. Desactivan hooks, helpers de diff, textconv, fsmonitor y lazy fetch. No elimines esas restricciones para forzar una captura. Dos observaciones detectan cambios visibles; no bloquean a un mutador hostil concurrente. Verificá nuevamente antes de consumir el sidecar.

Una vinculación vigente puede contener pruebas fallidas. No autentica autoría ni acredita que las comprobaciones se ejecutaron sobre ese commit. Dependencias, archivos ignorados y estado externo quedan fuera del árbol comprometido.

## Eventos explícitos de review

Con el ledger habilitado y un run existente:

```sh
lufy-ai review start --target /ruta/al/proyecto --run run-local --review revision-01 --json
lufy-ai review complete --target /ruta/al/proyecto --run run-local --review revision-01 --result passed --evidence /tmp/comprobaciones.json --json
lufy-ai context metrics --target /ruta/al/proyecto --json
```

`complete` admite `passed`, `failed`, `inconclusive` o `not_applicable` y exige coherencia con la conclusión de las comprobaciones. `passed` requiere una conclusión `supported`: fallos, incertidumbre, ausencia de comprobaciones aplicables o hallazgos críticos/altos lo impiden.

Un intento tiene un inicio y una conclusión. Los retries idénticos devuelven `duplicate_noop`; contenido distinto para la misma operación produce conflicto. `rework` y `reopen` requieren evidencia y un `--operation` estable por ocurrencia. Los IDs deben ser opacos, sin datos personales; se admiten 2–128 caracteres ASCII alfanuméricos y `._:-`.

Se guardan timestamps, nombres semánticos, resultado declarado y fingerprints, sin bodies de evidencia ni prompts. La CLI no ejecuta un modelo ni una review, no autentica un agente y no concede permiso de entrega. Los hooks instalados sólo orientan `SessionStart`: no producen estos eventos. Las métricas reflejan exclusivamente registros existentes. Ver [Run Ledger](run-ledger.md).
