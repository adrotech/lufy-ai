#!/usr/bin/env python3
"""Comparación opt-in de instalaciones A/B/C en fixtures; sin adopción automática."""

import argparse
import copy
import json
import os
import re
import shutil
import subprocess
import time
from pathlib import Path

import audit_work
import work

MINIMAL_POLICY = """# Agente principal

Responde en español y resuelve el pedido dentro del alcance autorizado.
Preserva los datos y archivos ajenos. Comprueba el resultado pertinente y reporta evidencia real y límites.
Consulta decisiones materiales sin respuesta disponible. No publiques ni hagas delivery sin autorización.
"""
ORDERS = ["ABC", "BCA", "CAB", "ABC", "BCA", "CAB"]
SOURCE_FILES = ["comparison.py", "work.py", "work-scenarios.json", "audit_work.py"]


def sources():
    base = Path(__file__).resolve().parent
    return {**{name: work.digest(base / name) for name in SOURCE_FILES},
            "work_policy": work.digest(base.parents[1] / work.POLICY)}


def followup(case, parent):
    """La segunda petición parte del artefacto entregado, conservando las fuentes previas."""
    result = copy.deepcopy(case)
    if case["id"] == "fix-filter":
        result["files"]["app.py"] = (parent / "app.py").read_text()
        result["prompt"] = ("Extendé select(rows, status): status puede ser un string exacto, una colección "
                            "de strings (list, tuple o set), o None. Una colección vacía devuelve []; None "
                            "conserva todos. Preservá registros, orden y comportamiento de strings. "
                            "Modificá sólo app.py y comprobá python3 -B test_app.py. Ya está autorizado.")
    elif case["id"] == "write-guide":
        result["files"]["guide.md"] = (parent / "guide.md").read_text()
        result["prompt"] = ("Actualizá guide.md para que el archivo exportado se llame gastos.csv: usá "
                            "report --format csv --output gastos.csv. Conservá propósito, instrucciones "
                            "para generarlo y ubicación. No cambies reference.md. Ya está autorizado.")
        result["required_literals"] = {"guide.md": ["report --format csv --output gastos.csv"]}
    else:
        raise ValueError("Seguimiento no definido")
    return result


def prepare(root, case, variant, binary, launch):
    root.mkdir(parents=True, exist_ok=False)
    for name, content in case["files"].items():
        work.relative_path(name)
        p = root / name
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(content)
    log = ""
    if variant in {"A", "C"}:
        argv = [str(binary), "install", "--target", str(root), "--tool", "codex", "--yes"]
        if variant == "C":
            argv.extend(["--delegation", "off"])
        p = subprocess.run(argv, cwd=launch, text=True, capture_output=True, timeout=30)
        log = p.stdout + p.stderr
        if p.returncode:
            raise RuntimeError("Falló la preparación de " + variant + ": " + log[-500:])
        config = root / ".lufy/config/project.yaml"
        text = config.read_text()
        text = re.sub(r'(parallel_execution:\n  enabled:) true', r'\1 false', text)
        config.write_text(text)
    else:
        (root / "AGENTS.md").write_text(MINIMAL_POLICY)
    with (root / "AGENTS.md").open("a") as stream:
        stream.write(work.fixture_constraints(case))
    return log


def metrics(result, events):
    return {**audit_work.usage(events), "duration_ms": result.get("duration_ms"),
            "human_interventions": None}


def run_study(repo, output, codex, baseline, current, model, timeout, live=False):
    repo, output = Path(repo).resolve(), Path(output).resolve()
    if output.exists() or output == repo or repo in output.parents:
        raise ValueError("El output debe ser externo y nuevo")
    output.mkdir(parents=True)
    launch = output / "launch"
    launch.mkdir()
    binaries = {"A": Path(baseline).resolve(), "C": Path(current).resolve()}
    binary_hashes = {key: work.digest(value) for key, value in binaries.items()}
    bin_dirs = {}
    for variant, binary in binaries.items():
        d = output / "bin" / variant
        d.mkdir(parents=True)
        (d / "lufy-ai").symlink_to(binary)
        bin_dirs[variant] = d
    fixed_sources = sources()
    cases = work.load_cases()
    plan = {"model_requested": model, "runtime_sha256": work.digest(codex), "sources": fixed_sources,
            "product_binary_sha256": binary_hashes, "timeout_seconds": timeout, "repetitions": 1,
            "max_model_turns": 24, "order": [{"case": c["id"], "variants": list(order)} for c, order in zip(cases, ORDERS)],
            "stop_policy": "Fallo común de lanzamiento/sandbox/proceso detiene invocaciones; timeout es resultado incompleto de la celda sin retries.",
            "limits": "Un par por celda; modelo solicitado fijo, identidad servida no atestada; costo humano/monetario no disponible.",
            "live": live}
    prepared = []
    # Todo el tratamiento y las entradas se fijan antes del primer modelo.
    for case, order in zip(cases, ORDERS):
        for variant in order:
            parent = output / variant / "initial"
            parent.mkdir(parents=True, exist_ok=True)
            root = parent / case["id"] / "fixture"
            begin = time.monotonic_ns()
            log = prepare(root, case, variant, binaries.get(variant), launch)
            (root.parent / "preparation.log").write_text(log)
            prepared.append({"case": case["id"], "variant": variant, "fixture": str(root),
                             "agents_sha256": work.digest(root / "AGENTS.md"), "snapshot": work.snapshot(root),
                             "preparation_ms": round((time.monotonic_ns() - begin) / 1000000)})
    plan["prepared"] = prepared
    (output / "plan.json").write_text(json.dumps(plan, ensure_ascii=False, indent=2) + "\n")
    if not live:
        return {"status": "prepared", "output": str(output), "model_turns": 0}
    results, stopped, parents = [], False, {}

    def execute(case, variant, stage, parent=None):
        nonlocal stopped
        if stopped:
            results.append({"case": case["id"], "variant": variant, "stage": stage, "status": "not_run", "reason": "Infraestructura previa incompleta"})
            return
        if sources() != fixed_sources or work.digest(codex) != plan["runtime_sha256"] or any(work.digest(binaries[v]) != h for v, h in binary_hashes.items()):
            stopped = True
            raise RuntimeError("Cambió una fuente o binario del estudio")
        folder = output / variant / stage
        folder.mkdir(parents=True, exist_ok=True)
        if stage == "followup":
            root = folder / case["id"] / "fixture"
            root.parent.mkdir(parents=True)
            source = Path(parents[(case["id"], variant)]["directory"]) / "fixture"
            shutil.copytree(source, root)
            manifest = root / ".lufy/managed-state/install-state.json"
            if manifest.is_file():
                state = json.loads(manifest.read_text())
                state["targetRoot"] = str(root)
                manifest.write_text(json.dumps(state, ensure_ascii=False, indent=2) + "\n")
        print(stage + " " + case["id"] + " " + variant, flush=True)
        result = work.execute_case(repo, folder, case, str(codex), timeout, model,
                                   prepared=True, collections=stage == "followup" and case["id"] == "fix-filter",
                                   path_prefix=bin_dirs.get(variant))
        directory = folder / case["id"]
        events = [json.loads(line) for line in (directory / "events.jsonl").read_text().splitlines() if line.strip()]
        row = {"case": case["id"], "variant": variant, "stage": stage, "status": result["status"],
               "directory": str(directory), "metrics": metrics(result, events), "timed_out": result["timed_out"],
               "returncode": result["returncode"], "launch_error": result["launch_error"],
               "reasons": result["reasons"], "changed_paths": result["changed_paths"],
               "record_sha256": work.digest(directory / "result.json"), "parent_record_sha256": parent}
        results.append(row)
        stopped = bool(result["launch_error"] or (result["returncode"] and not result["timed_out"]))
        parents[(case["id"], variant)] = row
        (output / "progress.json").write_text(json.dumps(results, ensure_ascii=False, indent=2) + "\n")
        print("  " + result["status"], flush=True)

    for case, order in zip(cases, ORDERS):
        for variant in order:
            execute(case, variant, "initial")
    for name in ["fix-filter", "write-guide"]:
        base = next(c for c in cases if c["id"] == name)
        for variant in "ABC":
            parent = parents.get((name, variant))
            if not parent or parent["status"] == "failed" or parent["timed_out"] or parent["returncode"] or not (Path(parent["directory"]) / "fixture" / base["allowed_writes"][0]).is_file():
                results.append({"case": name, "variant": variant, "stage": "followup", "status": "not_run", "reason": "Sin entrega completa del parent"})
                continue
            next_case = followup(base, Path(parent["directory"]) / "fixture")
            execute(next_case, variant, "followup", parent["record_sha256"])
    # Un snapshot distinto después de otro turno invalida la muestra, sin editar su registro.
    for row in results:
        if "directory" in row:
            r = json.loads((Path(row["directory"]) / "result.json").read_text())
            row["post_snapshot_stable"] = work.snapshot(Path(row["directory"]) / "fixture") == r["after_oracle"]
    summary = {"status": "inconclusive", "adopted_from_comparison": False, "results": results,
               "model_turns": sum("directory" in r and not r.get("launch_error") for r in results),
               "sources_stable": sources() == fixed_sources,
               "limits": plan["limits"], "protocol_sha256": work.digest(output / "plan.json")}
    (output / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n")
    return {"status": summary["status"], "output": str(output), "model_turns": summary["model_turns"]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--codex", type=Path, required=True)
    parser.add_argument("--baseline", type=Path, required=True)
    parser.add_argument("--current", type=Path, required=True)
    parser.add_argument("--model", required=True)
    parser.add_argument("--timeout", type=int, default=60)
    args = parser.parse_args()
    if args.timeout <= 0 or not args.model.strip():
        parser.error("Modelo y timeout deben ser válidos")
    print(json.dumps(run_study(Path(__file__).resolve().parents[2], args.output, args.codex,
                               args.baseline, args.current, args.model, args.timeout, args.live), ensure_ascii=False))


if __name__ == "__main__":
    main()
