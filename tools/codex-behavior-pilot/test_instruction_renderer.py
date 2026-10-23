"""Verifica distribución de texto; no mide conducta de modelos."""
import importlib.util
import fnmatch
from pathlib import Path
import re
import tempfile
import unittest

REPO = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("instruction_renderer", REPO / "scripts/render-codex-instructions.py")
renderer = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(renderer)

BEGIN = "<!-- lufy:proportional:start -->"
END = "<!-- lufy:proportional:end -->"
MIRROR = Path("tools/lufy-cli-go/internal/assets/embedded")


class InstructionRendererTests(unittest.TestCase):
    def test_repository_has_no_instruction_drift(self):
        self.assertFalse(renderer.render(REPO, write=False))

    def test_ci_runs_for_instruction_source_and_renderer_changes(self):
        workflow = (REPO / ".github/workflows/go-cli-install.yml").read_text(encoding="utf-8")
        for event in ("pull_request", "push"):
            section = re.search(r"(?ms)^  " + event + r":\n(.*?)(?=^  [a-z_]+:|^permissions:)", workflow).group(1)
            patterns = re.findall(r'^ {6}- "([^"]+)"$', section, re.M)
            for path in (".codex/agents/reviewer.toml", ".codex/instructions/proportional-workflow.md", "scripts/render-codex-instructions.py"):
                self.assertTrue(any(fnmatch.fnmatchcase(path, pattern) for pattern in patterns), (event, path))

    def fixture(self, root):
        source = Path(".codex/instructions/proportional-workflow.md")
        for prefix in (Path(), MIRROR):
            p = root / prefix / source
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_text("## Decisiones proporcionales\nRegla compartida.\n", encoding="utf-8")
            for name in ("orchestrator", "sdd-router", "implementer"):
                p = root / prefix / ".codex/agents" / (name + ".toml")
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_text('developer_instructions = """\nInstrucción particular.\n' + BEGIN + '\nViejo.\n' + END + '\nConservar cierre.\n"""\nsandbox_mode = "read-only"\n', encoding="utf-8")
        return source

    def test_check_is_read_only_and_write_preserves_role_and_permissions(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            role = root / ".codex/agents/implementer.toml"
            before = role.read_bytes()
            self.assertTrue(renderer.render(root, write=False))
            self.assertEqual(before, role.read_bytes())
            self.assertTrue(renderer.render(root, write=True))
            text = role.read_text(encoding="utf-8")
            self.assertIn("Instrucción particular.", text)
            self.assertIn("Conservar cierre.", text)
            self.assertIn('sandbox_mode = "read-only"', text)
            self.assertIn("Regla compartida.", text)
            self.assertFalse(renderer.render(root, write=False))
            self.assertFalse(renderer.render(root, write=True))

    def test_source_and_mirror_drift_are_detected_and_repaired(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = self.fixture(root)
            renderer.render(root, write=True)
            (root / source).write_text("## Decisiones proporcionales\nNueva regla.\n", encoding="utf-8")
            drift = renderer.render(root, write=False)
            self.assertEqual(7, len(drift))
            renderer.render(root, write=True)
            self.assertEqual((root / source).read_bytes(), (root / MIRROR / source).read_bytes())
            self.assertFalse(renderer.render(root, write=False))

    def test_invalid_markers_reject_before_any_write(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            role = root / ".codex/agents/implementer.toml"
            role.write_text(role.read_text(encoding="utf-8") + BEGIN, encoding="utf-8")
            before = {p: p.read_bytes() for p in root.rglob("*") if p.is_file()}
            with self.assertRaises(ValueError):
                renderer.render(root, write=True)
            self.assertEqual(before, {p: p.read_bytes() for p in before})

    def test_render_preserves_crlf_outside_shared_segment(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            role = root / ".codex/agents/implementer.toml"
            role.write_bytes(role.read_bytes().replace(b"\n", b"\r\n"))
            before = role.read_bytes()
            renderer.render(root, write=True)
            after = role.read_bytes()
            self.assertEqual(before.split(BEGIN.encode())[0], after.split(BEGIN.encode())[0])
            self.assertEqual(before.split(END.encode())[1], after.split(END.encode())[1])
            self.assertNotIn(b"\r\r\n", after)
            self.assertFalse(renderer.render(root, write=False))

    def test_symlink_rejected_without_following_target(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            role = root / ".codex/agents/implementer.toml"
            outside = root / "outside.toml"
            outside.write_bytes(role.read_bytes())
            original = outside.read_bytes()
            role.unlink()
            try:
                role.symlink_to(outside)
            except OSError:
                self.skipTest("symlink no disponible")
            with self.assertRaises(ValueError):
                renderer.render(root, write=True)
            self.assertEqual(original, outside.read_bytes())


if __name__ == "__main__":
    unittest.main()
