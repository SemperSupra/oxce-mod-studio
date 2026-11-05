import hashlib
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from native_gate import inspect_binary


class P:
    def __init__(self, stdout, code=0):
        self.stdout, self.stderr, self.returncode = stdout, '', code


DYNAMIC = '''Dynamic section at offset 0x2df0 contains 26 entries:
 0x0000000000000001 (NEEDED) Shared library: [libc.so.6]
 0x0000000000000001 (NEEDED) Shared library: [libSDL-1.2.so.0]
'''
HOST_CACHE = '''2 libs found in cache `/etc/ld.so.cache'
 libSDL-1.2.so.0 (libc6,x86-64) => /lib/x86_64-linux-gnu/libSDL-1.2.so.0
 libc.so.6 (libc6,x86-64) => /lib/x86_64-linux-gnu/libc.so.6
'''


class NativeGateTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.path = Path(self.tmp.name) / 'engine'
        self.path.write_bytes(b'\x7fELF' + b'\x00' * 40)
        self.digest = hashlib.sha256(self.path.read_bytes()).hexdigest()

    def inspect(self, *, readelf=None, cache=None):
        if readelf is None:
            readelf = lambda fd: P(DYNAMIC)
        if cache is None:
            cache = lambda: P(HOST_CACHE)
        # Avoid relying on a particular container's installed SDL libraries.
        with patch('native_gate.os.path.isfile', return_value=True):
            return inspect_binary(self.path, self.digest, readelf=readelf, library_cache=cache)

    def test_readelf_direct_import_positive_no_native_claim(self):
        r = self.inspect()
        self.assertEqual(r['status'], 'DIRECT_DEPS_IN_HOST_CACHE')
        self.assertEqual(r['direct_sonames'], ['libSDL-1.2.so.0', 'libc.so.6'])
        self.assertFalse(r['native_launched'])
        self.assertFalse(r['transitive_dependencies_verified'])
        self.assertFalse(r['native_gameplay_accepted'])

    def test_missing_direct_lib(self):
        r = self.inspect(cache=lambda: P('libc.so.6 (libc6) => /lib/libc.so.6'))
        self.assertEqual(r['reason'], 'DIRECT_DEPENDENCIES_NOT_IN_HOST_CACHE')
        self.assertEqual(r['missing_sonames'], ['libSDL-1.2.so.0'])

    def test_wrong_digest_fails_without_probe(self):
        r = inspect_binary(self.path, 'f'*64, readelf=lambda fd: self.fail('readelf executed'))
        self.assertEqual(r['reason'], 'SHA_MISMATCH')

    def test_malformed_digest(self):
        r = inspect_binary(self.path, 'FAKE', readelf=lambda fd: self.fail('readelf executed'))
        self.assertEqual(r['reason'], 'INVALID_EXPECTED_SHA')

    def test_reject_symlink(self):
        link = self.path.with_name('link')
        link.symlink_to(self.path)
        r = inspect_binary(link, self.digest, readelf=lambda fd: self.fail('readelf executed'))
        self.assertEqual(r['reason'], 'NOT_A_REGULAR_NONSYMLINK_FILE')

    def test_reject_non_elf(self):
        self.path.write_bytes(b'NOTELF')
        d = hashlib.sha256(self.path.read_bytes()).hexdigest()
        r = inspect_binary(self.path, d, readelf=lambda fd: self.fail('readelf executed'))
        self.assertEqual(r['reason'], 'NOT_ELF')

    def test_readelf_timeout_fails_closed(self):
        def timeout(fd):
            raise subprocess.TimeoutExpired('readelf', 15)
        r = self.inspect(readelf=timeout)
        self.assertEqual(r['reason'], 'READELF_ERROR:TimeoutExpired')

    def test_readelf_nonzero_fails_closed(self):
        r = self.inspect(readelf=lambda fd: P('invalid ELF', code=1))
        self.assertEqual(r['reason'], 'READELF_NONZERO')

    def test_ldconfig_nonzero_fails_closed(self):
        r = self.inspect(cache=lambda: P('broken', code=1))
        self.assertEqual(r['reason'], 'LDCONFIG_NONZERO')

    def test_no_imports_no_runtime_green(self):
        r = self.inspect(readelf=lambda fd: P('Dynamic section at offset 0x00 contains 0 entries'))
        self.assertEqual(r['reason'], 'NO_PROVEN_DIRECT_DYNAMIC_IMPORTS')

    def test_rename_after_hash_fails_closed(self):
        def swap(fd):
            self.path.rename(self.path.with_name('original'))
            self.path.write_bytes(b'\x7fELF' + b'Z' * 40)
            return P(DYNAMIC)
        r = self.inspect(readelf=swap)
        self.assertEqual(r['reason'], 'INPUT_CHANGED_DURING_PROBE')

    def test_modify_in_place_after_hash_fails_closed(self):
        def modify(fd):
            self.path.write_bytes(b'\x7fELF' + b'Z' * 40)
            return P(DYNAMIC)
        r = self.inspect(readelf=modify)
        self.assertEqual(r['reason'], 'INPUT_CHANGED_DURING_PROBE')

    def test_direct_probe_references_verified_open_fd(self):
        from native_gate import _readelf_dynamic
        calls = []
        def fake_run(cmd, **kwargs):
            calls.append((cmd, kwargs))
            return P(DYNAMIC)
        with patch('native_gate.subprocess.run', side_effect=fake_run):
            _readelf_dynamic(9)
        self.assertEqual(calls[0][0], ['/usr/bin/readelf', '-d', '/proc/self/fd/9'])
        self.assertEqual(calls[0][1]['pass_fds'], (9,))


if __name__ == '__main__':
    unittest.main()
