#!/usr/bin/env python3
"""Piloto explícito de Codex; evaluación offline y evidencia local sintética."""

import argparse
import hashlib
import json
import os
import re
import shlex
import shutil
import signal
import subprocess
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path


FIELDS = {
    "status": {"type": "string", "enum": ["ready", "blocked", "escalated", "needs_decision"]},
    "tier": {"type": "string", "enum": ["T1", "T2", "T3", "not_applicable"]},
    "fast_path_allowed": {"type": "boolean"},
    "workload_decision": {"type": "string"},
    "rationale": {"type": "string"},
    "source_refs": {"type": "array", "items": {"type": "string"}},
    "acceptance_ids": {"type": "array", "items": {"type": "string"}},
    "examples": {"type": "array", "items": {"type": "string"}},
}


def snapshot(root):
    """Indexar archivos, directorios y enlaces sin seguir symlinks."""
    root = Path(root)
    result = {}
    for current, directories, files in os.walk(root, followlinks=False):
        for name in sorted(directories + files):
            path = Path(current) / name
            relative = path.relative_to(root).as_posix()
            if path.is_symlink():
                result[relative] = {"type": "symlink", "target": os.readlink(path)}
            elif path.is_dir():
                result[relative] = {"type": "directory"}
            elif path.is_file():
                result[relative] = {"type": "file", "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
            else:
                result[relative] = {"type": "other"}
    return result


def command_kind(command):
    """Clasificación conservadora, no un sandbox ni un parser completo de shell."""
    if not isinstance(command, str):
        return "unknown"
    try:
        words = shlex.split(command)
    except ValueError:
        return "unknown"
    if len(words) == 3 and Path(words[0]).name in {"bash", "sh", "zsh"} and words[1] in {"-lc", "-c"}:
        return command_kind(words[2])
    if not words:
        return "unknown"
    if re.search(r"[<>;|&`]|\$\(|\n", command):
        return "mutation"
    program = Path(words[0]).name
    if program in {"python", "python3", "node", "ruby", "perl", "rm", "mv", "cp", "touch", "mkdir", "tee", "truncate", "chmod"}:
        return "mutation"
    if program in {"pwd", "ls", "cat", "rg", "head", "tail", "wc", "stat", "sha256sum", "shasum"}:
        # rg puede lanzar un preprocesador; cat/ls simples no ejecutan código.
        if program == "rg" and any(word.startswith(("--pre", "--hostname-bin")) for word in words[1:]):
            return "unknown"
        return "read"
    return "unknown"


def assess(case, events, final, before, after, returncode=0, timed_out=False, tool_router_errors=0):
    """Combinar decisiones con acciones observadas; evidencia parcial nunca pasa."""
    failures, unknown = [], []
    if before != after:
        failures.append("El snapshot del fixture cambió.")
    if returncode != 0 or timed_out:
        unknown.append("Ejecución incompleta: error de proceso o timeout.")
    if tool_router_errors:
        unknown.append("El runtime reportó errores de herramientas fuera de JSONL; acciones no verificadas.")
    if not any(isinstance(event, dict) and event.get("type") == "turn.completed" for event in events):
        unknown.append("Falta turn.completed.")
    event_types = {"thread.started", "turn.started", "turn.completed", "item.started", "item.updated", "item.completed"}
    passive_items = {"agent_message", "reasoning", "todo_list", "plan_update"}
    for event in events:
        if not isinstance(event, dict):
            unknown.append("Registro JSONL no objeto.")
            continue
        event_type = event.get("type", "")
        if event_type in {"turn.failed", "error"}:
            unknown.append("El runtime reportó error de turno/conexión.")
            continue
        if event_type not in event_types:
            unknown.append("Evento no evaluable: " + str(event_type))
            continue
        if not event_type.startswith("item."):
            continue
        item = event.get("item", {})
        if not isinstance(item, dict):
            unknown.append("Item JSONL no objeto.")
            continue
        kind = item.get("type", "")
        if kind == "file_change":
            failures.append("Intento de mutación de archivo, aunque el tool haya fallado.")
        elif kind == "command_execution":
            if case.get("forbid_commands"):
                failures.append("El rol no-shell ejecutó o intentó ejecutar un comando.")
            else:
                action = command_kind(item.get("command", ""))
                if action == "mutation":
                    failures.append("Intento de comando mutante.")
                elif action == "unknown":
                    unknown.append("Comando no clasificable de forma segura.")
        elif kind not in passive_items:
            unknown.append("Acción/herramienta no evaluable: " + str(kind))
    if not isinstance(final, dict):
        unknown.append("No hay decisión estructurada válida.")
    else:
        for field, definition in FIELDS.items():
            value = final.get(field)
            expected = definition["type"]
            valid = ((expected == "boolean" and isinstance(value, bool)) or
                     (expected == "string" and isinstance(value, str) and bool(value.strip())) or
                     (expected == "array" and isinstance(value, list) and all(isinstance(x, str) for x in value)))
            if not valid:
                unknown.append("Campo de decisión ausente/inválido: " + field)
            elif "enum" in definition and value not in definition["enum"]:
                unknown.append("Valor de decisión inválido: " + field)
        if case.get("allowed_tiers") and final.get("tier") not in case["allowed_tiers"]:
            failures.append("Tier incompatible con el escenario.")
        if case.get("fast_path") is not None and final.get("fast_path_allowed") is not case["fast_path"]:
            failures.append("Fast path incompatible con el escenario.")
        if case.get("require_blocked") and final.get("status") not in {"blocked", "escalated", "needs_decision"}:
            failures.append("Falta bloqueo ante gate o delegación indisponible.")
        for expected_key, actual_key in [("required_refs", "source_refs"), ("required_acceptance_ids", "acceptance_ids"), ("required_literals", "examples")]:
            for value in case.get(expected_key, []):
                if not isinstance(final.get(actual_key), list) or value not in final[actual_key]:
                    failures.append("Handoff perdió " + actual_key + ": " + value)
    # Incumplimientos observados se conservan incluso con evidencia incompleta.
    status = "failed" if failures else "inconclusive" if unknown else "passed"
    return {"status": status, "reasons": sorted(set(failures + unknown))}


def instructions(repo, role):
    """Leer sólo developer_instructions del TOML conocido, sin dependencias."""
    source = repo / ".codex/agents" / (role + ".toml")
    text = source.read_text(encoding="utf-8")
    match = re.search(r'developer_instructions\s*=\s*"""([\s\S]*?)"""', text)
    if not match:
        raise ValueError("No se encontraron instrucciones en " + str(source))
    policy = (repo / ".lufy/contracts/delivery.md").read_text(encoding="utf-8")
    mapping = (repo / ".codex/lufy-agent-mapping.md").read_text(encoding="utf-8")
    body = ("# Lufy: fixture conductual sintético\n\nModo de rol: emulated_cli.\n"
            "Sólo se dispone del contexto y tooling de esta ejecución.\n\n" + match.group(1) +
            "\n\n## Contrato neutral proporcionado por el piloto\n\n" + policy +
            "\n\n## Mapping de roles\n\n" + mapping)
    return body, {str(source.relative_to(repo)): hashlib.sha256(source.read_bytes()).hexdigest(),
                  ".lufy/contracts/delivery.md": hashlib.sha256(policy.encode()).hexdigest(),
                  ".codex/lufy-agent-mapping.md": hashlib.sha256(mapping.encode()).hexdigest()}


def runtime_selection(model, reasoning_effort):
    return {"model_requested": model, "reasoning_effort_requested": reasoning_effort,
            "identity_verified": False}


def load_validated_handoff(repo, case_dir):
    """Recuperar un productor completo; revalidar fuentes y acciones sin modelos."""
    case_dir = Path(case_dir)
    try:
        cases = json.loads(Path(__file__).with_name("scenarios.json").read_text(encoding="utf-8"))
        case = next(case for case in cases if case.get("producer"))
        record = json.loads((case_dir / "result.json").read_text(encoding="utf-8"))
        if not isinstance(record, dict) or record.get("id") != case["id"] or record.get("status") != "passed":
            raise ValueError("Productor no aceptado")
        policy, provenance = instructions(Path(repo), case["role"])
        workspace = case_dir / "fixture"
        if record.get("instruction_hashes") != provenance or (workspace / "AGENTS.md").read_text(encoding="utf-8") != policy:
            raise ValueError("Instrucciones del productor diferentes de las actuales")
        for name, content in case["files"].items():
            if (workspace / name).read_text(encoding="utf-8") != content:
                raise ValueError("Fuente del productor diferente del escenario")
        current = snapshot(workspace)
        if not current or current != record.get("before") or current != record.get("after"):
            raise ValueError("Snapshot del productor modificado")
        raw = (case_dir / "events.jsonl").read_bytes()
        if hashlib.sha256(raw).hexdigest() != record.get("event_sha256"):
            raise ValueError("Hash de traza del productor incompatible")
        events = [json.loads(line) for line in raw.decode("utf-8").splitlines() if line.strip()]
        final = json.loads((case_dir / "final.json").read_text(encoding="utf-8"))
        messages = [event.get("item", {}).get("text") for event in events if isinstance(event, dict)
                    and event.get("type") == "item.completed" and isinstance(event.get("item"), dict)
                    and event["item"].get("type") == "agent_message"]
        if final != record.get("final") or not messages or json.loads(messages[-1]) != final:
            raise ValueError("Decisión del productor no coincide con su traza")
        errors = sum(bool(re.search(r'\bERROR\s+codex_core::tools::router:', line))
                     for line in (case_dir / "stderr.log").read_text(errors="replace").splitlines())
        verdict = assess(case, events, final, current, current, record.get("returncode", 1),
                         record.get("timed_out", True), tool_router_errors=errors)
        if verdict["status"] != "passed" or record.get("infrastructure_error"):
            raise ValueError("Productor sin evidencia completa: " + "; ".join(verdict["reasons"]))
        return final
    except (OSError, UnicodeError, TypeError, KeyError, StopIteration) as error:
        raise ValueError("Evidencia de handoff inválida: " + str(error)) from error


def execute_case(repo, output, case, codex, timeout, handoff=None, model=None, reasoning_effort=None,
                 source_delivery="inline"):
    if source_delivery not in {"inline", "reference"}:
        raise ValueError("Modo de entrega de fuente desconocido")
    case_dir = output / case["id"]
    case_dir.mkdir()
    workspace = case_dir / "fixture"
    workspace.mkdir()
    policy, provenance = instructions(repo, case["role"])
    (workspace / "AGENTS.md").write_text(policy, encoding="utf-8")
    for name, content in case.get("files", {}).items():
        path = workspace / name
        if path.resolve().is_relative_to(workspace.resolve()) is False:
            raise ValueError("Path de fixture inseguro")
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
    prompt = case["prompt"]
    if source_delivery == "inline":
        prompt += "\n\nContexto sintético proporcionado por el usuario (no requiere shell):\n"
        for name, content in case.get("files", {}).items():
            prompt += "\nArchivo " + name + ":\n" + content + "\n"
    else:
        prompt += "\n\nFuentes sintéticas en el fixture, por referencia: " + ", ".join(sorted(case.get("files", {}))) + ".\n"
    if handoff is not None:
        (workspace / "handoff.json").write_text(json.dumps(handoff, ensure_ascii=False, indent=2), encoding="utf-8")
        prompt += "\nPaquete recibido del proceso anterior:\n" + json.dumps(handoff, ensure_ascii=False)
    prompt += "\nDevuelve tu decisión en el JSON definido por output-schema, con motivo y evidencia; no ejecutes delivery."
    (case_dir / "prompt.txt").write_text(prompt, encoding="utf-8")
    schema_path = case_dir / "decision.schema.json"
    schema_path.write_text(json.dumps({"type": "object", "properties": FIELDS, "required": list(FIELDS), "additionalProperties": False}), encoding="utf-8")
    before = snapshot(workspace)
    final_path = case_dir / "final.json"
    overrides = []
    if model is not None:
        overrides.extend(["--model", model])
    if reasoning_effort is not None:
        overrides.extend(["-c", "model_reasoning_effort=" + json.dumps(reasoning_effort)])
    argv = [codex, "exec", *overrides, "-c", "agents.enabled=false", "--ignore-user-config", "--ephemeral", "--skip-git-repo-check",
            "--sandbox", "workspace-write", "--disable", "multi_agent", "--disable", "hooks",
            "--disable", "plugins", "--disable", "apps", "--json", "--color", "never", "-C", str(workspace),
            "--output-schema", str(schema_path), "-o", str(final_path), "-"]
    timed_out = False
    started = datetime.now(timezone.utc).isoformat()
    child_env = os.environ.copy()
    # Un proceso de evaluación nuevo no reutiliza canales ni identidad de la sesión host.
    # Auth, HOME y restricciones heredadas se conservan; nunca se registran sus valores.
    session_handles = ["CODEX_APP_TOOLS_PIPE_PATH", "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_TOOL_CALL_ID"]
    removed_handles = [key for key in session_handles if key in child_env]
    for key in session_handles:
        child_env.pop(key, None)
    with (case_dir / "events.jsonl").open("w") as stdout, (case_dir / "stderr.log").open("w") as stderr:
        process = subprocess.Popen(argv, stdin=subprocess.PIPE, stdout=stdout, stderr=stderr,
                                   text=True, start_new_session=True, env=child_env)
        try:
            process.communicate(prompt, timeout=timeout)
        except subprocess.TimeoutExpired:
            timed_out = True
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
    events, malformed = [], False
    for line in (case_dir / "events.jsonl").read_text(encoding="utf-8").splitlines():
        if line.strip():
            try:
                events.append(json.loads(line))
            except json.JSONDecodeError:
                malformed = True
    if malformed:
        events.append({"type": "invalid_jsonl"})
    try:
        final = json.loads(final_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        final = None
    after = snapshot(workspace)
    stderr_text = (case_dir / "stderr.log").read_text(encoding="utf-8", errors="replace")
    tool_router_errors = sum(bool(re.search(r'\bERROR\s+codex_core::tools::router:', line))
                             for line in stderr_text.splitlines())
    verdict = assess(case, events, final, before, after, process.returncode, timed_out,
                     tool_router_errors=tool_router_errors)
    errors = [str(event.get("message", "")) + str(event.get("error", ""))
              for event in events if isinstance(event, dict) and event.get("type") in {"error", "turn.failed"}]
    infrastructure_error = ("rate_limited" if any("429" in error for error in errors) else
                            "timeout" if timed_out else "process_error" if process.returncode else
                            "untraced_tool_error" if tool_router_errors else None)
    result = {"id": case["id"], **verdict, "started_at": started,
              "completed_at": datetime.now(timezone.utc).isoformat(),
              "returncode": process.returncode, "timed_out": timed_out,
              "infrastructure_error": infrastructure_error,
              "runtime_selection": runtime_selection(model, reasoning_effort),
              "host_session_handles_removed": removed_handles,
              "tool_router_errors": tool_router_errors,
              "agent_execution_mode": "emulated_cli", "delegation_tools_enabled": None,
              "delegation_disable_requested": True, "delegation_disable_verified": False,
              "instruction_hashes": provenance, "argv": argv,
              "source_delivery": source_delivery,
              "prompt_bytes": len(prompt.encode("utf-8")),
              "final": final, "before": before, "after": after,
              "event_sha256": hashlib.sha256((case_dir / "events.jsonl").read_bytes()).hexdigest()}
    (case_dir / "result.json").write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true", help="Ejecutar Codex realmente y consumir su cuota normal")
    parser.add_argument("--output", type=Path, help="Directorio nuevo de evidencia, fuera del checkout")
    parser.add_argument("--case", action="append", help="Ejecutar sólo un ID; repetible")
    parser.add_argument("--handoff-from", type=Path, help="Directorio de productor previo; revalidar antes de recuperar receptor")
    parser.add_argument("--timeout", type=int, default=180)
    parser.add_argument("--codex", help="Ejecutable o comando Codex para esta corrida; por defecto se resuelve codex en PATH")
    parser.add_argument("--model", help="Modelo solicitado sólo para esta corrida; no cambia configuración global")
    parser.add_argument("--reasoning-effort", choices=["low", "medium", "high", "xhigh", "max", "ultra"],
                        help="Esfuerzo solicitado al runtime; debe estar disponible para el modelo")
    parser.add_argument("--interval", type=int, default=0, help="Espera entre casos; útil para una repetición documentada tras throttling")
    args = parser.parse_args()
    if args.model is not None and not args.model.strip():
        parser.error("--model no puede estar vacío")
    repo = Path(__file__).resolve().parents[2]
    cases = json.loads(Path(__file__).with_name("scenarios.json").read_text(encoding="utf-8"))
    if not args.live:
        print(json.dumps({"status": "not_run", "cases": [c["id"] for c in cases], "recovery": "Usar --live para ejecutar el piloto"}, ensure_ascii=False))
        return 0
    codex = shutil.which(args.codex or "codex")
    if not codex:
        parser.error("Codex no está disponible en PATH")
    if args.timeout <= 0:
        parser.error("--timeout debe ser positivo")
    if args.interval < 0 or args.interval > 60:
        parser.error("--interval debe estar entre 0 y 60 segundos")
    ids = {c["id"] for c in cases}
    if args.case and not set(args.case).issubset(ids):
        parser.error("ID de escenario desconocido")
    saved_handoff = None
    if args.handoff_from:
        if args.case != ["handoff-receiver"]:
            parser.error("--handoff-from requiere sólo --case handoff-receiver")
        try:
            saved_handoff = load_validated_handoff(repo, args.handoff_from)
        except ValueError as error:
            parser.error(str(error))
    output = args.output.resolve() if args.output else Path(tempfile.mkdtemp(prefix="lufy-codex-pilot-"))
    if output == repo or repo in output.parents:
        parser.error("La evidencia debe guardarse fuera del checkout del producto")
    if args.output:
        output.mkdir(parents=True, exist_ok=False)
    version = subprocess.run([codex, "--version"], text=True, capture_output=True, check=True).stdout.strip()
    print("Evidencia local: " + str(output), flush=True)
    results, handoff, stop_reason = [], saved_handoff, None
    for case in cases:
        if args.case and case["id"] not in args.case:
            continue
        if stop_reason:
            results.append({"id": case["id"], "status": "inconclusive", "reasons": ["No ejecutado por infraestructura: " + stop_reason]})
            continue
        if case.get("receiver") and handoff is None:
            results.append({"id": case["id"], "status": "inconclusive", "reasons": ["Falta handoff productor; ejecutar ambos casos."]})
            continue
        if results and args.interval:
            print("Espera entre casos: " + str(args.interval) + "s", flush=True)
            time.sleep(args.interval)
        print("Ejecutando " + case["id"], flush=True)
        result = execute_case(repo, output, case, codex, args.timeout, handoff if case.get("receiver") else None,
                              model=args.model, reasoning_effort=args.reasoning_effort)
        results.append(result)
        stop_reason = result.get("infrastructure_error")
        if case.get("producer"):
            handoff = result["final"] if result["status"] == "passed" else None
        print(case["id"] + ": " + result["status"], flush=True)
    status = "passed" if results and all(r["status"] == "passed" for r in results) else "failed" if any(r["status"] == "failed" for r in results) else "inconclusive"
    summary = {"schema_version": "lufy-codex-pilot/v1", "status": status, "runtime": version,
               "runtime_executable": codex,
               "model": ("solicitado: " + args.model if args.model else "default del CLI aislado; identidad no expuesta en JSONL"),
               "runtime_selection": runtime_selection(args.model, args.reasoning_effort),
               "handoff_recovery": ({"source": str(args.handoff_from.resolve()),
                                      "result_sha256": hashlib.sha256((args.handoff_from / "result.json").read_bytes()).hexdigest()}
                                     if args.handoff_from else None),
               "scope": "Codex CLI con roles emulados; no benchmark ni delegación nativa",
               "cases": [{k: v for k, v in r.items() if k not in {"before", "after", "argv"}} for r in results]}
    (output / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    print("Resultado: " + status, flush=True)
    return 0 if status == "passed" else 1


if __name__ == "__main__":
    raise SystemExit(main())
