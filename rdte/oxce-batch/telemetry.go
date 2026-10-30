package main
import (
 "bufio"
 "bytes"
 "encoding/json"
 "fmt"
)
type TacticalUnit struct{ID int `json:"id"`;Faction int `json:"faction"`;X int `json:"x"`;Y int `json:"y"`;Z int `json:"z"`;TU int `json:"tu"`;HP int `json:"hp"`;Status int `json:"status"`}
type TacticalFrame struct{Event string `json:"event"`;Turn int `json:"turn"`;Side int `json:"side"`;Units []TacticalUnit `json:"units"`}
type TacticalAssessment struct{
 SchemaVersion int
 EvidenceClass string
 NativeSnapshots int
 PlayerObserved bool
 EnemyObserved bool
 PlayerMoveObserved bool
 EnemyActionObserved bool
 EnemyMoved bool
 EnemyTUDecreased bool
 HostileFactionTurnObserved bool
 ReturnedToPlayerTurn bool
 MalformedFrames int
 FirstFailure string
 GameplayQualified bool
}
func parseTacticalFrames(raw []byte)([]TacticalFrame,int){
 frames:=[]TacticalFrame{};invalid:=0
 scanner:=bufio.NewScanner(bytes.NewReader(raw));scanner.Buffer(make([]byte,4096),4*1024*1024)
 for scanner.Scan(){
  var f TacticalFrame
  if json.Unmarshal(scanner.Bytes(),&f)!=nil||f.Event!="battle-snapshot"||f.Turn<1||len(f.Units)==0{invalid++;continue}
  seen:=map[int]bool{};valid:=true
  for _,u:=range f.Units{if seen[u.ID]||u.HP<0||u.TU<0{valid=false;break};seen[u.ID]=true}
  if valid{frames=append(frames,f)}else{invalid++}
 }
 if scanner.Err()!=nil{invalid++}
 return frames,invalid
}
func assessTacticalTelemetry(raw []byte)TacticalAssessment{
 frames,bad:=parseTacticalFrames(raw)
 out:=TacticalAssessment{SchemaVersion:1,EvidenceClass:"NATIVE_READ_ONLY_TACTICAL_UNIT_DELTA_NOT_RELEASE_OR_ACCESSIBILITY_ACCEPTANCE",NativeSnapshots:len(frames),MalformedFrames:bad,GameplayQualified:false}
 if bad>0{out.FirstFailure=fmt.Sprintf("malformed-native-snapshot:%d",bad);return out}
 if len(frames)<2{out.FirstFailure="insufficient-real-native-snapshots";return out}
 var prev *TacticalFrame
 sawEnemySide:=false
 for i:=range frames{
  f:=&frames[i]
  if f.Side==1{sawEnemySide=true;out.HostileFactionTurnObserved=true}
  if sawEnemySide&&f.Side==0&&f.Turn>=2{out.ReturnedToPlayerTurn=true}
  for _,u:=range f.Units{if u.Faction==0{out.PlayerObserved=true};if u.Faction==1{out.EnemyObserved=true}}
  if prev!=nil&&prev.Side==f.Side{
   before:=map[int]TacticalUnit{};for _,u:=range prev.Units{before[u.ID]=u}
   for _,u:=range f.Units{
    was,ok:=before[u.ID];if !ok||was.Faction!=u.Faction{continue}
    moved:=was.X!=u.X||was.Y!=u.Y||was.Z!=u.Z
    spent:=u.TU<was.TU
    if f.Side==0&&u.Faction==0&&moved{out.PlayerMoveObserved=true}
    if f.Side==1&&u.Faction==1&&(moved||spent){
     out.EnemyActionObserved=true
     if moved{out.EnemyMoved=true}
     if spent{out.EnemyTUDecreased=true}
    }
   }
  }
  prev=f
 }
 if !out.PlayerObserved||!out.EnemyObserved{out.FirstFailure="missing-opposing-factions";return out}
 if !out.HostileFactionTurnObserved{out.FirstFailure="hostile-side-never-observed";return out}
 if !out.EnemyActionObserved{out.FirstFailure="no-native-enemy-position-or-tu-delta";return out}
 return out
}
