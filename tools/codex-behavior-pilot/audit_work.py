#!/usr/bin/env python3
"""Auditar evidencia existente por familia, sin ejecutar modelos ni candidatos."""

import argparse
import hashlib
import json
import re
from pathlib import Path

import work

METRICS = ("input_tokens", "cached_input_tokens", "output_tokens", "reasoning_output_tokens")
FAMILIES = {"explain-code": "analysis", "propose-export": "proposal", "write-guide": "writing",
            "translate-guide": "translation", "fix-filter": "implementation", "analyze-join": "data_analysis"}


def read_regular(path, limit):
    path = Path(path)
    if any(part.is_symlink() for part in [path, *path.parents]):
        raise ValueError("La evidencia no puede ser un enlace simbólico")
    if not path.is_file() or path.stat().st_size > limit:
        raise ValueError("Evidencia ausente, no regular o excedida")
    return path.read_bytes()


def fingerprint(raw):
    return hashlib.sha256(raw).hexdigest()


def usage(events):
    turns = [e for e in events if isinstance(e, dict) and e.get("type") == "turn.completed"]
    result = {key: None for key in METRICS}
    for key in METRICS:
        values = [e.get("usage", {}).get(key) if isinstance(e.get("usage"), dict) else None for e in turns]
        if values and all(type(v) is int and v >= 0 for v in values):
            result[key] = sum(values)
    # Ausencia de telemetría humana, reloj o precio no equivale a costo cero.
    return {**result, "duration_ms": None, "human_interventions": None, "monetary_cost": None}


def audit(directory):
    directory = Path(directory)
    result = {"directory": str(directory), "case": None, "family": "unknown", "original_status": None,
              "audit_status": "inconclusive", "scope_checks_passed": False, "reasons": [],
              "metrics": usage([]), "fingerprints": {}, "comparison_eligible": False, "execution": "not_available"}
    try:
        raw = read_regular(directory / "result.json", 4 * 1024 * 1024)
        record = json.loads(raw)
        if not isinstance(record, dict):
            raise ValueError("El registro no es un objeto")
        case = next((c for c in work.load_cases() if c["id"] == record.get("id")), None)
        if case is None:
            raise ValueError("Caso no reconocido por este evaluador")
        result.update(case=case["id"], family=FAMILIES[case["id"]], original_status=record.get("status"))
        trace = read_regular(directory / "events.jsonl", 32 * 1024 * 1024)
        stderr = read_regular(directory / "stderr.log", 1024 * 1024).decode("utf-8")
        prompt = read_regular(directory / "prompt.txt", 1024 * 1024).decode("utf-8")
        result["fingerprints"] = {"record": fingerprint(raw), "trace": fingerprint(trace),
                                  "prompt": fingerprint(prompt.encode())}
        if fingerprint(trace) != record.get("trace_sha256"):
            raise ValueError("La traza cambió desde la captura")
        if prompt != case["prompt"]:
            raise ValueError("El pedido no coincide con el corpus del evaluador")
        inputs = record.get("inputs")
        if (not isinstance(inputs, dict) or not inputs or inputs != record.get("inputs_after")
                or not all(isinstance(v, str) and re.fullmatch(r"[a-f0-9]{64}", v) for v in inputs.values())):
            raise ValueError("Las fuentes/runtime no permanecieron estables en la captura")
        for key in ("before", "after", "after_oracle"):
            if not isinstance(record.get(key), dict) or not record[key]:
                raise ValueError("Falta snapshot de " + key)
        for name, body in case["files"].items():
            expected = {"type": "file", "sha256": fingerprint(body.encode())}
            if record["before"].get(name) != expected:
                raise ValueError("La entrada no coincide con el corpus: " + name)
        if work.snapshot(directory / "fixture") != record["after_oracle"]:
            raise ValueError("El fixture cambió después de la captura")
        events = [json.loads(line) for line in trace.decode().splitlines() if line.strip()]
        verdict = work.assess(case, events, record.get("final"), record["before"], record["after"],
                              record.get("returncode", 1), record.get("timed_out", True),
                              record.get("oracle"), stderr, directory / "fixture")
        problems = work.content_checks(directory / "fixture", case)
        if record["after"] != record["after_oracle"]:
            problems.append("El fixture cambió durante el oráculo")
        if problems:
            verdict["status"] = "failed"
            verdict["scope_checks_passed"] = False
            verdict["reasons"].extend(problems)
        # No reinterpretamos una prueba funcional declarada como una ejecución propia del auditor.
        complete = (type(record.get("returncode")) is int and record["returncode"] == 0
                    and record.get("timed_out") is False and not record.get("launch_error")
                    and any(isinstance(e, dict) and e.get("type") == "turn.completed" for e in events))
        result.update(execution="completed" if complete else "incomplete", audit_status=verdict["status"], scope_checks_passed=verdict["scope_checks_passed"],
                      reasons=verdict["reasons"], metrics=usage(events), policy_sha256=record.get("policy_sha256"),
                      runtime_sha256=inputs.get("runtime_sha256"), model_identity_verified=record.get("model_identity_verified") is True,
                      oracle_reexecuted=False, semantic_review="required" if case.get("manual_review") else "recorded_oracle")
    except (OSError, ValueError, TypeError, KeyError, StopIteration, UnicodeError) as error:
        result["reasons"].append(str(error))
    return result


def summarize(samples):
    families = []
    for case, family in FAMILIES.items():
        rows = [s for s in samples if s["case"] == case]
        families.append({"case": case, "family": family, "attempts": len(rows),
                         "statuses": [r["audit_status"] for r in rows],
                         "completed_attempts": sum(r.get("execution") == "completed" for r in rows),
                         "coverage": "observed" if rows else "not_run"})
    return {"status": "inconclusive", "adopted_from_comparison": False, "samples": samples, "families": families,
            "limitations": ["Sin pares A/B/C con condiciones fijadas no hay comparación causal.",
                            "needs_review requiere juicio semántico explícito; no es aprobación automática.",
                            "Los registros auditados conservan su veredicto original.",
                            "No hay evidencia de segunda modificación en este corpus."]}


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directories", nargs="*", type=Path, help="Directorios de casos ya capturados")
    args = parser.parse_args(argv)
    print(json.dumps(summarize([audit(path) for path in args.directories]), ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
