"""No native co-op executable involved; tests a fake GameClient semantics only."""
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from python_native_read import read_with_upstream_client


class FakeGameClient:
    def __init__(self):
        self.calls = []
    def ok(self, obj):
        self.calls.append(obj)
        return {'ok': True, 'state': 'synthetic-stub'}


class TestBridge(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.bin_dir = tempfile.TemporaryDirectory(prefix='b40-go-read-')
        cls.binary = str(Path(cls.bin_dir.name) / 'catalogcheck')
        subprocess.run(['go', 'build', '-trimpath', '-o', cls.binary, './cmd/catalogcheck'],
                       cwd=Path(__file__).resolve().parents[1], check=True, timeout=60)
    @classmethod
    def tearDownClass(cls):
        cls.bin_dir.cleanup()

    def test_reuses_existing_gameclient_interface(self):
        c = FakeGameClient()
        receipt = read_with_upstream_client(c, self.binary,
                            ('game.inspect_state', 'coop.inspect_checksum'))
        self.assertEqual([r['native_command'] for r in receipt], ['get_state', 'shared_checksum'])
        self.assertEqual([x['cmd'] for x in c.calls], ['get_state', 'shared_checksum'])
        self.assertEqual(receipt[0]['evidence_class'], 'RESPONSE_FROM_CALLER_NOT_ATTESTED')

    def test_control_refused_before_client_call(self):
        c = FakeGameClient()
        with self.assertRaises(subprocess.CalledProcessError):
            read_with_upstream_client(c, self.binary, ('session.host',))
        self.assertEqual(c.calls, [])


if __name__ == '__main__':
    unittest.main()
