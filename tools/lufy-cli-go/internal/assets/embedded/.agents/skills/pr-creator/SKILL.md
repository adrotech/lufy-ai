---
name: pr-creator
description: Draft GitHub PR title and body from diff, specs, validation evidence, risks, and traceability without performing delivery.
---

# PR Creator

Use when preparing PR content only.

1. Read the diff and available spec/task context.
2. Include summary, validation evidence, risks, and follow-ups.
3. Include ignored/internal path guard evidence when available: prefer `lufy-ai pr guard --base <base>`; fallback is `git diff --name-only <base>...HEAD -- | git check-ignore -v --no-index --stdin` plus manual review of `openspec/`, `.lufy/`, `.lufy-ai/`, `pr_review/`.
4. If ignored or internal paths are detected, mark the PR content as pending correction or document the user's explicit override. Explain that `.gitignore` does not prevent tracked/cherry-picked files from entering a PR.
5. Do not stage, commit, push, or create the PR.
6. Keep content in Spanish unless the repository template requires otherwise.
