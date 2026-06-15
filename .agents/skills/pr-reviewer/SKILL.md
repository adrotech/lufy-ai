---
name: pr-reviewer
description: Review a PR or branch and generate the Lufy HTML report in pr_review/ with scoring, desk check, evidence, risks, and final recommendation.
---

# PR Reviewer

Use when the user asks for a PR review, PR audit, branch review, or phrases like `pr review de owner/repo#N` from Codex.

This Codex-visible skill MUST follow the Lufy PR review contract, not a chat-only review. If `.opencode/skills/pr.reviewer/SKILL.md` exists, use it as the canonical detailed contract and follow it. If it is not readable, still follow the minimum contract below.

## Required Output Artifact

- Create `pr_review/` if it does not exist.
- Always write a self-contained HTML report named `pr_review/pr-review-<number>-<yyyyMMdd-HHmm>.html`.
- If the PR has no number, use `pr_review/pr-review-<slug>-<yyyyMMdd-HHmm>.html`.
- Do not overwrite existing reports; add `-2`, `-3`, etc. on collisions.
- Use `.opencode/skills/pr.reviewer/templates/report.html` when available, preserving the Notion-dark style markers: `--navy`, `--navy-deep`, `--surface`, `.gauge`, `.scoregrid`, `.issue`, `.final-summary`.
- The HTML must be self-contained: inline CSS, no external assets, no CDN, no required JavaScript.

## Review Scope

- Stay read-only for source code and GitHub: do not edit code, comment on GitHub, approve, reject, merge, push, or perform delivery.
- You may write only the HTML report under `pr_review/`.
- Collect PR metadata, diff, comments/reviews, checks, and local context when available with read-only commands such as `gh pr view`, `gh pr diff`, `gh pr checks`, `git status`, `git diff --name-only`, and `git diff --stat`.
- If evidence is unavailable, state `No disponible`, `No aplica`, or `Pendiente de confirmar`; do not invent checks, coverage, comments, benchmarks, monitors, or risks.

## Report Content Contract

The HTML report must include, in Spanish:

- Resumen ejecutivo.
- Metadata del PR.
- Veredicto and score from 0 to 100.
- Findings ordered by severity: `CRÍTICO`, `ALTO`, `MEDIO`, `BAJO`, `INFORMATIVO`.
- Buenas prácticas observadas.
- Análisis arquitectónico.
- Seguridad y privacidad.
- Pruebas y evidencia.
- Observabilidad y operación.
- Migraciones/configuración/contratos.
- Desk check y simulación with realistic scenarios and generic layers.
- Comentarios previos no resueltos.
- Action items priorizados.
- Limitaciones del review.
- Resumen final y recomendación.

Scoring weights:

| Dimensión | Peso |
| --- | --- |
| Arquitectura y diseño | 20% |
| Correctitud funcional y contratos | 20% |
| Pruebas y evidencia | 15% |
| Seguridad y privacidad | 15% |
| Observabilidad y operación | 10% |
| Mantenibilidad y complejidad | 10% |
| Desk check | 10% |

Verdict rules:

- `Aprobar`: score >= 80 and no blocking critical/high findings.
- `Pedir cambios`: score >= 50 or at least one correctable critical/high finding.
- `Rechazar`: score < 50, systemic risk, insufficient evidence for a risky change, or multiple critical findings.

## Final Chat Response

Do not paste the full HTML. Report the generated path as a clickable Markdown link and keep the `open` command as a fallback. Respond only with:

```markdown
Reporte generado: [pr_review/pr-review-<...>.html](pr_review/pr-review-<...>.html)
Abrir: `open pr_review/pr-review-<...>.html`

Resumen ejecutivo:
- <máximo 5 bullets>
```
