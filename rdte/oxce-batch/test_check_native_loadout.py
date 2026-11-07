import tempfile
from pathlib import Path
import yaml
import unittest
from check_native_loadout import classify

class NativeLoadoutTests(unittest.TestCase):
 def setUp(self):
  self.t=tempfile.TemporaryDirectory();self.addCleanup(self.t.cleanup)
  self.path=Path(self.t.name)/'_quick_.asav'
  self.units=[{'id':1,'faction':0},{'id':1000000,'faction':1}]
  self.items=[{'id':0,'type':'STR_RDTE_TOOL','owner':1},{'id':1,'type':'STR_RDTE_CELL','owner':1}]
  self.write()
 def write(self):
  self.path.write_text(yaml.safe_dump({'name':'_quick_.asav'})+'---\n'+yaml.safe_dump({'battleGame':{'units':self.units,'items':self.items}}))
 def test_positive_never_infers_shot_or_victory(self):
  r=classify(self.path);self.assertTrue(r['playerLoadoutObserved']);self.assertFalse(r['weaponActionVerified']);self.assertFalse(r['victoryQualified'])
 def test_hostile_only_tool_fails(self):
  self.items[0]['owner']=1000000;self.write();self.assertFalse(classify(self.path)['playerLoadoutObserved'])
 def test_missing_ammo_fails(self):
  self.items=self.items[:1];self.write();self.assertFalse(classify(self.path)['playerLoadoutObserved'])
 def test_hostile_only_ammo_fails(self):
  self.items[1]['owner']=1000000;self.write();self.assertFalse(classify(self.path)['playerLoadoutObserved'])
 def test_no_player_fails(self):
  self.units=self.units[1:];self.write()
  with self.assertRaisesRegex(ValueError,'no player'):classify(self.path)
 def test_duplicate_unit_refuses(self):
  self.units.append(self.units[0]);self.write()
  with self.assertRaisesRegex(ValueError,'duplicate'):classify(self.path)
 def test_unresolved_item_owner_refuses(self):
  self.items[0]['owner']=333;self.write()
  with self.assertRaisesRegex(ValueError,'unresolved'):classify(self.path)
 def test_symlink_refused(self):
  link=Path(self.t.name)/'alias.asav';link.symlink_to(self.path)
  with self.assertRaisesRegex(ValueError,'invalid original'):classify(link)
 def test_invalid_yaml_refused(self):
  self.path.write_text('{BROKEN YAML')
  with self.assertRaises(yaml.YAMLError):classify(self.path)
if __name__=='__main__':unittest.main()
