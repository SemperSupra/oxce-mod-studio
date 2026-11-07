package main
import (
 "encoding/json"
 "strings"
 "testing"
)
func frame(turn,side,enemyX,enemyTU int)[]byte{
 x:=TacticalFrame{Event:"battle-snapshot",Turn:turn,Side:side,Units:[]TacticalUnit{
  {ID:1,Faction:0,X:2,Y:2,Z:0,TU:61,HP:34},
  {ID:1000000,Faction:1,X:enemyX,Y:4,Z:0,TU:enemyTU,HP:24},
 }}
 v,_:=json.Marshal(x);return v
}
func TestActualNativeEnemyMoveRequiresDelta(t *testing.T){
 sample:=strings.Join([]string{string(frame(1,0,4,48)),string(frame(1,1,4,48)),string(frame(1,1,5,42)),string(frame(2,0,5,42))},"\n")
 r:=assessTacticalTelemetry([]byte(sample))
 if r.FirstFailure!=""||!r.EnemyActionObserved||!r.EnemyMoved||!r.EnemyTUDecreased||!r.ReturnedToPlayerTurn||r.GameplayQualified{t.Fatalf("bad classification %+v",r)}
}
func TestNoEnemyActionIsNeverInferredFromTurnOnly(t *testing.T){
 s:=strings.Join([]string{string(frame(1,0,4,48)),string(frame(1,1,4,48)),string(frame(2,0,4,48))},"\n")
 r:=assessTacticalTelemetry([]byte(s))
 if r.EnemyActionObserved||r.FirstFailure==""{t.Fatal("false AI action")}
}
func TestMissingCorruptOrDuplicateSnapshotsFailClosed(t *testing.T){
 for _,s:=range []string{"",string(frame(1,0,4,48)),"{bad json}\n"+string(frame(1,1,4,48)),string(frame(1,0,4,48))+"\n"+string(frame(1,0,4,48))}{
  r:=assessTacticalTelemetry([]byte(s))
  if r.EnemyActionObserved||r.FirstFailure==""{t.Fatalf("unqualified stream passed %+v",r)}
 }
}
