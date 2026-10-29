// Source-audited original Go inspection of the real co-op TestServer.
package main
import(
 "bufio"
 "encoding/json"
 "fmt"
 "net"
 "os"
 "path/filepath"
 "strconv"
 "strings"
 "time"
)
type Result struct{
 SchemaVersion int
 EvidenceClass string
 RequestedMode string
 RealMultiplayerQualified bool
 HostNewGameReached bool
 HostMenuReached bool
 HostTcpAccepted bool
 ClientJoinAccepted bool
 ClientLobbyReached bool
 HostLobbyReached bool
 JoinPopupsDismissed bool
 StartEligible bool
 CampaignStartClicked bool
 HostBasePlacementReached bool
 HostBaseNamed bool
 HostSavedBaseVisible bool
 ClientSeesHostBase bool
 ClientBaseNamed bool
 ClientSavedBaseVisible bool
 HostWorldObserved bool
 ClientWorldObserved bool
 HostGeo map[string]interface{}
 ClientGeo map[string]interface{}
 WorldClockEqual bool
 CampaignMode string
 CampaignModeConsistent bool
 ObservedSubsetEqual bool
 HostStrategicReady bool
 ClientStrategicReady bool
 ClientObservation string
 HostStates []string
 ClientStates []string
 FirstFailure string
 Commands []string
}
// The pinned co-op TestServer reports campaignType 0=Separate, 1=Shared.
// A separate-campaign base mismatch is not evidence of SHARED replication failure.
func campaignMode(host,client map[string]interface{}) (string,bool) {
  h,hOK:=host["campaignType"].(float64)
  c,cOK:=client["campaignType"].(float64)
  if !hOK||!cOK||h!=c{return "UNKNOWN",false}
  if h==0{return "SEPARATE",true}
  if h==1{return "SHARED",true}
  return "UNKNOWN",false
}
type Client struct{name string; conn net.Conn; reader *bufio.Reader}
func dial(path,name string)(*Client,error){
 b,e:=os.ReadFile(filepath.Join(path,"testserver_port.txt"));if e!=nil{return nil,e}
 p,e:=strconv.Atoi(strings.TrimSpace(string(b)));if e!=nil||p<1||p>65535{return nil,fmt.Errorf("invalid test port")}
 c,e:=net.DialTimeout("tcp",fmt.Sprintf("127.0.0.1:%d",p),2*time.Second);if e!=nil{return nil,e}
 return &Client{name:name,conn:c,reader:bufio.NewReader(c)},nil
}
func(c *Client) run(req map[string]interface{})(map[string]interface{},error){
 c.conn.SetDeadline(time.Now().Add(4*time.Second))
 b,_:=json.Marshal(req)
 if _,e:=c.conn.Write(append(b,'\n'));e!=nil{return nil,e}
 s,e:=c.reader.ReadString('\n');if e!=nil{return nil,e}
 var v map[string]interface{};if e=json.Unmarshal([]byte(s),&v);e!=nil{return nil,e}
 if v["ok"]!=true{return v,fmt.Errorf("command %v refused: %v",req["cmd"],v["error"])}
 return v,nil
}
func(c *Client) states()([]string,error){
 r,e:=c.run(map[string]interface{}{"cmd":"get_state"});if e!=nil{return nil,e}
 out:=[]string{}
 raw,ok:=r["states"].([]interface{});if !ok{return nil,fmt.Errorf("missing states array")}
 for _,v:=range raw{if name,ok:=v.(string);ok{out=append(out,name)}}
 return out,nil
}
func(c *Client) waitState(expected string)([]string,error){
 until:=time.Now().Add(12*time.Second)
 for time.Now().Before(until){
  v,e:=c.states();if e!=nil{return v,e}
  for _,s:=range v{if strings.Contains(s,expected){return v,nil}}
  time.Sleep(150*time.Millisecond)
 }
 latest,e:=c.states();if e!=nil{return latest,e}
 return latest,fmt.Errorf("timeout waiting for %s; states=%v",expected,latest)
}
func main(){
 if len(os.Args)!=4{fmt.Fprintln(os.Stderr,"usage: session_probe <host-user> <client-user> <out.json>");os.Exit(2)}
 mode:=os.Getenv("SF_COOP_MODE")
 if mode=="" {mode="separate"}
 r:=Result{SchemaVersion:1,EvidenceClass:"REAL_COOP_ENGINE_SESSION_PROGRESS_NOT_ACCEPTANCE",RequestedMode:mode,Commands:[]string{}}
 defer func(){out,_:=json.MarshalIndent(r,"","  ");os.WriteFile(os.Args[3],append(out,'\n'),0644);fmt.Println(string(out))}()
 if mode!="separate"&&mode!="shared" {r.FirstFailure="invalid-coop-mode";return}
 host,e:=dial(os.Args[1],"host");if e!=nil{r.FirstFailure="dial-host: "+e.Error();return};defer host.conn.Close()
 client,e:=dial(os.Args[2],"client");if e!=nil{r.FirstFailure="dial-client: "+e.Error();return};defer client.conn.Close()
 cmd:=func(c *Client,v map[string]interface{})(map[string]interface{},bool){
  r.Commands=append(r.Commands,c.name+":"+fmt.Sprint(v["cmd"]))
  out,e:=c.run(v);if e!=nil{r.FirstFailure=c.name+":"+e.Error();return out,false};return out,true
 }
 gameMode:="coop_separate"
 if mode=="shared" {gameMode="coop_shared"}
 if _,ok:=cmd(host,map[string]interface{}{"cmd":"open_new_game","mode":gameMode});!ok{return}
 r.HostStates,e=host.waitState("NewGameState");if e!=nil{r.FirstFailure="host-new-game-state: "+e.Error();return}
 r.HostNewGameReached=true
 if _,ok:=cmd(host,map[string]interface{}{"cmd":"newgame_ok"});!ok{return}
 r.HostStates,e=host.waitState("HostMenu");if e!=nil{r.FirstFailure="host-menu: "+e.Error();return}
 r.HostMenuReached=true
 if _,ok:=cmd(host,map[string]interface{}{"cmd":"host_tcp","server":"SignalFrontierRDTE","port":"47900","player":"HostPlayer"});!ok{return}
 r.HostTcpAccepted=true
 if _,ok:=cmd(client,map[string]interface{}{"cmd":"join_tcp","ip":"127.0.0.1","port":"47900","player":"ClientPlayer"});!ok{return}
 r.ClientJoinAccepted=true
 r.ClientStates,e=client.waitState("LobbyMenu");if e!=nil{r.FirstFailure="client-lobby: "+e.Error();return}
 r.ClientLobbyReached=true
 // The actual upstream co-op state machine raises a Profile notification
 // on both clients; close via the same real Profile::buttonOK handler players use.
 for _,c:=range []*Client{host,client}{
   if _,e:=c.waitState("Profile");e!=nil{r.FirstFailure=c.name+"-join-profile: "+e.Error();return}
   if _,ok:=cmd(c,map[string]interface{}{"cmd":"profile_ok"});!ok{return}
 }
 r.JoinPopupsDismissed=true
 r.HostStates,e=host.waitState("LobbyMenu")
 if e!=nil{r.FirstFailure="host-lobby: "+e.Error();return}
 r.HostLobbyReached=true
 var eligible bool
 deadline:=time.Now().Add(10*time.Second)
 for time.Now().Before(deadline){
   v,ok:=cmd(host,map[string]interface{}{"cmd":"lobby_state"});if !ok{return}
   if v["startEligible"]==true{eligible=true;break}
   time.Sleep(200*time.Millisecond)
 }
 if !eligible{r.FirstFailure="host-lobby-not-start-eligible";return}
 r.StartEligible=true
 if _,ok:=cmd(host,map[string]interface{}{"cmd":"lobby_action"});!ok{return}
 if _,e:=host.waitState("ConfirmStartCampaignState");e!=nil{r.FirstFailure="host-start-confirm: "+e.Error();return}
 v,ok:=cmd(host,map[string]interface{}{"cmd":"lobby_confirm_ok"});if !ok{return}
 if v["clicked"]!=true{r.FirstFailure="host-start-confirm-not-clicked";return}
 r.CampaignStartClicked=true
 r.HostStates,e=host.waitState("BuildNewBaseState");if e!=nil{r.FirstFailure="host-base-placement: "+e.Error();return}
 r.HostBasePlacementReached=true
 // OXCE BuildNewBaseState::placeAt uses original globe polygons; test 2 deg E, 0 lat
 // against our independently authored triangle, never an original retail map.
 if _,ok:=cmd(host,map[string]interface{}{"cmd":"place_first_base","lon":0.03490658503988659,"lat":0.0,"name":"Signal Ridge"});!ok{return}
 r.HostBaseNamed=true
 hostBases,ok:=cmd(host,map[string]interface{}{"cmd":"get_soldiers"});if !ok{return}
 list,hasList:=hostBases["bases"].([]interface{})
 if !hasList{r.FirstFailure="host-base-list-missing";return}
 for _,item:=range list{
  record,isObj:=item.(map[string]interface{})
  if isObj&&record["name"]=="Signal Ridge"{r.HostSavedBaseVisible=true;break}
 }
 if !r.HostSavedBaseVisible{r.FirstFailure="host-base-not-visible-in-real-save";return}
 r.HostStates,e=host.states();if e!=nil{r.FirstFailure="host-state-after-base: "+e.Error();return}
 r.ClientStates,e=client.states();if e!=nil{r.ClientObservation="client-state-unavailable: "+e.Error()} 
 clientBases,err:=client.run(map[string]interface{}{"cmd":"get_soldiers"})
 if err!=nil{r.ClientObservation="client-base-observe: "+err.Error()}else if bases,valid:=clientBases["bases"].([]interface{});valid{
  for _,item:=range bases{
   record,ok:=item.(map[string]interface{})
   if ok&&record["name"]=="Signal Ridge"{r.ClientSeesHostBase=true}
  }
 }
 // Source-derived hold: pinned SavedGame::saveCoopToMemory refuses any base
 // with lon==0 && lat==0, even on valid land. Host at 2 degrees in radians
 // is required for SHARED serialization; no bypass of engine save invariants.
 // The native separate-coop client must place its own first base before
 // the host's wait-for-bases dialogue can settle the campaign. Use the same
 // proven synthetic land triangle, offset from the host's base location.
 // TestServer::place_first_base delegates to Globe::insideLand in RADIANS,
 // so 5 degrees east must be 5*pi/180, not a numeric 5.0 radians.
 if mode=="separate" {
 r.ClientStates,e=client.waitState("BuildNewBaseState")
 if e!=nil{r.FirstFailure="client-first-base-state: "+e.Error();return}
 if _,ok:=cmd(client,map[string]interface{}{"cmd":"place_first_base","lon":0.08726646259971647,"lat":0.0,"name":"Echo Station"});!ok{return}
 r.ClientBaseNamed=true
 local,ok:=cmd(client,map[string]interface{}{"cmd":"get_soldiers"});if !ok{return}
 cb,exists:=local["bases"].([]interface{})
 if !exists{r.FirstFailure="client-base-list-missing";return}
 for _,item:=range cb{
  record,isObj:=item.(map[string]interface{})
  if isObj&&record["name"]=="Echo Station"{r.ClientSavedBaseVisible=true}
 }
 if !r.ClientSavedBaseVisible{r.FirstFailure="client-named-base-not-visible";return}
 } // SHARED has one authoritative host base; client must not place its own base.
 r.HostStates,e=host.states();if e!=nil{r.FirstFailure="post-client-base-host-state: "+e.Error();return}
 r.ClientStates,e=client.states();if e!=nil{r.FirstFailure="post-client-base-client-state: "+e.Error();return}
 // Allow the original host/client network state machines a short,
 // bounded chance to complete base-settlement acknowledgements.
 until:=time.Now().Add(4*time.Second)
 for time.Now().Before(until){
   hs,he:=host.states()
   cs,ce:=client.states()
   if he==nil&&ce==nil{
     r.HostStates=hs;r.ClientStates=cs
     r.HostStrategicReady=len(hs)>0&&!strings.Contains(hs[len(hs)-1],"CoopState")&&!strings.Contains(hs[len(hs)-1],"BuildNewBaseState")
     r.ClientStrategicReady=len(cs)>0&&!strings.Contains(cs[len(cs)-1],"CoopState")&&!strings.Contains(cs[len(cs)-1],"BuildNewBaseState")
     if r.HostStrategicReady&&r.ClientStrategicReady{break}
   }
   time.Sleep(250*time.Millisecond)
 }
 hostGeo,ge:=host.run(map[string]interface{}{"cmd":"geo_state"})
 if ge==nil{
  r.HostWorldObserved=true
  r.HostGeo=map[string]interface{}{"time":hostGeo["time"],"funds":hostGeo["funds"],"campaignType":hostGeo["campaignType"],"bases":hostGeo["bases"],"discoveredResearch":hostGeo["discoveredResearch"]}
 }else{r.ClientObservation+=" host-geo:"+ge.Error()}
 clientGeo,ge:=client.run(map[string]interface{}{"cmd":"geo_state"})
 if ge==nil{
  r.ClientWorldObserved=true
  r.ClientGeo=map[string]interface{}{"time":clientGeo["time"],"funds":clientGeo["funds"],"campaignType":clientGeo["campaignType"],"bases":clientGeo["bases"],"discoveredResearch":clientGeo["discoveredResearch"]}
 }else{r.ClientObservation+=" client-geo:"+ge.Error()}
 if r.HostWorldObserved&&r.ClientWorldObserved{
  h,_:=json.Marshal(r.HostGeo["time"]);c,_:=json.Marshal(r.ClientGeo["time"])
  r.WorldClockEqual=string(h)==string(c)
  r.CampaignMode,r.CampaignModeConsistent=campaignMode(r.HostGeo,r.ClientGeo)
  if mode=="shared" {
   if bases,ok:=r.ClientGeo["bases"].([]interface{});ok {for _,entry:=range bases {
    if base,ok:=entry.(map[string]interface{});ok&&base["name"]=="Signal Ridge" {r.ClientSeesHostBase=true;r.ClientSavedBaseVisible=true}
   }
   }
  }
  hh,_:=json.Marshal(r.HostGeo);cc,_:=json.Marshal(r.ClientGeo)
  r.ObservedSubsetEqual=string(hh)==string(cc)
  // Equality of a selected read-only projection is NEVER tactical or shared-world acceptance.
 }
 if mode=="shared" && (!r.ClientWorldObserved||!r.HostWorldObserved||!r.ClientSeesHostBase) && r.FirstFailure=="" {
   r.FirstFailure="shared-world-not-loaded-or-not-replicated: "+r.ClientObservation
 }
 // Still no shared world checksum or tactical turn-by-turn equivalence.
 // Distinct single-player saves in a co-op lobby are not proof of sync.
 // A host base in a real save is NOT proof of synchronized campaign,
 // client base replication, shared Battlescape or a playable release.
}
