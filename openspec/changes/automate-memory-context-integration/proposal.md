# Automate memory/context integration

## Why

Memory and context graph assets exist, but important integration signals depend too much on agents remembering skill instructions. Projects need visible health checks, best-effort OpenCode lifecycle wiring, and cleaner context query signal.

## What Changes

- Add project-level `context_graph.exclude` defaults for LUFY managed-state backups and ancestors.
- Exclude `.lufy/managed-state/backups/**` and `.lufy/managed-state/ancestors/**` from context graph discovery by default.
- Add an OpenCode local plugin that runs memory orientation on session creation and memory validation after memory edits on a best-effort basis.
- Extend `doctor` and `verify --deep` to report memory/context/hook integration status and recovery commands.
- Document lifecycle hooks and context graph exclusions.

## Impact

- Affects Go CLI config, context graph, doctor/verify reporting, managed OpenCode assets and documentation.
- Does not modify private `.lufy/memory` note contents.
- Does not require network services, LLMs or embedding providers.
