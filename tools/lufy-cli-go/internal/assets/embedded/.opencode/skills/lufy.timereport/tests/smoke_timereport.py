#!/usr/bin/env python3
"""Smoke validation for lufy.timereport using sanitized synthetic data."""

from __future__ import annotations

import os
from pathlib import Path
import sqlite3
import subprocess
import sys
import tempfile


ROOT = Path(__file__).resolve().parents[3]
SCRIPT = ROOT / "skills/lufy.timereport/scripts/generate_timereport.py"
SENSITIVE = ["USER_PROMPT_SECRET", "ASSISTANT_OUTPUT_SECRET", "FULL_TOOL_PAYLOAD", "diff --git", "session_diff"]
REQUIRED = ["LUFY Developer Impact Report", "Propiedades de la tarea", "Resumen ejecutivo", "Impacto diario", "Tiempo IA activo", "Tiempo humano activo", "Diagnóstico de tiempo", "Confianza scope", "Dónde se fue el tiempo", "Cómo reducir la próxima iteración", "Paso a paso", "Qué se hizo", "Por qué", "Primera / última actividad", "Duración activa", "IA", "Humano", "Aprendizajes y pivots", "LOC neto", "Commits", "Tool calls", "Top tools", "Subagents", "Skills", "Fases / timeline", "Stack detectado", "Metodología y limitaciones"]


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="lufy-timereport-smoke-") as tmp:
        tmp_path = Path(tmp)
        target = tmp_path / "repo"
        target.mkdir()
        (target / ".opencode").mkdir()
        (target / "go.mod").write_text("module fixture\n", encoding="utf-8")
        init_fixture_git(target)
        db_path = tmp_path / "opencode.db"
        create_fixture_db(db_path, target)

        output = tmp_path / "report.html"
        run([sys.executable, str(SCRIPT), "--db", str(db_path), "--target-dir", str(target), "--output", str(output)])
        html = output.read_text(encoding="utf-8")
        assert_report_shape(html)
        for expected in ("read", "apply_patch", "implementer", "lufy.timereport"):
            if expected not in html:
                raise AssertionError(f"Métrica esperada ausente en fixture legacy: {expected}")

        modern_db_path = tmp_path / "opencode-modern.db"
        create_modern_fixture_db(modern_db_path, target)
        modern_output = tmp_path / "modern-report.html"
        run([sys.executable, str(SCRIPT), "--db", str(modern_db_path), "--target-dir", str(target), "--output", str(modern_output)])
        modern_html = modern_output.read_text(encoding="utf-8")
        assert_report_shape(modern_html)
        for expected in ("bash", "skill", "openspec-apply-change", "validator"):
            if expected not in modern_html:
                raise AssertionError(f"Métrica esperada ausente en fixture moderno: {expected}")
        for unexpected in ("noise_tool", "Noise task"):
            if unexpected in modern_html:
                raise AssertionError(f"El scope task incluyó actividad de otra tarea: {unexpected}")

        repo_output = tmp_path / "modern-repo-report.html"
        run([sys.executable, str(SCRIPT), "--db", str(modern_db_path), "--target-dir", str(target), "--scope", "repo", "--output", str(repo_output)])
        repo_html = repo_output.read_text(encoding="utf-8")
        if "noise_tool" not in repo_html:
            raise AssertionError("El scope repo no incluyó actividad global del repositorio")

        child_output = tmp_path / "modern-child-report.html"
        run([sys.executable, str(SCRIPT), "--db", str(modern_db_path), "--target-dir", str(target), "--session-id", "s2-child", "--tier", "T2", "--change", "add-fixture", "--output", str(child_output)])
        child_html = child_output.read_text(encoding="utf-8")
        for expected in ("Task root", "s2-root", "T2", "add-fixture"):
            if expected not in child_html:
                raise AssertionError(f"Scope por session-id no renderizó contrato de tarea: {expected}")

        missing_db_output = tmp_path / "missing-db.html"
        run([sys.executable, str(SCRIPT), "--db", str(tmp_path / "missing.db"), "--target-dir", str(target), "--output", str(missing_db_output)])
        missing_html = missing_db_output.read_text(encoding="utf-8")
        if "OpenCode SQLite no disponible" not in missing_html:
            raise AssertionError("La degradación por DB faltante no quedó visible")

        (target / ".opencode/project.yaml").write_text(":::\n", encoding="utf-8")
        invalid_project_output = tmp_path / "invalid-project.html"
        run([sys.executable, str(SCRIPT), "--db", str(db_path), "--target-dir", str(target), "--output", str(invalid_project_output)])
        if "project.yaml" not in invalid_project_output.read_text(encoding="utf-8"):
            raise AssertionError("La degradación por project.yaml inválido no quedó visible")

        non_git = tmp_path / "non-git"
        non_git.mkdir()
        non_git_output = tmp_path / "non-git.html"
        run([sys.executable, str(SCRIPT), "--db", str(db_path), "--target-dir", str(non_git), "--output", str(non_git_output)])
        if "Git no disponible" not in non_git_output.read_text(encoding="utf-8"):
            raise AssertionError("La degradación por directorio no-Git no quedó visible")
    print("smoke_timereport: ok")
    return 0


def create_fixture_db(db_path: Path, target: Path) -> None:
    connection = sqlite3.connect(db_path)
    with connection:
        connection.executescript(
            """
            CREATE TABLE project (id TEXT, directory TEXT);
            CREATE TABLE workspace (id TEXT, directory TEXT);
            CREATE TABLE session (id TEXT, directory TEXT, created INTEGER, updated INTEGER);
            CREATE TABLE message (id TEXT, sessionID TEXT, role TEXT, created INTEGER, completed INTEGER, agent TEXT, text TEXT);
            CREATE TABLE part (id TEXT, sessionID TEXT, messageID TEXT, type TEXT, tool TEXT, command TEXT, agent TEXT, created INTEGER, state TEXT, input TEXT, output TEXT);
            CREATE TABLE event (id TEXT, sessionID TEXT, type TEXT, tool TEXT, created INTEGER, payload TEXT);
            """
        )
        connection.execute("INSERT INTO project VALUES (?, ?)", ("p1", str(target)))
        connection.execute("INSERT INTO workspace VALUES (?, ?)", ("w1", str(target)))
        connection.execute("INSERT INTO session VALUES (?, ?, ?, ?)", ("s1", str(target), 1_700_000_000, 1_700_000_900))
        connection.execute("INSERT INTO message VALUES (?, ?, ?, ?, ?, ?, ?)", ("m1", "s1", "user", 1_700_000_000, 1_700_000_001, "implementer", "USER_PROMPT_SECRET"))
        connection.execute("INSERT INTO message VALUES (?, ?, ?, ?, ?, ?, ?)", ("m2", "s1", "assistant", 1_700_000_010, 1_700_000_100, "implementer", "ASSISTANT_OUTPUT_SECRET"))
        connection.execute("INSERT INTO part VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", ("pt1", "s1", "m2", "tool", "read", "", "implementer", 1_700_000_020, '{"status":"completed","time":{"start":1700000020,"end":1700000030}}', "FULL_TOOL_PAYLOAD", "diff --git a/secret b/secret"))
        connection.execute("INSERT INTO part VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", ("pt2", "s1", "m2", "tool", "apply_patch", "lufy.timereport", "implementer", 1_700_000_200, '{"status":"completed"}', "{}", ""))
        connection.execute("INSERT INTO event VALUES (?, ?, ?, ?, ?, ?)", ("e1", "s1", "skill", "bash", 1_700_000_300, "session_diff"))


def create_modern_fixture_db(db_path: Path, target: Path) -> None:
    connection = sqlite3.connect(db_path)
    with connection:
        connection.executescript(
            """
            CREATE TABLE session (id TEXT, parent_id TEXT, directory TEXT, path TEXT, title TEXT, slug TEXT, agent TEXT, time_created INTEGER, time_updated INTEGER, model TEXT, cost REAL, tokens_input INTEGER, tokens_output INTEGER);
            CREATE TABLE message (id TEXT, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
            CREATE TABLE part (id TEXT, message_id TEXT, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
            CREATE TABLE event (id TEXT, aggregate_id TEXT, seq INTEGER, type TEXT, data TEXT);
            """
        )
        connection.execute(
            "INSERT INTO session VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
            ("s2-root", None, str(target), "", "Task root", "task-root", "orchestrator", 1_700_001_000_000, 1_700_001_900_000, '{"id":"fixture"}', 0.0, 100, 20),
        )
        connection.execute(
            "INSERT INTO session VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
            ("s2-child", "s2-root", str(target), "", "Generar timereport", "task-child", "validator", 1_700_001_100_000, 1_700_001_800_000, '{"id":"fixture"}', 0.0, 90, 18),
        )
        connection.execute(
            "INSERT INTO session VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
            ("noise-root", None, str(target), "", "Noise task", "noise-task", "implementer", 1_700_000_000_000, 1_700_000_900_000, '{"id":"fixture"}', 0.0, 20, 5),
        )
        connection.execute(
            "INSERT INTO message VALUES (?, ?, ?, ?, ?)",
            ("m3", "s2-root", 1_700_001_000_000, 1_700_001_001_000, '{"role":"user","time":{"created":1700001000},"summary":{"diffs":[]},"content":"USER_PROMPT_SECRET"}'),
        )
        connection.execute(
            "INSERT INTO message VALUES (?, ?, ?, ?, ?)",
            ("m4", "s2-child", 1_700_001_010_000, 1_700_001_100_000, '{"role":"assistant","agent":"validator","time":{"created":1700001010},"text":"ASSISTANT_OUTPUT_SECRET"}'),
        )
        connection.execute(
            "INSERT INTO part VALUES (?, ?, ?, ?, ?, ?)",
            ("pt3", "m4", "s2-child", 1_700_001_020_000, 1_700_001_030_000, '{"type":"tool","tool":"bash","state":{"status":"completed","input":{"command":"echo USER_PROMPT_SECRET"},"output":"diff --git secret","time":{"start":1700001020,"end":1700001030}}}'),
        )
        connection.execute(
            "INSERT INTO part VALUES (?, ?, ?, ?, ?, ?)",
            ("pt4", "m4", "s2-child", 1_700_001_200_000, 1_700_001_210_000, '{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"openspec-apply-change"},"output":"FULL_TOOL_PAYLOAD","time":{"start":1700001200,"end":1700001210}}}'),
        )
        connection.execute(
            "INSERT INTO message VALUES (?, ?, ?, ?, ?)",
            ("m-noise", "noise-root", 1_700_000_010_000, 1_700_000_020_000, '{"role":"assistant","agent":"implementer","time":{"created":1700000010},"text":"ASSISTANT_OUTPUT_SECRET"}'),
        )
        connection.execute(
            "INSERT INTO part VALUES (?, ?, ?, ?, ?, ?)",
            ("pt-noise", "m-noise", "noise-root", 1_700_000_030_000, 1_700_000_040_000, '{"type":"tool","tool":"noise_tool","state":{"status":"completed","time":{"start":1700000030,"end":1700000040}}}'),
        )
        connection.execute("INSERT INTO event VALUES (?, ?, ?, ?, ?)", ("e2", "s2-child", 1, "session.updated", '{"type":"step-finish","time":{"created":1700001300},"payload":"session_diff"}'))


def assert_report_shape(html: str) -> None:
    missing = [label for label in REQUIRED if label not in html]
    if missing:
        raise AssertionError(f"Secciones requeridas ausentes: {missing}")
    leaked = [token for token in SENSITIVE if token in html]
    if leaked:
        raise AssertionError(f"Contenido sensible filtrado al HTML: {leaked}")
    if "http://" in html or "https://" in html:
        raise AssertionError("El HTML contiene referencias remotas")


def init_fixture_git(target: Path) -> None:
    run_git(["init"], target)
    run_git(["add", "go.mod"], target)
    run_git(["commit", "-m", "fixture commit"], target)


def run(args: list[str]) -> None:
    env = os.environ.copy()
    completed = subprocess.run(args, check=False, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
    if completed.returncode != 0:
        raise AssertionError(f"Comando falló: {' '.join(args)}\nSTDOUT:{completed.stdout}\nSTDERR:{completed.stderr}")


def run_git(args: list[str], cwd: Path) -> None:
    env = os.environ.copy()
    env.update({
        "GIT_AUTHOR_NAME": "Fixture",
        "GIT_AUTHOR_EMAIL": "fixture@example.invalid",
        "GIT_COMMITTER_NAME": "Fixture",
        "GIT_COMMITTER_EMAIL": "fixture@example.invalid",
    })
    completed = subprocess.run(["git", *args], cwd=cwd, check=False, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
    if completed.returncode != 0:
        raise AssertionError(f"Git fixture falló: git {' '.join(args)}\nSTDOUT:{completed.stdout}\nSTDERR:{completed.stderr}")


if __name__ == "__main__":
    sys.exit(main())
