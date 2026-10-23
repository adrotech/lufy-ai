#!/usr/bin/env python3
"""Comprueba o materializa segmentos Codex; no modifica políticas ni permisos."""
import argparse
from pathlib import Path
import sys

BEGIN = "<!-- lufy:proportional:start -->"
END = "<!-- lufy:proportional:end -->"
SOURCE = Path(".codex/instructions/proportional-workflow.md")
MIRROR = Path("tools/lufy-cli-go/internal/assets/embedded")
ROLES = ("orchestrator", "sdd-router", "implementer")


def read_regular(root, relative):
    path = root / relative
    for part in (path,) + tuple(path.parents):
        if part == root.parent:
            break
        if part.is_symlink():
            raise ValueError("ruta simbólica no admitida: " + str(relative))
    if not path.is_file() or path.stat().st_size > 1024 * 1024:
        raise ValueError("archivo ausente, no regular o excedido: " + str(relative))
    with path.open("r", encoding="utf-8", newline="") as stream:
        return stream.read()


def render(root, write=False):
    root = Path(root).absolute()
    source = read_regular(root, SOURCE)
    if not source.strip() or any(token in source for token in (BEGIN, END, '"""', "\\")):
        raise ValueError("fuente compartida vacía o incompatible con TOML inline")
    desired = {MIRROR / SOURCE: source}
    for role in ROLES:
        relative = Path(".codex/agents") / (role + ".toml")
        body = read_regular(root, relative)
        if body.count(BEGIN) != 1 or body.count(END) != 1:
            raise ValueError("markers ausentes o ambiguos: " + str(relative))
        start, end = body.index(BEGIN), body.index(END)
        opener = 'developer_instructions = """'
        if body.count(opener) != 1:
            raise ValueError("developer_instructions ambiguo")
        field_start = body.index(opener) + len(opener)
        field_end = body.find('"""', field_start)
        if not field_start <= start < end < field_end:
            raise ValueError("markers fuera de developer_instructions")
        newline = "\r\n" if "\r\n" in body else "\n"
        inline = source.replace("\r\n", "\n").rstrip("\n").replace("\n", newline)
        updated = body[:start] + BEGIN + newline + inline + newline + body[end:]
        desired[relative] = updated
        desired[MIRROR / relative] = updated
    # Validar todos los archivos antes de la primera escritura.
    drift = [path for path, body in desired.items() if read_regular(root, path) != body]
    if write:
        for path in drift:
            with (root / path).open("w", encoding="utf-8", newline="") as stream:
                stream.write(desired[path])
    return drift


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument("--check", action="store_true", help="sólo comprobar (default)")
    mode.add_argument("--write", action="store_true", help="materializar segmentos y mirrors")
    args = parser.parse_args()
    try:
        drift = render(args.root, write=args.write)
    except (OSError, UnicodeError, ValueError) as error:
        print("Codex instructions: " + str(error), file=sys.stderr)
        return 1
    for path in drift:
        print(("rendered: " if args.write else "drift: ") + str(path))
    return int(bool(drift) and not args.write)


if __name__ == "__main__":
    sys.exit(main())
