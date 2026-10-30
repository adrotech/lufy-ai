# Estado del producto

Generación `v0.7.0` del flujo inline, revisada el 29/10/2026. La base de integración incluye [PR #256](https://github.com/adrotech/lufy-ai/pull/256), [PR #257](https://github.com/adrotech/lufy-ai/pull/257), [PR #258](https://github.com/adrotech/lufy-ai/pull/258) y [PR #259](https://github.com/adrotech/lufy-ai/pull/259): refactorización, evaluador, documentación y manual visual. Los artifacts, firmas y estado de publicación se consultan en [la release](https://github.com/adrotech/lufy-ai/releases/tag/v0.7.0).

| Área | Capacidad vigente |
| --- | --- |
| Trabajo inline | Configuración schema 2; evidencia según el resultado y delegación `off`, `on_request` o `auto`. |
| Instalación Codex | Project-local, 16 assets, cinco skills y bloque generado en `AGENTS.md`; manifest schema 4. |
| Recuperación | Backup schema 2 para estado 4, restore validado y rollback acotado cuando existe preimagen. |
| Contexto | Índice lexical v3, impacto Git y evaluación de límites de review. |
| Memoria y skills | Obsidian portable y registry derivado. |
| Evidencia y ledger | Registros explícitos opcionales; no ejecutan pruebas ni conceden permisos. |
| Asignación adaptativa | Experimental, deshabilitada por default; no ejecuta agentes. |
| Evaluador de ensayos | Gramática `command-effects/v2`, diagnósticos por evento y procedencia de helpers. |

## Retiro y decisiones preservadas

Se retiraron los motores SDD/OpenSpec, la selección de metodología por tiers, los adapters OpenCode/Claude Code, sus parsers, roles instalables y Result Contract. El Observatory de OpenCode no se presenta como funcionalidad de Codex. No se mantienen lectores de sus formatos anteriores.

Se conservan instalación por hashes, ownership, conflictos, backup/restore, configuración y planes de superficies. Contexto, memoria, revisión, ledger y colaboración se usan según necesidad. Las veinte casillas incompletas del inventario previo pertenecían a una plantilla del motor retirado: no son trabajo implementado ni un backlog nuevo. Los checkboxes históricos tampoco prueban cierre o entrega.

## Evidencia y límites

Los dos PRs integrados concluyeron con cinco checks remotos aprobados cada uno; ver [trazabilidad de entrega](propuestas/lufy-inline-2026-10-24/entrega.md). La mejora del evaluador registró 63 pruebas Python offline y cobertura Go 80,6%. Esos resultados corresponden a esos candidatos; una modificación posterior requiere comprobar lo que afecte.

El estudio comparativo completó 24 turnos con fixtures de seis familias. No demuestra superioridad general ni menor costo humano; tampoco cubre una app desde cero, seguridad, migraciones u operaciones productivas. La lectura offline posterior reconoció 13 comandos antes desconocidos y dejó 20 inconclusos entre 79 comandos completados. No reejecutó modelos ni cambió los veredictos históricos.

La instalación personal fue sustituida y verificada por separado del código distribuido. Su política `off`, sus backups y sus notas no son defaults ni assets públicos. La [historia](history/README.md) conserva condiciones y límites de cada informe.

## Publicación y seguimiento

La versión canónica vive en `RELEASE_VERSION`. Su publicación requiere versión y changelog coherentes, validación del candidato y promoción autorizada a `main`; el workflow crea el tag y publica artifacts con checksums, SBOM, provenance y firmas, y verifica la descarga publicada. Un tag por sí solo no demuestra que ese proceso terminó.

Para una nueva versión, seguir [la guía de releases](releasing.md). La instalación personal conserva ejecutable y estado propios; publicar una release no la actualiza automáticamente.
