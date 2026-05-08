# Delivery Policy

Canonical policy for lufy-ai agents, commands, and skills.

## Roles

- `orchestrator` coordinates and routes; must not edit files or run shell commands.
- `explorer` investigates impact and repository context read-only; must not edit files.
- `implementer` implements bounded changes and uses systemic workflow: initial context analysis, no repeated old-file rereads during normal implementation, bounded final reread of changed/affected old files, and validación agrupada at the end of a work block/proposal unless blocked, risky, or diagnosing; must not commit, push, create PRs, or update GitHub Projects.
- `validator` runs compile/test evidence and diagnoses failures read-only; must not edit files.
- `reviewer` reviews only; must not modify files.
- `delivery` owns Git/GH operations, PR creation, project sync, traceability comments, and final validation evidence.

## Branch And PR Rules

- Normal integration branch and default PR base: `develop`.
- Productive/stable branch: `main`; use it only for `develop` → `main` promotion, release, or explicitly authorized hotfix work.
- Normal work opens PRs from feature/fix/chore branches to `develop`.
- Promotion to production opens a PR from `develop` to `main` with release/promotion validation evidence.
- Stable release tags MUST match `v*` and be created from commits reachable from `origin/main`; do not tag unpromoted `develop` commits for stable releases.
- Protected PR source branches: `develop`, `main`, `master`, `development`.
- Never force push unless the user explicitly requests it.
- Report dirty or mixed worktrees before staging.
- With explicit delivery authorization, `delivery` can run `git status`, `git diff`, `git log`, `git add`, `git commit`, `git push`, and `gh` without intermediate prompts.
- Without explicit delivery authorization, return `blocked` with exact recovery instruction.

## Validation Tiers

- **Systemic analysis gate** for `explorer`/`implementer`: analyze existing files, dependencies, interconnections, feedback paths, and structure/behavior impact at the beginning of a coherent block/proposal.
- **Implementation gate** for `implementer`: do not reread old files repeatedly during normal implementation after the initial analysis. Reread old files only when they were modified/affected, conflicts appear, new evidence invalidates the initial analysis, scope changes, a blocker appears, or risk requires confirmation.
- **Block/proposal gate** for `implementer` and `validator`: run grouped validation at the end of all tasks in a coherent block/proposal, including tests and coverage when real commands exist for the scope. Do not run tests constantly during normal implementation.
- **Exception gate**: run focused rereads or validation earlier only when a blocker, risky change, feedback loop, or failure diagnosis requires it.
- **Final PR gate** for `delivery`: run the repository's real full validation suite when available (typecheck/compile, tests, coverage, linting as applicable).
- If change affects behavior, include functional evidence when practical.
- Never claim validation passed without command evidence.
- For this repo, the CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto for that CLI and must not fall back to legacy install paths.

## OpenSpec Task Closure

An OpenSpec task is complete only when:

1. Implementation and local validation evidence.
2. Commit.
3. Push.
4. OpenSpec/artifacts updated, including task checkbox.
5. GitHub Project or issue updated when tracking exists.
6. Issue comment with summary, commit ID/link, continuity, and trace markers.
7. PR when change is 100% complete or user explicitly requests it.

If any item is missing, report `blocked` or `sync_pending` with exact recovery command.

Tasks incompletas always block archive. Do not archive a change with unchecked tasks, even with user confirmation. `migrate-installer-to-go-cli` is explicitly blocked from archive while any tasks remain incomplete.

Current active/focus spec context: `install-managed-assets-with-hash-idempotency` covers managed assets, SHA-256, manifest-driven idempotency, backup/restore, and structural verify.

## Completed Change Gate

- If a change reaches 100% tasks complete, create PR before starting another change.
- Do not begin new change while completed-change PR is pending, unless explicitly authorized.
- Use `.opencode/templates/pr-evidence.md` or skill helpers for PR body.

## GitHub Project Sync

- Use `.opencode/skills/project-sync/project_sync.py` for board sync only when a downstream project installs that optional skill.
- Keep generated content in Spanish.
- Configure Project IDs through environment variables.
- On remote failures, return `sync_pending` with recovery command.

## Language And Format

- Human-facing delivery output, PR bodies, issue comments, and artifacts default to Spanish.
- Preserve technical identifiers, code symbols, filenames, routes, and CLI flags.
- Include ASCII diagrams in PR evidence unless change is documentation-only.
