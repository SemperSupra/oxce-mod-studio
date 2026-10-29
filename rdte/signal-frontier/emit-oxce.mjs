import {readFileSync,mkdirSync,writeFileSync} from 'node:fs';
import {join,resolve} from 'node:path';
// Public original authored contract translator, intentionally incomplete until
// native OXCE 8.6.5 proof. Never infer gameplay or full schema support from output.
export function key(id){if(!/^[a-z][a-z0-9-]*$/.test(id))throw Error('unsafe-identifier:'+id);return 'STR_SF_'+id.toUpperCase().replaceAll('-','_');}
export function emitSubset(c){
 const rules=['# Signal Frontier original test content — engine semantics NOT YET VERIFIED'];
 const supported=new Set(['charged-cell','pulse-emitter']);
 rules.push('research:');
 for(const r of c.research){
  rules.push('  - name: '+key(r.id),'    cost: '+r.cost,'    points: '+r.cost);
  if(r.requires.length)rules.push('    dependencies: ['+r.requires.map(key).join(', ')+']');
 }
 rules.push('items:');
 rules.push('  - type: '+key('charged-cell'),'    battleType: 2','    clipSize: 12','    power: 9','    damageType: 1');
 rules.push('  - type: '+key('pulse-emitter'),'    battleType: 1','    compatibleAmmo: ['+key('charged-cell')+']','    accuracySnap: 55','    tuSnap: 30','    invWidth: 1','    invHeight: 2');
 rules.push('manufacture:');
 for(const m of c.manufacture){
  if(!supported.has(m.output))continue;
  rules.push('  - name: '+key('manufacture-'+m.id),'    requires: ['+key(m.requiresResearch)+']',
    '    producedItems:','      '+key(m.output)+': '+m.quantity);
 }
 const unmapped=c.equipment.filter(x=>!supported.has(x.id)).map(x=>x.id);
 const languages=['en-US:'];
 for(const e of [...c.research,...c.equipment])languages.push('  '+key(e.id)+': '+JSON.stringify((e.label||e.id).replaceAll('-',' ')));
 return {ruleset:rules.join('\n')+'\n',language:languages.join('\n')+'\n',report:{
  engineLoadVerified:false,playable:false,kind:'PARTIAL_RULESET_TRANSLATION',
  sourceEngine:'MeridianOXC/OpenXcom@668880bb756852f0734cfe92e438e400080a2f4d',
  unmappedEquipment:unmapped,excludedDomains:['countries','regions','crafts','startingBase','maps','AI','campaign-scripts'] }};
}
if(process.argv[1]&&import.meta.url.endsWith(process.argv[1])){
 const c=JSON.parse(readFileSync(new URL('./campaign.json',import.meta.url)));
 const data=emitSubset(c);const out=resolve(process.argv[2]||'.out/sf01-subset');
 mkdirSync(out,{recursive:true});
 writeFileSync(join(out,'signal-frontier.rul'),data.ruleset);
 writeFileSync(join(out,'en-US.yml'),data.language);
 writeFileSync(join(out,'translation-report.json'),JSON.stringify(data.report,null,2)+'\n');
 console.log(JSON.stringify({out,unmapped:data.report.unmappedEquipment.length,engineLoadVerified:false}));
}
