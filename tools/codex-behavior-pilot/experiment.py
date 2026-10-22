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
    for field in ["before", "instruction_hashes", "runtime_selection"]:
        if baseline.get(field) != candidate.get(field):
            reasons.append("Fuentes, instrucciones o runtime diferentes: " + field)
    argv_left, argv_right = baseline.get("argv"), candidate.get("argv")
    if not (isinstance(argv_left, list) and argv_left and isinstance(argv_left[0], str) and argv_left[0].strip()
            and isinstance(argv_right, list) and argv_right and isinstance(argv_right[0], str) and argv_right[0].strip()
            and argv_left[0] == argv_right[0]):
        reasons.append("Ejecutable diferente o no verificable.")
    if all(value is not None for value in [*left.values(), *right.values()]):
        if right["prompt_bytes"] >= left["prompt_bytes"] or right["input_tokens"] >= left["input_tokens"]:
            reasons.append("No hay reducción estricta de prompt y tokens de entrada observados.")
        if right["duration_ms"] > left["duration_ms"]:
            reasons.append("La duración observada del candidato es mayor.")
    return {"eligible": not reasons, "reasons": sorted(set(reasons)), "baseline": left, "candidate": right}


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
    repo = Path(__file__).resolve().parents[2]
    output = args.output.resolve()
    if output == repo or repo in output.parents:
        parser.error("La evidencia debe quedar fuera del checkout")
    output.mkdir(parents=True, exist_ok=False)
    case = next(case for case in json.loads(Path(__file__).with_name("scenarios.json").read_text()) if case["id"] == "auto-chain")
    pairs, stopped = [], False
    for index in range(args.pairs):
        paths = {label: output / ("pair-" + str(index + 1)) / label for label in ["baseline", "candidate"]}
        for path in paths.values():
            path.mkdir(parents=True)
        order = ["baseline", "candidate"] if index % 2 == 0 else ["candidate", "baseline"]
        for label in order:
            print("Ejecutando par " + str(index + 1) + ": " + label, flush=True)
            result = execute_case(repo, paths[label], case, codex, args.timeout, model=args.model,
                                  reasoning_effort=args.reasoning_effort,
                                  source_delivery="inline" if label == "baseline" else "reference")
            if result["status"] != "passed":
                stopped = True
                break
            if args.interval:
                time.sleep(args.interval)
        pairs.append(compare_pair(paths["baseline"] / case["id"], paths["candidate"] / case["id"]))
        if stopped:
            break
    summary = {"schema_version": "lufy-codex-experiment/v1", "status": "evaluated", **summarize_pairs(pairs),
               "runtime": subprocess.run([codex, "--version"], capture_output=True, text=True, check=True).stdout.strip(),
               "runtime_executable": codex, "case": case["id"], "pairs": pairs,
               "stopped_on_incomplete_or_failed_case": stopped,
               "limits": "Muestra acotada; no demuestra coste, rendimiento global ni routing productivo."}
    (output / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({key: summary[key] for key in ["status", "decision", "adopted", "pair_count"]}), flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
