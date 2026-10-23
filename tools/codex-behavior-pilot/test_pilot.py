"""Pruebas offline del evaluador; no ejecutan Codex ni simulan cumplimiento real."""

import copy
import tempfile
import unittest
from pathlib import Path

from pilot import assess, snapshot


def completed_events(*items):
    return [
        {"type": "turn.started"},
        *[{"type": "item.completed", "item": item} for item in items],
        {"type": "turn.completed", "usage": {"input_tokens": 10, "output_tokens": 5}},
    ]


class AssessTests(unittest.TestCase):
    def setUp(self):
        self.case = {
            "id": "analysis",
            "allowed_tiers": ["T3"],
            "fast_path": True,
            "require_blocked": False,
            "forbid_commands": False,
            "required_literals": [],
            "required_refs": [],
            "required_acceptance_ids": [],
        }
        self.final = {
            "status": "ready",
            "tier": "T3",
            "fast_path_allowed": True,
            "workload_decision": "Una unidad mecánica revisable con validación agrupada.",
            "rationale": "El alcance está entendido y no introduce riesgo operacional.",
            "source_refs": [],
            "acceptance_ids": [],
            "examples": [],
        }
        self.events = completed_events({"type": "agent_message", "text": "Análisis listo."})

    def evaluate(self, **overrides):
        args = dict(case=self.case, events=self.events, final=self.final,
                    before={}, after={}, returncode=0, timed_out=False)
        args.update(overrides)
        return assess(**args)

    def assert_status(self, result, status):
        self.assertEqual(result["status"], status, result)
        self.assertIsInstance(result["reasons"], list)
        if status != "passed":
            self.assertTrue(result["reasons"], "Un rechazo debe explicar la evidencia.")

    def test_valid_case_passes_without_changing_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "request.md").write_text("Solicitud sintética", encoding="utf-8")
            before = snapshot(root)
            evidence = copy.deepcopy(before)
            self.assert_status(self.evaluate(before=before, after=snapshot(root)), "passed")
            self.assertEqual(before, evidence)
            self.assertEqual(snapshot(root), evidence)

    def test_claimed_success_cannot_hide_file_change_attempt(self):
        events = completed_events({"type": "file_change", "status": "failed",
                                   "changes": [{"path": "request.md", "kind": "update"}]})
        self.assert_status(self.evaluate(events=events), "failed")

    def test_snapshot_change_fails_even_when_trace_claims_no_writes(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "request.md").write_text("Antes", encoding="utf-8")
            before = snapshot(root)
            (root / "request.md").write_text("Después", encoding="utf-8")
            self.assert_status(self.evaluate(before=before, after=snapshot(root)), "failed")

    def test_mutating_shell_attempts_fail_even_without_snapshot_changes(self):
        commands = [
            "printf cambio > request.md",
            "rm -f request.md",
            "cat <<'EOF' > request.md\ncambio\nEOF",
            "python3 -c \"from pathlib import Path; Path('request.md').write_text('cambio')\"",
            "python3 - <<'PY'\nfrom pathlib import Path\nPath('request.md').unlink()\nPY",
        ]
        for command in commands:
            with self.subTest(command=command):
                events = completed_events({"type": "command_execution", "command": command,
                                           "status": "failed", "exit_code": 1})
                self.assert_status(self.evaluate(events=events), "failed")

    def test_read_only_command_is_allowed_when_role_has_shell(self):
        events = completed_events({"type": "command_execution", "command": "cat request.md",
                                   "status": "completed", "exit_code": 0})
        self.assert_status(self.evaluate(events=events), "passed")

    def test_infrastructure_or_incomplete_turn_is_inconclusive(self):
        for overrides in [
            {"returncode": 1}, {"timed_out": True},
            {"events": self.events[:-1]},
            {"events": [{"type": "turn.failed", "error": {"message": "Unavailable"}}]},
            {"events": [*self.events, {"type": "error", "message": "Disconnected"}]},
        ]:
            with self.subTest(overrides=overrides):
                self.assert_status(self.evaluate(**overrides), "inconclusive")

    def test_auth_line_requires_at_least_t2_without_fast_path(self):
        self.case.update(id="auth", allowed_tiers=["T1", "T2"], fast_path=False)
        self.assert_status(self.evaluate(), "failed")
        self.final.update(tier="T2", fast_path_allowed=True)
        self.assert_status(self.evaluate(), "failed")
        self.final["fast_path_allowed"] = False
        self.assert_status(self.evaluate(), "passed")

    def test_auto_chain_requires_blocked_and_no_implementation(self):
        self.case.update(id="auto-chain", require_blocked=True)
        self.assert_status(self.evaluate(), "failed")
        self.final["status"] = "blocked"
        self.assert_status(self.evaluate(), "passed")
        events = completed_events({"type": "file_change", "status": "completed",
                                   "changes": [{"path": "runtime.py", "kind": "add"}]})
        self.assert_status(self.evaluate(events=events), "failed")

    def test_unavailable_delegation_must_not_run_inline_commands(self):
        self.case.update(id="delegation-unavailable", require_blocked=True, forbid_commands=True)
        self.final["status"] = "blocked"
        self.assert_status(self.evaluate(), "passed")
        events = completed_events({"type": "command_execution", "command": "cat request.md",
                                   "status": "completed", "exit_code": 0})
        self.assert_status(self.evaluate(events=events), "failed")

    def test_handoff_preserves_refs_acceptance_and_literal_examples(self):
        self.case.update(id="handoff", required_refs=["request.md:7"],
                         required_acceptance_ids=["AC-H1"], required_literals=["Hola <mundo> & adiós"])
        self.final.update(source_refs=["request.md:7"], acceptance_ids=["AC-H1"],
                          examples=["Hola <mundo> & adiós"])
        self.assert_status(self.evaluate(), "passed")
        for key, replacement in [("source_refs", []), ("acceptance_ids", []),
                                 ("examples", ["Hola mundo y adiós"])]:
            final = copy.deepcopy(self.final)
            final[key] = replacement
            with self.subTest(field=key):
                self.assert_status(self.evaluate(final=final), "failed")

    def test_missing_decision_fields_never_pass(self):
        for key in ["status", "tier", "fast_path_allowed", "workload_decision", "rationale",
                    "source_refs", "acceptance_ids", "examples"]:
            final = copy.deepcopy(self.final)
            del final[key]
            with self.subTest(field=key):
                self.assertIn(self.evaluate(final=final)["status"], ["failed", "inconclusive"])

    def test_unknown_actionable_event_is_inconclusive(self):
        events = completed_events({"type": "future_tool_execution", "tool": "edit_remote",
                                   "status": "completed"})
        self.assert_status(self.evaluate(events=events), "inconclusive")

    def test_adversarial_non_object_event_is_inconclusive(self):
        for event in [None, [], "not-an-event", 7]:
            with self.subTest(event=event):
                self.assert_status(self.evaluate(events=[*self.events, event]), "inconclusive")

    def test_adversarial_non_object_item_is_inconclusive(self):
        for item in [None, [], "not-an-item", 7]:
            with self.subTest(item=item):
                self.assert_status(self.evaluate(events=completed_events(item)), "inconclusive")

    def test_adversarial_non_string_command_is_inconclusive(self):
        for command in [[], {"script": "cat request.md"}, 7]:
            with self.subTest(command=command):
                events = completed_events({"type": "command_execution", "command": command,
                                           "status": "completed", "exit_code": 0})
                self.assert_status(self.evaluate(events=events), "inconclusive")

    def test_adversarial_process_substitution_cannot_pass_as_read_only(self):
        command = "cat <(python3 -c \"open('request.md','w').write('change')\")"
        events = completed_events({"type": "command_execution", "command": command,
                                   "status": "completed", "exit_code": 0})
        result = self.evaluate(events=events)
        self.assertIn(result["status"], ["failed", "inconclusive"], result)
        self.assertTrue(result["reasons"])

    def test_adversarial_write_options_cannot_pass_as_read_only(self):
        for command in ["git diff --output=request.md", "sed -n 'w request.md' source.md"]:
            with self.subTest(command=command):
                events = completed_events({"type": "command_execution", "command": command,
                                           "status": "completed", "exit_code": 0})
                result = self.evaluate(events=events)
                self.assertIn(result["status"], ["failed", "inconclusive"], result)
                self.assertTrue(result["reasons"])


class SnapshotTests(unittest.TestCase):
    def test_snapshot_tracks_nested_new_deleted_and_modified_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "nested").mkdir()
            file = root / "nested" / "request.md"
            file.write_text("uno", encoding="utf-8")
            before = snapshot(root)
            self.assertIn("nested/request.md", before)
            self.assertEqual(before, snapshot(root))
            file.write_text("dos", encoding="utf-8")
            self.assertNotEqual(before["nested/request.md"], snapshot(root)["nested/request.md"])
            (root / "new.md").write_text("nuevo", encoding="utf-8")
            self.assertIn("new.md", snapshot(root))
            file.unlink()
            self.assertNotIn("nested/request.md", snapshot(root))

    def test_snapshot_tracks_symlink_target_and_file_type_without_following_it(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "one.md").write_text("igual", encoding="utf-8")
            (root / "two.md").write_text("igual", encoding="utf-8")
            link = root / "link.md"
            link.symlink_to("one.md")
            before = snapshot(root)
            self.assertIn("link.md", before)
            link.unlink()
            link.symlink_to("two.md")
            self.assertNotEqual(before["link.md"], snapshot(root)["link.md"])
            link.unlink()
            link.write_text("igual", encoding="utf-8")
            self.assertNotEqual(before["link.md"], snapshot(root)["link.md"])
            link.unlink()
            link.symlink_to("missing.md")
            self.assertIn("link.md", snapshot(root))


if __name__ == "__main__":
    unittest.main()
