"""Reconocer efectos declarados en una gramática POSIX acotada, sin ejecutar código.

No es un sandbox ni atesta binarios, configuración, entorno o permisos efectivos.
Python arbitrario, heredocs, expansiones y composiciones no soportadas quedan unknown.
"""
from pathlib import PurePosixPath
import re

VERSION = "command-effects/v2"
MAX_CHARS = 65536
MAX_TOKENS = 4096
MAX_DEPTH = 8


def _effect(kind, reason):
    return {"kind": kind, "reason": reason}


def _tokens(source):
    """Retener la diferencia entre operadores y argumentos que estaban entre comillas."""
    tokens, word = [], []
    active, quote, i = False, None, 0

    def flush():
        nonlocal active
        if active:
            tokens.append(("".join(word), False))
            word.clear()
            active = False

    while i < len(source):
        c = source[i]
        if c == "\x00":
            raise ValueError("invalid_character")
        if quote == "'":
            if c == "'":
                quote = None
            else:
                word.append(c)
        elif c == "\\":
            i += 1
            if i == len(source):
                raise ValueError("malformed_shell")
            following = source[i]
            if following != "\n":
                if quote == '"' and following not in '$`"\\':
                    word.append("\\")
                word.append(following)
                active = True
        elif c in "$`":
            raise ValueError("shell_expansion")
        elif quote == '"':
            if c == '"':
                quote = None
            else:
                word.append(c)
        elif c in "'\"":
            quote, active = c, True
        elif c in " \t\r":
            flush()
        elif c == "#" and not active:
            while i < len(source) and source[i] != "\n":
                i += 1
            continue
        elif c in ";&|<>\n":
            flush()
            op = c
            if i + 1 < len(source) and source[i + 1] == c and c != "\n":
                op += c
                i += 1
            if op == "<<":
                raise ValueError("unsupported_heredoc")
            tokens.append((op, True))
        elif c in "(){}~":
            raise ValueError("unsupported_shell_syntax")
        else:
            word.append(c)
            active = True
        if len(tokens) > MAX_TOKENS:
            raise ValueError("limit_exceeded")
        i += 1
    if quote:
        raise ValueError("malformed_shell")
    flush()
    if len(tokens) > MAX_TOKENS:
        raise ValueError("limit_exceeded")
    return tokens


def inspect_command(command, allowed=(), _depth=0):
    if not isinstance(command, str) or not command.strip():
        return _effect("unknown", "missing_command")
    if "\x00" in command:
        return _effect("unknown", "invalid_character")
    if len(command) > MAX_CHARS or _depth > MAX_DEPTH:
        return _effect("unknown", "limit_exceeded")
    try:
        tokens = _tokens(command)
    except ValueError as error:
        return _effect("unknown", str(error))
    groups, current, last = [], [], None
    for token in tokens:
        value, operator = token
        if operator and value in {";", "&&", "\n"}:
            if not current:
                if value == "\n" and last in {None, "\n", ";"}:
                    continue
                return _effect("unknown", "empty_sequence")
            groups.append(current)
            current = []
            last = value
        else:
            current.append(token)
            last = None
    if current:
        groups.append(current)
    elif last == "&&":
        return _effect("unknown", "incomplete_sequence")
    if not groups:
        return _effect("unknown", "missing_command")
    effects = [_simple(group, allowed, _depth) for group in groups]
    for kind in ("delivery", "mutation", "unknown", "check", "read"):
        for effect in effects:
            if effect["kind"] == kind:
                return effect
    return _effect("unknown", "unsupported_command")


def command_kind(command, allowed=()):
    return inspect_command(command, allowed)["kind"]


def _simple(tokens, allowed, depth):
    words = [value for value, operator in tokens if not operator]
    if not words or not words[0]:
        return _effect("unknown", "missing_command")
    program = PurePosixPath(words[0]).name
    if program == "git":
        effect = _git(words[1:])
        if effect["kind"] == "delivery":
            return effect
    elif program == "gh" and any(w in {"create", "merge", "release"} for w in words[1:]):
        return _effect("delivery", "delivery_command")
    if any(operator and value in {">", ">>"} for value, operator in tokens):
        return _effect("mutation", "output_redirection")
    if any(operator for _, operator in tokens):
        return _effect("unknown", "unsupported_composition")
    if program in {"sh", "bash", "zsh"} and len(words) == 3 and words[1] in {"-c", "-lc"}:
        return inspect_command(words[2], allowed, depth + 1)
    if program == "git":
        return effect
    if program in {"rm", "mv", "cp", "touch", "mkdir", "rmdir", "tee", "truncate", "chmod", "ln", "dd"}:
        return _effect("mutation", "mutation_command")
    if program == "sed" and any(w.startswith("-i") for w in words[1:]):
        return _effect("mutation", "sed_in_place")
    if isinstance(allowed, (list, tuple)):
        for declared in allowed:
            try:
                candidate = _tokens(declared) if isinstance(declared, str) else []
                if candidate and not any(operator for _, operator in candidate) and words == [w for w, _ in candidate]:
                    return _effect("check", "declared_check")
            except ValueError:
                continue
    if program == "sed":
        if len(words) == 4 and words[1] == "-n" and re.fullmatch(r"\d+(,\d+)?p", words[2]):
            return _effect("read", "numbered_read")
        return _effect("unknown", "unsupported_sed_script")
    if program in {"shasum", "sha256sum"}:
        return _checksum(program, words[1:])
    if program in {"pwd", "ls", "cat", "head", "tail", "wc", "stat", "rg", "nl"}:
        if program == "rg" and any(w.startswith(("--pre", "--hostname-bin")) for w in words[1:]):
            return _effect("unknown", "external_helper")
        return _effect("read", "literal_read")
    return _effect("unknown", "unsupported_command")


def _git(args):
    no_pager, configured = False, False
    while args and args[0].startswith("-"):
        option = args[0]
        if option in {"--no-pager", "--no-optional-locks"}:
            no_pager = no_pager or option == "--no-pager"
            args = args[1:]
        elif option in {"-c", "-C", "--git-dir", "--work-tree", "--namespace", "--config-env"}:
            if len(args) < 2:
                return _effect("unknown", "unsupported_git_configuration")
            configured, args = True, args[2:]
        elif option.startswith(("--git-dir=", "--work-tree=", "--namespace=", "--config-env=")) or option in {"--paginate", "-p", "--bare"}:
            configured, args = True, args[1:]
        else:
            return _effect("unknown", "unsupported_git_option")
    if not args:
        return _effect("unknown", "unsupported_git_command")
    subcommand, *options = args
    if subcommand in {"push", "commit", "tag"}:
        return _effect("delivery", "git_delivery")
    if subcommand in {"add", "checkout", "switch", "restore", "reset", "clean", "update-index", "apply", "merge", "rebase"}:
        return _effect("mutation", "git_mutation")
    if configured:
        return _effect("unknown", "unsupported_git_configuration")
    if subcommand == "status":
        allowed = {"--short", "-s", "--branch", "-b", "--porcelain", "--porcelain=v1", "--porcelain=v2",
                   "--porcelain=1", "--porcelain=2", "--untracked-files", "-u", "-uno", "-unormal", "-uall",
                   "--untracked-files=no", "--untracked-files=normal", "--untracked-files=all"}
        for option in options:
            if option == "--":
                break
            if option not in allowed:
                return _effect("unknown", "unsupported_git_option")
        return _effect("read", "git_status")
    if subcommand == "diff":
        flags = options[:options.index("--")] if "--" in options else options
        if any(w.startswith("--output") for w in flags):
            return _effect("mutation", "git_diff_output")
        if not no_pager or not {"--no-ext-diff", "--no-textconv"}.issubset(flags):
            return _effect("unknown", "git_helpers_not_disabled")
        safe = {"--no-ext-diff", "--no-textconv", "--stat", "--numstat", "--name-only", "--name-status",
                "--check", "--cached", "--staged", "--no-color", "--exit-code", "--quiet"}
        if any(w not in safe and not re.fullmatch(r"[A-Za-z0-9_./~^:-]+", w) for w in flags):
            return _effect("unknown", "unsupported_git_option")
        if any(w.startswith("-") and w not in safe for w in flags):
            return _effect("unknown", "unsupported_git_option")
        return _effect("read", "git_diff_without_helpers")
    return _effect("unknown", "unsupported_git_command")


def _checksum(program, args):
    i = 0
    while i < len(args):
        word = args[i]
        if word == "--":
            return _effect("read", "checksum_read")
        if program == "shasum" and word in {"-a", "--algorithm"}:
            i += 1
            if i == len(args) or args[i] not in {"1", "224", "256", "384", "512", "512224", "512256"}:
                return _effect("unknown", "unsupported_checksum_option")
        elif word.startswith("-") and word not in {"-b", "-t", "--binary", "--text"}:
            return _effect("unknown", "unsupported_checksum_option")
        i += 1
    return _effect("read", "checksum_read")
