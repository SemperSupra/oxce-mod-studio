"""Fail-closed preflight for recovered *native* OXCE binaries (stdlib only).

This does not launch OXCE or establish any gameplay / TestServer / Q4 grade.
Only inspect exact ELF runtime dependencies, SHA provenance and namespaces.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys

SHA_RE = re.compile(r'^[0-9a-f]{64}$')
MISSING_RE = re.compile(r'^\s*(\S+)\s+=>\s+not found\s*$', re.MULTILINE)
MAX_BINARY_BYTES = 128 * 1024 * 1024


def inspect_binary(path: Path, expected_sha: str, *, ldd=lambda p: subprocess.run(
    ['ldd', str(p)], text=True, capture_output=True, timeout=15, check=False)) -> dict:
    """Inspect a pinned ELF; do not execute it, connect, or alter any files."""
    result = {
        'schema': 'oxce.native-readiness.v1',
        'status': 'BLOCKED',
        'native_launched': False,
        'listener_verified': False,
        'native_gameplay_accepted': False,
        'evidence_grade': 'runtime-preflight-only',
        'binary': str(path),
        'expected_sha256': expected_sha,
        'sha256': None,
        'missing_sonames': [],
        'reason': '',
    }
    if not SHA_RE.fullmatch(expected_sha):
        result['reason'] = 'INVALID_EXPECTED_SHA'; return result
    if path.is_symlink() or not path.is_file():
        result['reason'] = 'NOT_A_REGULAR_NONSYMLINK_FILE'; return result
    if path.stat().st_size > MAX_BINARY_BYTES:
        result['reason'] = 'BINARY_TOO_LARGE'; return result
    with path.open('rb') as fh:
        sha = hashlib.file_digest(fh, 'sha256').hexdigest()
    result['sha256'] = sha
    if sha != expected_sha:
        result['reason'] = 'SHA_MISMATCH'; return result
    with path.open('rb') as fh:
        if fh.read(4) != b'\x7fELF':
            result['reason'] = 'NOT_ELF'; return result
    try:
        probe = ldd(path)
    except (OSError, subprocess.TimeoutExpired) as err:
        result['reason'] = f'LDD_ERROR:{type(err).__name__}'; return result
    output = (probe.stdout or '') + '\n' + (probe.stderr or '')
    missing = sorted(set(MISSING_RE.findall(output)))
    result['missing_sonames'] = missing
    if missing:
        result['reason'] = 'MISSING_RUNTIME_LIBRARIES'; return result
    if probe.returncode != 0:
        result['reason'] = 'LDD_NONZERO'; return result
    result['status'] = 'RUNTIME_DEPS_READY'
    result['reason'] = 'NATIVE_LAUNCH_AND_NETWORK_ISOLATION_STILL_NOT_TESTED'
    return result


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--binary', type=Path, required=True)
    p.add_argument('--expected-sha256', required=True)
    p.add_argument('--out', type=Path, help='atomic JSON receipt (no overwrite)')
    args = p.parse_args(argv)
    r = inspect_binary(args.binary, args.expected_sha256)
    payload = json.dumps(r, sort_keys=True, indent=2) + '\n'
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        with args.out.open('x', encoding='utf8') as f:
            f.write(payload)
    print(payload, end='')
    return 0 if r['status'] == 'RUNTIME_DEPS_READY' else 20

if __name__ == '__main__':
    sys.exit(main())
