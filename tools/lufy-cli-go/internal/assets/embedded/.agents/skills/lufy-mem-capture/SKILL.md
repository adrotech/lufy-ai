---
name: lufy-mem-capture
description: Capture durable lessons into portable Obsidian memory under .lufy/memory when the repo uses Lufy memory.
---

# Lufy Memory Capture

Use for durable decisions, reusable gotchas, delivery outcomes, or architecture lessons.

1. Check whether `.lufy/config/project.yaml` declares Obsidian memory.
2. Capture only durable knowledge, not routine turn summaries.
3. Use `lufy-ai memory capture --target <repo> --title <title> --type <decision|rule|flow|lesson|concept> [--link <slug>] <text>` instead of hand-editing Markdown when possible.
4. Treat explicit user corrections to AI decisions as durable `rule` or `lesson` memory.
5. Connect related existing notes with `lufy-ai memory connect` and run or recommend `lufy-ai memory validate` when memory files change.
