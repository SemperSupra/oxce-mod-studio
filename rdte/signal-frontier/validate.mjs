import {readFileSync} from 'node:fs';
export function validate(g){
 const errors=[];const require=(c,m)=>{if(!c)errors.push(m);};
 require(g.schemaVersion===1,'version');require(g.originalContent===true,'provenance');require(g.engineCompatible===false,'unqualified-engine');
 const sets={};
 for(const cat of ['sectors','roles','enemies','equipment','crafts','research','manufacture','missions','maps']){
  const entries=g[cat]||[];const ids=entries.map(e=>e.id);
  require(ids.every(Boolean),cat+':missing-id');require(new Set(ids).size===ids.length,cat+':duplicate-id');sets[cat]=new Set(ids);
 }
 const objectives=new Set(g.objectives||[]);
 for(const m of g.missions||[]){
  require(sets.sectors.has(m.sector),m.id+':sector');
  require(sets.maps.has(m.map),m.id+':map');
  require(Number.isSafeInteger(m.seed),m.id+':seed');
  for(const id of m.team||[]) require(sets.roles.has(id),m.id+':team:'+id);
  for(const id of m.enemies||[])require(sets.enemies.has(id),m.id+':enemy:'+id);
  for(const id of m.objectives||[])require(objectives.has(id),m.id+':objective:'+id);
  require(objectives.has(m.alternateWin),m.id+':alternate-win');
  for(const id of [...(m.successNext||[]),...(m.failureNext||[])])require(sets.missions.has(id),m.id+':next:'+id);
 }
 for(const item of g.manufacture||[]){
  require(sets.research.has(item.requiresResearch),item.id+':requiresResearch');
  require(sets.equipment.has(item.output),item.id+':output');
 }
 for(const topic of g.research||[])for(const dep of topic.requires||[])require(sets.research.has(dep),topic.id+':research-link');
 const maps=new Map((g.missions||[]).map(m=>[m.id,m]));
 for(const condition of ['successNext','failureNext']){
  const visited=new Set(),pending=['silent-relay'];
  while(pending.length){
   const id=pending.shift();
   if(visited.has(id))continue;
   visited.add(id);for(const next of maps.get(id)?.[condition]||[])pending.push(next);
  }
  require(visited.has('quiet-array-choice'),'unreachable-final:'+condition);
 }
 for(const locale of ['en-US','de','th'])for(const key of ['title','begin','win'])
  require(Boolean(g.locale?.[locale]?.[key]),'locale:'+locale+':'+key);
 require(g.designTargetMinutes?.[0]>0 && g.designTargetMinutes[1]>=g.designTargetMinutes[0],'session-length');
 return errors;
}
export function readCampaign(file){return JSON.parse(readFileSync(file,'utf8'));}
