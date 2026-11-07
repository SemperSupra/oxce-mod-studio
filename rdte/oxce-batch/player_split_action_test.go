package main
import (
 "encoding/json"
 "os"
 "strings"
 "testing"
)
func movementFrame(turn, side, unitID, x, tu int) string {
 f:=TacticalFrame{Event:"battle-snapshot", Turn:turn, Side:side, Units:[]TacticalUnit{
  {ID:unitID,Faction:0,X:x,Y:3,Z:0,TU:tu,HP:34},
  {ID:1000000,Faction:1,X:5,Y:5,Z:0,TU:48,HP:24},
 }}
 b,_:=json.Marshal(f);return string(b)
}
func replaySplit(frames []string) TacticalAssessment {
 // Appended genuine hostile action supplies the separate enemy-action contract.
 frames=append(frames, movementFrame(1,1,1,3,60),
   actionFrame(1,1,3,60,5,42), actionFrame(2,0,3,64,5,42))
 return assessTacticalTelemetry([]byte(strings.Join(frames,"\n")))
}
func TestSplitWalkTUThenTilePositionAdjacentSameTurn(t *testing.T){
 a:=replaySplit([]string{movementFrame(1,0,1,3,64),movementFrame(1,0,1,3,60),movementFrame(1,0,1,4,60)})
 if !a.PlayerMoveObserved||!a.PlayerTUDecreased||!a.PlayerMovementSpentTU||a.GameplayQualified {t.Fatalf("native split move not joined: %+v",a)}
}
func TestStaleSpendBeforeUnchangedFrameCannotBeReused(t *testing.T){
 a:=replaySplit([]string{movementFrame(1,0,1,3,64),movementFrame(1,0,1,3,60),movementFrame(1,0,1,3,60),movementFrame(1,0,1,4,60)})
 if a.PlayerMovementSpentTU {t.Fatalf("stale TU spend was misattributed: %+v",a)}
}
func TestTurnBoundaryNeverJoinsSpendToMove(t *testing.T){
 a:=replaySplit([]string{movementFrame(0,0,1,3,64),movementFrame(0,0,1,3,60),movementFrame(1,0,1,4,60)})
 if a.PlayerMovementSpentTU {t.Fatalf("cross-turn TU spend was misattributed: %+v",a)}
}
func TestFactionBoundaryNeverJoinsSpendToMove(t *testing.T){
 a:=replaySplit([]string{movementFrame(1,0,1,3,64),movementFrame(1,0,1,3,60),movementFrame(1,1,1,4,60)})
 if a.PlayerMovementSpentTU {t.Fatalf("cross-side TU spend was misattributed: %+v",a)}
}
func TestSeparatePlayerUnitsCannotShareSpend(t *testing.T){
 a:=replaySplit([]string{movementFrame(1,0,1,3,64),movementFrame(1,0,1,3,60),movementFrame(1,0,2,4,60)})
 if a.PlayerMovementSpentTU {t.Fatalf("cross-unit TU spend was misattributed: %+v",a)}
}
func TestNativePublishedPlayerActionSpendsBeforeMovement(t *testing.T){
 path:=os.Getenv("SF_PUBLIC_NATIVE_PLAYER_JSONL")
 if path=="" {t.Skip("pinned original-only native player telemetry not provided")}
 raw,err:=os.ReadFile(path)
 if err!=nil{t.Fatal(err)}
 r:=assessTacticalTelemetry(raw)
 if r.NativeSnapshots!=30||r.MalformedFrames!=0||r.FirstFailure!=""||
   !r.PlayerMoveObserved||!r.PlayerTUDecreased||!r.PlayerMovementSpentTU||
   !r.EnemyActionObserved||r.GameplayQualified {
  t.Fatalf("actual B65 native click evidence was not qualified within bounded scope: %+v",r)
 }
}
