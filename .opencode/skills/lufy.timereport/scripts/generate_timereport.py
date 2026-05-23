#!/usr/bin/env python3
"""Generate a local, sanitized LUFY time report as self-contained HTML."""

from __future__ import annotations

import argparse
import datetime as dt
import html
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import sys
import tempfile
from typing import Any

AI_GAP_CAP_SECONDS = 5 * 60
HUMAN_GAP_CAP_SECONDS = 10 * 60
DEFAULT_DB = Path.home() / ".local/share/opencode/opencode.db"

SENSITIVE_WORDS = ("session_diff", "diff --git", "BEGIN SECRET", "FULL_TOOL_PAYLOAD", "ASSISTANT_OUTPUT_SECRET", "USER_PROMPT_SECRET")


def main() -> int:
    args = parse_args()
    target_dir = Path(args.target_dir).expanduser().resolve()
    output = Path(args.output).expanduser() if args.output else Path(tempfile.gettempdir()) / f"lufy-timereport-{dt.datetime.now().strftime('%Y%m%d-%H%M%S')}.html"
    db_path = Path(args.db).expanduser()

    report = build_report(target_dir, db_path, args.since, args.until)
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
    return parser.parse_args()


def build_report(target_dir: Path, db_path: Path, since: str | None, until: str | None) -> dict[str, Any]:
    limitations: list[str] = []
    activity = collect_opencode_activity(db_path, target_dir, since, until, limitations)
    git = collect_git_metrics(target_dir, since, until, limitations)
    stack = collect_stack(target_dir, limitations)
    metrics = calculate_time_metrics(activity, limitations)
    phases = infer_phases(activity)
    return {"activity": activity, "git": git, "stack": stack, "metrics": metrics, "phases": phases, "limitations": limitations, "since": since, "until": until}


def collect_opencode_activity(db_path: Path, target_dir: Path, since: str | None, until: str | None, limitations: list[str]) -> dict[str, Any]:
    empty = {"sessions": [], "messages": [], "parts": [], "events": [], "tool_counts": {}, "subagents": {}, "skills": {}, "timestamps": []}
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
        sessions = collect_sessions(connection, tables, target_dir, limitations)
        session_ids = {string_value(row, ("id", "sessionID", "session_id")) for row in sessions}
        session_ids.discard(None)
        messages = collect_child_rows(connection, tables, "message", session_ids, limitations)
        parts = collect_child_rows(connection, tables, "part", session_ids, limitations)
        events = collect_child_rows(connection, tables, "event", session_ids, limitations)

    timestamps = filter_timestamps([*timestamps_from_rows(sessions), *timestamps_from_rows(messages), *timestamps_from_rows(parts), *timestamps_from_rows(events)], since, until)
    tool_counts = count_tools(parts, events)
    return {
        "sessions": sessions,
        "messages": messages,
        "parts": parts,
        "events": events,
        "tool_counts": tool_counts,
        "subagents": count_names([*messages, *parts], ("agent", "subagent", "agent_name")),
        "skills": count_skill_names([*parts, *events]),
        "timestamps": timestamps,
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
    rows = select_rows(connection, "session", safe_cols)
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


def collect_child_rows(connection: sqlite3.Connection, tables: set[str], table: str, session_ids: set[str], limitations: list[str]) -> list[dict[str, Any]]:
    if table not in tables or not session_ids:
        return []
    cols = columns(connection, table)
    safe_cols = safe_select_columns(cols)
    session_col = next((col for col in cols if col in ("sessionID", "session_id", "sessionId")), None)
    if not session_col:
        limitations.append(f"Tabla `{table}` sin columna sessionID/session_id; sección relacionada degradada.")
        return []
    placeholders = ",".join("?" for _ in session_ids)
    sql = f"SELECT {', '.join(quote_ident(c) for c in safe_cols)} FROM {quote_ident(table)} WHERE {quote_ident(session_col)} IN ({placeholders})"
    return [dict(row) for row in connection.execute(sql, sorted(session_ids)).fetchall()]


def safe_select_columns(cols: list[str]) -> list[str]:
    deny = {"text", "prompt", "output", "input", "content", "body", "diff", "session_diff", "arguments", "payload"}
    allowed = [col for col in cols if col.lower() not in deny]
    return allowed or [cols[0]]


def select_rows(connection: sqlite3.Connection, table: str, cols: list[str]) -> list[dict[str, Any]]:
    sql = f"SELECT {', '.join(quote_ident(c) for c in cols)} FROM {quote_ident(table)}"
    return [dict(row) for row in connection.execute(sql).fetchall()]


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
        for key in ("created", "updated", "created_at", "updated_at", "completed", "completed_at", "timestamp", "time", "start", "end"):
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
            if value and ("skill" in value or value.startswith("lufy.") or value.startswith("opsx-")):
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


def render_report(report: dict[str, Any], target_dir: Path, db_path: Path) -> str:
    activity = report["activity"]
    git = report["git"]
    metrics = report["metrics"]
    limitations = report["limitations"]
    sections = [
        f"<header><h1>LUFY Time Report</h1><p class='muted'>Target: {esc(str(target_dir))} · Generado: {esc(dt.datetime.now().isoformat(timespec='seconds'))}</p></header>",
        "<section><h2>Resumen de tiempo</h2><div class='grid'>" + card("Wall-clock", metrics["wall_clock"]) + card("AI working time", metrics["ai_time"]) + card("Tiempo humano activo", metrics["human_time"]) + "</div></section>",
        "<section><h2>Git</h2><div class='grid'>" + card("Commits", str(git["commits"])) + card("LOC neto", str(git["net_loc"])) + "</div></section>",
        "<section><h2>Tool calls</h2>" + card("Total", str(sum(activity["tool_counts"].values()))) + render_table("Top tools", activity["tool_counts"]) + "</section>",
        "<section><h2>Subagents</h2>" + render_table("Subagents", activity["subagents"]) + "</section>",
        "<section><h2>Skills</h2>" + render_table("Skills", activity["skills"]) + "</section>",
        "<section><h2>Fases / timeline</h2>" + render_pairs(report["phases"]) + "</section>",
        "<section><h2>Stack detectado</h2><p class='muted'>Fuente: " + esc(report["stack"]["source"]) + "</p>" + "".join(f"<span class='pill'>{esc(item)}</span>" for item in report["stack"]["items"]) + "</section>",
        "<section><h2>Metodología y limitaciones</h2>" + methodology(db_path) + render_limitations(limitations) + "</section>",
    ]
    return "\n".join(sections)


def card(label: str, value: str) -> str:
    return f"<div class='card'><div class='muted'>{esc(label)}</div><div class='metric'>{esc(value)}</div></div>"


def render_table(title: str, values: dict[str, int]) -> str:
    if not values:
        return f"<h3>{esc(title)}</h3><p class='warn'>No disponible</p>"
    rows = "".join(f"<tr><td>{esc(name)}</td><td>{count}</td></tr>" for name, count in sorted(values.items(), key=lambda item: item[1], reverse=True)[:10])
    return f"<h3>{esc(title)}</h3><table><thead><tr><th>Nombre</th><th>Conteo</th></tr></thead><tbody>{rows}</tbody></table>"


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
    return "<ul>" + "".join(f"<li>{esc(item)}</li>" for item in limitations) + "</ul>"


def esc(value: str) -> str:
    return html.escape(value, quote=True)


def format_seconds(seconds: float) -> str:
    if seconds < 0:
        return "No disponible"
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
