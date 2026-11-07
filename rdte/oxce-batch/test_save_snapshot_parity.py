"""Offline adversarial source tests; no native engine execution/acceptance."""
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

import yaml
sys.path.insert(0, str(Path(__file__).resolve().parent))
from save_snapshot_parity import evaluate


def unit(uid=1, tu=50, pos=(2, 3, 0), hp=33, faction=0, status=0, saved=False):
    if saved:
        return dict(id=uid, faction=faction, position=list(pos), tu=tu, health=hp, status=status)
    return dict(id=uid, faction=faction, x=pos[0], y=pos[1], z=pos[2], tu=tu, hp=hp, status=status)


class SaveParityTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.save = self.root / '_quick_.asav'
        self.trace = self.root / 'battle.jsonl'
        self.save_units = [unit(1, saved=True), unit(42, 33, (4, 5, 0), 22, 1, saved=True)]
        self.native_units = [unit(1), unit(42, 33, (4, 5, 0), 22, 1)]
        self.write()

    def write(self, *, save_turn=1, native_turn=1, save_side=0, native_side=0):
        data = {'battleGame': {'units': self.save_units, 'turn': save_turn, 'side': save_side}}
        self.save.write_text(yaml.safe_dump({'name': '_quick_.asav'}) + '---\n' + yaml.safe_dump(data))
        self.trace.write_text(json.dumps({'event':'battle-snapshot','turn':native_turn,
                                          'side':native_side,'units':self.native_units})+'\n')

    def test_baseline_and_no_false_loaded_state_claim(self):
        result = evaluate(self.save, self.trace)
        self.assertTrue(result['savedSnapshotEqualsLastNativeSnapshot'])
        self.assertFalse(result['loadedStateSemanticsQualified'])
        self.assertFalse(result['gameplayQualified'])

    def test_unserialized_side_on_real_oxce_save_is_disclosed_not_invented(self):
        header, game = list(yaml.safe_load_all(self.save.read_text()))
        del game['battleGame']['side']
        self.save.write_text(yaml.safe_dump(header)+'---\n'+yaml.safe_dump(game))
        result = evaluate(self.save, self.trace)
        self.assertTrue(result['savedSnapshotEqualsLastNativeSnapshot'])
        self.assertFalse(result['savedSideSerialized'])
        self.assertFalse(result['savedSideCompared'])
        self.native_units = [unit(1), unit(42, 33, (4, 5, 0), 22, 1)]
        self.trace.write_text(json.dumps({'event':'battle-snapshot','turn':1, 'side':1,
                                          'units':self.native_units})+'\n')
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])

    def test_position_mutation_refuses(self):
        self.save_units[0]['position'] = [9, 3, 0]
        self.write()
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])

    def test_tu_mutation_refuses(self):
        self.save_units[0]['tu'] = 55
        self.write()
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])

    def test_health_faction_status_mutations_refuse(self):
        for key, val in [('health', 4), ('faction', 1), ('status', 1)]:
            with self.subTest(key=key):
                original = self.save_units[0][key]
                self.save_units[0][key] = val
                self.write()
                self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])
                self.save_units[0][key] = original

    def test_turn_or_side_mismatch_refuses(self):
        self.write(save_turn=2)
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])
        self.write(save_side=1)
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])

    def test_extra_and_missing_units_refuse(self):
        self.native_units.pop()
        self.write()
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])
        self.native_units = [unit(1), unit(42, 33, (4, 5, 0), 22, 1), unit(99)]
        self.write()
        self.assertFalse(evaluate(self.save, self.trace)['savedSnapshotEqualsLastNativeSnapshot'])

    def test_duplicate_units_refuse(self):
        self.native_units.append(self.native_units[0])
        self.write()
        with self.assertRaisesRegex(ValueError, 'duplicate unit ID'):
            evaluate(self.save, self.trace)

    def test_malformed_battle_line_refuses(self):
        self.trace.write_text('{BAD}\n')
        with self.assertRaisesRegex(ValueError, 'malformed native battle line'):
            evaluate(self.save, self.trace)

    def test_save_without_correct_header_refuses(self):
        self.save.write_text(yaml.safe_dump({'name': 'fake.sav'})+'---\n'+
                             yaml.safe_dump({'battleGame':{'units':self.save_units,'turn':1,'side':0}}))
        with self.assertRaisesRegex(ValueError, 'not an OXCE quick-save'):
            evaluate(self.save, self.trace)

    def test_symlink_and_empty_file_refuse(self):
        other = self.root / 'other.asav'
        other.symlink_to(self.save)
        with self.assertRaisesRegex(ValueError, 'untrusted'):
            evaluate(other, self.trace)
        self.trace.write_bytes(b'')
        with self.assertRaisesRegex(ValueError, 'untrusted'):
            evaluate(self.save, self.trace)

    def test_missing_yaml_battle_game_refuses(self):
        self.save.write_text(yaml.safe_dump({'name': '_quick_.asav'})+'---\n'+yaml.safe_dump({'unrelated':1}))
        with self.assertRaisesRegex(ValueError, 'not an OXCE quick-save'):
            evaluate(self.save, self.trace)

    def test_wrong_event_type_refuses(self):
        self.trace.write_text(json.dumps({'event':'simulated','turn':1,'side':0,'units':self.native_units})+'\n')
        with self.assertRaisesRegex(ValueError, 'unexpected native event'):
            evaluate(self.save, self.trace)


if __name__ == '__main__':
    unittest.main()
