"""Tres controles offline del runner nativo con subprocess completamente simulado."""

import argparse
import contextlib
import hashlib
import io
import json
import os
import shlex
import tempfile
import unittest
from pathlib import Path
from unittest.mock import Mock, patch

import native
from test_native import lifecycle
from test_native_collaboration import callback


@unittest.skipUnless(os.name == "posix", "El runner native sólo soporta POSIX")
class NativeRunnerOfflineTests(unittest.TestCase):
    def fixture_run(self, wrong_nonce=False, hook_delivery="project", transport="legacy"):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        root = Path(temporary.name).resolve()
        repo = root / "repo"
        module = repo / "tools/codex-behavior-pilot/native.py"
        module.parent.mkdir(parents=True)
        module.write_text("# Ruta sintética del observador; no se ejecuta.\n", encoding="utf-8")
        source = repo / ".codex/agents/reviewer.toml"
        source.parent.mkdir(parents=True)
        source.write_text('name = "reviewer"\n', encoding="utf-8")
        (repo / ".codex/hooks.json").write_text(json.dumps({"hooks": {
            event: [{"hooks": [{"type": "command", "command": "synthetic-not-executed"}]}]
            for event in native.LIFECYCLE
        }}), encoding="utf-8")
        template = repo / "tools/lufy-cli-go/internal/resultcontract/testdata/valid-minimal.json"
        template.parent.mkdir(parents=True)
        template.write_text(json.dumps({"schema_version": "result-contract/v1"}), encoding="utf-8")
        before = {path.relative_to(repo).as_posix(): path.read_bytes() for path in repo.rglob("*") if path.is_file()}
        output = root / "evidence"
        args = argparse.Namespace(output=output, codex="fixture-codex", lufy="fixture-lufy",
                                  model="fixture-model", reasoning_effort="medium", timeout=10, hook_delivery=hook_delivery,
                                  transport=transport)
        ambient = {"HOME": "/fixture/home", "PATH": "/fixture/bin", "OPENAI_API_KEY": "synthetic-auth-canary",
                   "CODEX_PERMISSION_PROFILE": "synthetic-profile", "CODEX_SANDBOX": "synthetic-sandbox",
                   **{name: "synthetic-host-handle" for name in ["CODEX_APP_TOOLS_PIPE_PATH", "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_TOOL_CALL_ID"]}}

        def fake_popen(argv, **kwargs):
            process = Mock(returncode=0)

            def communicate(prompt, timeout):
                if transport == "collaboration":
                    self.assertIn("collaboration.list_agents", prompt)
                    self.assertIn("collaboration.followup_task", prompt)
                    self.assertIn("not_available", prompt)
                kwargs["stdout"].write(json.dumps({"type": "turn.completed"}) + "\n")
                kwargs["stdout"].flush()
                hooks = lifecycle()
                hooks.insert(-1, {"event": "Stop"})
                if transport == "collaboration":
                    target = "/root/fixture-reviewer"
                    expected = json.loads((output / "expected.json").read_text())
                    def inspection(payload):
                        return callback("list_agents", {"path_prefix": target}, {"agents": [
                            {"agent_name": target, "agent_status": {"completed": payload}},
                        ]}, expected)
                    hooks = [{"event": "SessionStart"},
                             callback("spawn_agent", {"agent_type": "reviewer"}, {"task_name": target}, expected),
                             inspection("null"),
                             callback("followup_task", {"target": target}, {"message": "queued"}, expected),
                             inspection(json.dumps(expected)), {"event": "Stop"}, {"event": "SessionEnd"}]
                observer = json.loads((output / "observer.json").read_text())
                for hook in hooks:
                    hook["run_nonce"] = "other_run" if wrong_nonce else observer["run_nonce"]
                    if hook["event"] in native.LIFECYCLE:
                        hook["lufy_exit_code"] = 0
                    for completed in hook.get("completed", []):
                        if completed.get("payload") == "valid":
                            completed["fingerprint"] = observer["expected_hash"]
                (output / "hooks.jsonl").write_text("".join(json.dumps(hook) + "\n" for hook in hooks), encoding="utf-8")

            process.communicate.side_effect = communicate
            return process

        def fake_run(argv, **kwargs):
            if argv[1] == "doctor":
                return Mock(returncode=0, stdout=b'{"codex_runtime":{"status":"available"}}')
            return Mock(returncode=0, stdout=b'{"status":"valid"}')

        with patch.object(native, "__file__", str(module)), \
                patch.object(native.shutil, "which", side_effect=lambda value: {"fixture-codex": "/fixture/bin/codex", "fixture-lufy": "/fixture/bin/lufy"}.get(value)), \
                patch.object(native.subprocess, "run", side_effect=fake_run) as validation, \
                patch.object(native.subprocess, "Popen", side_effect=fake_popen) as popen, \
                patch.dict(native.os.environ, ambient, clear=True), \
                contextlib.redirect_stdout(io.StringIO()):
            result = native.run_native(args)
            self.assertEqual(dict(native.os.environ), ambient)
        self.assertEqual(before, {path.relative_to(repo).as_posix(): path.read_bytes() for path in repo.rglob("*") if path.is_file()})
        popen.assert_called_once()
        self.assertEqual(validation.call_count, 2)
        self.assertEqual(validation.call_args_list[0].args[0][1], "doctor")
        return result, popen.call_args, validation.call_args_list[1], output, module, ambient

    def test_collaboration_registers_only_fixture_reviewer_without_loading_user_config(self):
        result, call, _, output, _, _ = self.fixture_run(transport="collaboration", hook_delivery="inline")
        argv = call.args[0]
        self.assertIn("--ignore-user-config", argv)
        overrides = [argv[i + 1] for i in range(len(argv) - 1) if argv[i] == "-c"]
        self.assertIn("agents.reviewer.config_file=" + json.dumps(str(output / "fixture/.codex/agents/reviewer.toml")), overrides)
        self.assertEqual(result["transport_profile"], "collaboration")
        self.assertEqual(result["cleanup_status"], "not_available")
        self.assertIs(result["native_execution_verified"], False)

    def test_collaboration_transport_requires_current_run_nonce_independently_of_cleanup(self):
        result = self.fixture_run(transport="collaboration")[0]
        self.assertIs(result["native_transport_verified"], True)
        self.assertIs(result["lufy_lifecycle_verified"], False)
        result = self.fixture_run(transport="collaboration", wrong_nonce=True)[0]
        self.assertIs(result["native_transport_verified"], False)
        self.assertEqual(result["reason"], "hook_run_mismatch")

    def test_read_only_native_invocation_preserves_hooks_without_ephemeral(self):
        _, call, validation, _, _, _ = self.fixture_run()
        argv = call.args[0]
        pairs = [argv[index:index + 2] for index in range(len(argv) - 1)]
        self.assertEqual(argv[:2], ["/fixture/bin/codex", "exec"])
        self.assertIn(["--sandbox", "read-only"], pairs)
        self.assertIn(["-c", 'approval_policy="never"'], pairs)
        self.assertIn(["-c", "agents.enabled=true"], pairs)
        self.assertIn(["--enable", "hooks"], pairs)
        self.assertIn(["--disable", "apps"], pairs)
        self.assertIn(["--disable", "plugins"], pairs)
        self.assertNotIn("--ephemeral", argv)
        self.assertNotIn("--dangerously-bypass-approvals-and-sandbox", argv)
        self.assertFalse(call.kwargs.get("shell", False))
        self.assertEqual(validation.args[0][:3], ["/fixture/bin/lufy", "result", "validate"])

    def test_hook_trust_and_observer_paths_stay_within_invocation_fixture(self):
        result, call, _, output, module, _ = self.fixture_run()
        self.assertIn("--dangerously-bypass-hook-trust", call.args[0])
        self.assertEqual(result["hook_trust"], "invocation_only_vetted_fixture")
        config_path = output / "observer.json"
        config = json.loads(config_path.read_text(encoding="utf-8"))
        self.assertEqual(Path(config["workspace"]), output / "fixture")
        self.assertEqual(Path(config["observations"]), output / "hooks.jsonl")
        hooks = json.loads((output / "fixture/.codex/hooks.json").read_text(encoding="utf-8"))["hooks"]
        for groups in hooks.values():
            command = groups[0]["hooks"][0]["command"]
            self.assertEqual(shlex.split(command)[1:], [str(module), "hook", "--config", str(config_path)])

    def test_child_env_and_public_metadata_preserve_privacy_and_uncertainty(self):
        result, call, _, output, _, ambient = self.fixture_run()
        child = call.kwargs["env"]
        handles = {"CODEX_APP_TOOLS_PIPE_PATH", "CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_TOOL_CALL_ID"}
        self.assertFalse(handles.intersection(child))
        for key in set(ambient) - handles:
            self.assertEqual(child[key], ambient[key])
        self.assertIs(result["runtime_identity_verified"], False)
        serialized = json.dumps(result)
        self.assertNotIn("synthetic-auth-canary", serialized)
        self.assertNotIn("synthetic-host-handle", serialized)
        self.assertEqual(result["event_sha256"], hashlib.sha256((output / "events.jsonl").read_bytes()).hexdigest())
        self.assertEqual(result["hooks_sha256"], hashlib.sha256((output / "hooks.jsonl").read_bytes()).hexdigest())

    def test_hooks_from_another_run_never_establish_success(self):
        self.assertEqual(self.fixture_run()[0]["status"], "passed")
        self.assertEqual(self.fixture_run(wrong_nonce=True)[0]["status"], "inconclusive")


    def test_inline_delivery_stays_invocation_scoped_without_project_hook_file(self):
        result, call, _, output, module, ambient = self.fixture_run(hook_delivery="inline")
        argv = call.args[0]
        inline = [argv[index + 1] for index in range(len(argv) - 1)
                  if argv[index] == "-c" and argv[index + 1].startswith("hooks=")]
        self.assertEqual(len(inline), 1)
        definition = json.loads((output / "fixture-hooks.json").read_text(encoding="utf-8"))
        self.assertEqual(inline[0], "hooks=" + native.toml_value(definition["hooks"]))
        self.assertFalse((output / "fixture/.codex/hooks.json").exists())
        self.assertEqual(result["hook_delivery"], "inline")
        self.assertEqual(result["hook_trust"], "invocation_only_vetted_fixture")
        for groups in definition["hooks"].values():
            command = groups[0]["hooks"][0]["command"]
            self.assertEqual(shlex.split(command)[1:], [str(module), "hook", "--config", str(output / "observer.json")])
        pairs = [argv[index:index + 2] for index in range(len(argv) - 1)]
        self.assertIn(["--sandbox", "read-only"], pairs)
        self.assertIn(["-c", 'approval_policy="never"'], pairs)
        self.assertNotIn("--ephemeral", argv)
        self.assertNotIn("--dangerously-bypass-approvals-and-sandbox", argv)
        self.assertFalse(call.kwargs.get("shell", False))
        self.assertEqual(call.kwargs["env"]["HOME"], ambient["HOME"])
        self.assertNotIn("synthetic-auth-canary", json.dumps(result))


if __name__ == "__main__":
    unittest.main()
