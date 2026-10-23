# Eventos explícitos de review en Codex

El host registra el inicio antes de revisar y la conclusión después de revisar. Estos comandos no ejecutan un modelo ni autentican la identidad de un agente. La validación estructural del Result Contract tampoco demuestra la calidad de la review. El reviewer conserva sus permisos de lectura; el host autorizado escribe únicamente metadata local.

```sh
lufy-ai review start --target <repo> --run <id-opaco> --review <intento-opaco> --json
lufy-ai review complete --target <repo> --run <id-opaco> --review <intento-opaco> --result passed --evidence <result-contract.json> --json
lufy-ai context metrics --target <repo> --json
```

Cada intento tiene un solo inicio y una sola conclusión. Un nuevo intento usa otro `--review`. Repetir la misma operación conserva el timestamp observado y devuelve `duplicate_noop`; cambiar su contenido o intentar otra conclusión devuelve conflicto. Los IDs aceptan 2–128 caracteres ASCII alfanuméricos y `._:-`; deben ser opacos y no contener datos personales. Run ID identifica el stream existente; review y operation se guardan hasheados. La clave de idempotencia separa los hashes de cada componente para evitar alias entre IDs que contengan `:`.

`rework` y `reopen` requieren inicio y un Result Contract válido. Cada ocurrencia usa `--operation` estable y distinto; un retry reutiliza ese valor. `complete` exige `--result passed|failed`; un contrato `blocked` o `escalated` no permite `passed`. El ledger debe estar habilitado; no se inicializa una instalación ni se migra su layout desde estos comandos.

Se persisten timestamps reales, nombres semánticos, outcome declarado y fingerprint canónico del contrato. No se guardan cuerpos del contrato, rutas de evidencia, prompts ni variables de entorno. Los eventos son `KindEvidence` sin checkpoint y conservan el estado de workflow. Un run dedicado a telemetría puede seguir sin estado terminal. No autorizan delivery ni vinculan todavía la evidencia a un candidato Git exacto.

`SubagentStart`, `SubagentStop` y `SessionEnd` siguen siendo lifecycle genérico. No se convierten en reviews. Las métricas derivan duración de pares explícitos y sólo reflejan lo registrado; no prueban actividad ausente ni costos o tokens desconocidos.

La validación offline usa el FileStore y el proyector reales: CAS, retries concurrentes, conflictos, privacidad, reloj inválido y preservación de checkpoints. El bloque 4 además instrumentó una review manual del productor en el worktree; la evidencia local distingue esa observación del protocolo nativo inconcluso del bloque 3.
