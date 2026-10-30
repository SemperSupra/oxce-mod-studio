#!/usr/bin/env node
// Independent opt-in read-only native tactical telemetry for the pinned GPL
// OXCE Game.cpp, applied after patch-engine.mjs. Zero sockets/actions here.
// Derived from public SavedBattleGame/BattleUnit APIs. Never write identities.
import {readFileSync,writeFileSync} from 'node:fs';
const path=process.argv[2];if(!path)throw Error('Game.cpp path required');
let s=readFileSync(path,'utf8');
const head='#include "../Savegame/SavedBattleGame.h"\n',space='namespace OpenXcom\n{\n';
const call='if(!_states.empty())sfRdteState(typeid(*_states.back()).name());';
for(const a of [head,space,call])if(!s.includes(a))throw Error('pinned telemetry anchor missing: '+a);
if(s.includes('SF_RDTE_TACTICAL_BEGIN'))throw Error('duplicate tactical telemetry');
s=s.replace(head,head+'#include "../Savegame/BattleUnit.h"\n');
const impl=`// SF_RDTE_TACTICAL_BEGIN
// A bounded, opt-in state observation independent of the semantics or
// accessibility UI. No changes to battle state; numeric data only.
static void sfRdteBattle(const Game* game)
{
 const char* target=std::getenv("OXCE_RDTE_BATTLE");
 if(!target || !*target || !game)return;
 SavedGame* saved=game->getSavedGame();
 if(!saved)return;
 SavedBattleGame* battle=saved->getSavedBattle();
 if(!battle)return;
 const std::vector<BattleUnit*> *units=battle->getUnits();
 if(!units)return;
 std::ostringstream stamp,record;
 stamp<<battle->getTurn()<<"|"<<static_cast<int>(battle->getSide());
 record<<"{\\"event\\":\\"battle-snapshot\\",\\"turn\\":"<<battle->getTurn()
       <<",\\"side\\":"<<static_cast<int>(battle->getSide())<<",\\"units\\":[";
 bool first=true;
 for(const BattleUnit* unit:*units)
 {
  if(!unit)continue;
  Position p=unit->getPosition();
  stamp<<"|"<<unit->getId()<<":"<<static_cast<int>(unit->getFaction())
       <<":"<<p.x<<","<<p.y<<","<<p.z
       <<":"<<unit->getTimeUnits()<<":"<<unit->getHealth()
       <<":"<<static_cast<int>(unit->getStatus());
  if(!first)record<<",";
  first=false;
  record<<"{\\"id\\":"<<unit->getId()
        <<",\\"faction\\":"<<static_cast<int>(unit->getFaction())
        <<",\\"x\\":"<<p.x<<",\\"y\\":"<<p.y<<",\\"z\\":"<<p.z
        <<",\\"tu\\":"<<unit->getTimeUnits()
        <<",\\"hp\\":"<<unit->getHealth()
        <<",\\"status\\":"<<static_cast<int>(unit->getStatus())<<"}";
 }
 record<<"]}";
 static std::string last;static unsigned emitted=0;
 if(last==stamp.str() || emitted>=1024)return;
 last=stamp.str();++emitted;
 std::ofstream out(target,std::ios::app);
 if(out)out<<record.str()<<std::endl;
}
// SF_RDTE_TACTICAL_END
`;
s=s.replace(space,space+impl);
s=s.replace(call,call+'\n\t\tsfRdteBattle(this);');
writeFileSync(path,s);
console.log('SF_RDTE_TACTICAL_READ_ONLY_PATCH_APPLIED');
