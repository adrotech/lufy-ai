"""Comparaciones pareadas offline; fixtures sintéticos, sin ejecutar modelos."""

import copy
import hashlib
import json
import tempfile
import unittest
from pathlib import Path

from experiment import PROVENANCE_FILES, compare_pair, summarize_pairs


class ExperimentTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        root = Path(temporary.name)
        self.baseline = root / "baseline" / "auto-chain"
        self.candidate = root / "candidate" / "auto-chain"
        self.write_fixture(self.baseline, "inline", "Fuente sintética explícita. " * 12, [70, 30], 4)
        self.write_fixture(self.candidate, "reference", "Lee runtime.py: fuente sintética.", [50, 20], 3)

    def write_fixture(self, directory, source_delivery, prompt, tokens, seconds):
        directory.mkdir(parents=True)
        snapshot = {"runtime.py": {"type": "file", "sha256": hashlib.sha256(b"synthetic").hexdigest()}}
        result = {
            "id": "auto-chain", "status": "passed", "reasons": [],
            "started_at": "2026-10-22T12:00:00+00:00",
            "completed_at": "2026-10-22T12:00:0" + str(seconds) + "+00:00",
            "returncode": 0, "timed_out": False, "infrastructure_error": None,
            "agent_execution_mode": "emulated_cli", "delegation_tools_enabled": None,
            "delegation_disable_requested": True, "delegation_disable_verified": False,
            "tool_router_errors": 0,
            "runtime_selection": {"model_requested": "fixture-model", "reasoning_effort_requested": "low", "identity_verified": False},
            "instruction_hashes": {".codex/agents/implementer.toml": "a" * 64, ".lufy/contracts/delivery.md": "b" * 64},
            "argv": ["/fixture/codex", "exec", "--disable", "multi_agent"],
            "source_delivery": source_delivery, "before": snapshot, "after": copy.deepcopy(snapshot),
            "final": {"status": "blocked", "tier": "T2", "fast_path_allowed": False,
                      "workload_decision": "Una unidad; falta aprobación post-plan.",
                      "rationale": "No se implementa sin aprobación explícita.",
                      "source_refs": [], "acceptance_ids": [], "examples": []},
        }
        provenance = {"schema_version": "lufy-codex-execution-provenance/v1",
                      "runtime": {"executable": "/fixture/codex", "sha256": "c" * 64, "version": "fixture-version"},
                      "sources": {path: "d" * 64 for path in PROVENANCE_FILES}}
        result.update(execution_provenance_before=provenance, execution_provenance_after=copy.deepcopy(provenance))
        events = [{"type": "turn.started"}]
        events.extend({"type": "turn.completed", "usage": {"input_tokens": value}} for value in tokens)
        self.write_events(directory, result, events)
        (directory / "prompt.txt").write_text(prompt, encoding="utf-8")
        (directory / "stderr.log").write_text("", encoding="utf-8")

    def write_events(self, directory, result, events):
        body = "".join(json.dumps(event) + "\n" for event in events).encode("utf-8")
        (directory / "events.jsonl").write_bytes(body)
        result["event_sha256"] = hashlib.sha256(body).hexdigest()
        (directory / "result.json").write_text(json.dumps(result), encoding="utf-8")

    def read_result(self, directory):
        return json.loads((directory / "result.json").read_text(encoding="utf-8"))

    def change_result(self, directory, **fields):
        result = self.read_result(directory)
        result.update(fields)
        (directory / "result.json").write_text(json.dumps(result), encoding="utf-8")

    def compare(self):
        return compare_pair(self.baseline, self.candidate)

    def assert_ineligible(self, comparison):
        self.assertIs(comparison["eligible"], False, comparison)
        self.assertIsInstance(comparison["reasons"], list)
        self.assertTrue(comparison["reasons"])

    def test_valid_pair_uses_observed_metrics_and_preserves_evidence(self):
        before = {str(path): path.read_bytes() for directory in [self.baseline, self.candidate]
                  for path in directory.iterdir()}
        comparison = self.compare()
        self.assertIs(comparison["eligible"], True, comparison)
        self.assertEqual(comparison["baseline"]["input_tokens"], 100)
        self.assertEqual(comparison["candidate"]["input_tokens"], 70)
        self.assertEqual(comparison["baseline"]["duration_ms"], 4000)
        self.assertEqual(comparison["candidate"]["duration_ms"], 3000)
        for label, directory in [("baseline", self.baseline), ("candidate", self.candidate)]:
            self.assertEqual(comparison[label]["prompt_bytes"], len((directory / "prompt.txt").read_bytes()))
        self.assertEqual(before, {str(path): path.read_bytes() for directory in [self.baseline, self.candidate]
                                 for path in directory.iterdir()})

    def test_missing_usage_is_not_invented_as_zero_token_saving(self):
        result = self.read_result(self.candidate)
        result["input_tokens"] = 0
        self.write_events(self.candidate, result, [{"type": "turn.completed"}])
        comparison = self.compare()
        self.assert_ineligible(comparison)
        self.assertIsNone(comparison["candidate"]["input_tokens"])

    def test_invalid_token_usage_is_unavailable_even_if_other_turn_has_usage(self):
        original = self.read_result(self.candidate)
        for invalid in [None, True, -1, "10", 1.5]:
            with self.subTest(tokens=invalid):
                self.write_events(self.candidate, copy.deepcopy(original), [
                    {"type": "turn.completed", "usage": {"input_tokens": 10}},
                    {"type": "turn.completed", "usage": {"input_tokens": invalid}},
                ])
                comparison = self.compare()
                self.assert_ineligible(comparison)
                self.assertIsNone(comparison["candidate"]["input_tokens"])

    def test_hash_mismatch_rejects_otherwise_successful_claim(self):
        with (self.candidate / "events.jsonl").open("a", encoding="utf-8") as file:
            file.write(json.dumps({"type": "turn.completed", "usage": {"input_tokens": 1}}) + "\n")
        self.assert_ineligible(self.compare())

    def test_runtime_instruction_source_and_executable_must_match_protocol(self):
        original = self.read_result(self.candidate)
        changes = [
            {"source_delivery": "inline"},
            {"instruction_hashes": {".codex/agents/implementer.toml": "c" * 64}},
            {"runtime_selection": {"model_requested": "other-fixture", "reasoning_effort_requested": "low", "identity_verified": False}},
            {"argv": ["/fixture/other-codex", "exec"]},
            {"id": "analysis"},
        ]
        for fields in changes:
            with self.subTest(fields=fields):
                self.change_result(self.candidate, **{**original, **fields})
                self.assert_ineligible(self.compare())
        self.change_result(self.candidate, **original)
        self.change_result(self.baseline, source_delivery="reference")
        self.assert_ineligible(self.compare())

    def test_snapshot_mismatch_or_observed_mutation_rejects_pair(self):
        original = self.read_result(self.candidate)
        altered = {"runtime.py": {"type": "file", "sha256": "d" * 64}}
        self.change_result(self.candidate, before=altered, after=altered)
        self.assert_ineligible(self.compare())
        self.change_result(self.candidate, **original)
        self.change_result(self.candidate, after=altered)
        self.assert_ineligible(self.compare())

    def test_trace_mutation_is_reassessed_despite_claimed_pass_and_unchanged_snapshot(self):
        result = self.read_result(self.candidate)
        self.write_events(self.candidate, result, [
            {"type": "item.completed", "item": {"type": "file_change", "status": "failed", "changes": [{"path": "runtime.py", "kind": "update"}]}},
            {"type": "turn.completed", "usage": {"input_tokens": 70}},
        ])
        self.assert_ineligible(self.compare())

    def test_stderr_router_error_is_reassessed_instead_of_trusting_cached_count(self):
        (self.candidate / "stderr.log").write_text(
            "ERROR codex_core::tools::router: error=collab spawn failed\n", encoding="utf-8")
        self.assert_ineligible(self.compare())

    def test_auto_chain_requires_blocked_without_fast_path_even_when_cached_status_passed(self):
        original = self.read_result(self.candidate)
        for fields in [{"status": "ready"}, {"fast_path_allowed": True}, {"tier": "T3"}]:
            final = {**original["final"], **fields}
            with self.subTest(final=final):
                self.change_result(self.candidate, final=final)
                self.assert_ineligible(self.compare())

    def test_incomplete_process_does_not_become_eligible_from_cached_pass(self):
        original = self.read_result(self.candidate)
        for fields in [{"returncode": 1}, {"timed_out": True}]:
            with self.subTest(fields=fields):
                self.change_result(self.candidate, **{**original, **fields})
                self.assert_ineligible(self.compare())

    def test_duration_missing_invalid_or_inverted_is_unavailable(self):
        original = self.read_result(self.candidate)
        for fields in [{"completed_at": None}, {"started_at": "not-a-date"},
                       {"completed_at": "2026-10-22T11:59:59+00:00"}]:
            with self.subTest(fields=fields):
                self.change_result(self.candidate, **{**original, **fields})
                comparison = self.compare()
                self.assert_ineligible(comparison)
                self.assertIsNone(comparison["candidate"]["duration_ms"])

    def test_eligibility_requires_strict_prompt_and_token_improvement_and_no_slower_duration(self):
        original = self.read_result(self.candidate)
        original_prompt = (self.candidate / "prompt.txt").read_bytes()
        (self.candidate / "prompt.txt").write_bytes((self.baseline / "prompt.txt").read_bytes())
        self.assert_ineligible(self.compare())
        (self.candidate / "prompt.txt").write_bytes(original_prompt)
        self.write_events(self.candidate, original, [{"type": "turn.completed", "usage": {"input_tokens": 100}}])
        self.assert_ineligible(self.compare())
        self.write_events(self.candidate, original, [{"type": "turn.completed", "usage": {"input_tokens": 70}}])
        self.change_result(self.candidate, completed_at="2026-10-22T12:00:05+00:00")
        self.assert_ineligible(self.compare())

    def test_zero_or_single_pair_keeps_baseline_without_automatic_adoption(self):
        for pairs in [[], [self.compare()]]:
            with self.subTest(pair_count=len(pairs)):
                summary = summarize_pairs(pairs)
                self.assertEqual(summary["decision"], "keep_baseline", summary)
                self.assertIs(summary["adopted"], False)
                self.assertEqual(summary["pair_count"], len(pairs))

    def test_two_eligible_pairs_only_support_sample_never_adopt_automatically(self):
        comparison = self.compare()
        summary = summarize_pairs([comparison, copy.deepcopy(comparison)])
        self.assertEqual(summary["decision"], "candidate_supported_in_sample", summary)
        self.assertIs(summary["adopted"], False)
        self.assertEqual(summary["pair_count"], 2)

    def test_one_bad_pair_keeps_baseline_even_with_two_other_eligible_pairs(self):
        good = self.compare()
        bad = {**good, "eligible": False, "reasons": ["Evidencia incompleta"]}
        summary = summarize_pairs([good, good, bad])
        self.assertEqual(summary["decision"], "keep_baseline", summary)
        self.assertIs(summary["adopted"], False)
        self.assertEqual(summary["pair_count"], 3)

    def test_adversarial_non_object_result_is_rejected_without_crash(self):
        for malformed in [None, [], "not-a-record", 7]:
            with self.subTest(record=malformed):
                (self.candidate / "result.json").write_text(json.dumps(malformed), encoding="utf-8")
                self.assert_ineligible(self.compare())

    def test_adversarial_malformed_or_non_object_jsonl_cannot_pass(self):
        original = self.read_result(self.candidate)
        for malformed in [None, [], {"type": []}, {"type": "item.completed", "item": {"type": []}}]:
            with self.subTest(event=malformed):
                self.write_events(self.candidate, copy.deepcopy(original), [
                    malformed, {"type": "turn.completed", "usage": {"input_tokens": 70}},
                ])
                self.assert_ineligible(self.compare())
        (self.candidate / "events.jsonl").write_text("{invalid JSON\n", encoding="utf-8")
        self.assert_ineligible(self.compare())

    def test_adversarial_matching_empty_provenance_is_not_verified_provenance(self):
        originals = {directory: self.read_result(directory) for directory in [self.baseline, self.candidate]}
        for field, missing in [("runtime_selection", {}), ("instruction_hashes", {}), ("argv", [])]:
            with self.subTest(field=field):
                for directory, original in originals.items():
                    self.change_result(directory, **{**original, field: missing})
                self.assert_ineligible(self.compare())

    def test_adversarial_matching_invalid_provenance_is_not_a_valid_runtime(self):
        originals = {directory: self.read_result(directory) for directory in [self.baseline, self.candidate]}
        changes = [
            {"runtime_selection": {"model_requested": "fixture-model"}},
            {"instruction_hashes": "not-a-hash-map"},
            {"instruction_hashes": {".codex/agents/implementer.toml": None}},
            {"argv": [None]},
            {"argv": [""]},
        ]
        for fields in changes:
            with self.subTest(fields=fields):
                for directory, original in originals.items():
                    self.change_result(directory, **{**original, **fields})
                self.assert_ineligible(self.compare())


if __name__ == "__main__":
    unittest.main()
