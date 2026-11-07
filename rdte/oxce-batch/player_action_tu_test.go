package main

import (
 "encoding/json"
 "os"
 "strings"
 "testing"
)

func actionFrame(turn, side, playerX, playerTU, enemyX, enemyTU int) string {
 f := TacticalFrame{Event:"battle-snapshot", Turn:turn, Side:side, Units:[]TacticalUnit{
  {ID:1, Faction:0, X:playerX, Y:2, Z:0, TU:playerTU, HP:34},
  {ID:1000000, Faction:1, X:enemyX, Y:4, Z:0, TU:enemyTU, HP:24},
 }}
 b, _ := json.Marshal(f)
 return string(b)
}

func testSequence(a, b string) TacticalAssessment {
 stream := strings.Join([]string{
  a, b,
  actionFrame(1,1,4,48,4,48),
  actionFrame(1,1,4,48,5,42),
  actionFrame(2,0,4,60,5,42),
 }, "\n")
 return assessTacticalTelemetry([]byte(stream))
}

func TestPlayerPositionAndTUDecrementSameTurnAndUnit(t *testing.T) {
 a := testSequence(actionFrame(1,0,2,61,4,48), actionFrame(1,0,3,57,4,48))
 if a.FirstFailure!="" || !a.PlayerMoveObserved || !a.PlayerTUDecreased ||
   !a.PlayerMovementSpentTU || !a.EnemyMoved || a.GameplayQualified {
  t.Fatalf("real evidence contract rejected: %+v", a)
 }
}

func TestPlayerMovementWithoutTUDecrementCannotPass(t *testing.T) {
 a := testSequence(actionFrame(1,0,2,61,4,48), actionFrame(1,0,3,61,4,48))
 if !a.PlayerMoveObserved || a.PlayerTUDecreased || a.PlayerMovementSpentTU || a.GameplayQualified {
  t.Fatalf("no-TU teleport accepted as movement action: %+v", a)
 }
}

func TestPlayerTUDecrementWithoutMovementCannotPass(t *testing.T) {
 a := testSequence(actionFrame(1,0,2,61,4,48), actionFrame(1,0,2,56,4,48))
 if a.PlayerMoveObserved || !a.PlayerTUDecreased || a.PlayerMovementSpentTU || a.GameplayQualified {
  t.Fatalf("TU-only action misreported as player movement: %+v", a)
 }
}

func TestCrossTurnRepositionIsNotAPlayerAction(t *testing.T) {
 a := testSequence(actionFrame(0,0,2,61,4,48), actionFrame(1,0,3,57,4,48))
 if a.PlayerMoveObserved || a.PlayerTUDecreased || a.PlayerMovementSpentTU {
  t.Fatalf("turn-transition position delta impersonated player click: %+v", a)
 }
}

func TestFactionAndUnitSwitchCannotImpersonatePlayerAction(t *testing.T) {
 a := testSequence(actionFrame(1,0,2,61,4,48), actionFrame(1,0,2,61,5,42))
 if a.PlayerMoveObserved || a.PlayerTUDecreased || a.PlayerMovementSpentTU {
  t.Fatalf("enemy delta was misreported as player action: %+v", a)
 }
}

func TestAuthenticPublicNativeSnapshotsStayEvidenceBound(t *testing.T) {
 path := os.Getenv("SF_PUBLIC_NATIVE_BATTLE_JSONL")
 if path=="" {t.Skip("separate immutable public native artifact not supplied")}
 raw, err := os.ReadFile(path)
 if err!=nil {t.Fatal(err)}
 r := assessTacticalTelemetry(raw)
 if r.MalformedFrames>0 || r.FirstFailure!="" || !r.EnemyActionObserved || r.GameplayQualified ||
    (r.PlayerMovementSpentTU && (!r.PlayerMoveObserved || !r.PlayerTUDecreased)) {
  t.Fatalf("native source telemetry contract not satisfied: %+v", r)
 }
}
