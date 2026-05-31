#!/usr/bin/env python3
"""Generate a local, sanitized LUFY time report as self-contained HTML."""

from __future__ import annotations

import argparse
import datetime as dt
import html
import json
import os
from pathlib import Path
import re
import sqlite3
import subprocess
import sys
import tempfile
from typing import Any

AI_GAP_CAP_SECONDS = 5 * 60
HUMAN_GAP_CAP_SECONDS = 10 * 60
DEFAULT_DB = Path.home() / ".local/share/opencode/opencode.db"
MONTHS_ES = ("ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic")

SENSITIVE_WORDS = ("session_diff", "diff --git", "BEGIN SECRET", "FULL_TOOL_PAYLOAD", "ASSISTANT_OUTPUT_SECRET", "USER_PROMPT_SECRET")


def main() -> int:
    args = parse_args()
    target_dir = Path(args.target_dir).expanduser().resolve()
    output = Path(args.output).expanduser() if args.output else Path(tempfile.gettempdir()) / f"lufy-timereport-{dt.datetime.now().strftime('%Y%m%d-%H%M%S')}.html"
    db_path = Path(args.db).expanduser()

    report = build_report(target_dir, db_path, args.since, args.until, args.scope, args.session_id, args.tier, args.change)
    body = render_report(report, target_dir, db_path)
    template = (Path(__file__).resolve().parents[1] / "templates/report.html").read_text(encoding="utf-8")
    html_doc = template.replace("{{BODY}}", body)

    assert_no_sensitive_fixture_tokens(html_doc)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(html_doc, encoding="utf-8")
    print(f"Reporte LUFY generado: {output}")
    if report["limitations"]:
        print("Limitaciones: " + "; ".join(report["limitations"]))
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Genera un reporte LUFY de tiempo/ROI en HTML autocontenido.")
    parser.add_argument("--output", help="Ruta del HTML de salida. Default: /tmp/lufy-timereport-<timestamp>.html")
    parser.add_argument("--target-dir", default=os.getcwd(), help="Repositorio/directorio objetivo. Default: cwd")
    parser.add_argument("--db", default=str(DEFAULT_DB), help="SQLite OpenCode. Default: ~/.local/share/opencode/opencode.db")
    parser.add_argument("--from", dest="since", help="Fecha inicial opcional para filtrar Git/actividad")
    parser.add_argument("--to", dest="until", help="Fecha final opcional para filtrar Git/actividad")
    parser.add_argument("--scope", choices=("task", "repo"), default="task", help="Alcance del reporte. Default: task; usar repo para actividad completa del repositorio.")
    parser.add_argument("--session-id", help="Sesión OpenCode a usar como ancla. Si es subagente, se reporta el árbol de la sesión raíz.")
    parser.add_argument("--tier", choices=("T1", "T2", "T3"), help="Tier original de la tarea cuando se conoce.")
    parser.add_argument("--change", help="OpenSpec/LUFY SDD change id asociado cuando existe.")
    return parser.parse_args()


def build_report(target_dir: Path, db_path: Path, since: str | None, until: str | None, scope: str, session_id: str | None, tier: str | None, change: str | None) -> dict[str, Any]:
    limitations: list[str] = []
    activity = collect_opencode_activity(db_path, target_dir, since, until, scope, session_id, tier, change, limitations)
    git_since, git_until = git_bounds(since, until, activity)
    git = collect_git_metrics(target_dir, git_since, git_until, limitations)
    stack = collect_stack(target_dir, limitations)
    metrics = calculate_time_metrics(activity, limitations)
    phases = infer_phases(activity)
    steps = build_step_timeline(activity)
    return {"activity": activity, "git": git, "stack": stack, "metrics": metrics, "phases": phases, "steps": steps, "limitations": limitations, "since": since, "until": until, "git_since": git_since, "git_until": git_until}


def collect_opencode_activity(db_path: Path, target_dir: Path, since: str | None, until: str | None, scope: str, session_id: str | None, tier: str | None, change: str | None, limitations: list[str]) -> dict[str, Any]:
    empty = {"sessions": [], "messages": [], "parts": [], "events": [], "tool_counts": {}, "subagents": {}, "skills": {}, "timestamps": [], "scope": default_scope(scope, tier, change)}
    if not db_path.exists():
        limitations.append(f"OpenCode SQLite no disponible en {db_path}; métricas OpenCode marcadas como No disponible.")
        return empty
    try:
        connection = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
        connection.row_factory = sqlite3.Row
    except sqlite3.Error as exc:
        limitations.append(f"No se pudo abrir SQLite OpenCode en modo read-only: {exc}.")
        return empty

    with connection:
        tables = get_tables(connection)
        for table in ("project", "workspace", "session", "message", "part", "event"):
            if table not in tables:
                limitations.append(f"Tabla OpenCode `{table}` no disponible; las métricas relacionadas degradan.")
        all_sessions = collect_sessions(connection, tables, target_dir, limitations)
        sessions, scope_info = select_scope_sessions(all_sessions, scope, session_id, tier, change, target_dir, limitations)
        session_ids = {string_value(row, ("id", "sessionID", "session_id")) for row in sessions}
        session_ids.discard(None)
        messages = collect_child_rows(connection, tables, "message", session_ids, limitations)
        parts = collect_child_rows(connection, tables, "part", session_ids, limitations)
        events = collect_child_rows(connection, tables, "event", session_ids, limitations)

    timestamps = filter_timestamps([*timestamps_from_rows(sessions), *timestamps_from_rows(messages), *timestamps_from_rows(parts), *timestamps_from_rows(events)], since, until)
    tool_counts = count_tools(parts, events)
    skills = count_skill_names([*parts, *events])
    methodology_info = detect_methodology(target_dir, change, skills, limitations)
    if not tier:
        scope_info["tier"] = infer_tier(change, methodology_info)
    scope_info.update({
        "session_count": len(sessions),
        "time_from": format_bound(min((ts for ts, _ in timestamps), default=None)),
        "time_to": format_bound(max((ts for ts, _ in timestamps), default=None)),
        "methodology": methodology_info,
    })
    return {
        "sessions": sessions,
        "messages": messages,
        "parts": parts,
        "events": events,
        "tool_counts": tool_counts,
        "subagents": count_names([*sessions, *messages, *parts], ("agent", "subagent", "agent_name", "mode")),
        "skills": skills,
        "timestamps": timestamps,
        "scope": scope_info,
    }


def default_scope(scope: str, tier: str | None, change: str | None) -> dict[str, Any]:
    return {
        "mode": scope,
        "anchor_session_id": "No disponible",
        "root_session_id": "No disponible",
        "root_title": "No disponible",
        "root_slug": "No disponible",
        "session_count": 0,
        "tier": tier or "No indicado",
        "tier_source": "explícito" if tier else "no indicado",
        "change": change or "No indicado",
        "methodology": {"id": "No detectada", "mode": "none", "change": change or "No indicado", "source": "no disponible"},
        "time_from": "No disponible",
        "time_to": "No disponible",
    }


def get_tables(connection: sqlite3.Connection) -> set[str]:
    rows = connection.execute("SELECT name FROM sqlite_master WHERE type='table'").fetchall()
    return {row[0] for row in rows}


def columns(connection: sqlite3.Connection, table: str) -> list[str]:
    return [row[1] for row in connection.execute(f"PRAGMA table_info({quote_ident(table)})").fetchall()]


def collect_sessions(connection: sqlite3.Connection, tables: set[str], target_dir: Path, limitations: list[str]) -> list[dict[str, Any]]:
    if "session" not in tables:
        return []
    cols = columns(connection, "session")
    safe_cols = safe_select_columns(cols)
    rows = [sanitize_structural_row(row) for row in select_rows(connection, "session", safe_cols)]
    directory_cols = [col for col in cols if col.lower() in {"directory", "path", "cwd", "root", "project_path", "workspace_path"}]
    target = str(target_dir)
    if directory_cols:
        filtered = [row for row in rows if any(path_matches(row.get(col), target) for col in directory_cols)]
        if filtered:
            return filtered
        limitations.append("No se encontraron sesiones OpenCode para el directorio objetivo; métricas de sesión pueden estar vacías.")
        return []
    limitations.append("Schema de `session` sin columna de directorio; no se mezclan sesiones sin filtro verificable.")
    return []


def select_scope_sessions(all_sessions: list[dict[str, Any]], scope: str, session_id: str | None, tier: str | None, change: str | None, target_dir: Path, limitations: list[str]) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    scope_info = default_scope(scope, tier, change)
    if scope == "repo":
        scope_info.update({"root_title": "Todas las sesiones del repositorio", "root_slug": "repo", "session_count": len(all_sessions)})
        return all_sessions, scope_info
    if not all_sessions:
        limitations.append("No hay sesiones OpenCode para determinar la tarea actual.")
        return [], scope_info

    by_id = {session_identifier(row): row for row in all_sessions if session_identifier(row)}
    if session_id:
        anchor = by_id.get(session_id)
        if not anchor:
            limitations.append(f"Session id solicitado no encontrado para {target_dir}: {session_id}.")
            return [], scope_info
    else:
        anchor = latest_root_session(all_sessions)
        if not anchor:
            limitations.append("No se pudo inferir la sesión raíz de la tarea; se usa la sesión más reciente disponible.")
            anchor = max(all_sessions, key=session_sort_ts)

    root = root_session(anchor, by_id, limitations)
    root_id = session_identifier(root)
    included = descendants(root_id, all_sessions) if root_id else [root]
    scope_info.update({
        "anchor_session_id": session_identifier(anchor) or "No disponible",
        "root_session_id": root_id or "No disponible",
        "root_title": string_value(root, ("title", "name", "summary")) or "Sin título",
        "root_slug": string_value(root, ("slug",)) or "Sin slug",
        "tier": tier or infer_tier(change, {}),
        "tier_source": "explícito" if tier else "inferido",
    })
    return included, scope_info


def latest_root_session(sessions: list[dict[str, Any]]) -> dict[str, Any] | None:
    roots = [row for row in sessions if not parent_identifier(row)]
    candidates = roots or sessions
    return max(candidates, key=session_sort_ts) if candidates else None


def root_session(anchor: dict[str, Any], by_id: dict[str, dict[str, Any]], limitations: list[str]) -> dict[str, Any]:
    current = anchor
    seen: set[str] = set()
    while True:
        current_id = session_identifier(current)
        parent_id = parent_identifier(current)
        if not parent_id:
            return current
        if parent_id in seen:
            limitations.append("Ciclo detectado en parent_id de sesiones OpenCode; se corta el ascenso de árbol.")
            return current
        seen.add(parent_id)
        parent = by_id.get(parent_id)
        if not parent:
            limitations.append(f"Parent session no disponible en el target para {current_id}; se usa la sesión ancla como raíz.")
            return current
        current = parent


def descendants(root_id: str | None, sessions: list[dict[str, Any]]) -> list[dict[str, Any]]:
    if not root_id:
        return []
    included_ids = {root_id}
    changed = True
    while changed:
        changed = False
        for row in sessions:
            row_id = session_identifier(row)
            parent_id = parent_identifier(row)
            if row_id and parent_id in included_ids and row_id not in included_ids:
                included_ids.add(row_id)
                changed = True
    return [row for row in sessions if session_identifier(row) in included_ids]


def session_identifier(row: dict[str, Any]) -> str | None:
    return string_value(row, ("id", "sessionID", "session_id", "sessionId"))


def parent_identifier(row: dict[str, Any]) -> str | None:
    return string_value(row, ("parent_id", "parentID", "parentId", "parent"))


def session_sort_ts(row: dict[str, Any]) -> float:
    values = [ts for ts, _ in timestamps_from_rows([row])]
    return max(values) if values else 0


def collect_child_rows(connection: sqlite3.Connection, tables: set[str], table: str, session_ids: set[str], limitations: list[str]) -> list[dict[str, Any]]:
    if table not in tables or not session_ids:
        return []
    cols = columns(connection, table)
    safe_cols = safe_select_columns(cols)
    session_col = next((col for col in cols if col in ("sessionID", "session_id", "sessionId", "aggregate_id")), None)
    if not session_col:
        limitations.append(f"Tabla `{table}` sin columna sessionID/session_id/aggregate_id; sección relacionada degradada.")
        return []
    placeholders = ",".join("?" for _ in session_ids)
    sql = f"SELECT {', '.join(quote_ident(c) for c in safe_cols)} FROM {quote_ident(table)} WHERE {quote_ident(session_col)} IN ({placeholders})"
    return [sanitize_structural_row(dict(row)) for row in connection.execute(sql, sorted(session_ids)).fetchall()]


def safe_select_columns(cols: list[str]) -> list[str]:
    deny = {"text", "prompt", "output", "input", "content", "body", "diff", "session_diff", "arguments", "payload"}
    allowed = [col for col in cols if col.lower() not in deny]
    return allowed or [cols[0]]


def select_rows(connection: sqlite3.Connection, table: str, cols: list[str]) -> list[dict[str, Any]]:
    sql = f"SELECT {', '.join(quote_ident(c) for c in cols)} FROM {quote_ident(table)}"
    return [dict(row) for row in connection.execute(sql).fetchall()]


def sanitize_structural_row(row: dict[str, Any]) -> dict[str, Any]:
    data = parse_json_object(row.get("data"))
    if data:
        copy_structural_value(row, data, "role")
        copy_structural_value(row, data, "type")
        copy_structural_value(row, data, "agent")
        copy_structural_value(row, data, "mode")
        copy_structural_value(row, data, "tool")
        copy_structural_value(row, data, "command")
        copy_structural_value(row, data, "name")
        if isinstance(data.get("time"), dict):
            row.setdefault("created", data["time"].get("created"))
            row.setdefault("updated", data["time"].get("updated"))
            row.setdefault("completed", data["time"].get("completed"))
        if isinstance(data.get("state"), dict):
            state = data["state"]
            if isinstance(state.get("time"), dict):
                row.setdefault("start", state["time"].get("start"))
                row.setdefault("end", state["time"].get("end"))
            state_input = state.get("input")
            if isinstance(state_input, dict):
                name = state_input.get("name")
                if isinstance(name, str) and name.strip():
                    row.setdefault("skill", name.strip())
    row.pop("data", None)
    return row


def parse_json_object(value: Any) -> dict[str, Any]:
    if not isinstance(value, str) or not value.strip().startswith("{"):
        return {}
    try:
        parsed = json.loads(value)
    except json.JSONDecodeError:
        return {}
    return parsed if isinstance(parsed, dict) else {}


def copy_structural_value(row: dict[str, Any], data: dict[str, Any], key: str) -> None:
    value = data.get(key)
    if isinstance(value, str) and value.strip():
        row.setdefault(key, value.strip())


def quote_ident(value: str) -> str:
    return '"' + value.replace('"', '""') + '"'


def path_matches(value: Any, target: str) -> bool:
    if not isinstance(value, str):
        return False
    try:
        return str(Path(value).expanduser().resolve()) == target
    except OSError:
        return value == target


def timestamps_from_rows(rows: list[dict[str, Any]]) -> list[tuple[float, str]]:
    result: list[tuple[float, str]] = []
    for row in rows:
        role = str(row.get("role") or row.get("type") or row.get("event") or "activity").lower()
        for key in ("created", "updated", "created_at", "updated_at", "time_created", "time_updated", "completed", "completed_at", "timestamp", "time", "start", "end"):
            value = row.get(key)
            result.extend((ts, role) for ts in parse_timestamp_value(value))
    return result


def parse_timestamp_value(value: Any) -> list[float]:
    if value is None:
        return []
    if isinstance(value, (int, float)):
        return [normalize_epoch(float(value))]
    if isinstance(value, str):
        stripped = value.strip()
        if not stripped:
            return []
        if stripped.startswith("{"):
            try:
                data = json.loads(stripped)
            except json.JSONDecodeError:
                return []
            result: list[float] = []
            for nested in ("created", "updated", "completed", "start", "end"):
                result.extend(parse_timestamp_value(data.get(nested)))
            return result
        try:
            return [normalize_epoch(float(stripped))]
        except ValueError:
            try:
                return [dt.datetime.fromisoformat(stripped.replace("Z", "+00:00")).timestamp()]
            except ValueError:
                return []
    return []


def normalize_epoch(value: float) -> float:
    return value / 1000 if value > 10_000_000_000 else value


def filter_timestamps(values: list[tuple[float, str]], since: str | None, until: str | None) -> list[tuple[float, str]]:
    since_ts = parse_bound(since)
    until_ts = parse_bound(until)
    return sorted((ts, role) for ts, role in values if (since_ts is None or ts >= since_ts) and (until_ts is None or ts <= until_ts))


def parse_bound(value: str | None) -> float | None:
    if not value:
        return None
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp()
    except ValueError:
        return None


def git_bounds(since: str | None, until: str | None, activity: dict[str, Any]) -> tuple[str | None, str | None]:
    if since or until:
        return since, until
    if activity.get("scope", {}).get("mode") != "task":
        return None, None
    timestamps = [ts for ts, _ in activity.get("timestamps", [])]
    if not timestamps:
        return None, None
    return format_git_bound(min(timestamps)), format_git_bound(max(timestamps))


def format_git_bound(timestamp: float) -> str:
    return dt.datetime.fromtimestamp(timestamp, dt.timezone.utc).isoformat()


def format_bound(timestamp: float | None) -> str:
    if timestamp is None:
        return "No disponible"
    return format_datetime_human(dt.datetime.fromtimestamp(timestamp, dt.timezone.utc))


def format_datetime_human(value: dt.datetime) -> str:
    local = value.astimezone()
    month = MONTHS_ES[local.month - 1]
    timezone = local.tzname() or "local"
    return f"{local.day:02d} {month} {local.year}, {local.hour:02d}:{local.minute:02d} {timezone}"


def format_range_text(start: Any, end: Any) -> str:
    return f"{format_maybe_datetime(start)} → {format_maybe_datetime(end)}"


def format_maybe_datetime(value: Any) -> str:
    if value is None:
        return "sin límite"
    if isinstance(value, (int, float)):
        return format_bound(float(value))
    if isinstance(value, str):
        parsed = parse_bound(value)
        if parsed is not None:
            return format_bound(parsed)
        return value
    return str(value)


def detect_methodology(target_dir: Path, change: str | None, skills: dict[str, int], limitations: list[str]) -> dict[str, str]:
    if change:
        found = find_change_path(target_dir, change)
        if found:
            return {"id": "openspec", "mode": "full/lite no inferido", "change": change, "source": str(found.relative_to(target_dir))}
        limitations.append(f"Change id indicado no encontrado en openspec/changes: {change}.")
        return {"id": "openspec", "mode": "indicado", "change": change, "source": "argumento --change"}
    skill_names = set(skills)
    if any(name.startswith("openspec-") or name.startswith("opsx-") for name in skill_names):
        return {"id": "openspec", "mode": "inferido", "change": "No disponible; pasar --change para precisión", "source": "skills detectados"}
    if any(name.startswith("lufy.sdd") or name.startswith("lufy-sdd") for name in skill_names) or (target_dir / "lufy-sdd").exists() or (target_dir / ".lufy-sdd").exists():
        return {"id": "lufy-sdd", "mode": "inferido", "change": "No disponible", "source": "señales locales"}
    return {"id": "none", "mode": "none", "change": "No aplica", "source": "sin señales SDD"}


def find_change_path(target_dir: Path, change: str) -> Path | None:
    base = target_dir / "openspec" / "changes"
    direct = base / change
    if direct.exists():
        return direct
    archive = base / "archive"
    if archive.exists():
        for path in archive.rglob(change):
            if path.is_dir():
                return path
    return None


def infer_tier(change: str | None, methodology: dict[str, str]) -> str:
    if change or methodology.get("id") in {"openspec", "lufy-sdd"}:
        return "T1/T2 inferido"
    return "T3/Express o sin spec inferido"


def count_tools(parts: list[dict[str, Any]], events: list[dict[str, Any]]) -> dict[str, int]:
    counts: dict[str, int] = {}
    for row in [*parts, *events]:
        name = string_value(row, ("tool", "tool_name", "name"))
        if name:
            counts[name] = counts.get(name, 0) + 1
    return counts


def count_names(rows: list[dict[str, Any]], keys: tuple[str, ...]) -> dict[str, int]:
    counts: dict[str, int] = {}
    for row in rows:
        name = string_value(row, keys)
        if name:
            counts[name] = counts.get(name, 0) + 1
    return counts


def count_skill_names(rows: list[dict[str, Any]]) -> dict[str, int]:
    counts: dict[str, int] = {}
    for row in rows:
        for key in ("skill", "command", "name", "type"):
            value = string_value(row, (key,))
            if value and ("skill" in value or value.startswith("lufy.") or value.startswith("opsx-") or value.startswith("openspec-")):
                counts[value] = counts.get(value, 0) + 1
    return counts


def string_value(row: dict[str, Any], keys: tuple[str, ...]) -> str | None:
    for key in keys:
        value = row.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()
    return None


def calculate_time_metrics(activity: dict[str, Any], limitations: list[str]) -> dict[str, str]:
    timestamps = activity["timestamps"]
    if len(timestamps) < 2:
        limitations.append("Timestamps insuficientes para calcular tiempos con precisión.")
        return {"wall_clock": "No disponible", "ai_time": "No disponible", "human_time": "No disponible"}
    all_ts = [ts for ts, _ in timestamps]
    ai_ts = [ts for ts, role in timestamps if role in {"assistant", "tool", "event", "step-finish", "activity"}]
    human_ts = [ts for ts, role in timestamps if role == "user"]
    return {
        "wall_clock": format_seconds(max(all_ts) - min(all_ts)),
        "ai_time": format_seconds(capped_intervals(ai_ts, AI_GAP_CAP_SECONDS)) if len(ai_ts) > 1 else "Estimación parcial",
        "human_time": format_seconds(capped_intervals(human_ts, HUMAN_GAP_CAP_SECONDS)) if len(human_ts) > 1 else "Estimación parcial",
    }


def capped_intervals(values: list[float], cap: int) -> float:
    ordered = sorted(values)
    return sum(min(max(0, right - left), cap) for left, right in zip(ordered, ordered[1:]))


def collect_git_metrics(target_dir: Path, since: str | None, until: str | None, limitations: list[str]) -> dict[str, Any]:
    if not run_git(target_dir, ["rev-parse", "--is-inside-work-tree"])[0]:
        limitations.append("Git no disponible o el target no es un repositorio; commits y LOC neto degradan.")
        return {"available": False, "commits": "No disponible", "net_loc": "No disponible"}
    args = ["log", "--numstat", "--format=commit:%H"]
    if since:
        args.append(f"--since={since}")
    if until:
        args.append(f"--until={until}")
    ok, output = run_git(target_dir, args)
    if not ok:
        limitations.append("No se pudo consultar Git log read-only; commits y LOC neto degradan.")
        return {"available": False, "commits": "No disponible", "net_loc": "No disponible"}
    commits = 0
    added = 0
    deleted = 0
    for line in output.splitlines():
        if line.startswith("commit:"):
            commits += 1
            continue
        parts = line.split("\t")
        if len(parts) >= 3 and parts[0].isdigit() and parts[1].isdigit():
            added += int(parts[0])
            deleted += int(parts[1])
    return {"available": True, "commits": commits, "net_loc": added - deleted, "added": added, "deleted": deleted}


def run_git(cwd: Path, args: list[str]) -> tuple[bool, str]:
    try:
        completed = subprocess.run(["git", *args], cwd=cwd, check=False, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except (OSError, subprocess.SubprocessError):
        return False, ""
    return completed.returncode == 0, completed.stdout


def collect_stack(target_dir: Path, limitations: list[str]) -> dict[str, Any]:
    project_yaml = target_dir / ".opencode/project.yaml"
    if project_yaml.exists():
        try:
            parsed = parse_simple_project_yaml(project_yaml.read_text(encoding="utf-8"))
        except OSError as exc:
            limitations.append(f"No se pudo leer .opencode/project.yaml: {exc}.")
            parsed = []
        if parsed:
            return {"source": ".opencode/project.yaml", "items": parsed}
        limitations.append(".opencode/project.yaml ausente de metadata stack simple o inválido; se usa heurística.")
    heuristics = []
    signals = {"Go": "go.mod", "Node/TypeScript": "package.json", "Python": "pyproject.toml", "Rust": "Cargo.toml", "OpenSpec": "openspec", "OpenCode": ".opencode"}
    for label, rel in signals.items():
        if (target_dir / rel).exists():
            heuristics.append(label)
    return {"source": "heurística" if heuristics else "No configurado", "items": heuristics or ["No configurado"]}


def parse_simple_project_yaml(text: str) -> list[str]:
    items: list[str] = []
    for line in text.splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if any(stripped.startswith(prefix) for prefix in ("stack:", "tooling:", "language:", "framework:")):
            value = stripped.split(":", 1)[1].strip().strip('"\'')
            if value and value not in {"|", ">"}:
                items.append(value)
        elif stripped.startswith("-"):
            value = stripped[1:].strip().strip('"\'')
            if value:
                items.append(value)
    return items[:12]


def infer_phases(activity: dict[str, Any]) -> list[tuple[str, int]]:
    tools = activity["tool_counts"]
    phases = {
        "exploración": sum(tools.get(name, 0) for name in ("read", "grep", "glob", "webfetch")),
        "implementación": sum(tools.get(name, 0) for name in ("apply_patch", "edit", "write")),
        "validación": sum(count for name, count in tools.items() if any(token in name for token in ("bash", "test", "validate"))),
        "delivery-readiness": sum(count for name, count in tools.items() if "git" in name or "gh" in name),
    }
    return [(name, count) for name, count in phases.items() if count > 0] or [("No disponible", 0)]


def build_step_timeline(activity: dict[str, Any]) -> list[dict[str, Any]]:
    events = structural_timeline_events(activity)
    if not events:
        return []
    buckets_by_phase: dict[str, dict[str, Any]] = {}
    for event in sorted(events, key=lambda item: item["start"]):
        phase = event["phase"]
        if phase not in buckets_by_phase:
            buckets_by_phase[phase] = new_step_bucket(event)
        else:
            merge_step_event(buckets_by_phase[phase], event)
    buckets = sorted(buckets_by_phase.values(), key=lambda item: item["first_seen"])
    return [finalize_step(index + 1, bucket) for index, bucket in enumerate(buckets[:24])]


def structural_timeline_events(activity: dict[str, Any]) -> list[dict[str, Any]]:
    events: list[dict[str, Any]] = []
    for row in activity.get("messages", []):
        bounds = row_bounds(row)
        if not bounds:
            continue
        role = str(row.get("role") or row.get("type") or "message").lower()
        actor = string_value(row, ("agent", "mode")) or role
        phase = "input humano" if role == "user" else "razonamiento IA"
        events.append(timeline_event(bounds, phase, role, actor, "message", None, None))
    for row in activity.get("parts", []):
        bounds = row_bounds(row)
        if not bounds:
            continue
        tool = string_value(row, ("tool", "tool_name", "name")) or "tool"
        skill = string_value(row, ("skill", "command"))
        actor = string_value(row, ("agent", "mode")) or "tool"
        phase = classify_timeline_phase(tool, skill, row)
        if phase == "ejecución" and tool in {None, "tool"} and not skill:
            continue
        events.append(timeline_event(bounds, phase, "tool", actor, "tool", tool, skill))
    for row in activity.get("events", []):
        bounds = row_bounds(row)
        if not bounds:
            continue
        tool = string_value(row, ("tool", "tool_name", "name"))
        skill = string_value(row, ("skill", "command"))
        event_type = string_value(row, ("type", "event")) or "event"
        phase = classify_timeline_phase(tool or event_type, skill, row)
        if phase == "ejecución" and not tool and not skill:
            continue
        events.append(timeline_event(bounds, phase, "event", event_type, event_type, tool, skill))
    return events


def row_bounds(row: dict[str, Any]) -> tuple[float, float] | None:
    values = [ts for ts, _ in timestamps_from_rows([row])]
    if not values:
        return None
    return min(values), max(values)


def timeline_event(bounds: tuple[float, float], phase: str, role: str, actor: str, kind: str, tool: str | None, skill: str | None) -> dict[str, Any]:
    start, end = capped_event_bounds(bounds, role)
    return {
        "start": start,
        "end": end,
        "phase": phase,
        "role": role,
        "actor": actor,
        "kind": kind,
        "tool": tool,
        "skill": skill,
    }


def classify_timeline_phase(tool: str | None, skill: str | None, row: dict[str, Any]) -> str:
    value = " ".join(str(item or "").lower() for item in (tool, skill, row.get("type"), row.get("command")))
    tokens = set(re.split(r"[^a-z0-9]+", value))
    if any(token in tokens for token in ("read", "grep", "glob", "search", "list", "inspect", "explore")):
        return "exploración"
    if "apply_patch" in value or any(token in tokens for token in ("edit", "write", "implement", "sync")):
        return "implementación"
    if any(token in tokens for token in ("test", "validate", "verify", "smoke", "coverage", "build")):
        return "validación"
    if any(token in tokens for token in ("git", "github", "gh", "commit", "push", "pr", "delivery")):
        return "delivery-readiness"
    if any(token in tokens for token in ("openspec", "opsx", "sdd", "skill")):
        return "workflow/spec"
    return "ejecución"


def capped_event_bounds(bounds: tuple[float, float], role: str) -> tuple[float, float]:
    start, end = bounds
    cap = HUMAN_GAP_CAP_SECONDS if role == "user" else AI_GAP_CAP_SECONDS
    if end - start > cap:
        return start, start + cap
    return start, end


def new_step_bucket(event: dict[str, Any]) -> dict[str, Any]:
    active_seconds = event_active_seconds(event)
    return {
        "start": event["start"],
        "end": event["end"],
        "first_seen": event["start"],
        "phase": event["phase"],
        "roles": [event["role"]],
        "actors": [event["actor"]],
        "kinds": [event["kind"]],
        "tools": [event["tool"]] if event["tool"] else [],
        "skills": [event["skill"]] if event["skill"] else [],
        "ai_points": ai_points_for_event(event),
        "human_points": human_points_for_event(event),
        "active_seconds": active_seconds,
        "ai_seconds": active_seconds if event["role"] in {"assistant", "tool", "event", "step-finish", "activity"} else 0,
        "human_seconds": active_seconds if event["role"] == "user" else 0,
        "event_count": 1,
    }


def merge_step_event(bucket: dict[str, Any], event: dict[str, Any]) -> None:
    bucket["start"] = min(bucket["start"], event["start"])
    bucket["end"] = max(bucket["end"], event["end"])
    bucket["roles"].append(event["role"])
    bucket["actors"].append(event["actor"])
    bucket["kinds"].append(event["kind"])
    if event["tool"]:
        bucket["tools"].append(event["tool"])
    if event["skill"]:
        bucket["skills"].append(event["skill"])
    bucket["ai_points"].extend(ai_points_for_event(event))
    bucket["human_points"].extend(human_points_for_event(event))
    active_seconds = event_active_seconds(event)
    bucket["active_seconds"] += active_seconds
    if event["role"] in {"assistant", "tool", "event", "step-finish", "activity"}:
        bucket["ai_seconds"] += active_seconds
    if event["role"] == "user":
        bucket["human_seconds"] += active_seconds
    bucket["event_count"] += 1


def event_active_seconds(event: dict[str, Any]) -> float:
    return max(0, event["end"] - event["start"])


def ai_points_for_event(event: dict[str, Any]) -> list[float]:
    if event["role"] in {"assistant", "tool", "event", "step-finish", "activity"}:
        return [event["start"], event["end"]]
    return []


def human_points_for_event(event: dict[str, Any]) -> list[float]:
    return [event["start"], event["end"]] if event["role"] == "user" else []


def finalize_step(index: int, bucket: dict[str, Any]) -> dict[str, Any]:
    return {
        "index": index,
        "phase": bucket["phase"],
        "what": step_description(bucket),
        "why": step_reason(bucket),
        "start": format_bound(bucket["start"]),
        "end": format_bound(bucket["end"]),
        "wall_clock": format_active_seconds(bucket["active_seconds"], bucket["event_count"]),
        "ai_time": format_role_active_seconds(bucket["ai_seconds"]),
        "human_time": format_role_active_seconds(bucket["human_seconds"]),
        "actors": summarize_values(bucket["actors"]),
        "tools": summarize_values([*bucket["tools"], *bucket["skills"]]),
        "events": bucket["event_count"],
    }


def step_description(bucket: dict[str, Any]) -> str:
    phase = bucket["phase"]
    if phase == "input humano":
        return "Solicitud, aclaración o decisión del usuario registrada como input humano."
    if phase == "razonamiento IA":
        return "Razonamiento, planificación o síntesis del agente sin exponer contenido conversacional."
    if phase == "exploración":
        return "Lectura, búsqueda e inspección estructural del proyecto."
    if phase == "implementación":
        return "Aplicación de cambios, sincronización o actualización de artefactos gestionados."
    if phase == "validación":
        return "Ejecución de validaciones, smoke tests, builds o verificaciones estructurales."
    if phase == "workflow/spec":
        return "Uso de skills, comandos SDD/OpenSpec o artefactos de metodología."
    if phase == "delivery-readiness":
        return "Preparación o inspección de estado Git/GitHub para delivery."
    return "Ejecución técnica estructural detectada por OpenCode."


def step_reason(bucket: dict[str, Any]) -> str:
    phase = bucket["phase"]
    if phase == "input humano":
        return "El usuario guió el objetivo, confirmó cambios o corrigió el alcance."
    if phase == "razonamiento IA":
        return "La IA necesitó planificar, interpretar evidencia y decidir el siguiente paso seguro."
    if phase == "exploración":
        return "Antes de modificar, la IA revisó contexto local para reducir supuestos y seguir patrones del repo."
    if phase == "implementación":
        return "La IA aplicó cambios para transformar la decisión en artefactos ejecutables o instalables."
    if phase == "validación":
        return "La IA buscó evidencia de que el cambio funciona y no rompe contratos existentes."
    if phase == "workflow/spec":
        return "La IA usó metodología o skills para preservar trazabilidad y consistencia del flujo."
    if phase == "delivery-readiness":
        return "La IA preparó evidencia o estado para que el cambio pueda revisarse e integrarse."
    return "El evento aporta actividad técnica, pero la fuente local no expone una intención más precisa."


def summarize_values(values: list[str]) -> str:
    counts: dict[str, int] = {}
    for value in values:
        if value:
            counts[value] = counts.get(value, 0) + 1
    if not counts:
        return "No disponible"
    return ", ".join(f"{name} ({count})" for name, count in sorted(counts.items(), key=lambda item: item[1], reverse=True)[:4])


def format_duration_points(values: list[float], cap: int) -> str:
    if not values:
        return "0m"
    seconds = capped_intervals(values, cap) if len(values) > 1 else 0
    if seconds <= 0:
        return "<1m"
    return format_seconds(seconds)


def format_active_seconds(seconds: float, event_count: int) -> str:
    if seconds <= 0:
        return "<1m" if event_count > 0 else "0m"
    return format_seconds(seconds)


def format_role_active_seconds(seconds: float) -> str:
    if seconds <= 0:
        return "0m"
    return format_seconds(seconds)


def render_report(report: dict[str, Any], target_dir: Path, db_path: Path) -> str:
    activity = report["activity"]
    git = report["git"]
    metrics = report["metrics"]
    limitations = report["limitations"]
    sections = [
        f"<header><h1>LUFY Developer Impact Report</h1><p class='muted'>Target: {esc(str(target_dir))} · Generado: {esc(format_datetime_human(dt.datetime.now(dt.timezone.utc)))}</p></header>",
        render_scope(activity["scope"], report),
        render_executive_summary(report),
        "<section><h2>Impacto diario</h2><div class='grid'>" + card("Tiempo calendario", metrics["wall_clock"]) + card("Tiempo IA activo", metrics["ai_time"]) + card("Tiempo humano activo", metrics["human_time"]) + card("Tool calls", str(sum(activity["tool_counts"].values()))) + "</div></section>",
        render_steps(report["steps"]),
        render_learnings_and_pivots(report),
        "<section><h2>Git</h2><div class='grid'>" + card("Commits", str(git["commits"])) + card("LOC neto", str(git["net_loc"])) + "</div></section>",
        "<section><h2>Tool calls</h2>" + card("Total", str(sum(activity["tool_counts"].values()))) + render_table("Top tools", activity["tool_counts"]) + "</section>",
        "<section><h2>Subagents</h2>" + render_table("Subagents", activity["subagents"]) + "</section>",
        "<section><h2>Skills</h2>" + render_table("Skills", activity["skills"]) + "</section>",
        "<section><h2>Fases / timeline</h2>" + render_pairs(report["phases"]) + "</section>",
        "<section><h2>Stack detectado</h2><p class='muted'>Fuente: " + esc(report["stack"]["source"]) + "</p>" + "".join(f"<span class='pill'>{esc(item)}</span>" for item in report["stack"]["items"]) + "</section>",
        "<section><h2>Metodología y limitaciones</h2>" + methodology(db_path) + render_limitations(limitations) + "</section>",
    ]
    return "\n".join(sections)


def render_scope(scope: dict[str, Any], report: dict[str, Any]) -> str:
    methodology_info = scope.get("methodology", {})
    cards = [
        card("Scope", str(scope.get("mode", "No disponible"))),
        card("Tier", str(scope.get("tier", "No disponible"))),
        card("Sesiones incluidas", str(scope.get("session_count", "0"))),
        card("Metodología", str(methodology_info.get("id", "No detectada"))),
    ]
    rows = {
        "Tarea raíz": str(scope.get("root_title", "No disponible")),
        "Slug": str(scope.get("root_slug", "No disponible")),
        "Root session": str(scope.get("root_session_id", "No disponible")),
        "Anchor session": str(scope.get("anchor_session_id", "No disponible")),
        "Ventana": f"{scope.get('time_from', 'No disponible')} → {scope.get('time_to', 'No disponible')}",
        "Git window": format_range_text(report.get("git_since"), report.get("git_until")),
        "Spec/change": str(methodology_info.get("change", scope.get("change", "No disponible"))),
        "Fuente metodología": str(methodology_info.get("source", "No disponible")),
        "Fuente tier": str(scope.get("tier_source", "No disponible")),
    }
    details = "".join(f"<tr><td>{esc(key)}</td><td>{esc(value)}</td></tr>" for key, value in rows.items())
    return "<section><h2>Propiedades de la tarea</h2><div class='grid'>" + "".join(cards) + "</div><div class='table-wrap'><table class='property-table'><tbody>" + details + "</tbody></table></div></section>"


def render_executive_summary(report: dict[str, Any]) -> str:
    activity = report["activity"]
    scope = activity["scope"]
    metrics = report["metrics"]
    git = report["git"]
    tool_total = sum(activity["tool_counts"].values())
    summary = (
        "<div class='callout blue'><div class='callout-title'>Qué pidió el usuario</div>"
        f"<p>Trabajar sobre la tarea <strong>{esc(str(scope.get('root_title', 'No disponible')))}</strong> "
        f"en scope <code>{esc(str(scope.get('mode', 'task')))}</code>, con tier <code>{esc(str(scope.get('tier', 'No disponible')))}</code>.</p></div>"
        "<div class='callout green'><div class='callout-title'>Qué aportó la IA</div>"
        f"<p>Ejecutó {tool_total} eventos de tools/skills, produjo evidencia local y dejó trazabilidad de tiempo: "
        f"{esc(metrics['ai_time'])} de IA activa y {esc(metrics['human_time'])} de intervención humana estimada.</p></div>"
        "<div class='callout purple'><div class='callout-title'>Resultado para revisar</div>"
        f"<p>El reporte combina actividad OpenCode, Git y metadata local. Git muestra {esc(str(git['commits']))} commits y "
        f"{esc(str(git['net_loc']))} LOC neto dentro del rango seleccionado.</p></div>"
    )
    return "<section><h2>Resumen ejecutivo</h2>" + summary + "</section>"


def render_steps(steps: list[dict[str, Any]]) -> str:
    if not steps:
        return "<section><h2>Paso a paso</h2><p class='warn'>No disponible</p></section>"
    rows = []
    for step in steps:
        rows.append(
            "<tr>"
            f"<td>{step['index']}</td>"
            f"<td>{esc(step['phase'])}</td>"
            f"<td>{esc(step['what'])}<br><span class='muted'>Actores: {esc(step['actors'])}<br>Tools/skills: {esc(step['tools'])}</span></td>"
            f"<td>{esc(step['why'])}</td>"
            f"<td>{esc(step['start'])}<br><span class='muted'>{esc(step['end'])}</span></td>"
            f"<td>{esc(step['wall_clock'])}</td>"
            f"<td>{esc(step['ai_time'])}</td>"
            f"<td>{esc(step['human_time'])}</td>"
            f"<td>{step['events']}</td>"
            "</tr>"
        )
    table = (
        "<div class='table-wrap'><table><thead><tr><th>#</th><th>Paso</th><th>Qué se hizo</th><th>Por qué</th><th>Primera / última actividad</th>"
        "<th>Duración activa</th><th>IA</th><th>Humano</th><th>Eventos</th></tr></thead><tbody>"
        + "".join(rows)
        + "</tbody></table></div>"
    )
    return "<section><h2>Paso a paso</h2><p class='muted'>Timeline estructural sanitizado de la tarea, agregado por fase y ordenado por primera aparición. La duración muestra actividad estimada del paso, no el tiempo calendario entre primera y última aparición. No incluye prompts, outputs, argumentos completos ni diffs.</p>" + table + "</section>"


def render_learnings_and_pivots(report: dict[str, Any]) -> str:
    activity = report["activity"]
    scope = activity["scope"]
    methodology_info = scope.get("methodology", {})
    learning_items = [
        f"Stack detectado desde {report['stack']['source']}: {', '.join(str(item) for item in report['stack']['items'])}.",
        f"Metodología detectada: {methodology_info.get('id', 'none')} ({methodology_info.get('source', 'sin fuente')}).",
        "El reporte puede reconstruir actividad, tiempos y herramientas desde fuentes locales sanitizadas.",
    ]
    if activity["skills"]:
        learning_items.append("Skills relevantes: " + ", ".join(sorted(activity["skills"])[:5]) + ".")
    pivot_items = [
        "Los pivots explícitos requieren un journal estructurado de tarea para distinguir correcciones por usuario, errores de validación o cambios de estrategia.",
        "Con las fuentes actuales se muestran cambios de fase y evidencia, pero no se inventan motivos conversacionales no registrados.",
    ]
    return (
        "<section><h2>Aprendizajes y pivots</h2>"
        "<div class='callout green'><div class='callout-title'>Aprendizajes detectados</div>" + render_list(learning_items) + "</div>"
        "<div class='callout orange'><div class='callout-title'>Pivots / correcciones</div>" + render_list(pivot_items) + "</div>"
        "</section>"
    )


def card(label: str, value: str) -> str:
    return f"<div class='card'><div class='muted'>{esc(label)}</div><div class='metric'>{esc(value)}</div></div>"


def render_table(title: str, values: dict[str, int]) -> str:
    if not values:
        return f"<h3>{esc(title)}</h3><p class='warn'>No disponible</p>"
    rows = "".join(f"<tr><td>{esc(name)}</td><td>{count}</td></tr>" for name, count in sorted(values.items(), key=lambda item: item[1], reverse=True)[:10])
    return f"<h3>{esc(title)}</h3><div class='table-wrap'><table><thead><tr><th>Nombre</th><th>Conteo</th></tr></thead><tbody>{rows}</tbody></table></div>"


def render_pairs(values: list[tuple[str, int]]) -> str:
    return "".join(f"<span class='pill'>{esc(name)}: {count}</span>" for name, count in values)


def methodology(db_path: Path) -> str:
    return (
        f"<p>Fuentes locales read-only: SQLite OpenCode ({esc(str(db_path))}), Git y .opencode/project.yaml opcional.</p>"
        f"<p>Heurísticas: AI gap cap {AI_GAP_CAP_SECONDS // 60} min; humano gap cap {HUMAN_GAP_CAP_SECONDS // 60} min. "
        "El reporte usa allowlist estructural y excluye prompts, outputs, payloads completos, contenidos de archivos, diffs y snapshots de cambios de sesión.</p>"
    )


def render_limitations(limitations: list[str]) -> str:
    if not limitations:
        return "<p>Sin limitaciones detectadas para las fuentes disponibles.</p>"
    return render_list(limitations)


def render_list(items: list[str]) -> str:
    return "<ul>" + "".join(f"<li>{esc(item)}</li>" for item in items) + "</ul>"


def esc(value: str) -> str:
    return html.escape(value, quote=True)


def format_seconds(seconds: float) -> str:
    if seconds < 0:
        return "No disponible"
    if 0 < seconds < 30:
        return "<1m"
    minutes = int(round(seconds / 60))
    hours, mins = divmod(minutes, 60)
    if hours:
        return f"{hours}h {mins}m"
    return f"{mins}m"


def assert_no_sensitive_fixture_tokens(html_doc: str) -> None:
    for token in SENSITIVE_WORDS:
        if token in html_doc:
            raise RuntimeError(f"El HTML contiene un token sensible de fixture: {token}")
    if "http://" in html_doc or "https://" in html_doc:
        raise RuntimeError("El HTML contiene referencias remotas")


if __name__ == "__main__":
    sys.exit(main())
