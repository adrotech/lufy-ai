# Actualización documental del flujo inline

Fecha: 29/10/2026. Base inspeccionada: `origin/develop`, commit `3b6470a71f982f426c2fa17dee013d9f27275878`. Trabajo en `codex/docs-inline-develop`, en el checkout aislado; se conserva el checkout personal con cambios ajenos.

## Resultado

Se reorganizaron las guías alrededor de uso, autonomía, configuración, comandos, memoria/contexto, evidencia, diagnóstico y publicación. Se actualizaron arquitectura, estado, contribuciones, seguridad, CI, ADRs e índices de la investigación. Se corrigieron las referencias de integración pendiente de los PR #256/#257 y el schema del manifest en el asset distribuido.

Se retiraron cuatro páginas: `docs/estado-del-trabajo.md`, `docs/codex-candidate-evidence.md`, `docs/codex-review-events.md` y `docs/codex-runtime-diagnostics.md`. Su información vigente se consolidó en estado, evidencia y diagnóstico. Los informes fechados, sus datos y sus veredictos se conservaron; se añadió navegación que permite interpretar su fecha.

Las cuatro skills de memoria quedaron en español, con sus mirrors exactos. Se corrigieron dos textos de ayuda de la CLI: default Codex y descripción de install. No se cambió comportamiento de comandos, schema, permisos o valores por defecto.

## Validación observada

| Comprobación | Resultado |
| --- | --- |
| `scripts/validate.sh` | Exit 0: 63 pruebas Python offline; Go con coverage 80,6%, vet, build, workflows, versiones y paridad. ShellCheck omitido por no estar instalado. |
| `go test ./internal/cli` y build del binario, después de corregir la ayuda | Exit 0. |
| `scripts/check-harness-coupling.sh` | Payload de 17 archivos editoriales, paridad exacta; 16 assets seleccionables. |
| `scripts/check-doc-release-version.sh --require-changelog` | Exit 0; versión canónica sin modificar. |
| Enlaces Markdown locales | Destinos y anchors comprobados mediante script temporal, incluyendo páginas nuevas y mirrors. |
| Ejemplos CLI en fixtures externos | 33 comandos generales y seis de evidencia/candidate/ledger/review completados con exit 0. |
| Skills traducidas | Frontmatter plano, nombres, descripciones, placeholders y paridad comprobados estáticamente; skills reconocidas por el registry instalado en el fixture. |
| Validador auxiliar `quick_validate.py` | No disponible: el Python local y el runtime auxiliar carecen de PyYAML. No se declara ejecutado correctamente. |
| `git diff --check origin/develop` | Sin errores de whitespace. |

Los smokes observaron dry-run sin crear assets, instalación/verificación, delegación explícita `off` y default `auto`, memoria, contexto, sync, backup/restore y binding sobre candidato limpio. Los intentos iniciales de `plan` sin Git o sin superficies fueron rechazados; esas precondiciones quedaron documentadas. La comprobación de delegación se hizo con JSON de `info`, porque YAML serializa `off` entre comillas.

La evidencia temporal está bajo `/private/tmp/lufy-docs-20261029`: ayudas capturadas, validación, scripts, salidas y fixtures sintéticos. No contiene nuevas ejecuciones de modelos ni modifica la instalación personal. Los commits de fixtures son locales, sin remotos ni entrega del producto.

No se actualizó `RELEASE_VERSION`, no se creó tag, release ni promoción a `main`. La publicación requiere preparar y comprobar un candidato de versión según [la guía vigente](../releasing.md).
