import hashlib
from pathlib import Path
import subprocess
import tempfile
import unittest

from native_gate import inspect_binary

class P:
    def __init__(self, stdout, code=0):
        self.stdout, self.stderr, self.returncode = stdout, '', code

class NativeGateTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.path = Path(self.tmp.name) / 'engine'
        self.path.write_bytes(b'\x7fELF' + b'\x00' * 40)
        self.digest = hashlib.sha256(self.path.read_bytes()).hexdigest()

    def test_positive_ldd_no_claim_of_execution(self):
        r = inspect_binary(self.path, self.digest, ldd=lambda _: P('libSDL-1.2.so.0 => /system/path (0x00)'))
        self.assertEqual(r['status'], 'RUNTIME_DEPS_READY')
        self.assertFalse(r['native_launched'])
        self.assertFalse(r['native_gameplay_accepted'])

    def test_missing_runtime_libraries(self):
        r = inspect_binary(self.path, self.digest, ldd=lambda _: P('libSDL-1.2.so.0 => not found\nlibSDL_net-1.2.so.0 => not found'))
        self.assertEqual(r['reason'], 'MISSING_RUNTIME_LIBRARIES')
        self.assertEqual(r['missing_sonames'], ['libSDL-1.2.so.0', 'libSDL_net-1.2.so.0'])

    def test_mismatched_digest_rejected_without_ldd(self):
        def deny(_): self.fail('invoked ldd on wrong binary')
        r = inspect_binary(self.path, 'f'*64, ldd=deny)
        self.assertEqual(r['reason'], 'SHA_MISMATCH')

    def test_malformed_digest(self):
        r = inspect_binary(self.path, 'FAKE', ldd=lambda _: self.fail('ldd'))
        self.assertEqual(r['reason'], 'INVALID_EXPECTED_SHA')

    def test_reject_symlink(self):
        link = self.path.with_name('link')
        link.symlink_to(self.path)
        r = inspect_binary(link, self.digest, ldd=lambda _: self.fail('ldd'))
        self.assertEqual(r['reason'], 'NOT_A_REGULAR_NONSYMLINK_FILE')

    def test_reject_non_elf(self):
        self.path.write_bytes(b'NOTELF')
        d = hashlib.sha256(self.path.read_bytes()).hexdigest()
        r = inspect_binary(self.path, d, ldd=lambda _: self.fail('ldd'))
        self.assertEqual(r['reason'], 'NOT_ELF')

    def test_ldd_timeout_fails_closed(self):
        def timeout(_): raise subprocess.TimeoutExpired('ldd', 15)
        r = inspect_binary(self.path, self.digest, ldd=timeout)
        self.assertEqual(r['reason'], 'LDD_ERROR:TimeoutExpired')

    def test_ldd_nonzero_fails_closed(self):
        r = inspect_binary(self.path, self.digest, ldd=lambda _: P('bad ELF', code=1))
        self.assertEqual(r['reason'], 'LDD_NONZERO')

if __name__ == '__main__': unittest.main()
