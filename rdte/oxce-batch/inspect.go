package main
import (
 "bufio"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "strings"
)
type Control struct {Role string `json:"role"`; Label string `json:"label"`}
type Frame struct {Class string `json:"class"`; Controls []Control `json:"controls"`}
type Scenario struct {
 Name string `json:"name"`
 NativeEngineObserved bool `json:"nativeEngineObserved"`
 TargetObserved bool `json:"targetObserved"`
 ScreenClasses []string `json:"screenClasses"`
 VisibleLabels []string `json:"visibleLabels"`
 ErrorScreen bool `json:"errorScreen"`
 FirstFatal string `json:"firstFatal,omitempty"`
 ScreenshotSHA256 string `json:"screenshotSHA256,omitempty"`
}
type Receipt struct {
 SchemaVersion int `json:"schemaVersion"`
 EvidenceClass string `json:"evidenceClass"`
 EngineBaseline string `json:"engineBaseline"`
 NativeScenarios []Scenario `json:"nativeScenarios"`
 GameplayQualified bool `json:"gameplayQualified"`
 MultiplayerQualified bool `json:"multiplayerQualified"`
 AccessibilityQualified bool `json:"accessibilityQualified"`
}
func unique(vals []string)[]string{
 out:=[]string{};seen:=map[string]bool{}
 for _,v:=range vals{if !seen[v]{seen[v]=true;out=append(out,v)}}
 return out
}
func analyze(root,name,target string)Scenario{
 s:=Scenario{Name:name,ScreenClasses:[]string{},VisibleLabels:[]string{}}
 folder:=filepath.Join(root,name)
 if f,e:=os.Open(filepath.Join(folder,"trace.jsonl"));e==nil{
  scanner:=bufio.NewScanner(f);scanner.Buffer(make([]byte,4096),1024*1024)
  for scanner.Scan(){
   var x Frame
   if json.Unmarshal(scanner.Bytes(),&x)!=nil{continue}
   if x.Class!=""{s.ScreenClasses=append(s.ScreenClasses,x.Class)}
   for _,c:=range x.Controls{
    if c.Label!=""{s.VisibleLabels=append(s.VisibleLabels,c.Label)}
    if strings.Contains(c.Label,"ERROR:")||strings.Contains(c.Label,"Video intro not found"){s.ErrorScreen=true}
   }
  }
  f.Close()
 }
 s.ScreenClasses=unique(s.ScreenClasses)
 s.VisibleLabels=unique(s.VisibleLabels)
 for _,v:=range s.ScreenClasses{
  if strings.Contains(v,"MainMenuState"){s.NativeEngineObserved=true}
  if strings.Contains(v,target){s.TargetObserved=true}
 }
 if name=="boot"{s.TargetObserved=s.NativeEngineObserved && len(s.VisibleLabels)>0}
 if f,e:=os.Open(filepath.Join(folder,"user/openxcom.log"));e==nil{
  scanner:=bufio.NewScanner(f)
  for scanner.Scan(){line:=scanner.Text();if strings.Contains(line,"[FATAL]")&&s.FirstFatal==""{s.FirstFatal=line}}
  f.Close()
 }
 if b,e:=os.ReadFile(filepath.Join(folder,"screen.png"));e==nil{
  h:=sha256.Sum256(b);s.ScreenshotSHA256=hex.EncodeToString(h[:])
 }
 return s
}
func batch(root string)Receipt{
 tasks:=[]struct{name,target string}{{"boot","MainMenuState"},{"new-game","NewGameState"},{"new-battle","NewBattleState"},{"new-battle-confirm","BriefingState"},{"new-battle-briefing-continue","InventoryState"},{"new-game-confirm","BuildNewBaseState"}}
 r:=Receipt{SchemaVersion:1,EvidenceClass:"ACTUAL_MODIFIED_OXCE_BATCHED_MENU_TRANSITIONS_NOT_GAMEPLAY",EngineBaseline:"MeridianOXC/OpenXcom@668880bb756852f0734cfe92e438e400080a2f4d",NativeScenarios:[]Scenario{}}
 for _,task:=range tasks{r.NativeScenarios=append(r.NativeScenarios,analyze(root,task.name,task.target))}
 return r
}
func main(){
 if len(os.Args)==4&&os.Args[1]=="--tactical"{
  raw,e:=os.ReadFile(os.Args[2]);receipt:=assessTacticalTelemetry(raw)
  if e!=nil{receipt.FirstFailure="native-telemetry-unavailable: "+e.Error()}
  out,err:=json.MarshalIndent(receipt,"","  ");if err!=nil{panic(err)}
  if err=os.WriteFile(os.Args[3],append(out,byte(10)),0644);err!=nil{panic(err)}
  fmt.Println(string(out));if receipt.FirstFailure!=""{os.Exit(1)};return
 }
 if len(os.Args)!=3{fmt.Fprintln(os.Stderr,"usage: inspect <root> <out.json>");os.Exit(2)}
 r:=batch(os.Args[1]);b,e:=json.MarshalIndent(r,"","  ");if e!=nil{panic(e)}
 if e=os.WriteFile(os.Args[2],append(b,byte(10)),0644);e!=nil{panic(e)}
 fmt.Println(string(b))
 for _,x:=range r.NativeScenarios{if !x.TargetObserved||x.ErrorScreen||x.FirstFatal!=""{fmt.Fprintln(os.Stderr,"FAILED gate",x.Name);os.Exit(1)}}
}
