"""Contrato de selección runtime opcional; todos los procesos están simulados."""

import contextlib
import copy
import hashlib
import io
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import Mock, patch

import pilot


def selection(model=None, effort=None):
    return {"model_requested": model, "reasoning_effort_requested": effort,
            "identity_verified": False}


class ExecuteRuntimeOptionsTests(unittest.TestCase):
    def run_fixture(self, model=None, effort=None, explicit=False, capture_process_kwargs=False):
        """Producir JSONL/final sintéticos, sin invocar un binario real."""
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            output = root / "evidence"
            output.mkdir()
            final = {
                "status": "ready", "tier": "T3", "fast_path_allowed": True,
                "workload_decision": "Una unidad revisable.", "rationale": "Análisis acotado.",
                "source_refs": [], "acceptance_ids": [], "examples": [],
            }

            def fake_popen(argv, **kwargs):
                process = Mock(returncode=0)

                def communicate(prompt, timeout):
                    kwargs["stdout"].write(json.dumps({"type": "turn.completed"}) + "\n")
                    kwargs["stdout"].flush()
                    Path(argv[argv.index("-o") + 1]).write_text(json.dumps(final), encoding="utf-8")

                process.communicate.side_effect = communicate
                return process

            case = {"id": "runtime-fixture", "role": "sdd-router", "prompt": "Analiza sin escribir."}
            runtime = {"model": model, "reasoning_effort": effort} if explicit else {}
            with patch.object(pilot, "instructions", return_value=("# Instrucciones sintéticas", {})), \
                    patch.object(pilot.subprocess, "Popen", side_effect=fake_popen) as popen:
                result = pilot.execute_case(root, output, case, "/fixture/codex", 10, **runtime)
            persisted = json.loads((output / case["id"] / "result.json").read_text(encoding="utf-8"))
            self.assertEqual(persisted["runtime_selection"], result["runtime_selection"])
            self.assertEqual(result["argv"], popen.call_args.args[0])
            self.assertIsInstance(result["argv"], list)
            self.assertFalse(popen.call_args.kwargs.get("shell", False))
            return (result, popen.call_args.kwargs) if capture_process_kwargs else result

    def test_execute_default_does_not_inject_model_or_reasoning(self):
        result = self.run_fixture()
        self.assertEqual(result["runtime_selection"], selection())
        self.assertNotIn("--model", result["argv"])
        self.assertFalse(any("model_reasoning_effort" in part for part in result["argv"]))

    def test_execute_forwards_explicit_options_as_separate_argv(self):
        model = "fixture-model;literal"
        result = self.run_fixture(model, "low", explicit=True)
        argv = result["argv"]
        self.assertEqual(result["runtime_selection"], selection(model, "low"))
        self.assertEqual(argv[argv.index("--model") + 1], model)
        self.assertIn(["-c", 'model_reasoning_effort="low"'],
                      [argv[index:index + 2] for index in range(len(argv) - 1)])

    def test_execute_model_and_effort_are_independently_optional(self):
        for model, effort in [("fixture-model", None), (None, "high")]:
            with self.subTest(model=model, effort=effort):
                result = self.run_fixture(model, effort, explicit=True)
                argv = result["argv"]
                self.assertEqual(result["runtime_selection"], selection(model, effort))
                self.assertEqual("--model" in argv, model is not None)
                self.assertEqual(any("model_reasoning_effort" in part for part in argv), effort is not None)

    def test_execute_explicitly_disables_agents_without_claiming_verification(self):
        result = self.run_fixture()
        argv = result["argv"]
        self.assertIn(["-c", "agents.enabled=false"],
                      [argv[index:index + 2] for index in range(len(argv) - 1)])
        self.assertIn(["--disable", "apps"],
                      [argv[index:index + 2] for index in range(len(argv) - 1)])
        self.assertIs(result["delegation_disable_requested"], True)
        self.assertIs(result["delegation_disable_verified"], False)
        self.assertIsNone(result["delegation_tools_enabled"])

    def test_execute_isolates_host_session_handles_without_removing_ambient_auth_or_sandbox(self):
        handles = ["CODEX_APP_TOOLS_PIPE_PATH", "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_TOOL_CALL_ID"]
        # Sólo valores sintéticos: nunca leer ni imprimir credenciales reales.
        ambient = {
            "HOME": "/fixture/home", "PATH": "/fixture/bin",
            "OPENAI_API_KEY": "synthetic-not-a-secret", "CODEX_HOME": "/fixture/codex-home",
            "CODEX_PERMISSION_PROFILE": "fixture-profile", "CODEX_SANDBOX": "fixture-sandbox",
            **{name: "synthetic-host-handle" for name in handles},
        }
        with patch.dict(pilot.os.environ, ambient, clear=True):
            _, kwargs = self.run_fixture(capture_process_kwargs=True)
            self.assertIn("env", kwargs, "Popen debe recibir un entorno aislado explícito.")
            child = kwargs["env"]
            self.assertIsInstance(child, dict)
            self.assertIsNot(child, pilot.os.environ)
            for name in handles:
                self.assertNotIn(name, child)
            for name in set(ambient) - set(handles):
                self.assertEqual(child.get(name), ambient[name], "Debe preservar variable ambiental " + name)
            self.assertEqual(dict(pilot.os.environ), ambient, "No debe mutar el entorno del host.")


class MainRuntimeOptionsTests(unittest.TestCase):
    def run_main_fixture(self, flags, expect_error=False, resolved_executable="/fixture/codex",
                         expected_executable_query=None):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            module = root / "repo" / "tools" / "codex-behavior-pilot" / "pilot.py"
            module.parent.mkdir(parents=True)
            module.with_name("scenarios.json").write_text(json.dumps([
                {"id": "runtime-fixture", "role": "sdd-router", "prompt": "Analiza."}
            ]), encoding="utf-8")
            output = root / "evidence"
            result = {"id": "runtime-fixture", "status": "passed", "final": {}}
            with patch.object(pilot, "__file__", str(module)), \
                    patch.object(pilot.shutil, "which", return_value=resolved_executable) as which, \
                    patch.object(pilot.subprocess, "run", return_value=Mock(stdout="codex-cli fixture\n")) as run, \
                    patch.object(pilot.subprocess, "Popen") as popen, \
                    patch.object(pilot, "execute_case", return_value=result) as execute, \
                    patch("sys.argv", ["pilot.py", "--live", "--output", str(output), *flags]), \
                    contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
                if expect_error:
                    with self.assertRaises(SystemExit) as error:
                        pilot.main()
                    self.assertEqual(error.exception.code, 2)
                    self.assertFalse(output.exists(), "Input inválido no debe crear evidencia.")
                    execute.assert_not_called()
                    run.assert_not_called()
                    popen.assert_not_called()
                    if expected_executable_query is not None:
                        which.assert_called_once_with(expected_executable_query)
                    return None
                self.assertEqual(pilot.main(), 0)
                summary = json.loads((output / "summary.json").read_text(encoding="utf-8"))
                execute.assert_called_once()
                popen.assert_not_called()
                if expected_executable_query is not None:
                    which.assert_called_once_with(expected_executable_query)
                    self.assertEqual(run.call_args.args[0], [resolved_executable, "--version"])
                    self.assertEqual(execute.call_args.args[3], resolved_executable)
                return summary, execute.call_args

    def test_main_records_requested_selection_without_claiming_identity(self):
        summary, call = self.run_main_fixture(["--model", "fixture-model", "--reasoning-effort", "low"])
        self.assertEqual(summary["runtime_selection"], selection("fixture-model", "low"))
        self.assertEqual(call.kwargs["model"], "fixture-model")
        self.assertEqual(call.kwargs["reasoning_effort"], "low")

    def test_main_default_preserves_runtime_selection_as_none(self):
        summary, call = self.run_main_fixture([])
        self.assertEqual(summary["runtime_selection"], selection())
        self.assertIsNone(call.kwargs["model"])
        self.assertIsNone(call.kwargs["reasoning_effort"])

    def test_main_accepts_each_supported_reasoning_effort(self):
        for effort in ["low", "medium", "high", "xhigh", "max", "ultra"]:
            with self.subTest(effort=effort):
                summary, call = self.run_main_fixture(["--reasoning-effort", effort])
                self.assertEqual(summary["runtime_selection"], selection(None, effort))
                self.assertIsNone(call.kwargs["model"])
                self.assertEqual(call.kwargs["reasoning_effort"], effort)

    def test_main_rejects_empty_model_before_creating_output(self):
        for value in ["", " ", "\t\n"]:
            with self.subTest(model=value):
                self.run_main_fixture(["--model", value], expect_error=True)

    def test_main_rejects_unsupported_reasoning_before_creating_output(self):
        for value in ["", "extreme", "LOW"]:
            with self.subTest(effort=value):
                self.run_main_fixture(["--reasoning-effort", value], expect_error=True)

    def test_codex_selector_defaults_to_codex_and_records_resolved_executable(self):
        summary, _ = self.run_main_fixture([], expected_executable_query="codex")
        self.assertEqual(summary["runtime_executable"], "/fixture/codex")
        self.assertEqual(summary["runtime_selection"], selection())

    def test_codex_selector_accepts_path_or_command_without_changing_model_selection(self):
        for requested, resolved in [
            ("codex-app-fixture", "/fixture/app-codex"),
            ("/fixture/App Folder/codex", "/fixture/App Folder/codex"),
        ]:
            with self.subTest(requested=requested):
                summary, _ = self.run_main_fixture(
                    ["--codex", requested, "--model", "fixture-model", "--reasoning-effort", "low"],
                    resolved_executable=resolved, expected_executable_query=requested)
                self.assertEqual(summary["runtime_executable"], resolved)
                self.assertEqual(summary["runtime_selection"], selection("fixture-model", "low"))

    def test_codex_selector_missing_executable_rejects_before_output_or_process(self):
        self.run_main_fixture(["--codex", "missing-codex-fixture"], expect_error=True,
                              resolved_executable=None,
                              expected_executable_query="missing-codex-fixture")


class ValidatedHandoffTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        self.case_dir = self.root / "evidence" / "handoff-producer"
        self.workspace = self.case_dir / "fixture"
        self.workspace.mkdir(parents=True)
        self.policy = "# Instrucciones sintéticas del productor\n"
        self.provenance = {".codex/agents/orchestrator.toml": "a" * 64,
                           ".lufy/contracts/delivery.md": "b" * 64}
        (self.workspace / "AGENTS.md").write_text(self.policy, encoding="utf-8")
        cases = json.loads(Path(pilot.__file__).with_name("scenarios.json").read_text(encoding="utf-8"))
        case = next(case for case in cases if case["id"] == "handoff-producer")
        for name, content in case["files"].items():
            (self.workspace / name).write_text(content, encoding="utf-8")
        self.final = {
            "status": "ready", "tier": "T2", "fast_path_allowed": False,
            "workload_decision": "Handoff documental de una unidad.",
            "rationale": "Se preservan referencias y ejemplos sin implementar.",
            "source_refs": ["request.md"], "acceptance_ids": ["AC-H1", "AC-H2"],
            "examples": ["Hola <mundo> & adiós", '{"email":"","password":""}'],
        }
        self.events = [{"type": "item.completed", "item": {"type": "agent_message", "text": json.dumps(self.final)}},
                       {"type": "turn.completed"}]
        raw = "".join(json.dumps(event) + "\n" for event in self.events).encode("utf-8")
        (self.case_dir / "events.jsonl").write_bytes(raw)
        (self.case_dir / "stderr.log").write_text("", encoding="utf-8")
        (self.case_dir / "final.json").write_text(json.dumps(self.final), encoding="utf-8")
        snapshot = pilot.snapshot(self.workspace)
        self.record = {
            "id": "handoff-producer", "status": "passed", "reasons": [],
            "returncode": 0, "timed_out": False, "infrastructure_error": None,
            "before": snapshot, "after": copy.deepcopy(snapshot), "final": self.final,
            "event_sha256": hashlib.sha256(raw).hexdigest(), "instruction_hashes": self.provenance,
            "tool_router_errors": 0, "source_delivery": "inline",
            "agent_execution_mode": "emulated_cli", "delegation_tools_enabled": None,
            "delegation_disable_requested": True, "delegation_disable_verified": False,
            "runtime_selection": selection(), "argv": ["/fixture/codex", "exec"],
        }
        self.write_record(self.record)

    def write_record(self, record):
        (self.case_dir / "result.json").write_text(json.dumps(record), encoding="utf-8")

    def load_handoff(self):
        with patch.object(pilot, "instructions", return_value=(self.policy, self.provenance)):
            return pilot.load_validated_handoff(self.root, self.case_dir)

    def test_validated_handoff_returns_verified_final_without_mutating_evidence(self):
        before = {str(path): path.read_bytes() for path in self.case_dir.rglob("*") if path.is_file()}
        self.assertEqual(self.load_handoff(), self.final)
        self.assertEqual(before, {str(path): path.read_bytes() for path in self.case_dir.rglob("*") if path.is_file()})

    def test_validated_handoff_rejects_incomplete_or_inconclusive_producer(self):
        for fields in [{"returncode": 1}, {"timed_out": True}, {"status": "inconclusive"}, {"final": None}]:
            with self.subTest(fields=fields):
                self.write_record({**self.record, **fields})
                with self.assertRaises(ValueError):
                    self.load_handoff()

    def test_validated_handoff_rejects_trace_hash_mismatch(self):
        with (self.case_dir / "events.jsonl").open("a", encoding="utf-8") as file:
            file.write(json.dumps({"type": "turn.completed"}) + "\n")
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_rejects_modified_source_snapshot(self):
        (self.workspace / "request.md").write_text("Fuente reemplazada tras el productor.", encoding="utf-8")
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_rejects_changed_current_instruction_hashes(self):
        self.provenance = {**self.provenance, ".codex/agents/orchestrator.toml": "c" * 64}
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_reassesses_forbidden_action_despite_cached_pass(self):
        raw = "".join(json.dumps(event) + "\n" for event in [
            {"type": "item.completed", "item": {"type": "file_change", "status": "failed", "changes": []}},
            {"type": "turn.completed"},
        ]).encode("utf-8")
        (self.case_dir / "events.jsonl").write_bytes(raw)
        self.write_record({**self.record, "event_sha256": hashlib.sha256(raw).hexdigest()})
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_rejects_final_diverging_from_completed_trace(self):
        final = {**self.final, "rationale": "Texto reemplazado después de finalizar el turno."}
        self.write_record({**self.record, "final": final})
        (self.case_dir / "final.json").write_text(json.dumps(final), encoding="utf-8")
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_rejects_missing_source_even_with_consistent_snapshots(self):
        (self.workspace / "request.md").unlink()
        snapshot = pilot.snapshot(self.workspace)
        self.write_record({**self.record, "before": snapshot, "after": snapshot})
        with self.assertRaises(ValueError):
            self.load_handoff()

    def test_validated_handoff_rejects_altered_agents_despite_consistent_snapshots(self):
        (self.workspace / "AGENTS.md").write_text("Instrucciones sustituidas.", encoding="utf-8")
        snapshot = pilot.snapshot(self.workspace)
        self.write_record({**self.record, "before": snapshot, "after": snapshot})
        with self.assertRaises(ValueError):
            self.load_handoff()


if __name__ == "__main__":
    unittest.main()
