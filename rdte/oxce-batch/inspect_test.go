package main
import (
 "os"
 "path/filepath"
 "testing"
)
func TestEmptyFailClosed(t *testing.T){
 r:=batch(t.TempDir())
 if len(r.NativeScenarios)!=6||r.GameplayQualified||r.MultiplayerQualified||r.AccessibilityQualified{t.Fatal("false equivalence")}
 for _,s:=range r.NativeScenarios{if s.TargetObserved{t.Fatal("absent native trace accepted")}}
}
func TestMenuTraceAndErrorControl(t *testing.T){
 root:=t.TempDir();d:=filepath.Join(root,"boot")
 if e:=os.MkdirAll(d,0700);e!=nil{t.Fatal(e)}
 trace:=`{"class":"N8OpenXcom13MainMenuStateE","controls":[{"role":"button","label":"New Game"}]}`+string(byte(10))
 if e:=os.WriteFile(filepath.Join(d,"trace.jsonl"),[]byte(trace),0600);e!=nil{t.Fatal(e)}
 s:=analyze(root,"boot","MainMenuState")
 if !s.NativeEngineObserved||!s.TargetObserved||len(s.VisibleLabels)!=1{t.Fatalf("unexpected %+v",s)}
 s=analyze(root,"new-game","NewGameState");if s.TargetObserved{t.Fatal("unobserved new game accepted")}
 s=analyze(root,"new-battle-briefing-continue","InventoryState");if s.TargetObserved{t.Fatal("unobserved inventory accepted")}
}
func TestNoSyntheticScreenshotIsInvented(t *testing.T){
 s:=analyze(t.TempDir(),"boot","MainMenuState")
 if s.ScreenshotSHA256!=""{t.Fatal("invented screenshot")}
}


 // Live public run 41030217895 had MainMenuState in native trace and a real
 // screenshot, but zero emitted Control.Label fields. That is a label
 // observability hold, NOT evidence the menu failed to launch. Keep source
 // reachability and unqualified UI/accessibility scientifically separate.
 func TestBootNativeMenuWithoutControlLabels(t *testing.T) {
  root:=t.TempDir();d:=filepath.Join(root,"boot")
  if err:=os.MkdirAll(d,0700);err!=nil{t.Fatal(err)}
  trace:=`{"class":"N8OpenXcom13MainMenuStateE","controls":[]}`+"\n"
  if err:=os.WriteFile(filepath.Join(d,"trace.jsonl"),[]byte(trace),0600);err!=nil{t.Fatal(err)}
  s:=analyze(root,"boot","MainMenuState")
  if !s.NativeEngineObserved||!s.TargetObserved {t.Fatalf("lost genuine native menu reachability: %+v",s)}
  if s.UILabelTelemetryObserved||s.UILabelTelemetryGrade!="LABEL_TELEMETRY_MISSING_VISUAL_QA_HELD" {
   t.Fatalf("invented label or visual acceptance: %+v",s)
  }
  if s.ScreenshotSHA256!="" {t.Fatal("invented screenshot evidence")}
  r:=batch(root)
  if r.GameplayQualified||r.MultiplayerQualified||r.AccessibilityQualified {t.Fatal("promoted research to human/gameplay acceptance")}
 }
 func TestControlLabelsDoNotReplaceNativeMenuState(t *testing.T) {
  root:=t.TempDir();d:=filepath.Join(root,"boot")
  if err:=os.MkdirAll(d,0700);err!=nil{t.Fatal(err)}
  trace:=`{"class":"N8OpenXcom10StartStateE","controls":[{"role":"button","label":"New Game"}]}`+"\n"
  if err:=os.WriteFile(filepath.Join(d,"trace.jsonl"),[]byte(trace),0600);err!=nil{t.Fatal(err)}
  s:=analyze(root,"boot","MainMenuState")
  if s.NativeEngineObserved||s.TargetObserved {t.Fatalf("invented main menu from label telemetry: %+v",s)}
  if !s.UILabelTelemetryObserved||s.UILabelTelemetryGrade!="LABEL_TELEMETRY_PRESENT_NOT_VISUAL_QA" {t.Fatalf("lost control label signal: %+v",s)}
 }
 func TestNativeBootErrorStillFailsIndependently(t *testing.T) {
  root:=t.TempDir();d:=filepath.Join(root,"boot")
  if err:=os.MkdirAll(d,0700);err!=nil{t.Fatal(err)}
  trace:=`{"class":"N8OpenXcom13MainMenuStateE","controls":[{"role":"status","label":"ERROR: missing resource"}]}`+"\n"
  if err:=os.WriteFile(filepath.Join(d,"trace.jsonl"),[]byte(trace),0600);err!=nil{t.Fatal(err)}
  s:=analyze(root,"boot","MainMenuState")
  if !s.TargetObserved||!s.ErrorScreen {t.Fatalf("suppressed independent native error gate: %+v",s)}
 }
