package main
import (
 "os"
 "path/filepath"
 "testing"
)
func TestEmptyFailClosed(t *testing.T){
 r:=batch(t.TempDir())
 if len(r.NativeScenarios)!=3||r.GameplayQualified||r.MultiplayerQualified||r.AccessibilityQualified{t.Fatal("false equivalence")}
 for _,s:=range r.NativeScenarios{if s.TargetObserved{t.Fatal("absent native trace accepted")}}
}
func TestMenuTraceAndErrorControl(t *testing.T){
 root:=t.TempDir();d:=filepath.Join(root,"boot")
 if e:=os.MkdirAll(d,0700);e!=nil{t.Fatal(e)}
 trace:=`{"class":"N8OpenXcom13MainMenuStateE","controls":[{"role":"button","label":"New Game"}]}`+"
"
 if e:=os.WriteFile(filepath.Join(d,"trace.jsonl"),[]byte(trace),0600);e!=nil{t.Fatal(e)}
 s:=analyze(root,"boot","MainMenuState")
 if !s.NativeEngineObserved||!s.TargetObserved||len(s.VisibleLabels)!=1{t.Fatalf("unexpected %+v",s)}
 s=analyze(root,"new-game","NewGameState");if s.TargetObserved{t.Fatal("unobserved new game accepted")}
}
func TestNoSyntheticScreenshotIsInvented(t *testing.T){
 s:=analyze(t.TempDir(),"boot","MainMenuState")
 if s.ScreenshotSHA256!=""{t.Fatal("invented screenshot")}
}
