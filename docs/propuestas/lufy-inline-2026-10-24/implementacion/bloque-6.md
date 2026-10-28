# Bloque 6: retiro de motores y formatos anteriores

Estado: implementado y validado; instalación personal actualizada. Integración Git y publicación pendientes de autorización. Trabajo inline, sin subagentes ni nuevas ejecuciones de modelos.

## Resultado

Se retiraron de la fuente los motores OpenSpec/Lufy SDD, sus comandos, parsers, adaptadores de metodología, router y roles instalables, adapters OpenCode/Claude Code y configuraciones asociadas. También se eliminaron los consumidores y pruebas específicos de esos motores, la tarea de sincronización OpenSpec en CI y los experimentos de roles anteriores. El Observatory no fue migrado.

El Context Graph conserva archivos, símbolos, referencias, impacto Git y eventos opcionales. Usa extractores v3; ya no interpreta artifacts/markers SDD ni ofrece `context coverage` o `context trace`. `context review` utiliza el diff directo y los límites declarados; el índice es orientativo. La asignación adaptativa permanece experimental y deshabilitada en el proyecto personal; su fallback refiere al agente principal, sin router instalado.

El snapshot previo del bloque tenía 1.181 archivos: se retiraron 824. El payload embebido contiene 17 archivos editoriales, incluidos los 16 assets instalables y la plantilla AGENTS. No se distribuye el contenido retirado. El inventario de 406 documentos históricos y el snapshot privado conservan las fuentes. Se extrajeron decisiones vigentes en `docs/estado-del-trabajo.md` y ADRs; se actualizaron arquitectura, instalación y estado. Los informes fechados conservan sus resultados con un aviso histórico. No se convirtieron checkboxes históricos en evidencia de cierre: las veinte casillas incompletas pertenecían a una plantilla, no a una iniciativa pendiente del snapshot. El backlog local del checkout original permanece intacto.

## Recuperación y formatos

Configuración schema 2, manifest schema 4 y backup schema 2 con `runtimeStateSchema: 4`. No hay normalizador ni migrador de las generaciones anteriores. Un restore público rechaza generación incompatible, hashes o payload de configuración/estado inválido antes de escribir. El rollback interno conserva preimágenes opacas para deshacer una operación sobre estado dañado. La retención sólo poda backups reconocidos de la generación actual.

La regresión inicial demostró que un backup anterior era aceptado: [RED](bloque-6-backup-red.log). Las pruebas finales cubren rechazo sin escritura, recuperación de estado dañado, rollback, retención y archivos ajenos. La revisión final también retiró dos rutas de conversión del AGENTS.md completo y el alias `@lufy-ia.harness.md`: se rechazan sin conversión; la desinstalación preserva ese texto ajeno. El wrapper consulta `version --runtime-contract` antes de instalar y upgrade exige la misma generación en los checksums. Esa metadata identifica compatibilidad, no constituye una firma de autenticidad.

## Evidencia ejecutada

| Comprobación | Resultado |
| --- | --- |
| `GOCACHE=/private/tmp/lufy-inline-go-cache scripts/validate.sh` | Aprobada: whitespace contra origin/develop, PR guard, 51 pruebas Python offline, pinning, versiones de docs/release, YAML, paridad del payload, tests Go, cobertura **80,6% ≥ 80%**, vet y build. [Log](bloque-6-validation.log). |
| `tools/lufy-cli-go/scripts/smoke-install.sh` | Instalación, repetición, backup y restore aprobados. [Log](bloque-6-smoke-install.log). |
| `tools/lufy-cli-go/scripts/smoke-wrapper.sh` | Preview e instalación aprobados; CLI de generación anterior bloqueada antes de instalar. [Log](bloque-6-smoke-wrapper.log). |
| `python3 /private/tmp/lufy-retirement-20261028/lifecycle.py` | Siete recorridos aprobados: preview read-only, install/repeat/sync, uninstall preservando archivos propios, restore actual, rechazo de config antigua, upgrade actual y bloqueo del anterior. [Resultados](bloque-6-lifecycle-result.json). |
| `python3 /private/tmp/lufy-retirement-20261028/replace_personal.py rehearse` | Drift previo bloqueado, interrupciones antes/después del estado, restore conjunto, reanudación/repetición y restore completo después de sync aprobados. [Resultados](bloque-6-rehearsal-result.json). |
| `python3 /private/tmp/lufy-retirement-20261028/rehearse_revision.py` | Recuperación conjunta desde una revisión parcial conocida aprobada. [Resultado](bloque-6-revision-rehearsal-result.json). |
| Build `GOOS=linux/windows GOARCH=amd64` | Compilación cruzada aprobada; no equivale a ejecución de tests en esos sistemas. |
| `gofmt -l` y `bash -n` para cada script vigente | Sin diferencias/problemas. ShellCheck no estaba instalado y se informó omitido. |
| Revisión de enlaces documentales vigentes | 21 documentos, sin enlaces locales rotos. |
| `git apply --reverse --check bloques-1-a-6-inline.patch` | Aprobado. Patch acumulado de fuente con nuevos archivos incluidos. [Metadata](bloque-6-patch.json). |

Las 51 pruebas Python sustituyen el recuento anterior de 182 porque se eliminaron las pruebas de los motores de roles retirados. Permanecen las de trabajo por familias, auditoría y comparación. No se ejecutaron modelos ni se cambió la conclusión inconclusa del bloque 5.

## Instalación personal

Se reemplazaron conjuntamente tres ejecutables, wrapper y manifest en el target personal. El candidato definitivo tuvo una revisión posterior de ejecutables/metadata tras retirar las últimas rutas de compatibilidad, validada y ensayada antes de aplicarse. El snapshot anterior original se conservó, con postimages exactos para recuperar también una revisión interrumpida.

- Target: `/Users/adrianrojas/Desktop/projects/lufy-ai`.
- Backup conjunto privado: `/Users/adrianrojas/.local/share/lufy-ai/backups/retirement-20261028` (directorio 0700).
- Snapshot: 668 archivos, 14.615.172 bytes comprimidos; incluye ejecutables, configuración, manifest, assets, memoria y proyecciones de la instalación anterior.
- Candidato instalado SHA-256: `306c90c65a6eac53370870773d506d34555426031e586af3bcf1e75266820f64` en las tres rutas.
- Estado final: schema 4, ejecución inline, delegación off, 16 assets, cinco skills, cero drift/missing. `sync` y `verify --deep --quiet` aprobados. [Resultado](bloque-6-live-result.json).
- De 7.868 archivos protegidos se conservaron 7.867 idénticos; el único cambio deliberado al momento del control fue el índice de memoria por una lección nueva. Memoria final: 18 notas, cero backlinks rotos. Después se agregó este informe y se actualizó el índice de la propuesta.
- HEAD `d7f6b090d2c1ec6323ffd6abcb6812bb9847bbba`, rama `codex/feat-adaptive-role-allocation` y staging vacío conservados.

La implementación completa permanece en `/private/tmp/lufy-inline-20261026`, rama `codex/lufy-inline`, y en el patch acumulado adjunto. El código Go, `.opencode`, specs y trabajo local del checkout original se preservaron; no se les aplicó la eliminación masiva de la fuente aislada. Los informes permiten revisar el resultado sin sobrescribir ese trabajo.

## Límites y siguiente frontera

No hay commit, push, PR, integración a develop ni release. La recuperación privada es puntual para este snapshot y target, no un migrador soportado del producto. Los tests se ejecutaron en macOS; Linux y Windows sólo se compilaron. ShellCheck y CI remoto no fueron ejecutados. El resultado acredita el retiro técnico y la instalación personal; no demuestra superioridad universal del nuevo proceso, menor costo humano ni cobertura de todas las familias propuestas.
