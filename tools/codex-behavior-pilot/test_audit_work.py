"""El auditor conserva incertidumbre y rechaza evidencia alterada; no invoca modelos."""

import hashlib
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import audit_work
import work


class AuditTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name).resolve()
        self.case = work.load_cases()[0]
        work.prepare_fixture(self.root / "fixture", self.case, "Política del test")
        final = "Multiplica por 1.21, redondea a dos decimales y para 100 devuelve 121."
        self.events = [{"type": "item.completed", "item": {"type": "agent_message", "text": final}},
                       {"type": "turn.completed", "usage": {"input_tokens": 10, "output_tokens": 5}}]
        snap = work.snapshot(self.root / "fixture")
        self.record = {"id": self.case["id"], "status": "inconclusive", "before": snap, "after": snap,
                       "after_oracle": snap, "returncode": 0, "timed_out": False, "final": final,
                       "inputs": {"runtime_sha256": "a" * 64}, "inputs_after": {"runtime_sha256": "a" * 64}}
        (self.root / "prompt.txt").write_text(self.case["prompt"])
        (self.root / "stderr.log").write_text("")
        self.write()

    def write(self):
        raw = "\n".join(json.dumps(e) for e in self.events).encode()
        (self.root / "events.jsonl").write_bytes(raw)
        self.record["trace_sha256"] = hashlib.sha256(raw).hexdigest()
        (self.root / "result.json").write_text(json.dumps(self.record))

    def test_semantic_review_is_not_automatic_acceptance_and_record_is_immutable(self):
        before = (self.root / "result.json").read_bytes()
        with patch.object(work, "execute_case", side_effect=AssertionError("No ejecutar modelos")):
            result = audit_work.audit(self.root)
        self.assertEqual(result["original_status"], "inconclusive")
        self.assertEqual(result["audit_status"], "needs_review")
        self.assertFalse(result["comparison_eligible"])
        self.assertEqual(before, (self.root / "result.json").read_bytes())

    def test_new_command_analysis_keeps_original_verdict_and_missing_costs(self):
        self.events.insert(0, {"type": "item.completed", "item": {"type": "command_execution", "command": "git status --short"}})
        self.write()
        before = (self.root / "result.json").read_bytes()
        result = audit_work.audit(self.root)
        self.assertEqual(result["original_status"], "inconclusive")
        self.assertEqual(result["audit_status"], "needs_review")
        self.assertEqual(result["classifier_version"], work.CLASSIFIER_VERSION)
        self.assertEqual(result["command_observations"][0]["reason"], "git_status")
        self.assertFalse(result["comparison_eligible"])
        self.assertFalse(result["oracle_reexecuted"])
        self.assertIsNone(result["metrics"]["human_interventions"])
        self.assertEqual(before, (self.root / "result.json").read_bytes())

    def test_unknown_metrics_stay_missing_and_booleans_are_not_token_counts(self):
        result = audit_work.audit(self.root)
        self.assertEqual(result["metrics"]["input_tokens"], 10)
        for key in ["duration_ms", "monetary_cost", "human_interventions", "cached_input_tokens"]:
            self.assertIsNone(result["metrics"][key])
        self.assertIsNone(audit_work.usage([{"type": "turn.completed", "usage": {"input_tokens": True}}])["input_tokens"])

    def test_modified_trace_snapshot_prompt_or_provenance_never_passes(self):
        for kind in ["trace", "snapshot", "prompt", "provenance"]:
            with self.subTest(kind=kind), tempfile.TemporaryDirectory() as raw:
                import shutil
                copy = Path(raw).resolve() / "copy"
                shutil.copytree(self.root, copy)
                if kind == "trace": (copy / "events.jsonl").write_text("{}")
                elif kind == "snapshot": (copy / "fixture/price.py").write_text("alterado")
                elif kind == "prompt": (copy / "prompt.txt").write_text("otro pedido")
                else:
                    record = json.loads((copy / "result.json").read_text())
                    record["inputs_after"] = {"runtime_sha256": "b" * 64}
                    (copy / "result.json").write_text(json.dumps(record))
                result = audit_work.audit(copy)
                self.assertEqual(result["audit_status"], "inconclusive")
                self.assertFalse(result["scope_checks_passed"])

    def test_failed_attempt_is_not_hidden_by_unchanged_files(self):
        self.events.insert(0, {"type": "item.completed", "item": {"type": "file_change", "changes": []}})
        self.write()
        self.assertEqual(audit_work.audit(self.root)["audit_status"], "failed")

    def test_missing_turn_is_not_quality_failure_or_success(self):
        self.events = self.events[:1]
        self.record["returncode"] = 1
        self.write()
        result = audit_work.audit(self.root)
        self.assertEqual(result["audit_status"], "inconclusive")
        self.assertIsNone(result["metrics"]["input_tokens"])

    def test_symlinks_missing_or_invalid_records_are_inconclusive(self):
        for value in [None, [], {"id": "inventado"}]:
            (self.root / "result.json").write_text(json.dumps(value))
            self.assertEqual(audit_work.audit(self.root)["audit_status"], "inconclusive")
        link = self.root / "link"
        link.symlink_to(self.root, target_is_directory=True)
        self.assertEqual(audit_work.audit(link)["audit_status"], "inconclusive")

    def test_convincing_but_false_text_is_not_automatically_approved(self):
        for text in ["Para 100 devuelve 121.", "El descuento 1.21 transforma 100 en 79, con precisión garantizada."]:
            with self.subTest(text=text):
                self.record["final"] = text
                self.events[0]["item"]["text"] = text
                self.write()
                result = audit_work.audit(self.root)
                self.assertEqual(result["audit_status"], "needs_review")
                self.assertEqual(result["semantic_review"], "required")

    def test_empty_report_does_not_adopt_and_keeps_all_families_not_run(self):
        summary = audit_work.summarize([])
        self.assertEqual(summary["status"], "inconclusive")
        self.assertFalse(summary["adopted_from_comparison"])
        self.assertEqual(len(summary["families"]), 6)
        self.assertTrue(all(f["coverage"] == "not_run" for f in summary["families"]))


if __name__ == "__main__":
    unittest.main()
