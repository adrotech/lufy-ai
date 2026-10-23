"""Orquestación offline; comparación y ejecución se simulan deliberadamente."""

import contextlib
import copy
import io
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import experiment


class ExperimentRunnerTests(unittest.TestCase):
    def run_fixture(self, first_status="passed", drift_after_first=False, initial_failure=False,
                    error_at=None, incomplete_at=None):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        root = Path(temporary.name).resolve()
        repo = root / "repo"
        module = repo / "tools/codex-behavior-pilot/experiment.py"
        module.parent.mkdir(parents=True)
        module.write_text("# Módulo sintético, nunca ejecutado.\n", encoding="utf-8")
        module.with_name("scenarios.json").write_text(json.dumps([
            {"id": "auto-chain", "role": "implementer", "prompt": "No hay aprobación post-plan."}
        ]), encoding="utf-8")
        output = root / "evidence"
        stable = {"schema_version": "lufy-codex-execution-provenance/v1",
                  "runtime": {"executable": str(root / "codex"), "sha256": "a" * 64,
                              "version": "codex-cli fixture-version"},
                  "sources": {path: "b" * 64 for path in experiment.PROVENANCE_FILES}}
        capture_count = 0
        invoked = []

        def capture(*args):
            nonlocal capture_count
            capture_count += 1
            if initial_failure:
                raise ValueError("synthetic unavailable provenance")
            value = copy.deepcopy(stable)
            if drift_after_first and capture_count == 3:
                value["runtime"]["sha256"] = "c" * 64
            return value

        def execute(repo, directory, case, codex, timeout, **options):
            invoked.append((directory.parent.name, directory.name, options))
            if len(invoked) == error_at:
                raise OSError("synthetic process creation failure")
            case_dir = directory / case["id"]
            case_dir.mkdir()
            status = first_status if len(invoked) == 1 else "passed"
            if len(invoked) == incomplete_at:
                status = "inconclusive"
            result = {"id": case["id"], "status": status}
            (case_dir / "result.json").write_text(json.dumps(result), encoding="utf-8")
            return result

        def compare(baseline, candidate):
            # Aislar main; la verificación semántica real vive en test_experiments.
            for directory in [baseline, candidate]:
                path = directory / "result.json"
                if path.exists():
                    record = json.loads(path.read_text(encoding="utf-8"))
                    self.assertIn("execution_provenance_before", record)
                    self.assertIn("execution_provenance_after", record)
            return {"eligible": True, "comparable": True, "reasons": [],
                    "baseline": {}, "candidate": {}}

        argv = ["experiment.py", "--live", "--output", str(output), "--codex", "fixture-codex",
                "--model", "fixture-model", "--reasoning-effort", "medium", "--interval", "0", "--pairs", "2"]
        with patch.object(experiment, "__file__", str(module)), \
                patch.object(experiment.shutil, "which", return_value=str(root / "codex")), \
                patch.object(experiment, "capture_provenance", side_effect=capture) as captured, \
                patch.object(experiment, "execute_case", side_effect=execute) as executed, \
                patch.object(experiment, "compare_pair", side_effect=compare) as compared, \
                patch.object(experiment.subprocess, "run") as process, \
                patch.object(experiment.subprocess, "Popen") as popen, \
                patch("sys.argv", argv), \
                contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
            if initial_failure:
                with self.assertRaises(SystemExit) as error:
                    experiment.main()
                self.assertEqual(error.exception.code, 2)
                self.assertFalse(output.exists())
                executed.assert_not_called()
                compared.assert_not_called()
                captured.assert_called_once()
                summary = None
            else:
                self.assertEqual(experiment.main(), 0)
                summary = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            process.assert_not_called()
            popen.assert_not_called()
        return summary, invoked, capture_count, output, stable

    def test_first_inconclusive_or_case_provenance_drift_stops_further_invocations(self):
        for first_status, drift in [("inconclusive", False), ("passed", True)]:
            with self.subTest(first_status=first_status, drift=drift):
                summary, invoked, _, output, _ = self.run_fixture(first_status, drift)
                self.assertEqual(len(invoked), 1)
                self.assertEqual(summary["cases_invoked"], 1)
                self.assertIs(summary["stopped_on_incomplete_or_failed_case"], True)
                self.assertIs(summary["adopted"], False)
                self.assertEqual(summary["decision"], "keep_baseline")
                expected = "provenance_changed_during_case" if drift else "case_incomplete_or_failed"
                self.assertEqual(summary["stop_reason"], expected)
                self.assertFalse((output / "pair-1/candidate/auto-chain/result.json").exists())
                if drift:
                    self.assertIs(summary["provenance_stable"], False)
                    self.assertIs(summary["pairs"][0]["eligible"], False)

    def test_initial_capture_failure_creates_no_output_and_invokes_no_case(self):
        summary, invoked, capture_count, _, _ = self.run_fixture(initial_failure=True)
        self.assertIsNone(summary)
        self.assertEqual(invoked, [])
        self.assertEqual(capture_count, 1)

    def test_two_pairs_alternate_order_and_preserve_pre_post_provenance_without_adopting(self):
        summary, invoked, capture_count, output, stable = self.run_fixture()
        self.assertEqual([(pair, label) for pair, label, _ in invoked], [
            ("pair-1", "baseline"), ("pair-1", "candidate"),
            ("pair-2", "candidate"), ("pair-2", "baseline"),
        ])
        self.assertEqual([options["source_delivery"] for _, _, options in invoked],
                         ["inline", "reference", "reference", "inline"])
        self.assertEqual(capture_count, 10)
        self.assertEqual(summary["cases_invoked"], 4)
        self.assertEqual(summary["pair_count"], 2)
        self.assertEqual(summary["decision"], "candidate_supported_in_sample")
        self.assertIs(summary["adopted"], False)
        self.assertIs(summary["provenance_stable"], True)
        for pair, label, options in invoked:
            self.assertEqual(options["model"], "fixture-model")
            record = json.loads((output / pair / label / "auto-chain/result.json").read_text(encoding="utf-8"))
            self.assertEqual(record["execution_provenance_before"], stable)
            self.assertEqual(record["execution_provenance_after"], stable)

    def test_execution_oserror_writes_honest_summary_and_stops_after_first_attempt(self):
        summary, invoked, _, output, _ = self.run_fixture(error_at=1)
        self.assertEqual(len(invoked), 1)
        self.assertEqual(summary["cases_invoked"], 1)
        self.assertEqual(summary["stop_reason"], "case_execution_error")
        self.assertIs(summary["stopped_on_incomplete_or_failed_case"], True)
        self.assertEqual(summary["decision"], "keep_baseline")
        self.assertIs(summary["adopted"], False)
        self.assertTrue((output / "summary.json").is_file())
        self.assertIs(summary["pairs"][0]["eligible"], False)

    def test_incomplete_second_pair_never_supports_candidate_even_if_comparator_claims_eligible(self):
        summary, invoked, _, _, _ = self.run_fixture(incomplete_at=3)
        self.assertEqual(len(invoked), 3)
        self.assertEqual(summary["cases_invoked"], 3)
        self.assertEqual(summary["pair_count"], 2)
        self.assertEqual(summary["stop_reason"], "case_incomplete_or_failed")
        self.assertEqual(summary["decision"], "keep_baseline")
        self.assertIs(summary["adopted"], False)
        self.assertIs(summary["pairs"][-1]["eligible"], False)


if __name__ == "__main__":
    unittest.main()
