"""Source-level readiness only; requires native saved-inventory proof later."""
from pathlib import Path
import tempfile
import unittest
import yaml

ROOT=Path(__file__).resolve().parent


def inspect(path):
    data=yaml.safe_load(Path(path).read_text())
    if not isinstance(data,dict) or list(data)!=['items']:
        raise ValueError('only original synthetic item overrides admitted')
    items=data['items']
    if not isinstance(items,list) or len(items)!=2:
        raise ValueError('requires exactly two item overrides')
    expected={'STR_RDTE_TOOL','STR_RDTE_CELL'}
    if {i.get('type') for i in items if isinstance(i,dict)}!=expected:
        raise ValueError('wrong or duplicate item set')
    for item in items:
        if not isinstance(item,dict) or set(item)!={'type','bigSprite','floorSprite','handSprite'}:
            raise ValueError('unadmitted or missing item field')
        for key in ('bigSprite','floorSprite','handSprite'):
            if type(item[key]) is not int or not 0<=item[key]<=1:
                raise ValueError('unbound synthetic sprite reference')
    return {i['type']:i for i in items}


class Tests(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup)
        self.path=Path(self.tmp.name)/'overlay.rul'
        self.original=ROOT.joinpath('loadout-overlay.rul').read_text()
        self.path.write_text(self.original)
    def test_real_overlay_has_inventory_sprite_offset(self):
        r=inspect(self.path)
        self.assertEqual(set(r),{'STR_RDTE_TOOL','STR_RDTE_CELL'})
        self.assertEqual(r['STR_RDTE_TOOL']['bigSprite'],0)
        self.assertEqual(r['STR_RDTE_CELL']['bigSprite'],1)
    def test_negative_big_sprite_refused(self):
        self.path.write_text(self.original.replace('bigSprite: 0','bigSprite: -1'))
        with self.assertRaisesRegex(ValueError,'unbound'):inspect(self.path)
    def test_duplicate_item_refused(self):
        self.path.write_text(self.original.replace('type: STR_RDTE_CELL','type: STR_RDTE_TOOL'))
        with self.assertRaisesRegex(ValueError,'wrong or duplicate'):inspect(self.path)
    def test_unadmitted_category_refused(self):
        self.path.write_text(self.original+'units: []\n')
        with self.assertRaisesRegex(ValueError,'only original'):inspect(self.path)
    def test_item_extra_key_refused(self):
        self.path.write_text(self.original.replace('bigSprite: 0','bigSprite: 0\n    power: 1000'))
        with self.assertRaisesRegex(ValueError,'unadmitted'):inspect(self.path)
    def test_item_missing_field_refused(self):
        self.path.write_text(self.original.replace('    handSprite: 1\n',''))
        with self.assertRaisesRegex(ValueError,'unadmitted'):inspect(self.path)
    def test_boolean_sprite_refused(self):
        self.path.write_text(self.original.replace('bigSprite: 1','bigSprite: true'))
        with self.assertRaisesRegex(ValueError,'unbound'):inspect(self.path)

if __name__=='__main__':unittest.main()
