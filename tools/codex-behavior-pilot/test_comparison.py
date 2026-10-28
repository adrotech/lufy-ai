"""Calibración sintética de comparación y seguimientos; nunca ejecuta modelos."""

import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import comparison
import work


class FollowupTests(unittest.TestCase):
    def test_followup_uses_delivered_artifact_without_changing_parent(self):
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            (root / "app.py").write_text("def select(rows, status): return list(rows)\n")
            case = next(c for c in work.load_cases() if c["id"] == "fix-filter")
            before = (root / "app.py").read_bytes()
            followup = comparison.followup(case, root)
            self.assertEqual(followup["files"]["app.py"], before.decode())
            self.assertIn("colección", followup["prompt"])
            self.assertEqual((root / "app.py").read_bytes(), before)
            self.assertNotEqual(case["files"]["app.py"], "entrega cambiada")

    def test_writing_followup_changes_requested_literal_preserving_reference(self):
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            (root / "guide.md").write_text("Entrega inicial del modelo")
            case = next(c for c in work.load_cases() if c["id"] == "write-guide")
            followup = comparison.followup(case, root)
            self.assertEqual(followup["files"]["guide.md"], "Entrega inicial del modelo")
            self.assertEqual(followup["files"]["reference.md"], case["files"]["reference.md"])
            self.assertEqual(case["required_literals"]["guide.md"], ["report --format csv --output movimientos.csv"])
            self.assertEqual(followup["required_literals"]["guide.md"], ["report --format csv --output gastos.csv"])

    def test_collection_oracle_rejects_old_contract_and_accepts_full_new_contract(self):
        sources = [
            ("def select(rows, status): return [r for r in rows if status is None or r['status']==status]\n", "failed"),
            ("def select(rows, status):\n if status is None: return list(rows)\n allowed = {status} if isinstance(status, str) else set(status)\n return [r for r in rows if r['status'] in allowed]\n", "passed"),
            ("def select(rows, status):\n if status is None: return list(rows)\n return [r for r in rows if r['status'] in status]\n", "failed"),
        ]
        for source, expected in sources:
            with self.subTest(expected=expected), tempfile.TemporaryDirectory() as raw:
                fixture = Path(raw)
                (fixture / "app.py").write_text(source)
                # Fuentes de test conocidas; producción proporciona siempre sandbox.
                self.assertEqual(work._evaluate_filter(fixture, [], collections=True)["status"], expected)

    def test_prepared_fixture_cannot_overwrite_existing_result(self):
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            case = work.load_cases()[0]
            folder = root / case["id"]
            folder.mkdir()
            work.prepare_fixture(folder / "fixture", case, "Política")
            (folder / "result.json").write_text("no sobrescribir")
            with patch.object(work, "provenance", return_value={}), patch.object(work.subprocess, "Popen") as launch:
                with self.assertRaises(ValueError):
                    work.execute_case(root, root, case, "/runtime", 60, prepared=True)
            launch.assert_not_called()
            self.assertEqual((folder / "result.json").read_text(), "no sobrescribir")

    def test_preview_prepares_all_cells_without_launching_models(self):
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw).resolve()
            binary = root / "binary"
            binary.write_text("binario sintético")

            def prepare(fixture, case, variant, cli, launch):
                fixture.parent.mkdir(parents=True)
                work.prepare_fixture(fixture, case, "Política " + variant)
                return "fixture sintético"

            with patch.object(comparison, "prepare", side_effect=prepare), patch.object(work, "execute_case") as execute:
                result = comparison.run_study(root / "repo", root / "output", binary, binary, binary, "modelo-fijado", 60)
            execute.assert_not_called()
            plan = json.loads((root / "output/plan.json").read_text())
            self.assertEqual(len(plan["prepared"]), 18)
            self.assertEqual(plan["max_model_turns"], 24)
            self.assertEqual(result["model_turns"], 0)
            self.assertTrue(all(set(row["variants"]) == {"A", "B", "C"} for row in plan["order"]))


if __name__ == "__main__":
    unittest.main()
