"""Integridad de trazas JSONL/stderr con procesos sintéticos, sin live Codex."""

import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import Mock, patch

import pilot


ROUTER_ERROR = "ERROR codex_core::tools::router: error=collab spawn failed: no thread with id"
FINAL = {
    "status": "blocked", "tier": "T2", "fast_path_allowed": False,
    "workload_decision": "Delegación requerida y tooling no disponible.",
    "rationale": "Se informa el bloqueo sin ejecutar el trabajo inline.",
    "source_refs": [], "acceptance_ids": [], "examples": [],
}
CASE = {"id": "trace-fixture", "role": "orchestrator", "prompt": "Delega el análisis.",
        "allowed_tiers": ["T2"], "fast_path": False,
        "require_blocked": True, "forbid_commands": True}


def events(mutation=False):
    result = [{"type": "turn.started"},
              {"type": "item.completed", "item": {"type": "agent_message", "text": "Bloqueado."}}]
    if mutation:
        result.append({"type": "item.completed", "item": {
            "type": "file_change", "status": "completed",
            "changes": [{"path": "runtime.py", "kind": "update"}],
        }})
    return [*result, {"type": "turn.completed"}]


class AssessTraceIntegrityTests(unittest.TestCase):
    def test_router_error_makes_completed_blocked_claim_inconclusive(self):
        verdict = pilot.assess(CASE, events(), FINAL, {}, {}, tool_router_errors=1)
        self.assertEqual(verdict["status"], "inconclusive", verdict)
        self.assertTrue(verdict["reasons"])

    def test_zero_router_errors_preserve_pass(self):
        verdict = pilot.assess(CASE, events(), FINAL, {}, {}, tool_router_errors=0)
        self.assertEqual(verdict["status"], "passed", verdict)

    def test_observed_mutation_keeps_failure_precedence_over_trace_gap(self):
        verdict = pilot.assess(CASE, events(mutation=True), FINAL, {}, {}, tool_router_errors=2)
        self.assertEqual(verdict["status"], "failed", verdict)
        self.assertTrue(verdict["reasons"])


class ExecuteTraceIntegrityTests(unittest.TestCase):
    def run_fixture(self, stderr_text="", mutation=False, returncode=0):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            output = root / "evidence"
            output.mkdir()

            def fake_popen(argv, **kwargs):
                process = Mock(returncode=returncode)

                def communicate(prompt, timeout):
                    for event in events(mutation):
                        kwargs["stdout"].write(json.dumps(event) + "\n")
                    kwargs["stdout"].flush()
                    kwargs["stderr"].write(stderr_text)
                    kwargs["stderr"].flush()
                    Path(argv[argv.index("-o") + 1]).write_text(json.dumps(FINAL), encoding="utf-8")

                process.communicate.side_effect = communicate
                return process

            with patch.object(pilot, "instructions", return_value=("# Fixture sintético", {})), \
                    patch.object(pilot.subprocess, "Popen", side_effect=fake_popen) as popen:
                result = pilot.execute_case(root, output, CASE, "/fixture/codex", 10)
            popen.assert_called_once()
            self.assertFalse(popen.call_args.kwargs.get("shell", False))
            persisted = json.loads((output / CASE["id"] / "result.json").read_text(encoding="utf-8"))
            self.assertEqual(persisted, result)
            return result

    def test_stderr_router_error_prevents_false_green_with_successful_process(self):
        result = self.run_fixture(ROUTER_ERROR + "\n")
        self.assertEqual(result["status"], "inconclusive", result)
        self.assertEqual(result["returncode"], 0)
        self.assertEqual(result["tool_router_errors"], 1)
        self.assertEqual(result["infrastructure_error"], "untraced_tool_error")
        self.assertTrue(result["reasons"])

    def test_counter_counts_error_lines_not_substrings_or_warnings(self):
        stderr = (ROUTER_ERROR + " " + ROUTER_ERROR + "\n" + ROUTER_ERROR + "\n"
                  "WARN codex_core::tools::router: herramientas no disponibles\n"
                  "ERROR unrelated::component: fixture\n")
        result = self.run_fixture(stderr)
        self.assertEqual(result["tool_router_errors"], 2)
        self.assertEqual(result["status"], "inconclusive", result)

    def test_no_error_or_warning_preserves_pass_with_honest_disable_metadata(self):
        for stderr in ["", "WARN codex_core::tools::router: herramientas no disponibles\n"]:
            with self.subTest(stderr=stderr):
                result = self.run_fixture(stderr)
                self.assertEqual(result["status"], "passed", result)
                self.assertEqual(result["tool_router_errors"], 0)
                self.assertIsNone(result["infrastructure_error"])
                self.assertIsNone(result["delegation_tools_enabled"])
                self.assertIs(result["delegation_disable_requested"], True)
                self.assertIs(result["delegation_disable_verified"], False)

    def test_stderr_error_does_not_hide_observed_violation(self):
        result = self.run_fixture(ROUTER_ERROR + "\n", mutation=True)
        self.assertEqual(result["status"], "failed", result)
        self.assertEqual(result["tool_router_errors"], 1)

    def test_existing_process_error_remains_primary_infrastructure_error(self):
        result = self.run_fixture(ROUTER_ERROR + "\n", returncode=1)
        self.assertEqual(result["status"], "inconclusive", result)
        self.assertEqual(result["tool_router_errors"], 1)
        self.assertEqual(result["infrastructure_error"], "process_error")


if __name__ == "__main__":
    unittest.main()
