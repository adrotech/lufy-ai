"""Evidencia estructurada del perfil collaboration, sin modelos ni red."""

import copy
import json
import unittest

import native


def callback(name, inputs, response, expected):
    return native.normalize_hook({"hook_event_name": "PostToolUse", "tool_name": "collaboration." + name,
                                  "tool_input": inputs, "tool_response": response}, native.digest(expected))


class CollaborationProfileTests(unittest.TestCase):
    def setUp(self):
        self.target = "/root/synthetic-reviewer"
        self.payload = {"schema_version": "result-contract/v1", "synthetic": True}
        self.spawn = callback("spawn_agent", {"agent_type": "reviewer", "task_name": "synthetic-reviewer"},
                              {"task_name": self.target}, self.payload)
        self.empty = self.inspection("null")
        self.followup = callback("followup_task", {"target": self.target, "message": "private-canary"},
                                 {"message": "queued"}, self.payload)
        self.valid = self.inspection(json.dumps(self.payload))
        self.hooks = [self.spawn, self.empty, self.followup, self.valid]
        self.snapshot = {"fixture": {"sha256": "a" * 64}}

    def inspection(self, completed):
        return callback("list_agents", {"path_prefix": self.target}, {"agents": [
            {"agent_name": self.target, "agent_status": {"completed": completed}},
        ]}, self.payload)

    def assess(self, **kw):
        args = dict(events=[{"type": "turn.completed"}], hooks=self.hooks,
                    before=self.snapshot, after=copy.deepcopy(self.snapshot))
        args.update(kw)
        return native.assess_collaboration(**args)

    def test_normalizer_preserves_identity_and_payload_classification_without_bodies(self):
        self.assertEqual(self.spawn["receivers"], [native.identifier(self.target)])
        self.assertEqual(self.spawn["role_requested"], "reviewer")
        self.assertEqual(self.followup["target_ids"], [native.identifier(self.target)])
        self.assertEqual(self.empty["completed"][0]["payload"], "empty")
        self.assertEqual(self.valid["completed"][0]["fingerprint"], native.digest(self.payload))
        self.assertNotIn("private-canary", json.dumps(self.hooks))
        self.assertNotIn(self.target, json.dumps(self.hooks))

    def test_transport_passes_but_missing_close_never_passes_integral_case(self):
        result = self.assess()
        self.assertEqual(result["transport_status"], "passed", result)
        self.assertEqual(result["status"], "inconclusive")
        self.assertEqual(result["cleanup_status"], "not_available")
        self.assertEqual(result["role_status"], "requested")

    def test_exact_child_callback_is_required_to_observe_role(self):
        start = {"event": "SubagentStart", "agent_type": "reviewer", "agent_id": native.identifier(self.target)}
        self.assertEqual(self.assess(hooks=[start, *self.hooks])["role_status"], "observed")
        start["agent_id"] = native.identifier("unmapped-uuid")
        self.assertEqual(self.assess(hooks=[start, *self.hooks])["role_status"], "requested")

    def test_missing_reordered_wrong_target_and_extra_recovery_fail_closed(self):
        cases = [self.hooks[:2] + self.hooks[3:], self.hooks[1:],
                 [self.spawn, self.followup, self.empty, self.valid],
                 [self.spawn, self.empty, self.followup, self.followup, self.valid],
                 [self.spawn, self.empty, self.followup, self.empty]]
        changed = copy.deepcopy(self.hooks)
        changed[2]["target_ids"] = [native.identifier("other-child")]
        cases.append(changed)
        for hooks in cases:
            with self.subTest(hooks=hooks):
                self.assertEqual(self.assess(hooks=hooks)["transport_status"], "inconclusive")

    def test_wrong_role_failed_calls_and_fingerprint_are_rejected(self):
        for index, field, value in [(0, "role_requested", "default"), (0, "success", False),
                                    (2, "success", False), (3, "success", False)]:
            hooks = copy.deepcopy(self.hooks); hooks[index][field] = value
            with self.subTest(index=index, field=field):
                self.assertEqual(self.assess(hooks=hooks)["transport_status"], "inconclusive")
        hooks = copy.deepcopy(self.hooks)
        hooks[3]["completed"][0]["fingerprint"] = "invalid"
        self.assertEqual(self.assess(hooks=hooks)["transport_status"], "inconclusive")

    def test_other_agents_or_invalid_completed_payload_are_not_accepted(self):
        hooks = copy.deepcopy(self.hooks)
        hooks[3]["completed"].append({"agent_id": native.identifier("other"), "payload": "valid"})
        self.assertEqual(self.assess(hooks=hooks)["transport_status"], "inconclusive")
        hooks[3] = self.inspection("invalid-private-canary")
        self.assertNotIn("invalid-private-canary", json.dumps(hooks))
        self.assertEqual(self.assess(hooks=hooks)["transport_status"], "inconclusive")

    def test_contradictory_role_and_malformed_inspection_never_pass(self):
        start = {"event": "SubagentStart", "agent_type": "implementer", "agent_id": native.identifier(self.target)}
        self.assertEqual(self.assess(hooks=[start, *self.hooks])["transport_status"], "inconclusive")
        invalid = callback("list_agents", {"path_prefix": self.target}, {"agents": [None, {
            "agent_name": self.target, "agent_status": {"completed": json.dumps(self.payload)},
        }]}, self.payload)
        self.assertEqual(self.assess(hooks=[*self.hooks[:-1], invalid])["transport_status"], "inconclusive")

    def test_other_running_agent_is_not_hidden_by_normalizer(self):
        inspection = callback("list_agents", {"path_prefix": self.target}, {"agents": [
            {"agent_name": self.target, "agent_status": {"completed": json.dumps(self.payload)}},
            {"agent_name": "other-agent", "agent_status": "running"},
        ]}, self.payload)
        self.assertEqual(self.assess(hooks=[*self.hooks[:-1], inspection])["transport_status"], "inconclusive")

    def test_unknown_trace_and_observed_violations_cannot_be_hidden_by_transport(self):
        self.assertEqual(self.assess(events=[{"type": "new_unknown"}])["transport_status"], "inconclusive")
        self.assertEqual(self.assess(events=[{"type": "item.completed", "item": {"type": "command_execution"}}])["status"], "failed")
        self.assertEqual(self.assess(after={"fixture": "changed"})["status"], "failed")
        self.assertEqual(self.assess(hooks=[*self.hooks, {"event": "PreToolUse", "tool": "exec_command"}])["status"], "failed")

    def test_observed_concatenated_hook_names_are_exact_aliases_only(self):
        for name in native.COLLABORATION_TOOLS:
            record = native.normalize_hook({"hook_event_name": "PostToolUse", "tool_name": "collaboration" + name,
                                            "tool_input": {}, "tool_response": {}}, native.digest(self.payload))
            self.assertEqual(record["tool"], name)
        unknown = native.normalize_hook({"hook_event_name": "PostToolUse", "tool_name": "evilcollaborationspawn_agent"}, native.digest(self.payload))
        self.assertNotEqual(unknown["tool"], "spawn_agent")
        shell = native.normalize_hook({"hook_event_name": "PreToolUse", "tool_name": "functionsexec_command"}, native.digest(self.payload))
        self.assertEqual(self.assess(hooks=[*self.hooks, shell])["status"], "failed")

    def test_empty_followup_ack_requires_later_structured_state_and_never_masks_error(self):
        for response in [None, "", {}]:
            followup = callback("followup_task", {"target": self.target}, response, self.payload)
            self.assertEqual(self.assess(hooks=[self.spawn, self.empty, followup, self.valid])["transport_status"], "passed")
            self.assertEqual(self.assess(hooks=[self.spawn, self.empty, followup])["transport_status"], "inconclusive")
        for response in [{"error": "failed"}, {"isError": True}, {"status": "failed"}]:
            followup = callback("followup_task", {"target": self.target}, response, self.payload)
            self.assertEqual(self.assess(hooks=[self.spawn, self.empty, followup, self.valid])["transport_status"], "inconclusive")


if __name__ == "__main__":
    unittest.main()
