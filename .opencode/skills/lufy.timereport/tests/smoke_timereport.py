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
REQUIRED = ["Wall-clock", "AI working time", "Tiempo humano activo", "LOC neto", "Commits", "Tool calls", "Top tools", "Subagents", "Skills", "Fases / timeline", "Stack detectado", "Metodología y limitaciones"]


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
        missing = [label for label in REQUIRED if label not in html]
        if missing:
            raise AssertionError(f"Secciones requeridas ausentes: {missing}")
        leaked = [token for token in SENSITIVE if token in html]
        if leaked:
            raise AssertionError(f"Contenido sensible filtrado al HTML: {leaked}")
        if "http://" in html or "https://" in html:
            raise AssertionError("El HTML contiene referencias remotas")

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
