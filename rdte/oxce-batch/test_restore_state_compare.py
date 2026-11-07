import json
from pathlib import Path
import sys
import tempfile
import unittest
import yaml
sys.path.insert(0,str(Path(__file__).resolve().parent))
from restore_state_compare import evaluate

class RestoreTests(unittest.TestCase):
    def setUp(self):
        tmp=tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.root=Path(tmp.name)
        self.before=self.root/'pre.asav';self.after=self.root/'post.asav'
        self.battle=self.root/'battle.jsonl';self.screen=self.root/'trace.jsonl'
        self.commands=self.root/'commands.txt';self.log=self.root/'openxcom.log'
        self.units=[{'id':1,'faction':0,'position':[2,3,0],'tu':54,'health':29,'status':0}]
        self.write()

    def write(self, afterTurn=1, beforeTurn=1, secondTurn=True, postUnits=None):
        def doc(turn,units): return yaml.safe_dump({'name':'_quick_.asav'})+'---\n'+yaml.safe_dump({'battleGame':{'units':units,'turn':turn}})
        self.before.write_text(doc(beforeTurn,self.units))
        self.after.write_text(doc(afterTurn,postUnits if postUnits is not None else self.units))
        base={'event':'battle-snapshot','side':0,'turn':1,'units':[{'id':1,'faction':0,'x':2,'y':3,'z':0,'tu':54,'hp':29,'status':0}]}
        latter={**base,'turn':2} if secondTurn else base
        self.battle.write_text(json.dumps(base)+'\n'+json.dumps(latter)+'\n')
        states=['BattlescapeState','SaveGameState','BattlescapeState','NextTurnState','BattlescapeState','LoadGameState','BattlescapeState','SaveGameState']
        self.screen.write_text('\n'.join(json.dumps({'event':'screen-class','class':s}) for s in states)+'\n')
        self.commands.write_text('CLICK 320 204\nKEY 286\nCLICK 510 304\nCLICK 320 200\nKEY 290\nKEY 286\n')
        self.log.write_text('[INFO] game\n')
    def e(self):return evaluate(self.before,self.after,self.battle,self.screen,self.commands,self.log)
    def test_positive_does_not_upgrade_full_gameplay(self):
        r=self.e();self.assertTrue(r['restoredObservableUnitStateSubsetQualified']);self.assertFalse(r['fullSaveRestoreSemanticsQualified'])
    def test_missing_intervening_native_turn_refused(self):
        self.write(secondTurn=False);self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_wrong_restore_turn_refused(self):
        self.write(afterTurn=2);self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_mutated_unit_refused(self):
        u=[dict(self.units[0],tu=51)];self.write(postUnits=u)
        r=self.e();self.assertFalse(r['restoredObservableUnitStateSubsetQualified']);self.assertEqual(r['changedUnits'],[1])
    def test_missing_load_state_refused(self):
        self.screen.write_text('\n'.join(x for x in self.screen.read_text().splitlines() if 'LoadGameState' not in x)+'\n')
        self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_missing_postload_save_refused(self):
        lines=self.screen.read_text().splitlines();self.screen.write_text('\n'.join(lines[:-1])+'\n')
        self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_command_out_of_order_refused(self):
        self.commands.write_text('KEY 290\nKEY 286\nCLICK 510 304\nCLICK 320 200\nKEY 286\n')
        self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_log_fatal_refused(self):
        self.log.write_text('[FATAL] oops\n');self.assertFalse(self.e()['restoredObservableUnitStateSubsetQualified'])
    def test_yaml_or_trace_corruption_refused(self):
        self.after.write_text('wrong');
        with self.assertRaises(ValueError):self.e()
        self.write();self.battle.write_text('{BROKEN}');
        with self.assertRaises(ValueError):self.e()
    def test_symlink_quicksave_refused(self):
        self.after.unlink();self.after.symlink_to(self.before)
        with self.assertRaises(ValueError):self.e()
    def test_wrong_native_kind_refused(self):
        data=[json.loads(s) for s in self.battle.read_text().splitlines()];data[0]['event']='fake';self.battle.write_text('\n'.join(json.dumps(x) for x in data)+'\n')
        with self.assertRaises(ValueError):self.e()

if __name__=='__main__':unittest.main()
