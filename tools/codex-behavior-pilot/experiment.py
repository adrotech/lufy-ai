#!/usr/bin/env python3
"""Ensayo acotado de contexto; ninguna decisión adopta cambios automáticamente."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import time
from datetime import datetime

from pilot import assess, execute_case

PROVENANCE_FILES = (
    "tools/codex-behavior-pilot/pilot.py", "tools/codex-behavior-pilot/experiment.py",
    "tools/codex-behavior-pilot/scenarios.json", ".codex/agents/implementer.toml",
    ".lufy/contracts/delivery.md", ".codex/lufy-agent-mapping.md",
    ".codex/instructions/proportional-workflow.md", "scripts/render-codex-instructions.py",
)


def capture_provenance(repo, codex):
    """Observación local; no atesta identidad del modelo ni bloquea archivos."""
    binary = Path(codex).resolve(strict=True)
    if not binary.is_file() or binary.stat().st_size > 512 * 1024 * 1024:
        raise ValueError("Ejecutable no regular o excedido")
    checksum, size = hashlib.sha256(), 0
    with binary.open("rb") as stream:
        for chunk in iter(lambda: stream.read(65536), b""):
            size += len(chunk)
            if size > 512 * 1024 * 1024:
                raise ValueError("Ejecutable excedido")
            checksum.update(chunk)
    version = subprocess.run([str(binary), "--version"], stdout=subprocess.PIPE,
                             stderr=subprocess.DEVNULL, text=True, check=True, timeout=5).stdout.strip()
    if not version or len(version) > 4096:
        raise ValueError("Versión no verificable")
    sources = {}
    for relative in PROVENANCE_FILES:
        source = Path(repo) / relative
        if source.is_symlink() or not source.is_file() or source.stat().st_size > 1024 * 1024:
            raise ValueError("Fuente de procedencia no regular o excedida")
        sources[relative] = hashlib.sha256(source.read_bytes()).hexdigest()
    return {"schema_version": "lufy-codex-execution-provenance/v1",
            "runtime": {"executable": str(binary), "sha256": checksum.hexdigest(), "version": version},
            "sources": sources}


def valid_provenance(value):
    if (not isinstance(value, dict) or set(value) != {"schema_version", "runtime", "sources"}
            or value.get("schema_version") != "lufy-codex-execution-provenance/v1"):
        return False
    runtime, sources = value.get("runtime"), value.get("sources")
    if (not isinstance(runtime, dict) or set(runtime) != {"executable", "sha256", "version"}
            or not isinstance(sources, dict) or set(sources) != set(PROVENANCE_FILES)):
        return False
    hashes = [runtime.get("sha256"), *sources.values()]
    return (all(isinstance(item, str) and re.fullmatch(r"[0-9a-f]{64}", item) for item in hashes)
            and all(isinstance(runtime.get(key), str) and runtime[key].strip()
                    for key in ["executable", "version"]))


def execution_profile(argv):
    if not isinstance(argv, list) or not argv or not all(isinstance(item, str) and item for item in argv):
        return None
    result, seen, index = [], set(), 0
    while index < len(argv):
        item = argv[index]
        result.append(item)
        if item in {"-C", "--output-schema", "-o"}:
            if item in seen or index + 1 >= len(argv) or argv[index + 1].startswith("-"):
                return None
            seen.add(item)
            result.append("<case-path>")
            index += 1
        index += 1
    return result


def _read_sample(directory, case):
    metrics = {"prompt_bytes": None, "input_tokens": None, "duration_ms": None}
    reasons = []
    record = {}
    try:
        record = json.loads((directory / "result.json").read_text(encoding="utf-8"))
        if not isinstance(record, dict):
            raise ValueError("resultado no objeto")
        raw = (directory / "events.jsonl").read_bytes()
        events = [json.loads(line) for line in raw.decode("utf-8").splitlines() if line.strip()]
        if hashlib.sha256(raw).hexdigest() != record.get("event_sha256"):
            reasons.append("Hash de JSONL incompatible.")
        errors = sum(bool(re.search(r'\bERROR\s+codex_core::tools::router:', line))
                     for line in (directory / "stderr.log").read_text(errors="replace").splitlines())
        if not isinstance(record.get("before"), dict) or not record["before"]:
            reasons.append("Snapshot de fuente ausente.")
        verdict = assess(case, events, record.get("final"), record.get("before"), record.get("after"),
                         record.get("returncode", 1), record.get("timed_out", True), tool_router_errors=errors)
        if verdict["status"] != "passed":
            reasons.extend(verdict["reasons"] or ["Caso sin evidencia de aceptación."])
        metrics["prompt_bytes"] = len((directory / "prompt.txt").read_bytes())
        turns = [event for event in events if isinstance(event, dict) and event.get("type") == "turn.completed"]
        values = [(event.get("usage") or {}).get("input_tokens")
                  if isinstance(event.get("usage"), dict) else None for event in turns]
        if values and all(type(value) is int and value >= 0 for value in values):
            metrics["input_tokens"] = sum(values)
        try:
            start = datetime.fromisoformat(record["started_at"])
            end = datetime.fromisoformat(record["completed_at"])
            elapsed = (end - start).total_seconds()
            if start.tzinfo is not None and end.tzinfo is not None and elapsed >= 0:
                metrics["duration_ms"] = round(elapsed * 1000)
        except (KeyError, TypeError, ValueError, OverflowError):
            pass
    except (OSError, ValueError, TypeError, KeyError):
        if not isinstance(record, dict):
            record = {}
        reasons.append("Muestra ausente o inválida.")
    if any(value is None for value in metrics.values()):
        reasons.append("Métricas observadas incompletas; no se sustituyen por cero.")
    return record, metrics, reasons


def compare_pair(baseline_case_dir, candidate_case_dir):
    cases = json.loads(Path(__file__).with_name("scenarios.json").read_text(encoding="utf-8"))
    case = next(case for case in cases if case["id"] == "auto-chain")
    baseline, left, reasons = _read_sample(Path(baseline_case_dir), case)
    candidate, right, errors = _read_sample(Path(candidate_case_dir), case)
    reasons.extend(errors)
    for record, mode in [(baseline, "inline"), (candidate, "reference")]:
        if record.get("id") != case["id"] or record.get("source_delivery") != mode:
            reasons.append("Caso o protocolo de entrega incompatible.")
        hashes = record.get("instruction_hashes")
        selection = record.get("runtime_selection")
        valid_hashes = (isinstance(hashes, dict) and bool(hashes)
                        and all(isinstance(key, str) and key.strip() and isinstance(value, str)
                                and re.fullmatch(r'[0-9a-f]{64}', value) for key, value in hashes.items()))
        valid_selection = (isinstance(selection, dict)
                           and isinstance(selection.get("model_requested"), str)
                           and bool(selection["model_requested"].strip())
                           and isinstance(selection.get("reasoning_effort_requested"), str)
                           and selection.get("reasoning_effort_requested") in {"low", "medium", "high", "xhigh", "max", "ultra"}
                           and selection.get("identity_verified") is False)
        if not valid_hashes or not valid_selection:
            reasons.append("Procedencia de instrucciones o runtime ausente.")
        before, after = record.get("execution_provenance_before"), record.get("execution_provenance_after")
        if not valid_provenance(before) or not valid_provenance(after) or before != after:
            reasons.append("Procedencia observada del ejecutable/fuentes ausente o cambiante.")
        else:
            observed_argv = record.get("argv")
            if not isinstance(observed_argv, list) or not observed_argv or before["runtime"]["executable"] != observed_argv[0]:
                reasons.append("Ejecutable observado incompatible con argv.")
    for field in ["before", "instruction_hashes", "runtime_selection", "execution_provenance_before"]:
        if baseline.get(field) != candidate.get(field):
            reasons.append("Fuentes, instrucciones o runtime diferentes: " + field)
    argv_left, argv_right = baseline.get("argv"), candidate.get("argv")
    if not (isinstance(argv_left, list) and argv_left and isinstance(argv_left[0], str) and argv_left[0].strip()
            and isinstance(argv_right, list) and argv_right and isinstance(argv_right[0], str) and argv_right[0].strip()
            and argv_left[0] == argv_right[0]):
        reasons.append("Ejecutable diferente o no verificable.")
    profile_left, profile_right = execution_profile(argv_left), execution_profile(argv_right)
    if profile_left is None or profile_right is None or profile_left != profile_right:
        reasons.append("Perfil de ejecución diferente o inválido.")
    comparable = not reasons
    if all(value is not None for value in [*left.values(), *right.values()]):
        if right["prompt_bytes"] >= left["prompt_bytes"] or right["input_tokens"] >= left["input_tokens"]:
            reasons.append("No hay reducción estricta de prompt y tokens de entrada observados.")
        if right["duration_ms"] > left["duration_ms"]:
            reasons.append("La duración observada del candidato es mayor.")
    return {"eligible": not reasons, "comparable": comparable, "reasons": sorted(set(reasons)), "baseline": left, "candidate": right}


def summarize_pairs(pairs):
    supported = len(pairs) >= 2 and all(pair.get("eligible") is True for pair in pairs)
    return {"decision": "candidate_supported_in_sample" if supported else "keep_baseline",
            "adopted": False, "pair_count": len(pairs)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--output", type=Path)
    parser.add_argument("--codex", default="codex")
    parser.add_argument("--model")
    parser.add_argument("--reasoning-effort", choices=["low", "medium", "high", "xhigh", "max", "ultra"], default="medium")
    parser.add_argument("--pairs", type=int, choices=[1, 2], default=2)
    parser.add_argument("--timeout", type=int, default=180)
    parser.add_argument("--interval", type=int, default=10)
    args = parser.parse_args()
    if not args.live:
        print(json.dumps({"status": "not_run", "case": "auto-chain", "pairs": args.pairs,
                          "variants": ["inline", "reference"], "adopted": False}))
        return 0
    if not args.output or not args.model or not args.model.strip() or args.timeout <= 0 or not 0 <= args.interval <= 60:
        parser.error("Requiere output nuevo, modelo explícito, timeout positivo e intervalo de 0 a 60")
    codex = shutil.which(args.codex)
    if not codex:
        parser.error("Ejecutable Codex no disponible")
    codex = str(Path(codex).resolve())
    repo = Path(__file__).resolve().parents[2]
    output = args.output.resolve()
    if output == repo or repo in output.parents:
        parser.error("La evidencia debe quedar fuera del checkout")
    try:
        initial_provenance = capture_provenance(repo, codex)
    except (OSError, ValueError, subprocess.SubprocessError):
        parser.error("Procedencia del runtime/fuentes no disponible; no se ejecutan modelos")
    output.mkdir(parents=True, exist_ok=False)
    case = next(case for case in json.loads(Path(__file__).with_name("scenarios.json").read_text()) if case["id"] == "auto-chain")
    pairs, stopped, provenance_failed, invoked, stop_reason = [], False, False, 0, None
    for index in range(args.pairs):
        paths = {label: output / ("pair-" + str(index + 1)) / label for label in ["baseline", "candidate"]}
        for path in paths.values():
            path.mkdir(parents=True)
        order = ["baseline", "candidate"] if index % 2 == 0 else ["candidate", "baseline"]
        for label in order:
            try:
                before = capture_provenance(repo, codex)
            except (OSError, ValueError, subprocess.SubprocessError):
                before = None
            if before != initial_provenance:
                stopped, provenance_failed, stop_reason = True, True, "provenance_changed_before_case"
                break
            print("Ejecutando par " + str(index + 1) + ": " + label, flush=True)
            invoked += 1
            try:
                result = execute_case(repo, paths[label], case, codex, args.timeout, model=args.model,
                                      reasoning_effort=args.reasoning_effort,
                                      source_delivery="inline" if label == "baseline" else "reference")
            except (OSError, ValueError, subprocess.SubprocessError):
                stopped, stop_reason = True, "case_execution_error"
                break
            try:
                after = capture_provenance(repo, codex)
            except (OSError, ValueError, subprocess.SubprocessError):
                after = None
            result.update(execution_provenance_before=before, execution_provenance_after=after)
            (paths[label] / case["id"] / "result.json").write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
            if before != after:
                stopped, provenance_failed, stop_reason = True, True, "provenance_changed_during_case"
                break
            if result["status"] != "passed":
                stopped = True
                stop_reason = "case_incomplete_or_failed"
                break
            if args.interval:
                time.sleep(args.interval)
        pairs.append(compare_pair(paths["baseline"] / case["id"], paths["candidate"] / case["id"]))
        if stopped:
            pairs[-1].update(eligible=False, comparable=False)
            pairs[-1]["reasons"] = sorted(set(pairs[-1]["reasons"] + ["Caso incompleto/fallido o procedencia inválida."]))
        if not pairs[-1]["comparable"] and not stopped:
            stopped, stop_reason = True, "pair_not_comparable"
        if stopped:
            break
    try:
        final_provenance = capture_provenance(repo, codex)
    except (OSError, ValueError, subprocess.SubprocessError):
        final_provenance = None
    if initial_provenance != final_provenance:
        stopped, provenance_failed, stop_reason = True, True, "provenance_changed_after_run"
    if provenance_failed:
        for pair in pairs:
            pair.update(eligible=False, comparable=False)
            pair["reasons"] = sorted(set(pair["reasons"] + ["Procedencia global cambiante o no disponible."]))
    summary = {"schema_version": "lufy-codex-experiment/v1", "status": "evaluated", **summarize_pairs(pairs),
               "runtime": initial_provenance["runtime"]["version"],
               "runtime_executable": codex, "case": case["id"], "pairs": pairs,
               "execution_provenance_before": initial_provenance, "execution_provenance_after": final_provenance,
               "provenance_stable": not provenance_failed, "cases_invoked": invoked, "stop_reason": stop_reason,
               "stopped_on_incomplete_or_failed_case": stopped,
               "limits": "Muestra acotada; no demuestra coste, rendimiento global ni routing productivo."}
    (output / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({key: summary[key] for key in ["status", "decision", "adopted", "pair_count"]}), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
