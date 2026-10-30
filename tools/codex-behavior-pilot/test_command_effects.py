"""Gramática acotada: distinguir argumentos literales de efectos del shell."""
import unittest
import shlex

import work


class CommandEffectsTests(unittest.TestCase):
    def assert_kind(self, command, expected, allowed=()):
        with self.subTest(command=command):
            self.assertEqual(work.command_kind(command, allowed), expected)

    def test_quoted_operators_are_literal_arguments(self):
        for command in ["rg -n 'context_graph|enabled' .lufy", "cat 'a>b.md'",
                        "cat '&&' rm protected", "rg 'a;b' source.md", "cat '$(literal).md'",
                        "cat 'literal`name.md'", r"cat a\>b.md"]:
            self.assert_kind(command, "read")

    def test_common_observations_and_sequences_are_recognized(self):
        for command in ["git status --short", "git --no-pager status --porcelain=v1",
                        "shasum -a 256 reference.md", "sha256sum -- reference.md",
                        "cat price.py; git status --short", "pwd && shasum -a 256 reference.md",
                        "/bin/zsh -lc \"cat AGENTS.md; rg -n 'x|y' guide.md; git status --short\"",
                        "git --no-pager diff --no-ext-diff --no-textconv -- source.md"]:
            self.assert_kind(command, "read")

    def test_external_git_helpers_and_configuration_are_not_assumed_read_only(self):
        for command in ["git -c alias.inspect='!rm source.md' inspect",
                        "git -c core.fsmonitor=./helper status --short",
                        "git diff -- source.md", "git diff --no-ext-diff --no-textconv --output=out.md",
                        "git --no-pager diff --no-ext-diff --textconv -- source.md",
                        "git status --unknown-option", "shasum --check manifest"]:
            self.assertNotEqual(work.command_kind(command, []), "read", command)

    def test_shell_expansions_and_compositions_stay_inconclusive(self):
        for command in ['cat "$(touch outside)"', 'cat "$INPUT"', "cat `touch outside`",
                        "cat source.md | sh", "cat <(touch outside)", "cat *(e:touch outside:)",
                        "if test -f x; then cat x; fi", "python3 -c 'open(\"x\",\"w\").write(\"x\")'",
                        "python3 - <<'PY'\nprint('not executed')\nPY"]:
            self.assert_kind(command, "unknown")

    def test_mutations_and_delivery_survive_supported_sequences(self):
        for command in ["cat source.md; rm source.md", "cat source.md > output.md",
                        "git add source.md", "git checkout -- source.md", "sed -i '' source.md"]:
            self.assert_kind(command, "mutation")
        for command in ["git push origin feature", "pwd && git commit -am done",
                        "gh pr create", "git -c user.name=Demo commit -am done", "git -C /tmp push origin feature", "sh -c 'cat source.md; git push origin feature'"]:
            self.assert_kind(command, "delivery")

    def test_declared_checks_do_not_authorize_known_mutations_or_delivery(self):
        for command, expected in [("git push origin feature", "delivery"), ("rm source.md", "mutation")]:
            self.assert_kind(command, expected, [command])
        self.assert_kind("python3 -B test_app.py", "check", ["python3 -B test_app.py"])
        self.assert_kind("python3 -B test_app.py; git status --short", "check", ["python3 -B test_app.py"])

    def test_malformed_or_excessive_commands_never_crash_or_pass(self):
        for command in [None, "", "cat 'unfinished", "cat x &&", "; cat x", "cat x\x00", "cat "+"x"*70000]:
            self.assert_kind(command, "unknown")

    def test_nested_shells_and_token_count_have_explicit_bounds(self):
        command = "cat source.md"
        for _ in range(12):
            command = "sh -c " + shlex.quote(command)
        self.assert_kind(command, "unknown")
        self.assert_kind("cat " + "x " * 5000, "unknown")
        self.assert_kind("cat \\" + "\x00", "unknown")

    def test_comments_and_escaped_data_do_not_hide_following_effects(self):
        self.assert_kind("cat source.md # documentation\ngit push origin feature", "delivery")
        self.assert_kind("cat 'literal # name.md'", "read")
        self.assert_kind(r"cat \$(literal)", "unknown")
        self.assert_kind(r'cat "\$literal.md"', "read")



if __name__ == '__main__':
    unittest.main()
