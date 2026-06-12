---
name: lufy-timereport
description: Generate or plan a local time/activity/ROI report for Lufy work with privacy by default.
---

# Lufy Time Report

Use when the user asks for a time report, ROI report, or activity summary.

1. Prefer local, read-only sources and avoid secrets.
2. For OpenCode installs, OpenCode SQLite may be a source when available.
3. For Codex installs, report source limitations unless Codex JSONL/app-server data was provided.
4. Mark missing metrics as unavailable instead of inventing evidence.
