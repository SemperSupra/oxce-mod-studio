"""Nonexecuting, fail-closed *direct-dependency* preflight for pinned Linux ELF.

Never execute the candidate ELF, its interpreter, ldd, or an OXCE game. This
verifies only direct DT_NEEDED names in the host ldconfig cache. It does NOT
prove transitive dependencies, actual loading, game startup, or network safety.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import subprocess
import sys

SHA_RE = re.compile(r'^[0-9a-f]{64}$')
NEEDED_RE = re.compile(r'\(NEEDED\)\s+Shared library:\s+\[([^\]\r\n]+)\]')
CACHE_RE = re.compile(r'^\s*([^\s()]+)\s+\(([^)]+)\)\s+=>\s+(/\S+)\s*$', re.MULTILINE)
MAX_BINARY_BYTES = 128 * 1024 * 1024
MAX_TOOL_OUTPUT = 4 * 1024 * 1024


def _readelf_dynamic(fd: int):
    """Pass only a verified open FD, never the potentially replaceable path."""
    return subprocess.run(
        ['/usr/bin/readelf', '-d', f'/proc/self/fd/{fd}'], pass_fds=(fd,),
        text=True, capture_output=True, timeout=15, check=False, env={'LC_ALL': 'C'},
    )


def _host_cache():
    """Query trusted host loader cache, not an interpreter for the target."""
    return subprocess.run(
        ['/sbin/ldconfig', '-p'], text=True, capture_output=True,
        timeout=10, check=False, env={'LC_ALL': 'C'},
    )


def _hash_fd(fd: int) -> str:
    os.lseek(fd, 0, os.SEEK_SET)
    digest = hashlib.sha256()
    while True:
        block = os.read(fd, 1024 * 1024)
        if not block:
            break
        digest.update(block)
    return digest.hexdigest()


def _fingerprint(st):
    return (st.st_dev, st.st_ino, st.st_size, st.st_mtime_ns, st.st_ctime_ns)


def inspect_binary(path: Path, expected_sha: str, *, readelf=None, library_cache=None) -> dict:
    """Hash one no-follow FD; inspect its direct imports without executing it."""
    result = {
        'schema': 'oxce.native-readiness.v2',
        'status': 'BLOCKED',
        'native_launched': False,
        'listener_verified': False,
        'native_gameplay_accepted': False,
        'transitive_dependencies_verified': False,
        'evidence_grade': 'static-direct-elf-preflight-only',
        'parser': 'readelf-dt-needed-via-open-fd',
        'binary': str(path),
        'expected_sha256': expected_sha,
        'sha256': None,
        'binary_abi': None,
        'direct_sonames': [],
        'missing_sonames': [],
        'reason': '',
    }
    if not SHA_RE.fullmatch(expected_sha):
        result['reason'] = 'INVALID_EXPECTED_SHA'; return result
    path = Path(path)
    try:
        initial = path.lstat()
    except OSError:
        result['reason'] = 'NOT_A_REGULAR_NONSYMLINK_FILE'; return result
    if not stat.S_ISREG(initial.st_mode):
        result['reason'] = 'NOT_A_REGULAR_NONSYMLINK_FILE'; return result
    if initial.st_size > MAX_BINARY_BYTES:
        result['reason'] = 'BINARY_TOO_LARGE'; return result
    if not hasattr(os, 'O_NOFOLLOW') or not Path('/proc/self/fd').is_dir():
        result['reason'] = 'LINUX_NOFOLLOW_FD_UNAVAILABLE'; return result
    fd = None
    try:
        try:
            fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | getattr(os, 'O_CLOEXEC', 0))
        except OSError:
            result['reason'] = 'OPEN_NOFOLLOW_FAILED'; return result
        opened = os.fstat(fd)
        if not stat.S_ISREG(opened.st_mode) or _fingerprint(initial) != _fingerprint(opened):
            result['reason'] = 'INPUT_CHANGED_DURING_PROBE'; return result
        if opened.st_size > MAX_BINARY_BYTES:
            result['reason'] = 'BINARY_TOO_LARGE'; return result
        digest = _hash_fd(fd)
        result['sha256'] = digest
        if digest != expected_sha:
            result['reason'] = 'SHA_MISMATCH'; return result
        header = os.pread(fd, 20, 0)
        if header[:4] != b'\x7fELF':
            result['reason'] = 'NOT_ELF'; return result
        # The qualified public runner and pinned OXCE binary are ELF64/x86-64.
        # A similarly named 32-bit library is not proof of direct loadability.
        if len(header) < 20 or header[4] != 2 or header[5] != 1 or int.from_bytes(header[18:20], 'little') != 62:
            result['reason'] = 'UNSUPPORTED_ELF_ABI'; return result
        result['binary_abi'] = 'elf64-x86_64'
        probe = readelf if readelf is not None else _readelf_dynamic
        try:
            proc = probe(fd)
        except (OSError, subprocess.TimeoutExpired) as err:
            result['reason'] = f'READELF_ERROR:{type(err).__name__}'; return result
        if proc.returncode != 0:
            result['reason'] = 'READELF_NONZERO'; return result
        output = (proc.stdout or '') + '\n' + (proc.stderr or '')
        if len(output) > MAX_TOOL_OUTPUT:
            result['reason'] = 'READELF_OUTPUT_TOO_LARGE'; return result
        needed = sorted(set(NEEDED_RE.findall(output)))
        result['direct_sonames'] = needed
        if not needed:
            result['reason'] = 'NO_PROVEN_DIRECT_DYNAMIC_IMPORTS'; return result
        cache_probe = library_cache if library_cache is not None else _host_cache
        try:
            cached = cache_probe()
        except (OSError, subprocess.TimeoutExpired) as err:
            result['reason'] = f'LDCONFIG_ERROR:{type(err).__name__}'; return result
        if cached.returncode != 0:
            result['reason'] = 'LDCONFIG_NONZERO'; return result
        listing = cached.stdout or ''
        if len(listing) > MAX_TOOL_OUTPUT:
            result['reason'] = 'LDCONFIG_OUTPUT_TOO_LARGE'; return result
        # Do not accept 32-bit-only entries for an ELF64/x86-64 engine.
        present = {soname for soname, flags, file_path in CACHE_RE.findall(listing)
                   if 'x86-64' in flags.split(',') and os.path.isfile(file_path)}
        missing = sorted(set(needed) - present)
        result['missing_sonames'] = missing
        # Check *again* after all subprocesses: no pathname swap or in-place
        # mutation can turn a passing static receipt into another file's claim.
        try:
            final_path = path.lstat()
            final_fd = os.fstat(fd)
        except OSError:
            result['reason'] = 'INPUT_CHANGED_DURING_PROBE'; return result
        if _fingerprint(final_path) != _fingerprint(initial) or _fingerprint(final_fd) != _fingerprint(opened) or _hash_fd(fd) != digest:
            result['reason'] = 'INPUT_CHANGED_DURING_PROBE'; return result
        if missing:
            result['reason'] = 'DIRECT_DEPENDENCIES_NOT_IN_HOST_CACHE'; return result
        result['status'] = 'DIRECT_DEPS_IN_HOST_CACHE'
        result['reason'] = 'TRANSITIVE_LOAD_AND_NATIVE_GAMEPLAY_NOT_VERIFIED'
        return result
    finally:
        if fd is not None:
            os.close(fd)


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--binary', type=Path, required=True)
    p.add_argument('--expected-sha256', required=True)
    p.add_argument('--out', type=Path, help='exclusive JSON receipt path (no overwrite)')
    args = p.parse_args(argv)
    r = inspect_binary(args.binary, args.expected_sha256)
    payload = json.dumps(r, sort_keys=True, indent=2) + '\n'
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        with args.out.open('x', encoding='utf8') as f:
            f.write(payload)
    print(payload, end='')
    return 0 if r['status'] == 'DIRECT_DEPS_IN_HOST_CACHE' else 20

if __name__ == '__main__':
    sys.exit(main())
