"""First-party regression tests; do not pretend these fixtures are native OXCE."""
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
import zipfile
import yaml

SPEC=importlib.util.spec_from_file_location('native_snap_probe',Path(__file__).with_name('native_snap_probe.py'))
probe=importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(probe)


def make_save(*,player_tu=64,enemy_hp=24,ammo=4,turn=1,enemy=(5,5,0)):
    units=[{'id':1,'faction':0,'position':[2,2,0],'tu':player_tu,'health':44,'status':0},
           {'id':1000000,'faction':1,'position':list(enemy),'tu':48,'health':enemy_hp,'status':0}]
    items=[{'id':0,'type':'STR_RDTE_TOOL','owner':1,'inventoryslot':'STR_RIGHT_HAND','ammoItem':1},
           {'id':1,'type':'STR_RDTE_CELL','ammoqty':ammo}]
    return {'turn':turn,'selectedUnit':1,'width':10,'length':10,'height':2,'units':units,'items':items}


def make_case(root, *,post_tu=39,post_ammo=3,post_hp=20):
    root.mkdir(parents=True,exist_ok=True)
    for filename,battle in [('pre-shot.asav',make_save()),
                            ('post-shot.asav',make_save(player_tu=post_tu,ammo=post_ammo,enemy_hp=post_hp))]:
        (root/filename).write_text(yaml.safe_dump_all([{'rulesets':['rdte-synthetic-master']},{'battleGame':battle}],sort_keys=True),encoding='utf-8')
    states=['MainMenuState','BattlescapeState','SaveGameState','BattlescapeState',
            'ActionMenuState','BattlescapeState','SaveGameState','BattlescapeState']
    (root/'trace.jsonl').write_text(''.join(json.dumps({'event':'screen-class','class':'N8OpenXcom'+s+'E'})+'\n' for s in states))
    (root/'commands.txt').write_text('KEY 286\nKEY 101\nKEY 50\nCLICK 320 204\nKEY 286\n')
    (root/'user').mkdir(exist_ok=True)
    (root/'user'/'openxcom.log').write_text('[INFO] synthetic test doubles only\n')
    return root

class ShotTests(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory(prefix='oxce-shot-')
        self.addCleanup(self.tmp.cleanup)
        self.root=make_case(Path(self.tmp.name)/'native')
    def test_map_click_calibrated_from_source_and_actual_successful_walk(self):
        b,_=probe.save_game(self.root/'pre-shot.asav')
        self.assertEqual((probe.target_from_calibration(b)['targetMailboxX'],probe.target_from_calibration(b)['targetMailboxY']),(320,204))
    def test_nearby_floor_target_rejects_any_occupying_unit(self):
        self.modify_battle('pre-shot.asav',lambda b:b['units'].append(
            {'id':2,'faction':0,'position':[2,3,0],'tu':60,'health':35,'status':0}))
        with self.assertRaisesRegex(ValueError,'occupied'):probe.assess(self.root)
    def test_previous_real_native_target_turn_is_not_misclassified_as_shot(self):
        archive=Path('/mnt/data/oxce-native-batch-42341948390.zip')
        if not archive.is_file():self.skipTest('B74 real native RED artifact not mounted')
        with zipfile.ZipFile(archive) as z:
            for dst,src in [('pre-shot.asav','first-native-snap-shot/pre-shot.asav'),
                            ('post-shot.asav','first-native-snap-shot/post-shot.asav'),
                            ('trace.jsonl','first-native-snap-shot/trace.jsonl'),
                            ('commands.txt','first-native-snap-shot/commands.txt'),
                            ('user/openxcom.log','first-native-snap-shot/user/openxcom.log')]:
                (self.root/dst).write_bytes(z.read(src))
        with self.assertRaisesRegex(ValueError,'sequence'):probe.assess(self.root)
    def test_native_stochastic_selected_spawn_uses_relative_candidate(self):
        self.modify_battle('pre-shot.asav',lambda b:b['units'][0].update(position=[3,3,0]))
        b,_=probe.save_game(self.root/'pre-shot.asav')
        t=probe.target_from_calibration(b)
        self.assertEqual(t['selectedShooterTile'],[3,3,0])
        self.assertEqual(t['targetTile'],[3,4,0])
        self.assertEqual((t['targetMailboxX'],t['targetMailboxY']),(320,204))
    def test_actual_b75_failed_calibration_now_prepares_relative_target(self):
        archive=Path('/mnt/data/oxce-native-batch-42342827628.zip')
        if not archive.is_file():self.skipTest('B75 native pre-shot evidence unavailable')
        with zipfile.ZipFile(archive) as z:
            saved=z.read('first-native-snap-shot/pre-shot.asav')
        (self.root/'pre-shot.asav').write_bytes(saved)
        b,_=probe.save_game(self.root/'pre-shot.asav')
        t=probe.target_from_calibration(b)
        self.assertEqual(t['targetTile'],[3,4,0])
    def test_native_relative_target_out_of_bounds_rejected(self):
        self.modify_battle('pre-shot.asav',lambda b:b['units'][0].update(position=[3,9,0]))
        with self.assertRaisesRegex(ValueError,'outside map'):probe.assess(self.root)
    def test_native_accepted_shot_requires_shooter_position_unchanged(self):
        self.modify_battle('post-shot.asav',lambda b:b['units'][0].update(position=[2,3,0]))
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_positive_synthetic_contract_still_refuses_full_gameplay(self):
        r=probe.assess(self.root)
        self.assertTrue(r['nativeShotWithTUAndAmmoSpent'])
        self.assertTrue(r['nativeEnemyDamageObserved'])
        for key in ('gameplayQualified','winningDebriefQualified','stockEngineQualified','signalFrontierMvpQualified','humanApproved'):
            self.assertFalse(r[key])
    def modify_battle(self,filename,fn):
        p=self.root/filename;docs=list(yaml.safe_load_all(p.read_text()));fn(docs[1]['battleGame']);p.write_text(yaml.safe_dump_all(docs))
    def test_no_ammo_spend_must_fail(self):
        self.modify_battle('post-shot.asav',lambda b:b['items'][1].update(ammoqty=4))
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_no_tu_spend_must_fail(self):
        self.modify_battle('post-shot.asav',lambda b:b['units'][0].update(tu=64))
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_no_damage_may_be_a_real_miss_but_not_enemy_effect(self):
        self.modify_battle('post-shot.asav',lambda b:b['units'][1].update(health=24))
        r=probe.assess(self.root)
        self.assertTrue(r['nativeShotWithTUAndAmmoSpent'])
        self.assertFalse(r['nativeEnemyDamageObserved'])
    def test_wrong_turn_fails(self):
        self.modify_battle('post-shot.asav',lambda b:b.update(turn=2))
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_different_clip_fails_closed(self):
        self.modify_battle('post-shot.asav',lambda b:b['items'][0].update(ammoItem=123))
        with self.assertRaises(ValueError):probe.assess(self.root)
    def test_wrong_enemy_position_blocks_calibration(self):
        self.modify_battle('pre-shot.asav',lambda b:b['units'][1].update(position=[5,4,0]))
        with self.assertRaisesRegex(ValueError,'hostile placement'):probe.assess(self.root)
    def test_unselected_shooter_blocks_calibration(self):
        self.modify_battle('pre-shot.asav',lambda b:b.update(selectedUnit=77))
        with self.assertRaisesRegex(ValueError,'wrong selected'):probe.assess(self.root)
    def test_missing_action_menu_fails(self):
        p=self.root/'trace.jsonl';p.write_text(''.join(s+'\n' for s in p.read_text().splitlines() if 'ActionMenuState' not in s))
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_wrong_keyboard_or_order_fails(self):
        p=self.root/'commands.txt';p.write_text('KEY 286\nKEY 50\nKEY 101\nCLICK 320 204\nKEY 286\n')
        with self.assertRaisesRegex(ValueError,'sequence'):probe.assess(self.root)
    def test_missing_native_log_fails(self):
        (self.root/'user'/'openxcom.log').unlink()
        with self.assertRaises(FileNotFoundError):probe.assess(self.root)
    def test_native_fatal_prevents_qualification(self):
        (self.root/'user'/'openxcom.log').write_text('[FATAL] original-only shot failed\n')
        self.assertFalse(probe.assess(self.root)['nativeShotWithTUAndAmmoSpent'])
    def test_symlink_to_native_quicksave_rejected(self):
        p=self.root/'pre-shot.asav';p.unlink();p.symlink_to(self.root/'post-shot.asav')
        with self.assertRaises(ValueError):probe.assess(self.root)
    def test_malformed_yaml_fails_closed(self):
        (self.root/'pre-shot.asav').write_text('not: [ a: bad\n')
        with self.assertRaises(yaml.YAMLError):probe.assess(self.root)
    def test_corrupt_trace_fails_closed(self):
        (self.root/'trace.jsonl').write_text('{oops}\n')
        with self.assertRaises(json.JSONDecodeError):probe.assess(self.root)
    def test_duplicate_unit_ids_rejected(self):
        self.modify_battle('pre-shot.asav',lambda b:b['units'].append(dict(b['units'][0])))
        with self.assertRaisesRegex(ValueError,'duplicate'):probe.assess(self.root)
    def test_actual_armed_public_native_archive_cannot_count_as_a_shot(self):
        archive=Path('/mnt/data/oxce-native-batch-42331085644.zip')
        if not archive.is_file():self.skipTest('archived authentic public original-only evidence unavailable')
        with zipfile.ZipFile(archive) as z:
            snap=z.read('armed-player/user/rdte-synthetic-master/_quick_.asav')
            for n in ('pre-shot.asav','post-shot.asav'):(self.root/n).write_bytes(snap)
            for n,member in [('trace.jsonl','armed-player/trace.jsonl'),('commands.txt','armed-player/commands.txt'),('user/openxcom.log','armed-player/user/openxcom.log')]:
                (self.root/n).write_bytes(z.read(member))
        with self.assertRaises(ValueError):probe.assess(self.root)

if __name__=='__main__':unittest.main()
