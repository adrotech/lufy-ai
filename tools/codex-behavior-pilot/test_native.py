"""Contrato offline del piloto nativo; no invoca modelos ni interpreta transcripts."""

import copy
import unittest

from native import assess_native


def tool(name, **evidence):
    return {"event": "PostToolUse", "tool": name, "success": True, **evidence}


def lifecycle():
    return [
        {"event": "SessionStart"},
        tool("spawn_agent", receivers=["child"]),
        {"event": "SubagentStart", "agent_type": "reviewer", "agent_id": "child"},
        tool("wait", target_ids=["child"], completed=[{"agent_id": "child", "payload": "empty"}]),
        tool("send_input", target_ids=["child"]),
        {"event": "SubagentStop", "agent_id": "child"},
        tool("wait", target_ids=["child"], completed=[
            {"agent_id": "child", "payload": "valid", "fingerprint": "b" * 64},
        ]),
        {"event": "SubagentStop", "agent_id": "child"},
        tool("close_agent", target_ids=["child"]),
        {"event": "SessionEnd"},
    ]


class NativeAssessmentTests(unittest.TestCase):
    def setUp(self):
        self.hooks = lifecycle()
        self.events = [{"type": "turn.completed"}]
        self.before = {"request.md": {"type": "file", "sha256": "a" * 64}}
        self.after = copy.deepcopy(self.before)

    def assess(self, **overrides):
        args = {"events": self.events, "hooks": self.hooks,
                "before": self.before, "after": self.after,
                "returncode": 0, "timed_out": False, "tool_router_errors": 0}
        args.update(overrides)
        return assess_native(**args)

    def assert_status(self, result, expected):
        self.assertEqual(result["status"], expected, result)

    def test_complete_native_recovery_passes_without_changing_evidence(self):
        hooks = copy.deepcopy(self.hooks)
        before = copy.deepcopy(self.before)
        self.assert_status(self.assess(), "passed")
        self.assertEqual(self.hooks, hooks)
        self.assertEqual(self.before, before)
        self.assertEqual(self.after, before)

    def test_each_required_lifecycle_or_tool_record_is_required(self):
        for index in range(len(self.hooks)):
            with self.subTest(missing=index):
                hooks = self.hooks[:index] + self.hooks[index + 1:]
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_wrong_role_or_child_identifiers_never_pass(self):
        changes = [
            (2, "agent_type", "implementer"),
            (1, "receivers", ["other-child"]),
            (2, "agent_id", "other-child"),
            (4, "target_ids", ["other-child"]),
            (5, "agent_id", "other-child"),
            (6, "target_ids", ["other-child"]),
            (7, "agent_id", "other-child"),
            (8, "target_ids", ["other-child"]),
        ]
        for index, field, value in changes:
            with self.subTest(index=index, field=field):
                hooks = copy.deepcopy(self.hooks)
                hooks[index][field] = value
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_initial_completion_and_recovered_completion_must_be_same_child(self):
        for index in [3, 6]:
            hooks = copy.deepcopy(self.hooks)
            hooks[index]["completed"][0]["agent_id"] = "other-child"
            with self.subTest(wait_index=index):
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_recovery_order_empty_wait_send_valid_wait_close_is_required(self):
        orders = [
            [0, 1, 2, 8, 3, 4, 5, 6, 7, 9],
            [0, 1, 2, 4, 3, 5, 6, 7, 8, 9],
            [0, 1, 2, 3, 6, 5, 4, 7, 8, 9],
        ]
        for order in orders:
            with self.subTest(order=order):
                self.assert_status(self.assess(hooks=[self.hooks[index] for index in order]), "inconclusive")

    def test_invalid_initial_or_recovery_payload_is_not_success(self):
        for index, payload in [(3, "valid"), (3, "invalid"), (6, "empty"), (6, "invalid")]:
            with self.subTest(index=index, payload=payload):
                hooks = copy.deepcopy(self.hooks)
                hooks[index]["completed"][0]["payload"] = payload
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_recovered_payload_requires_real_hex_fingerprint(self):
        for value in [None, "", "b" * 63, "z" * 64, True, []]:
            with self.subTest(fingerprint=value):
                hooks = copy.deepcopy(self.hooks)
                hooks[6]["completed"][0]["fingerprint"] = value
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_failed_tool_calls_do_not_establish_recovery_or_closure(self):
        for index in [1, 3, 4, 6, 8]:
            with self.subTest(index=index):
                hooks = copy.deepcopy(self.hooks)
                hooks[index]["success"] = False
                self.assert_status(self.assess(hooks=hooks), "inconclusive")

    def test_mutation_overrides_even_incomplete_execution(self):
        changed = {"request.md": {"type": "file", "sha256": "c" * 64}}
        self.assert_status(self.assess(after=changed, timed_out=True), "failed")
        for name in ["Bash", "apply_patch"]:
            with self.subTest(tool=name):
                hooks = [*self.hooks, tool(name)]
                self.assert_status(self.assess(hooks=hooks, returncode=1), "failed")

    def test_process_errors_timeouts_router_errors_and_incomplete_cli_are_inconclusive(self):
        for overrides in [
            {"returncode": 1}, {"timed_out": True}, {"tool_router_errors": 1},
            {"events": []}, {"events": [{"type": "turn.failed"}]},
            {"events": [*self.events, {"type": "error", "message": "synthetic unavailable"}]},
        ]:
            with self.subTest(overrides=overrides):
                self.assert_status(self.assess(**overrides), "inconclusive")

    def test_empty_or_missing_source_snapshots_cannot_prove_no_mutations(self):
        for value in [{}, None, [], "not-a-snapshot"]:
            with self.subTest(snapshot=value):
                self.assert_status(self.assess(before=value, after=value), "inconclusive")

    def test_non_object_inputs_fail_closed_without_crashing(self):
        for field in ["hooks", "events"]:
            for value in [None, {}, "not-a-sequence", 7]:
                with self.subTest(field=field, value=value):
                    self.assert_status(self.assess(**{field: value}), "inconclusive")
        for malformed in [None, [], "not-an-object", 7, {"event": []}]:
            with self.subTest(hook=malformed):
                self.assert_status(self.assess(hooks=[*self.hooks, malformed]), "inconclusive")
        for malformed in [None, [], "not-an-object", {"type": []}]:
            with self.subTest(event=malformed):
                self.assert_status(self.assess(events=[*self.events, malformed]), "inconclusive")

    def test_malformed_hook_evidence_and_unknown_actions_cannot_pass(self):
        malformed_values = [(1, "receivers", None), (3, "completed", None),
                            (3, "completed", [None]), (4, "target_ids", "child"),
                            (6, "completed", [{"agent_id": [], "payload": "valid"}])]
        for index, field, value in malformed_values:
            with self.subTest(index=index, field=field):
                hooks = copy.deepcopy(self.hooks)
                hooks[index][field] = value
                self.assert_status(self.assess(hooks=hooks), "inconclusive")
        for unknown in [{"event": "FutureHook"}, tool("future_tool")]:
            with self.subTest(unknown=unknown):
                self.assert_status(self.assess(hooks=[*self.hooks, unknown]), "inconclusive")

    def test_unrecognized_error_item_is_not_ignored(self):
        events = [*self.events, {"type": "item.completed", "item": {"type": "error", "message": "synthetic execution failed"}}]
        self.assert_status(self.assess(events=events), "inconclusive")

    def test_unknown_or_missing_item_type_cannot_hide_unobserved_tools(self):
        for item in [{"type": "web_search"}, {"type": "future_tool"}, {}]:
            with self.subTest(item=item):
                self.assert_status(self.assess(events=[*self.events, {"type": "item.completed", "item": item}]), "inconclusive")


if __name__ == "__main__":
    unittest.main()
