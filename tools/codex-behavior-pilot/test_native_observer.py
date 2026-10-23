"""Triangulación offline de normalización privada y precedencia de acciones."""

import copy
import hashlib
import json
import unittest

import native
from test_native import lifecycle


class NativeObserverTests(unittest.TestCase):
    def setUp(self):
        self.before = {"request.md": {"type": "file", "sha256": "a" * 64}}
        self.payload = {"schema_version": "result-contract/v1", "status": "validated",
                        "executive_summary": "SYNTHETIC_SECRET_CANARY"}
        self.expected = native.digest(self.payload)
        self.child = "synthetic-private-agent-id"

    def wait_hook(self, response=None):
        if response is None:
            response = {"status": {self.child: {"completed": json.dumps(self.payload)}}}
        return {"hook_event_name": "PostToolUse", "tool_name": "collab.wait",
                "tool_input": {"ids": [self.child]}, "tool_response": response}

    def assess(self, events, **extra):
        return native.assess_native(events, lifecycle(), self.before, copy.deepcopy(self.before), **extra)

    def test_non_string_hook_event_is_invalid_without_crashing(self):
        for value in [[], {}, None, 7, True]:
            with self.subTest(event=value):
                result = native.normalize_hook({"hook_event_name": value}, self.expected)
                self.assertEqual(result["event"], "invalid", result)

    def test_malformed_cli_item_type_is_inconclusive_without_crashing(self):
        events = [{"type": "item.completed", "item": {"type": []}}, {"type": "turn.completed"}]
        self.assertEqual(self.assess(events)["status"], "inconclusive")

    def test_observed_cli_command_keeps_failure_precedence_over_runtime_and_bad_events(self):
        command = {"type": "item.completed", "item": {"type": "command_execution", "command": "cat request.md"}}
        cases = [
            ([command, {"type": "turn.completed"}], {"returncode": 1}),
            ([command, {"type": "turn.completed"}], {"timed_out": True}),
            ([command, {"type": "turn.completed"}], {"tool_router_errors": 1}),
            ([{"type": []}, command, {"type": "turn.completed"}], {}),
            ([None, command, {"type": "turn.completed"}], {"returncode": 1}),
        ]
        for events, extra in cases:
            with self.subTest(events=events, extra=extra):
                self.assertEqual(self.assess(events, **extra)["status"], "failed")

    def test_matching_result_contract_hash_preserves_only_fingerprint_not_body(self):
        result = native.normalize_hook(self.wait_hook(), self.expected)
        child_hash = hashlib.sha256(self.child.encode()).hexdigest()
        self.assertEqual(result["completed"], [{"agent_id": child_hash, "payload": "valid", "fingerprint": self.expected}])
        self.assertEqual(result["target_ids"], [child_hash])
        serialized = json.dumps(result)
        self.assertNotIn("SYNTHETIC_SECRET_CANARY", serialized)
        self.assertNotIn("executive_summary", serialized)
        self.assertNotIn(self.child, serialized)

    def test_nonmatching_hash_is_invalid_without_a_verified_fingerprint(self):
        result = native.normalize_hook(self.wait_hook(), native.digest({"different": True}))
        self.assertEqual(result["completed"][0]["payload"], "invalid")
        self.assertNotIn("fingerprint", result["completed"][0])
        self.assertNotIn("SYNTHETIC_SECRET_CANARY", json.dumps(result))

    def test_single_text_content_wrapper_decodes_same_as_direct_response(self):
        response = self.wait_hook()["tool_response"]
        wrapped = {"content": [{"type": "text", "text": json.dumps(response)}]}
        direct = native.normalize_hook(self.wait_hook(), self.expected)
        self.assertEqual(native.normalize_hook(self.wait_hook(wrapped), self.expected), direct)

    def test_spawn_and_lifecycle_agent_ids_are_pseudonymized_consistently(self):
        child_hash = hashlib.sha256(self.child.encode()).hexdigest()
        spawned = native.normalize_hook({
            "hook_event_name": "PostToolUse", "tool_name": "collab.spawn_agent",
            "tool_input": {"agent_type": "reviewer"}, "tool_response": {"agent_id": self.child},
        }, self.expected)
        started = native.normalize_hook({"hook_event_name": "SubagentStart", "agent_id": self.child,
                                         "agent_type": "reviewer"}, self.expected)
        self.assertEqual(spawned["receivers"], [child_hash])
        self.assertEqual(started["agent_id"], child_hash)
        self.assertEqual(started["agent_type"], "reviewer")
        self.assertNotIn(self.child, json.dumps([spawned, started]))

    def test_arbitrary_non_object_hook_inputs_return_invalid(self):
        for data in [None, [], 7, "not-json", True]:
            with self.subTest(data=data):
                self.assertEqual(native.normalize_hook(data, self.expected)["event"], "invalid")

    def test_malformed_tool_evidence_cannot_become_a_valid_payload(self):
        for value in [None, [], 7, "not-json", {"content": [None]}]:
            with self.subTest(response=value):
                hook = self.wait_hook()
                hook["tool_response"] = value
                result = native.normalize_hook(hook, self.expected)
                self.assertFalse(any(item.get("payload") == "valid" for item in result.get("completed", [])))
                self.assertNotIn("SYNTHETIC_SECRET_CANARY", json.dumps(result))


    def test_error_flags_and_error_status_never_normalize_as_success(self):
        for error in [{"is_error": True}, {"timed_out": True},
                      {"status": "error"}, {"status": "failed"}, {"status": "not_found"}]:
            with self.subTest(error=error):
                response = {**self.wait_hook()["tool_response"], **error}
                normalized = native.normalize_hook(self.wait_hook(response), self.expected)
                self.assertIs(normalized["success"], False, normalized)
                self.assertNotIn("SYNTHETIC_SECRET_CANARY", json.dumps(normalized))


if __name__ == "__main__":
    unittest.main()
