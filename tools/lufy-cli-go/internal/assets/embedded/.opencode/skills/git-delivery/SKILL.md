---
name: git-delivery
description: Standardize authorized Git delivery for staging relevant files, creating Conventional Commits, pushing the current or a new safe branch, and returning Result Contract evidence to delivery/orchestrator.
---

# Git Delivery

Use this skill from the `delivery` agent after explicit user authorization for commit and push. This skill defines the delivery procedure; `delivery` still owns Git/GH operations, permission prompts, validation gates, and final reporting.

## Source Of Truth

- Follow `AGENTS.md` and `.opencode/policies/delivery.md`; if this skill conflicts with either, the policy wins.
- Do not create PRs unless the user explicitly authorized PR creation in addition to commit/push.
- Do not force push unless the user explicitly requested force push.
- Do not stage unrelated files. If the worktree contains mixed or unclear changes, stop and return `blocked` with the exact file selection needed.
- If `.lufy/config/project.yaml` is unavailable, report workflow limits as `not_available` and continue when no other gate blocks delivery.

## Preflight

1. Inspect branch and worktree:
   - `git status --short`
   - `git branch --show-current`
   - `git rev-parse --abbrev-ref --symbolic-full-name @{u}` when an upstream may exist.
2. Identify relevant files from the validated handoff, user request, diff, OpenSpec change, or explicit file list.
3. If relevant files cannot be separated from unrelated local changes, return `blocked`; do not guess.
4. Run or verify required final validation evidence before commit. Prefer already grouped evidence from `validator`; for this repo's managed assets/Go CLI scope, prefer `scripts/validate.sh` when applicable.
5. Before committing pending PR-bound changes, include `git diff --check origin/develop` unless the target base is explicitly different.

## Branch Rules

- Protected branches are `develop`, `main`, `master`, and `development`.
- If the current branch is protected, create a new branch before staging.
- If the current branch is not protected, use the current branch.
- Default base is `develop` unless the user explicitly requested another base.
- Derive new branch names from the commit type, scope, and summary:
  - Pattern: `<type>/<scope-or-summary-slug>`
  - Example: `fix/clients-commercial-status-filter`
  - Use lowercase ASCII, hyphens, and no trailing punctuation.
  - If the branch exists, append a short unique suffix.
- Set upstream explicitly on first push:
  - `git push -u origin <branch>`
- If upstream already exists, push to the current branch without changing upstream:
  - `git push`

## Commit Template

Use Conventional Commits:

```text
<type>(<scope>): <summary>

<body, optional>

Validation:
- <command> - <result>
```

Commit fields:

- `type`: choose from `fix`, `feat`, `docs`, `test`, `refactor`, `chore`, `ci`, `build`, `perf`, or `revert`.
- `scope`: use a short product or area name such as `lufy`, `clients`, `assets`, `workflow`, `delivery`, `openspec`, or the clearest affected package.
- `summary`: imperative or outcome-focused Spanish when natural; preserve technical identifiers. Keep it concise.
- `body`: omit when the title is enough. Include why/risk notes only when useful.
- Validation lines must only include real commands and observed outcomes.

Default behavior is one new commit containing all relevant files. Split commits only when the user explicitly requested separate commits or the validated handoff clearly separates independent delivery units.

## Staging And Commit

1. Stage only selected relevant paths with explicit `git add <path>...`.
2. Re-check staged scope with `git diff --cached --stat` and, when useful, `git diff --cached --name-only`.
3. If staged scope includes unrelated files, unstage only the wrong paths and return or continue with corrected scope.
4. Create the commit with the Conventional Commit title and optional body.
5. Capture the new commit SHA and subject:
   - `git log -1 --format=%h%x09%s`

## Push Evidence

Before pushing, collect local commits that will be included:

- If upstream exists: `git log --oneline @{u}..HEAD`
- If no upstream exists and the branch was created from `develop`: `git log --oneline origin/develop..HEAD`

After commit and before push, run the committed PR whitespace gate:

- `git diff --check origin/develop...HEAD`

Push according to branch rules. After push, verify upstream/remote state:

- `git rev-parse --abbrev-ref --symbolic-full-name @{u}`
- `git status --short --branch`

## Required Result Contract Fields

Return Result Contract envelope v1 to `delivery`. Include at minimum:

```yaml
schema_version: result-contract/v1
status: delivered | delivery_pending | blocked
executive_summary: <commit/push result in Spanish>
artifacts:
  changed:
    - <staged paths or none>
  referenced:
    - branch: <branch>
    - upstream: <upstream or none>
    - pr: not_created_without_authorization | <url>
evidence:
  commands:
    - command: git status --short
      result: passed | failed | blocked | not_run
      notes: <key output>
    - command: <validation command>
      result: passed | failed | blocked | not_run
      notes: <key output>
    - command: git diff --check origin/develop...HEAD
      result: passed | failed | blocked | not_run
      notes: <key output>
    - command: git push...
      result: passed | failed | blocked | not_run
      notes: <key output>
delivery_package:
  branch: <branch>
  upstream: <origin/branch or none>
  branch_created: true | false
  pr_created: true | false
  pr_authorization: authorized | not_authorized
  commits:
    - short_sha: <short sha>
      subject: <commit subject>
risks:
  - <risk or none>
next_recommended:
  owner: orchestrator | delivery | user | none
  action: <next action>
```

Use `delivery_pending` when commit/push succeeded but PR creation or remote checks are still explicitly pending. Use `delivered` for authorized commit/push-only delivery when no PR was authorized or required. Use `blocked` for missing authorization, branch safety, validation failure, mixed worktree, or push failure.
