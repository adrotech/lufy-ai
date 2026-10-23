#!/usr/bin/env python3
"""Regresión opt-in del transporte nativo Codex; sin inferencias desde transcripts."""

import argparse
import hashlib
import json
import os
import re
import shlex
import shutil
import signal
import subprocess
import sys
from pathlib import Path

from pilot import snapshot

LIFECYCLE = {"SessionStart", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"}
TOOLS = {"spawn_agent", "wait", "send_input", "close_agent"}
COLLABORATION_TOOLS = {"spawn_agent", "followup_task", "list_agents", "wait_agent"}
FORBIDDEN = {"Bash", "apply_patch", "exec_command", "write_stdin"}
TOOL_ALIASES = {**{"collaboration" + name: name for name in COLLABORATION_TOOLS},
                **{"functions" + name: name for name in FORBIDDEN}}
TRUST_WARNING = "`--dangerously-bypass-hook-trust` is enabled. Enabled hooks may run without review for this invocation."


def trace_gate(events, hooks, before, after, returncode, timed_out, tool_router_errors, allowed_tools):
    def result(status, reason):
        return {"status": status, "reason": reason}
    if isinstance(before, dict) and before and isinstance(after, dict) and before != after:
        return result("failed", "source_mutation")
    if isinstance(hooks, list) and any(isinstance(h, dict) and h.get("tool") in FORBIDDEN
                                       for h in hooks if isinstance(h, dict) and isinstance(h.get("tool"), str)):
        return result("failed", "forbidden_tool_observed")
    if isinstance(events, list):
        for event in events:
            item = event.get("item") if isinstance(event, dict) else None
            if isinstance(item, dict) and isinstance(item.get("type"), str) and item["type"] in {"command_execution", "file_change"}:
                return result("failed", "forbidden_action_observed")
    if not isinstance(events, list) or not isinstance(hooks, list):
        return result("inconclusive", "trace_invalid")
    if not isinstance(before, dict) or not before or not isinstance(after, dict) or not after:
        return result("inconclusive", "source_evidence_missing")
    if returncode or timed_out or tool_router_errors:
        return result("inconclusive", "runtime_incomplete")
    event_types = {"thread.started", "turn.started", "turn.completed", "turn.failed", "error",
                   "item.started", "item.updated", "item.completed"}
    for event in events:
        if not isinstance(event, dict) or not isinstance(event.get("type"), str) or event["type"] not in event_types:
            return result("inconclusive", "event_unknown")
        item = event.get("item", {})
        if not isinstance(item, dict) or ("type" in item and not isinstance(item["type"], str)):
            return result("inconclusive", "item_invalid")
        if event["type"].startswith("item.") and item.get("type") not in {"agent_message", "reasoning", "error", "collab_agent_tool_call", "collab_tool_call"}:
            return result("inconclusive", "item_type_unknown")
        if item.get("type") == "error" and item.get("message") != TRUST_WARNING:
            return result("inconclusive", "error_item_unknown")
        if event["type"] in {"turn.failed", "error"}:
            return result("inconclusive", "runtime_error")
    if not any(e["type"] == "turn.completed" for e in events):
        return result("inconclusive", "turn_incomplete")
    known = LIFECYCLE | {"PreToolUse", "PostToolUse"}
    for h in hooks:
        if not isinstance(h, dict) or not isinstance(h.get("event"), str) or h["event"] not in known:
            return result("inconclusive", "hook_unknown")
        if h["event"] in {"PreToolUse", "PostToolUse"} and (not isinstance(h.get("tool"), str) or h["tool"] not in allowed_tools):
            return result("inconclusive", "tool_unknown")
    return None


def assess_native(events, hooks, before, after, returncode=0, timed_out=False, tool_router_errors=0):
    def result(status, reason):
        return {"status": status, "reason": reason}
    gate = trace_gate(events, hooks, before, after, returncode, timed_out, tool_router_errors, TOOLS)
    if gate:
        return gate
    starts = [h for h in hooks if h["event"] == "SubagentStart"]
    if len(starts) != 1 or starts[0].get("agent_type") != "reviewer" or not isinstance(starts[0].get("agent_id"), str) or not starts[0]["agent_id"]:
        return result("inconclusive", "native_reviewer_missing")
    child = starts[0]["agent_id"]
    if not any(h["event"] == "SessionStart" for h in hooks) or not any(h["event"] == "SessionEnd" for h in hooks):
        return result("inconclusive", "session_hooks_missing")
    stops = [h for h in hooks if h["event"] == "SubagentStop"]
    if len(stops) != 2 or any(h.get("agent_id") != child for h in stops):
        return result("inconclusive", "child_stops_missing")
    calls = [h for h in hooks if h["event"] == "PostToolUse"]
    if [h["tool"] for h in calls] != ["spawn_agent", "wait", "send_input", "wait", "close_agent"]:
        return result("inconclusive", "recovery_order_invalid")
    if any(h.get("success") is not True for h in calls) or calls[0].get("receivers") != [child]:
        return result("inconclusive", "tool_result_invalid")
    if any(h.get("target_ids") != [child] for h in calls[1:]):
        return result("inconclusive", "recovery_target_mismatch")
    for call, expected in [(calls[1], "empty"), (calls[3], "valid")]:
        completed = call.get("completed")
        if not isinstance(completed, list) or len(completed) != 1 or not isinstance(completed[0], dict):
            return result("inconclusive", "wait_payload_missing")
        payload = completed[0]
        if payload.get("agent_id") != child or payload.get("payload") != expected:
            return result("inconclusive", "wait_payload_invalid")
        if expected == "valid" and (not isinstance(payload.get("fingerprint"), str) or not re.fullmatch(r"[0-9a-f]{64}", payload["fingerprint"])):
            return result("inconclusive", "payload_fingerprint_missing")
    return result("passed", "native_recovery_observed")


def assess_collaboration(events, hooks, before, after, returncode=0, timed_out=False, tool_router_errors=0):
    result = {"status": "inconclusive", "reason": "collaboration_evidence_missing",
              "transport_status": "inconclusive", "role_status": "not_available",
              "cleanup_status": "not_available"}
    gate = trace_gate(events, hooks, before, after, returncode, timed_out, tool_router_errors, COLLABORATION_TOOLS)
    if gate:
        result.update(gate)
        return result
    calls = [h for h in hooks if h["event"] == "PostToolUse"]
    spawns = [h for h in calls if h["tool"] == "spawn_agent"]
    followups = [h for h in calls if h["tool"] == "followup_task"]
    if len(spawns) != 1 or len(followups) != 1 or any(
            h.get("success") is not True and not (
                (h["tool"] == "followup_task" and h.get("acknowledgement") == "empty") or
                (h["tool"] == "wait_agent" and h.get("wait_outcome") == "timed_out"))
            for h in calls):
        return result
    spawn, followup = spawns[0], followups[0]
    receivers = spawn.get("receivers")
    if spawn.get("role_requested") != "reviewer" or not isinstance(receivers, list) or len(receivers) != 1:
        return result
    child = receivers[0]
    if not isinstance(child, str) or not re.fullmatch(r"[0-9a-f]{64}", child):
        return result
    result["role_status"] = "requested"
    starts = [h for h in hooks if h["event"] == "SubagentStart"]
    if len(starts) > 1 or any(h.get("agent_type") != "reviewer" for h in starts):
        return result
    if len(starts) == 1 and starts[0].get("agent_type") == "reviewer" and starts[0].get("agent_id") == child:
        result["role_status"] = "observed"
    if followup.get("target_ids") != [child] or calls.index(spawn) >= calls.index(followup):
        return result
    empty_seen = valid_seen = pending_wait_timeout = False
    for index, call in enumerate(calls):
        if call["tool"] == "wait_agent" and call.get("wait_outcome") == "timed_out":
            pending_wait_timeout = True
        if call["tool"] != "list_agents":
            continue
        if call.get("target_ids") != [child] or index < calls.index(spawn):
            return result
        if call.get("inspection_valid") is not True or call.get("agent_ids") not in ([], [child]):
            return result
        completed = call.get("completed")
        if not isinstance(completed, list) or len(completed) > 1:
            return result
        if not completed:
            continue
        payload = completed[0]
        if not isinstance(payload, dict) or payload.get("agent_id") != child:
            return result
        if index < calls.index(followup):
            if payload.get("payload") != "empty":
                return result
            empty_seen = True
        else:
            if payload.get("payload") != "valid" or not isinstance(payload.get("fingerprint"), str) or not re.fullmatch(r"[0-9a-f]{64}", payload["fingerprint"]):
                return result
            valid_seen = True
            pending_wait_timeout = False
    if empty_seen and valid_seen and not pending_wait_timeout:
        result.update(transport_status="passed", reason="cleanup_not_available")
    return result


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()


def decode(value):
    if isinstance(value, str):
        try:
            return json.loads(value)
        except (ValueError, TypeError):
            return value
    if isinstance(value, dict) and isinstance(value.get("content"), list):
        content = value["content"]
        if len(content) == 1 and isinstance(content[0], dict) and content[0].get("type") == "text":
            return decode(content[0].get("text"))
    return value


def identifier(value):
    return hashlib.sha256(value.encode()).hexdigest() if isinstance(value, str) and value else ""


def normalize_hook(data, expected_hash):
    if not isinstance(data, dict):
        return {"event": "invalid"}
    event = data.get("hook_event_name")
    if not isinstance(event, str) or event not in LIFECYCLE | {"PreToolUse", "PostToolUse"}:
        return {"event": "invalid"}
    record = {"event": event}
    for key in ["agent_id", "agent_type"]:
        if key in data:
            record[key] = identifier(data[key]) if key == "agent_id" else data[key]
    if record["event"] not in {"PreToolUse", "PostToolUse"}:
        return record
    name = data.get("tool_name", "")
    record["tool"] = TOOL_ALIASES.get(name, name.rsplit(".", 1)[-1]) if isinstance(name, str) else "invalid"
    inputs, response = decode(data.get("tool_input")), decode(data.get("tool_response"))
    if record["tool"] == "followup_task" and event == "PostToolUse" and "tool_response" in data:
        record["acknowledgement"] = "empty" if response is None or response == "" or response == {} else "response"
    inputs = inputs if isinstance(inputs, dict) else {}
    response = response if isinstance(response, dict) else {}
    status = response.get("status")
    failed_status = isinstance(status, str) and status in {"error", "failed", "not_found"}
    record["success"] = (bool(response) and not any(k in response for k in ["error", "error_code"])
                         and response.get("isError") is not True and response.get("is_error") is not True
                         and response.get("timed_out") is not True and not failed_status)
    if record["tool"] == "wait_agent" and event == "PostToolUse":
        failed = (any(k in response for k in ["error", "error_code"])
                  or response.get("isError") is True or response.get("is_error") is True or failed_status)
        timeout = response.get("timed_out")
        record["wait_outcome"] = ("error" if failed else "timed_out" if timeout is True
                                  else "completed" if timeout is False else "unknown")
        record["success"] = record["wait_outcome"] == "completed"
    ids = inputs.get("ids", inputs.get("task_ids"))
    if ids is None:
        ids = [inputs.get("target", inputs.get("path_prefix", inputs.get("id", inputs.get("agent_id", inputs.get("task_id")))))]
    record["target_ids"] = [identifier(i) for i in ids] if isinstance(ids, list) else []
    if record["tool"] == "spawn_agent":
        record["receivers"] = [identifier(response.get("task_name", response.get("agent_id", response.get("task_id"))))]
        if "agent_type" in inputs:
            record["role_requested"] = inputs["agent_type"] if isinstance(inputs["agent_type"], str) else "invalid"
    if record["tool"] in {"wait", "list_agents"}:
        completed = []
        states = response.get("status")
        if isinstance(states, dict):
            for agent, state in states.items():
                if isinstance(state, dict) and "completed" in state:
                    completed.append((agent, state["completed"]))
        tasks = response.get("task_results")
        if isinstance(tasks, list):
            for task in tasks:
                if isinstance(task, dict) and task.get("state") == "completed":
                    completed.append((task.get("task_id"), task.get("task_result")))
        agents = response.get("agents")
        if record["tool"] == "list_agents":
            record["inspection_valid"] = isinstance(agents, list)
            record["agent_ids"] = []
            for agent in agents if isinstance(agents, list) else []:
                if not isinstance(agent, dict) or not isinstance(agent.get("agent_name"), str) or not agent["agent_name"]:
                    record["inspection_valid"] = False
                    continue
                record["agent_ids"].append(identifier(agent["agent_name"]))
                state = agent.get("agent_status")
                if isinstance(state, dict) and set(state) == {"completed"}:
                    completed.append((agent["agent_name"], state["completed"]))
                elif not isinstance(state, str) or state not in {"running", "pending_init"}:
                    record["inspection_valid"] = False
        record["completed"] = []
        for agent, raw in completed:
            payload = decode(raw)
            state = "empty" if payload is None or payload == "" else "invalid"
            if isinstance(payload, dict) and digest(payload) == expected_hash:
                state = "valid"
            item = {"agent_id": identifier(agent), "payload": state}
            if state == "valid":
                item["fingerprint"] = expected_hash
            record["completed"].append(item)
    return record


def observe(config_path):
    config = json.loads(Path(config_path).read_text())
    raw = sys.stdin.buffer.read(65537)
    try:
        data = json.loads(raw) if len(raw) <= 65536 else None
    except ValueError:
        data = None
    record = normalize_hook(data, config["expected_hash"])
    output = b"" if record["event"] in {"PreToolUse", "PostToolUse"} else b'{"continue":true}'
    if record["event"] in LIFECYCLE:
        try:
            process = subprocess.run([config["lufy"], "lifecycle", "codex"], input=raw,
                                     capture_output=True, timeout=8, cwd=config["workspace"])
            record["lufy_exit_code"] = process.returncode
            output = process.stdout if process.returncode == 0 else output
        except (OSError, subprocess.TimeoutExpired):
            record["lufy_exit_code"] = -1
    record["run_nonce"] = config["run_nonce"]
    fd = os.open(config["observations"], os.O_CREAT | os.O_APPEND | os.O_WRONLY, 0o600)
    try:
        os.write(fd, (json.dumps(record, ensure_ascii=False) + "\n").encode())
    finally:
        os.close(fd)
    if record["event"] == "PreToolUse" and record.get("tool") in FORBIDDEN:
        output = b'{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Fixture nativo: shell y edicion prohibidos"}}'
    sys.stdout.buffer.write(output)


def source_snapshot(workspace):
    return {k: v for k, v in snapshot(workspace).items()
            if k != ".lufy/skill-registry.json" and k != ".lufy/runtime" and not k.startswith(".lufy/runtime/")}


def toml_value(value):
    if isinstance(value, dict):
        return "{" + ",".join(json.dumps(k) + "=" + toml_value(v) for k, v in value.items()) + "}"
    if isinstance(value, list):
        return "[" + ",".join(toml_value(v) for v in value) + "]"
    if isinstance(value, (str, bool, int)):
        return json.dumps(value)
    raise ValueError("Tipo no permitido en el override de hooks")


def run_native(args):
    if os.name != "posix":
        raise ValueError("El runner live requiere POSIX; la evaluación offline es portable")
    repo = Path(__file__).resolve().parents[2]
    transport = getattr(args, "transport", "legacy")
    if transport not in {"legacy", "collaboration"}:
        raise ValueError("Perfil de transporte desconocido")
    output = args.output.resolve()
    if output.is_relative_to(repo.resolve()):
        raise ValueError("La evidencia debe quedar fuera del checkout")
    output.mkdir(mode=0o700)
    workspace = output / "fixture"
    (workspace / ".codex/agents").mkdir(parents=True)
    (workspace / ".lufy/config").mkdir(parents=True)
    source = repo / ".codex/agents/reviewer.toml"
    product_hooks = repo / ".codex/hooks.json"
    definitions = json.loads(product_hooks.read_text())
    if any(not definitions.get("hooks", {}).get(event) for event in LIFECYCLE):
        raise ValueError("Definición lifecycle Lufy incompleta")
    shutil.copyfile(product_hooks, output / "product-hooks.json")
    shutil.copyfile(source, workspace / ".codex/agents/reviewer.toml")
    (workspace / "AGENTS.md").write_text("Fixture sintético. No ejecutes shell ni edites archivos. Sólo delegación nativa, espera, recuperación y cierre. No emules reviewer.\n")
    (workspace / ".codex/config.toml").write_text("[features]\nmulti_agent=true\nhooks=true\n")
    (workspace / ".lufy/config/project.yaml").write_text("schema_version: 1\ntool: codex\nrun_ledger:\n  enabled: true\n")
    template = repo / "tools/lufy-cli-go/internal/resultcontract/testdata/valid-minimal.json"
    contract = json.loads(template.read_text())
    contract["executive_summary"] = "Transporte sintético recuperado; no representa una revisión productiva."
    contract["artifacts"] = {"changed": ["none"], "referenced": ["fixture"]}
    contract["next_recommended"] = {"owner": "none", "action": "none"}
    contract["skill_resolution"] = {"local_skills_used": ["none"], "selected_skill_paths": ["none"], "bootstrap_recommended": False, "notes": "not_applicable"}
    expected = output / "expected.json"
    expected.write_text(json.dumps(contract, ensure_ascii=False))
    lufy = shutil.which(args.lufy)
    codex = shutil.which(args.codex)
    if not lufy or not codex:
        raise ValueError("Ejecutable explícito no disponible")
    probe = subprocess.run([lufy, "doctor", "--target", str(workspace), "--codex", codex, "--json"],
                           stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, timeout=10)
    runtime = json.loads(probe.stdout).get("codex_runtime")
    if not isinstance(runtime, dict) or runtime.get("status") != "available":
        raise ValueError("Runtime no diagnosticado antes del ensayo")
    validation = subprocess.run([lufy, "result", "validate", "--file", str(expected), "--role", "reviewer", "--json"], capture_output=True, timeout=10)
    if validation.returncode:
        raise ValueError("Envelope sintético no validado antes del ensayo")
    (output / "expected-validation.json").write_bytes(validation.stdout)
    observer_config = output / "observer.json"
    nonce = os.urandom(16).hex()
    observer_config.write_text(json.dumps({"lufy": lufy, "workspace": str(workspace), "observations": str(output / "hooks.jsonl"),
                                           "run_nonce": nonce, "expected_hash": digest(contract)}))
    command = shlex.join([sys.executable, str(Path(__file__).resolve()), "hook", "--config", str(observer_config)])
    hooks = {"hooks": {event: [{"hooks": [{"type": "command", "command": command, "timeout": 3 if event == "SessionEnd" else 8}]}]
                       for event in sorted(LIFECYCLE | {"PreToolUse", "PostToolUse"})}}
    hook_delivery = getattr(args, "hook_delivery", "project")
    hook_definition = (workspace / ".codex/hooks.json") if hook_delivery == "project" else output / "fixture-hooks.json"
    hook_definition.write_text(json.dumps(hooks))
    inline_hooks = ["-c", "hooks=" + toml_value(hooks["hooks"])] if hook_delivery == "inline" else []
    before = source_snapshot(workspace)
    provenance_paths = {"runner": Path(__file__).resolve(), "reviewer": source, "hooks": product_hooks,
                        "template": template, "observer": observer_config, "expected": expected, "fixture_hooks": hook_definition}
    provenance = {key: hashlib.sha256(path.read_bytes()).hexdigest() for key, path in provenance_paths.items()}
    prompt = ("Ensayo sintético de integración. No uses shell, ediciones, plan ni otras herramientas. "
              "Invoca exactamente un subagente nativo reviewer, sin fork de historia. Su primer encargo es responder literalmente null, sin herramientas. "
              "Espera su resultado. Ese null es un handoff inválido deliberado: recupéralo UNA vez enviando al MISMO agente un segundo encargo. "
              "El segundo encargo es responder exactamente el siguiente JSON, sin herramientas ni markdown: " + json.dumps(contract, ensure_ascii=False) +
              "\nEspera el segundo resultado y cierra ese agente. Si no están disponibles las herramientas o reviewer exacto, declara blocked sin emularlo. "
              "Al finalizar responde sólo una breve descripción de lo observado; tu texto no es evidencia de capacidades.")
    role_overrides = []
    if transport == "collaboration":
        role_overrides = ["-c", "agents.reviewer.description=" + json.dumps("Reviewer del fixture sintético Lufy"),
                          "-c", "agents.reviewer.config_file=" + json.dumps(str(workspace / ".codex/agents/reviewer.toml"))]
        prompt = ("Ensayo sintético autorizado de delegación nativa. No uses shell, archivos, exec ni otras herramientas. "
                  "Usa sólo collaboration.spawn_agent, collaboration.wait_agent, collaboration.list_agents y collaboration.followup_task. "
                  "Invoca exactamente UN agente con agent_type reviewer, fork_turns none y task_name native_reviewer. "
                  "Su primer encargo: responder exactamente null, sin herramientas. Conserva el task_name canónico devuelto. "
                  "Espera mediante wait_agent si hace falta y llama collaboration.list_agents con path_prefix IGUAL al nombre canónico para observar completed null. "
                  "Sólo después envía UNA recuperación mediante collaboration.followup_task al MISMO target. El segundo encargo es responder literalmente este JSON, sin herramientas ni markdown: "
                  + json.dumps(contract, ensure_ascii=False) +
                  "\nEspera y llama otra vez collaboration.list_agents con ese mismo path_prefix para observar completed con el JSON. "
                  "No envíes más recuperaciones ni crees otros agentes. Este perfil no expone cierre: informa cleanup not_available; completed no equivale a close. "
                  "Si faltan herramientas o reviewer exacto, declara blocked sin emularlo. Tu texto final no prueba transporte ni capacidades.")
    argv = [codex, "exec", "--ignore-user-config", "--skip-git-repo-check", "--sandbox", "read-only",
            "--enable", "multi_agent", "--enable", "hooks", "--disable", "apps", "--disable", "plugins",
            "-c", "agents.enabled=true", "-c", "approval_policy=\"never\"", "-c", "model_reasoning_effort=" + json.dumps(args.reasoning_effort),
            "-c", "projects." + json.dumps(str(workspace)) + ".trust_level=\"trusted\"",
            *role_overrides, *inline_hooks, "--dangerously-bypass-hook-trust", "--model", args.model, "--json", "-C", str(workspace), "-"]
    env = os.environ.copy()
    for key in ["CODEX_APP_TOOLS_PIPE_PATH", "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_TOOL_CALL_ID"]:
        env.pop(key, None)
    timed_out = False
    with (output / "events.jsonl").open("w") as stdout, (output / "stderr.log").open("w") as stderr:
        process = subprocess.Popen(argv, stdin=subprocess.PIPE, stdout=stdout, stderr=stderr, text=True, env=env, start_new_session=True)
        try:
            process.communicate(prompt, timeout=args.timeout)
        except subprocess.TimeoutExpired:
            timed_out = True
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
    def read_events(path):
        if not path.exists():
            return []
        try:
            return [json.loads(line) for line in path.read_text().splitlines() if line.strip()]
        except ValueError:
            return [None]
    events = read_events(output / "events.jsonl")
    observations = read_events(output / "hooks.jsonl")
    stderr = (output / "stderr.log").read_text(errors="replace")
    router_errors = sum(bool(re.search(r"\bERROR\s+codex_core::tools::router:", line)) for line in stderr.splitlines())
    after = source_snapshot(workspace)
    assessor = assess_collaboration if transport == "collaboration" else assess_native
    verdict = assessor(events, observations, before, after, process.returncode, timed_out, router_errors)
    lifecycle_records = [h for h in observations if isinstance(h, dict) and isinstance(h.get("event"), str) and h["event"] in LIFECYCLE]
    observed_events = {h["event"] for h in lifecycle_records}
    lifecycle_verified = LIFECYCLE.issubset(observed_events) and all(h.get("lufy_exit_code") == 0 for h in lifecycle_records)
    if verdict["status"] == "passed" and not lifecycle_verified:
        verdict = {"status": "inconclusive", "reason": "lufy_lifecycle_not_verified"}
    correlation = bool(observations) and all(isinstance(h, dict) and h.get("run_nonce") == nonce for h in observations)
    if verdict["status"] == "passed" and not correlation:
        verdict = {"status": "inconclusive", "reason": "hook_run_mismatch"}
    if transport == "collaboration" and verdict.get("transport_status") == "passed" and not correlation:
        verdict.update(transport_status="inconclusive", reason="hook_run_mismatch")
    if any(hashlib.sha256(path.read_bytes()).hexdigest() != provenance[key] for key, path in provenance_paths.items()):
        verdict = {"status": "failed", "reason": "external_source_mutation"}
    report = {"schema_version": "codex-native-regression/v1", **verdict, "returncode": process.returncode,
              "timed_out": timed_out, "tool_router_errors": router_errors, "lufy_lifecycle_verified": lifecycle_verified,
              "runtime_requested": {"executable": codex, "model": args.model, "reasoning_effort": args.reasoning_effort},
              "runtime_probe": runtime, "source_hashes": provenance, "instrumented_hooks": True, "hook_delivery": hook_delivery,
              "hook_run_verified": correlation, "capability_overrides_verified": False,
              "observed_lifecycle": sorted(observed_events), "native_execution_verified": verdict["status"] == "passed",
              "runtime_identity_verified": False, "hook_trust": "invocation_only_vetted_fixture", "argv": argv,
              "before": before, "after": after, "event_sha256": hashlib.sha256((output / "events.jsonl").read_bytes()).hexdigest(),
              "hooks_sha256": hashlib.sha256((output / "hooks.jsonl").read_bytes()).hexdigest() if (output / "hooks.jsonl").exists() else None,
              "instruction_sha256": hashlib.sha256(source.read_bytes()).hexdigest(), "template_sha256": hashlib.sha256(template.read_bytes()).hexdigest()}
    if transport == "collaboration":
        report.update(schema_version="codex-native-regression/v2", transport_profile=transport,
                      cleanup_status="not_available", role_status=verdict.get("role_status", "not_available"),
                      transport_status=verdict.get("transport_status", "inconclusive"),
                      native_execution_verified=False,
                      native_transport_verified=verdict.get("transport_status") == "passed")
    (output / "result.json").write_text(json.dumps(report, ensure_ascii=False, indent=2))
    return report


def main():
    if len(sys.argv) > 1 and sys.argv[1] == "hook":
        parser = argparse.ArgumentParser()
        parser.add_argument("hook")
        parser.add_argument("--config", required=True)
        observe(parser.parse_args().config)
        return 0
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--codex", required=True)
    parser.add_argument("--lufy", required=True)
    parser.add_argument("--hook-delivery", choices=["project", "inline"], default="project")
    parser.add_argument("--transport", choices=["legacy", "collaboration"], default="legacy")
    parser.add_argument("--model", required=True)
    parser.add_argument("--reasoning-effort", default="medium", choices=["low", "medium", "high", "xhigh", "max", "ultra"])
    parser.add_argument("--timeout", type=int, default=240)
    args = parser.parse_args()
    if not args.live or not 1 <= args.timeout <= 600 or not args.model.strip():
        parser.error("Requiere --live explícito, modelo y timeout entre 1 y 600 segundos")
    try:
        report = run_native(args)
    except (ValueError, OSError, subprocess.TimeoutExpired) as error:
        parser.exit(2, str(error) + "\n")
    print(json.dumps({"status": report["status"], "reason": report["reason"], "output": str(args.output)}, ensure_ascii=False))
    return 0 if report["status"] == "passed" else 1


if __name__ == "__main__":
    raise SystemExit(main())
