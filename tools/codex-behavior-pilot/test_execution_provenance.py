"""Procedencia experimental offline; no inicia modelos ni usa fuentes privadas."""

import copy
import hashlib
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import experiment
import test_experiments as fixtures


EXPECTED_FILES = (
    "tools/codex-behavior-pilot/pilot.py",
    "tools/codex-behavior-pilot/experiment.py",
    "tools/codex-behavior-pilot/scenarios.json",
    ".codex/agents/implementer.toml",
    ".lufy/contracts/delivery.md",
    ".codex/lufy-agent-mapping.md",
    ".codex/instructions/proportional-workflow.md",
    "scripts/render-codex-instructions.py",
)


class ExecutionProvenanceTests(unittest.TestCase):
    def setUp(self):
        self.helper = fixtures.ExperimentTests()
        self.helper.setUp()
        self.addCleanup(self.helper.doCleanups)
        self.provenance = {
            "schema_version": "lufy-codex-execution-provenance/v1",
            "runtime": {"executable": "/fixture/codex", "sha256": "c" * 64,
                        "version": "codex-cli 0.163.0-alpha.10.2"},
            "sources": {path: "d" * 64 for path in EXPECTED_FILES},
        }
        for directory in [self.helper.baseline, self.helper.candidate]:
            self.helper.change_result(directory, execution_provenance_before=copy.deepcopy(self.provenance),
                                      execution_provenance_after=copy.deepcopy(self.provenance),
                                      argv=self.argv(directory))

    def argv(self, directory):
        return ["/fixture/codex", "exec", "--sandbox", "workspace-write", "--model", "fixture-model",
                "--disable", "apps", "-C", str(directory / "fixture"),
                "--output-schema", str(directory / "schema.json"), "-o", str(directory / "final.json")]

    def assert_rejected(self):
        comparison = self.helper.compare()
        self.assertIs(comparison["eligible"], False, comparison)
        self.assertTrue(comparison["reasons"])

    def test_stable_provenance_normalizes_only_three_temporary_path_values(self):
        self.assertNotEqual(self.argv(self.helper.baseline), self.argv(self.helper.candidate))
        result = self.helper.compare()
        self.assertIs(result["eligible"], True, result)
        self.assertEqual(result["baseline"]["input_tokens"], 100)
        self.assertEqual(result["candidate"]["input_tokens"], 70)

    def test_binary_or_source_hash_drift_within_case_or_across_pair_rejects(self):
        original = self.helper.read_result(self.helper.candidate)
        for component in ["runtime", "sources"]:
            for within_case in [True, False]:
                with self.subTest(component=component, within_case=within_case):
                    changed = copy.deepcopy(self.provenance)
                    if component == "runtime":
                        changed["runtime"]["sha256"] = "e" * 64
                    else:
                        changed["sources"][EXPECTED_FILES[0]] = "e" * 64
                    self.helper.change_result(self.helper.candidate, **original)
                    self.helper.change_result(self.helper.candidate, execution_provenance_after=changed,
                        execution_provenance_before=copy.deepcopy(self.provenance) if within_case else changed)
                    self.assert_rejected()

    def test_absent_incomplete_or_malformed_provenance_never_passes(self):
        original = self.helper.read_result(self.helper.candidate)
        bad = [None, {}, {**self.provenance, "schema_version": "unknown/v1"}]
        missing_source = copy.deepcopy(self.provenance)
        del missing_source["sources"][EXPECTED_FILES[-1]]
        bad.append(missing_source)
        bad_hash = copy.deepcopy(self.provenance)
        bad_hash["runtime"]["sha256"] = True
        bad.append(bad_hash)
        for invalid in bad:
            with self.subTest(provenance=invalid):
                self.helper.change_result(self.helper.candidate, **original)
                self.helper.change_result(self.helper.candidate, execution_provenance_before=invalid,
                                          execution_provenance_after=invalid)
                self.assert_rejected()
        for directory in [self.helper.baseline, self.helper.candidate]:
            record = self.helper.read_result(directory)
            record.pop("execution_provenance_before", None)
            record.pop("execution_provenance_after", None)
            (directory / "result.json").write_text(json.dumps(record), encoding="utf-8")
        self.assert_rejected()

    def test_argv_sandbox_model_and_other_isolation_flags_are_not_normalized_away(self):
        original = self.argv(self.helper.candidate)
        changed_sandbox = list(original)
        changed_sandbox[changed_sandbox.index("--sandbox") + 1] = "danger-full-access"
        changed_model = list(original)
        changed_model[changed_model.index("--model") + 1] = "other-fixture-model"
        for argv in [changed_sandbox, changed_model, original + ["--enable", "apps"],
                     original + ["-c", 'model_reasoning_effort="high"']]:
            with self.subTest(argv=argv):
                self.helper.change_result(self.helper.candidate, argv=argv)
                self.assert_rejected()

    def test_capture_provenance_hashes_exact_sources_and_canonical_binary_without_model(self):
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory).resolve()
            for path in EXPECTED_FILES:
                source = repo / path
                source.parent.mkdir(parents=True, exist_ok=True)
                source.write_text("PRIVATE_SOURCE_CANARY " + path, encoding="utf-8")
            executable = repo / "bin" / "synthetic-codex"
            executable.parent.mkdir()
            executable.write_bytes(b"PRIVATE_SYNTHETIC_BINARY_CANARY")
            executable.chmod(0o755)
            before = {path.relative_to(repo).as_posix(): path.read_bytes() for path in repo.rglob("*") if path.is_file()}
            process = subprocess.CompletedProcess([], 0, stdout="codex-cli 0.163.0-alpha.10.2\n", stderr="")
            with patch.object(experiment.subprocess, "run", return_value=process) as probe:
                result = experiment.capture_provenance(repo, str(executable))
            self.assertEqual(tuple(experiment.PROVENANCE_FILES), EXPECTED_FILES)
            self.assertEqual(result["schema_version"], "lufy-codex-execution-provenance/v1")
            self.assertEqual(result["runtime"]["executable"], str(executable.resolve()))
            self.assertEqual(result["runtime"]["sha256"], hashlib.sha256(executable.read_bytes()).hexdigest())
            self.assertTrue(result["runtime"]["version"].strip())
            self.assertEqual(result["sources"], {path: hashlib.sha256((repo / path).read_bytes()).hexdigest() for path in EXPECTED_FILES})
            probe.assert_called_once()
            self.assertEqual(probe.call_args.args[0], [str(executable.resolve()), "--version"])
            self.assertFalse(probe.call_args.kwargs.get("shell", False))
            self.assertGreater(probe.call_args.kwargs["timeout"], 0)
            self.assertNotIn("PRIVATE_SOURCE_CANARY", json.dumps(result))
            self.assertNotIn("PRIVATE_SYNTHETIC_BINARY_CANARY", json.dumps(result))
            self.assertEqual(before, {path.relative_to(repo).as_posix(): path.read_bytes() for path in repo.rglob("*") if path.is_file()})

    def test_replaced_binary_at_same_path_changes_observed_fingerprint(self):
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory).resolve()
            for path in EXPECTED_FILES:
                file = repo / path
                file.parent.mkdir(parents=True, exist_ok=True)
                file.write_text("synthetic source", encoding="utf-8")
            executable = repo / "synthetic-codex"
            executable.write_bytes(b"binary-one")
            executable.chmod(0o755)
            process = subprocess.CompletedProcess([], 0, stdout="codex-cli 0.163.0-alpha.10.2\n", stderr="")
            with patch.object(experiment.subprocess, "run", return_value=process) as probe:
                first = experiment.capture_provenance(repo, str(executable))
                executable.write_bytes(b"binary-two")
                second = experiment.capture_provenance(repo, str(executable))
            self.assertEqual(probe.call_count, 2)
            self.assertEqual(first["runtime"]["executable"], second["runtime"]["executable"])
            self.assertNotEqual(first["runtime"]["sha256"], second["runtime"]["sha256"])


if __name__ == "__main__":
    unittest.main()
