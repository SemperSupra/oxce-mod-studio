package audience
import(
 "strings"
 "testing"
)
func TestAudienceParity(t *testing.T){
 for _,who:=range []Audience{Human,Automation,Agent}{
  m:=New(4701)
  for i,op:=range []string{"scan-relay","end-turn","recover-recorder"}{
   out,err:=m.Apply(Action{Audience:who,RequestID:"r",Operation:op,ExpectedSeq:uint64(i)})
   if err!=nil{t.Fatalf("%s/%s: %v",who,op,err)}
   if out.EngineExecuted||out.EvidenceClass!="CONTRACT_SIMULATOR_NOT_OXCE"{t.Fatal("false engine claim")}
   if i==2&&!out.State.Recovered{t.Fatal("recovery missing")}
  }
 }
}
func TestReplayStable(t *testing.T){
 a,b:=New(4701),New(4701)
 for i,op:=range []string{"scan-relay","end-turn","recover-recorder"}{
  req:=Action{Audience:Agent,RequestID:"fixed",Operation:op,ExpectedSeq:uint64(i)}
  x,err:=a.Apply(req);if err!=nil{t.Fatal(err)}
  y,err:=b.Apply(req);if err!=nil{t.Fatal(err)}
  if x.SHA256!=y.SHA256{t.Fatal("non-deterministic state")}
 }
}
func TestStaleAndInvalidFailClosed(t *testing.T){
 m:=New(4701);prev:=m.Snapshot("")
 tests:=[]Action{
  {Audience:"anonymous",RequestID:"r",Operation:"scan-relay",ExpectedSeq:0},
  {Audience:Agent,RequestID:"",Operation:"scan-relay",ExpectedSeq:0},
  {Audience:Agent,RequestID:"r",Operation:"scan-relay",ExpectedSeq:1},
  {Audience:Agent,RequestID:"r",Operation:"recover-recorder",ExpectedSeq:0},
  {Audience:Agent,RequestID:"r",Operation:"fire-weapon",ExpectedSeq:0},
 }
 for _,req:=range tests{
  _,err:=m.Apply(req);if err==nil{t.Fatalf("accepted %+v",req)}
  if m.Snapshot("").SHA256!=prev.SHA256{t.Fatal("mutated on rejected action")}
 }
}
func TestReadOnlySnapshot(t *testing.T){
 m:=New(5)
 a:=m.Snapshot("a");b:=m.Snapshot("b")
 if a.SHA256!=b.SHA256||m.state.Sequence!=0{t.Fatal("inspect mutated state")}
 if !strings.Contains(a.EvidenceClass,"NOT_OXCE"){t.Fatal("missing evidence boundary")}
}
