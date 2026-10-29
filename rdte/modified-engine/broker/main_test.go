package main
import("bytes";"encoding/json";"net/http/httptest";"os";"path/filepath";"testing")
func TestLocalOnlyControlledInputAndSeat(t *testing.T){
 d:=t.TempDir();b:=&Broker{mailbox:filepath.Join(d,"commands"),trace:filepath.Join(d,"trace"),token:"local-secret",activeSeat:"human"}
 h:=b.Handler([]byte("<main>hello</main>"))
 post:=func(path string,v interface{},auth string)*httptest.ResponseRecorder{
  var body []byte;if s,ok:=v.(string);ok{body=[]byte(s)}else{body,_=json.Marshal(v)}
  req:=httptest.NewRequest("POST",path,bytes.NewReader(body));req.Header.Set("Authorization",auth)
  rr:=httptest.NewRecorder();h.ServeHTTP(rr,req);return rr
 }
 a:=Action{Seat:"human",Kind:"key",Key:"enter",Expected:0}
 if rr:=post("/api/action",a,"");rr.Code!=401{t.Fatal("missing auth",rr.Code)}
 if rr:=post("/api/action",a,"Bearer local-secret");rr.Code!=200{t.Fatal("valid action",rr.Body)}
 raw,_:=os.ReadFile(b.mailbox);if string(raw)!="KEY 13\n"{t.Fatal("mailbox",string(raw))}
 if rr:=post("/api/action",a,"Bearer local-secret");rr.Code!=409{t.Fatal("stale was accepted")}
 if rr:=post("/api/seat","agent","Bearer local-secret");rr.Code!=200{t.Fatal("seat switch")}
 if rr:=post("/api/action",Action{Seat:"human",Kind:"key",Key:"enter",Expected:2},"Bearer local-secret");rr.Code!=409{t.Fatal("nonseat accepted")}
 if rr:=post("/api/action",Action{Seat:"agent",Kind:"key",Key:"f4",Expected:2},"Bearer local-secret");rr.Code!=200{t.Fatal("agent action",rr.Body)}
 raw,_=os.ReadFile(b.mailbox);if string(raw)!="KEY 13\nKEY 285\n"{t.Fatal("wrong transcript",string(raw))}
}
func TestInspectReadOnlyAndUnqualified(t *testing.T){
 d:=t.TempDir();trace:=filepath.Join(d,"trace")
 os.WriteFile(trace,[]byte("{\"event\":\"screen-class\",\"class\":\"StartState\"}\n"),0600)
 b:=&Broker{mailbox:filepath.Join(d,"input"),trace:trace,token:"t",activeSeat:"human"}
 h:=b.Handler(nil);req:=httptest.NewRequest("GET","/api/state",nil);rr:=httptest.NewRecorder();h.ServeHTTP(rr,req)
 var v Snapshot;if err:=json.Unmarshal(rr.Body.Bytes(),&v);err!=nil{t.Fatal(err)}
 if v.ScreenClass!="StartState"||v.Multiplayer||v.Revision!=0{t.Fatalf("false engine or multiplayer acceptance: %+v",v)}
}

func TestNativeAccessibleTextAndControlModel(t *testing.T){
 d:=t.TempDir()
 trace:=filepath.Join(d,"trace")
 os.WriteFile(trace,[]byte("{\"event\":\"screen\",\"class\":\"MainMenuState\",\"scaleX\":2,\"scaleY\":2,\"controls\":[{\"role\":\"button\",\"label\":\"New game\",\"x\":10,\"y\":20,\"w\":60,\"h\":15},{\"role\":\"text\",\"label\":\"Signal Frontier\",\"x\":10,\"y\":3,\"w\":120,\"h\":15}]}\n"),0600)
 b:=&Broker{mailbox:filepath.Join(d,"commands"),trace:trace,token:"secret"}
 got:=b.view()
 if got.ScreenClass!="MainMenuState"||len(got.Controls)!=2||got.Controls[0].Label!="New game"||got.ScaleX!=2{t.Fatalf("wrong semantics %+v",got)}
 if got.Multiplayer{t.Fatal("local UI falsely asserts real multiplayer")}
}
