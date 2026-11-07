package main
import (
 "os"
 "strings"
 "testing"
)
// Native battle state begins at turn zero. This is a legal first frame, not
// malformed input; negative turns and corrupt frames remain rejected.
func TestNativeTurnZeroAndNegativeTurnControl(t *testing.T) {
 s:=strings.Join([]string{string(frame(0,0,4,48)),string(frame(1,1,4,48)),string(frame(1,1,5,42)),string(frame(2,0,5,42))},"\n")
 r:=assessTacticalTelemetry([]byte(s))
 if r.FirstFailure!=""||r.MalformedFrames!=0||r.NativeSnapshots!=4||!r.EnemyMoved||!r.EnemyTUDecreased||!r.ReturnedToPlayerTurn||r.GameplayQualified {
  t.Fatalf("valid native turn-zero was wrongly rejected: %+v",r)
 }
 bad:=assessTacticalTelemetry(frame(-1,0,4,48))
 if bad.MalformedFrames!=1||bad.FirstFailure!="malformed-native-snapshot:1"{t.Fatalf("negative native turn admitted: %+v",bad)}
}
// The public synthetic artifact's battle.jsonl can be supplied independently
// by a pinned $0 artifact replay; never read proprietary original media here.
func TestNativePublishedSyntheticReplayWhenAvailable(t *testing.T){
 path:=os.Getenv("SF_PUBLIC_NATIVE_BATTLE_JSONL")
 if path==""{t.Skip("no externally pinned public synthetic telemetry")}
 raw,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 r:=assessTacticalTelemetry(raw)
 if r.NativeSnapshots<10||r.MalformedFrames!=0||r.FirstFailure!=""||!r.PlayerObserved||!r.EnemyObserved||!r.EnemyMoved||!r.EnemyTUDecreased||!r.ReturnedToPlayerTurn||r.GameplayQualified{
  t.Fatalf("published public synthetic native action replay failed: %+v",r)
 }
}
