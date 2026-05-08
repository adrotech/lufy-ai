# Backlog de mejoras de lufy-ai

Este backlog captura oportunidades de mejora derivadas del reporte externo `lufy-ai-improvement-report.md` sobre arquitectura, supply chain, install state, CI, UX y gobernanza. No es una proposal OpenSpec: cada ola o item puede convertirse luego en una proposal acotada.

## Criterios de priorización

- **P1**: impacto alto en seguridad, trazabilidad, release o bugs latentes que pueden afectar usuarios.
- **P2**: mejora importante de robustez, CI, gobernanza o mantenibilidad.
- **P3**: mejora útil de DX, documentación, UX o deuda menor.

## Ola 1 — Hardening estructural

Objetivo: estabilizar la fuente de verdad de assets, path safety, trazabilidad de install state y escrituras críticas antes de ampliar distribución.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-001` | P1 | Single source of truth de assets | A1, G8 | `hardening-structural-foundation` |
| `BL-002` | P1 | Persistir versión real en `install-state.json` y backups | A3, A5 | `hardening-structural-foundation` |
| `BL-003` | P2 | Fix `EnsureRelativeSafe` para semántica Windows | A2, G3, G9 | `hardening-structural-foundation` |
| `BL-004` | P2 | `copyFile` atómico en install/sync/backup | D1 | `hardening-structural-foundation` |
| `BL-005` | P2 | `SourceRootFingerprint` calculado desde catálogo | A4 | `hardening-structural-foundation` |

## Ola 2 — Supply chain y release

Objetivo: hacer defendible la cadena de release pública con firma, provenance, SBOM y políticas de tagging más seguras.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-006` | P1 | Firmar artifacts y checksums con cosign keyless | B1 | `harden-supply-chain-release` |
| `BL-007` | P1 | Generar SLSA provenance | B1 | `harden-supply-chain-release` |
| `BL-008` | P1 | Pinear GitHub Actions a commit SHA | B2 | `harden-supply-chain-release` |
| `BL-009` | P2 | SBOM por release | B8 | `harden-supply-chain-release` |
| `BL-010` | P1 | Auto-tag con labels para skip/minor/major | E1 | `harden-supply-chain-release` |
| `BL-011` | P2 | Retry/backoff ante race de `next_tag` | E2 | `harden-supply-chain-release` |
| `BL-012` | P2 | Release notes generadas | E7, K6 | `harden-supply-chain-release` |
| `BL-013` | P2 | Reducir permisos de `auto-release-tag.yml` | B5 | `harden-supply-chain-release` |
| `BL-014` | P2 | Sanitizar PR title en tag annotation | B6 | `harden-supply-chain-release` |

## Ola 3 — CI y calidad

Objetivo: mejorar la evidencia automática sin inventar toolchains globales y cubrir plataformas distribuidas.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-015` | P2 | Coverage Go con `-coverprofile` y threshold inicial | G1 | `improve-ci-quality-gates` |
| `BL-016` | P2 | `golangci-lint` con reglas mínimas | G6 | `improve-ci-quality-gates` |
| `BL-017` | P2 | Shellcheck para scripts bootstrap/install | G7 | `improve-ci-quality-gates` |
| `BL-018` | P2 | Matrix CI linux/macOS/Windows para tests/smokes | E8, G3 | `improve-ci-quality-gates` |
| `BL-019` | P2 | E2E real post-release contra GitHub Releases | G2 | `improve-ci-quality-gates` |
| `BL-020` | P3 | Golden tests para output del plan | G5 | `improve-ci-quality-gates` |
| `BL-021` | P3 | Test de runtime para `cmd/lufy-ai/main.go` | F8 | `improve-ci-quality-gates` |

## Ola 4 — Modelo de install state y UX del CLI

Objetivo: hacer el estado instalado más auditable, menos destructivo y más automatizable.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-022` | P1 | Preservar `mcp.engram.timeout`/`enabled` en `opencode.json` | C1 | `improve-install-state-ux` |
| `BL-023` | P2 | Namespace owned en `opencode.json` (`x-lufy-ai`) | C2 | `improve-install-state-ux` |
| `BL-024` | P2 | Registrar writes de `merge-json` en install state | C4 | `improve-install-state-ux` |
| `BL-025` | P3 | Checksum propio de `install-state.json` | C3 | `improve-install-state-ux` |
| `BL-026` | P3 | Lock de concurrencia `.lufy-ai/.lock` | C5 | `improve-install-state-ux` |
| `BL-027` | P2 | Retention de backups | C6 | `improve-install-state-ux` |
| `BL-028` | P2 | `verify` derivado del catálogo | F4 | `improve-install-state-ux` |
| `BL-029` | P2 | Reportar archivos extra en directorios gestionados como info | C7 | `improve-install-state-ux` |
| `BL-030` | P2 | Output JSON estructurado (`--json`) | F3 | `improve-install-state-ux` |
| `BL-031` | P2 | `lufy-ai status` | H2 | `improve-install-state-ux` |
| `BL-032` | P2 | Errores accionables consistentes con `ActionableError` | H1 | `improve-install-state-ux` |

## Ola 5 — Robustez del CLI y bootstrap

Objetivo: mejorar portabilidad, reproducibilidad y UX sin rediseñar el producto.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-033` | P2 | `bootstrap.sh` con retry/timeouts de `curl` | F1 | `improve-bootstrap-robustness` |
| `BL-034` | P2 | Verificar artifacts tar/zip desde Go en vez de shell parsing | B3 | `improve-bootstrap-robustness` |
| `BL-035` | P3 | Warning visible para `--version latest` | B4 | `improve-bootstrap-robustness` |
| `BL-036` | P3 | Evaluar `cobra` para help/autocomplete | F2 | `improve-cli-dx` |
| `BL-037` | P3 | Flags `--quiet` y `--verbose` | F5 | `improve-cli-dx` |
| `BL-038` | P3 | Manejo de `null` en `objectAt` | F6 | `improve-cli-dx` |
| `BL-039` | P3 | No enmascarar errores de `EvalSymlinks` | F7 | `improve-cli-dx` |
| `BL-040` | P3 | Warning para `--target ""` literal | H5 | `improve-cli-dx` |

## Ola 6 — Gobernanza y documentación

Objetivo: preparar el proyecto para contribución externa, reportes de seguridad y documentación de decisiones.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-041` | P2 | `CONTRIBUTING.md` | I1 | `community-governance-docs` |
| `BL-042` | P2 | `SECURITY.md` | I3 | `community-governance-docs` |
| `BL-043` | P2 | `CODEOWNERS` | B7, I4 | `community-governance-docs` |
| `BL-044` | P3 | `CODE_OF_CONDUCT.md` | I2 | `community-governance-docs` |
| `BL-045` | P3 | PR/issue templates | I5 | `community-governance-docs` |
| `BL-046` | P2 | `CHANGELOG.md` o changelog automatizado | I6 | `community-governance-docs` |
| `BL-047` | P3 | ADRs para decisiones de política | I9 | `community-governance-docs` |
| `BL-048` | P2 | `docs/status.md` para matriz feature/estado | I8 | `community-governance-docs` |
| `BL-049` | P3 | Split de README hacia `docs/architecture.md` y `docs/security.md` | I7 | `community-governance-docs` |

## Ola 7 — Observabilidad futura

Objetivo: dejar registrados temas útiles para una etapa posterior, sin bloquear hardening/release.

| ID | Prioridad | Tema | Hallazgos | Propuesta futura |
| --- | --- | --- | --- | --- |
| `BL-050` | P3 | Logging estructurado con niveles y posible JSON | J1 | `improve-observability` |
| `BL-051` | P3 | Métricas de tiempo de operaciones | J2 | `improve-observability` |
| `BL-052` | P2 | Deep verify opcional para plugins/configs | J3, K1 | `improve-observability` |
| `BL-053` | P3 | Check funcional de Engram (`engram --version`) | J4 | `improve-observability` |
| `BL-054` | P3 | ID único de install | K4 | `improve-observability` |

## Orden recomendado para proposals

1. `hardening-structural-foundation`: `BL-001` a `BL-005`.
2. `harden-supply-chain-release`: `BL-006` a `BL-014`.
3. `improve-ci-quality-gates`: `BL-015` a `BL-021`.
4. `improve-install-state-ux`: `BL-022` a `BL-032`.
5. `improve-bootstrap-robustness`: `BL-033` a `BL-035`.
6. `community-governance-docs`: `BL-041` a `BL-049`.

## Notas de aplicación

- No mezclar todas las olas en una sola proposal; cada una toca subsistemas distintos y requiere validación diferente.
- Mantener `develop` como rama de integración y `main` como rama productiva/release.
- Para cada proposal, correr análisis sistémico inicial y validación final agrupada según `AGENTS.md` y `.opencode/policies/delivery.md`.
- Tests y coverage deben ejecutarse solo si existe toolchain real para el alcance de la proposal; si no, reportar evidencia estática/documental.
