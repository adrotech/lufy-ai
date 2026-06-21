---
name: lufy-mem-connect
description: Connect existing Lufy Obsidian memory notes with safe backlinks and avoid broken links.
---

# Lufy Memory Connect

Use when memory notes need links, map updates, or relationship cleanup.

1. Search existing `.lufy/memory` notes before adding links.
2. Prefer `lufy-ai memory connect --target <repo> [--bidirectional] <from-slug> <to-slug>` for safe links.
3. Rebuild backlinks with `lufy-ai memory index` if relationships were edited manually.
4. Do not invent notes or stronger evidence than repo files/commands.
5. Validate memory structure after edits.
