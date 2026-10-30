// A fresh-OS-process test: reload the prior host-owned original-game save,
// rehost the existing seat, join a distinct client, and verify same campaign.
// It does not infer full multiplayer gameplay qualification.
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
type peer struct{name string;conn net.Conn;reader *bufio.Reader}
func dial(path,name string)(*peer,error){
 until:=time.Now().Add(14*time.Second)
 for time.Now().Before(until){
  b,err:=os.ReadFile(filepath.Join(path,"testserver_port.txt"))
  if err==nil{
   port,e:=strconv.Atoi(strings.TrimSpace(string(b)))
   if e==nil&&port>0&&port<=65535 {
    c,e:=net.DialTimeout("tcp",fmt.Sprintf("127.0.0.1:%d",port),400*time.Millisecond)
    if e==nil{return &peer{name,c,bufio.NewReader(c)},nil}
   }
  }
  time.Sleep(200*time.Millisecond)
 }
 return nil,fmt.Errorf("%s fresh native test server not reachable",name)
}
func(p *peer)cmd(cmd string,extra map[string]interface{})(map[string]interface{},error){
 req:=map[string]interface{}{"cmd":cmd};for k,v:=range extra{req[k]=v}
 raw,_:=json.Marshal(req)
 p.conn.SetDeadline(time.Now().Add(5*time.Second))
 if _,e:=p.conn.Write(append(raw,'\n'));e!=nil{return nil,e}
 line,e:=p.reader.ReadString('\n');if e!=nil{return nil,e}
 var out map[string]interface{};if e=json.Unmarshal([]byte(line),&out);e!=nil{return nil,e}
 if out["ok"]!=true{return out,fmt.Errorf("%s:%s rejected: %v",p.name,cmd,out["error"])}
 return out,nil
}
func(p *peer)waitState(sub string)(map[string]interface{},error){
 deadline:=time.Now().Add(16*time.Second)
 for time.Now().Before(deadline){
  r,e:=p.cmd("get_state",nil);if e!=nil{return nil,e}
  if v,ok:=r["states"].([]interface{});ok{
   for _,s:=range v{if strings.Contains(fmt.Sprint(s),sub){return r,nil}}
  }
  time.Sleep(175*time.Millisecond)
 }
 return nil,fmt.Errorf("%s timed out waiting for %s",p.name,sub)
}
type receipt struct{
 SchemaVersion int
 EvidenceClass string
 Mode string
 NewProcessHostsRespond bool
 NativeSaveLoaded bool
 HostIdentityPreserved bool
 ResumeLobbyMode int
 ClientRejoined bool
 ResumeCommandAccepted bool
 NativeHoldReleaseAccepted bool
 NativeHoldDialogCode int
 NativeHoldDialogText string
 LastHostStates []interface{}
 LastClientStates []interface{}
 HostWorldObserved bool
 ClientWorldObserved bool
 ModeConsistent bool
 SharedWorldSubsetEqual bool
 HostClockAdvancedAfterRestart bool
 ClientClockAdvancedAfterRestart bool
 ResumedClockEqual bool
 FirstFailure string
 GameplayQualified bool
}
func classifyMode(host,client map[string]interface{},mode string)(bool,bool){
 m:=float64(0);if mode=="shared"{m=1}
 a,okA:=host["campaignType"].(float64)
 b,okB:=client["campaignType"].(float64)
 consistent:=okA&&okB&&a==m&&b==m
 if !consistent{return false,false}
 if mode!="shared"{return true,false}
 // Comparable structured strategic snapshot only, not full game checksum.
 x:=map[string]interface{}{"bases":host["bases"],"funds":host["funds"],"time":host["time"]}
 y:=map[string]interface{}{"bases":client["bases"],"funds":client["funds"],"time":client["time"]}
 bx,_:=json.Marshal(x);by,_:=json.Marshal(y)
 return true,string(bx)==string(by)
}
func execute(hostPath,clientPath,saveFile,mode string)receipt{
 r:=receipt{SchemaVersion:1,EvidenceClass:"FRESH_OS_PROCESSES_REAL_COOP_SAVED_WORLD_REHOST_REJOIN_NOT_TACTICAL_GAMEPLAY",Mode:mode,GameplayQualified:false}
 fail:=func(s string)receipt{r.FirstFailure=s;return r}
 if mode!="shared"&&mode!="separate"{return fail("invalid-mode")}
 host,e:=dial(hostPath,"host");if e!=nil{return fail(e.Error())};defer host.conn.Close()
 client,e:=dial(clientPath,"client");if e!=nil{return fail(e.Error())};defer client.conn.Close()
 for _,p:=range []*peer{host,client}{
  pong,e:=p.cmd("ping",nil);if e!=nil||pong["pong"]!=true{return fail(p.name+":fresh-native-ping-failed")}
 }
 r.NewProcessHostsRespond=true
 if _,e=host.cmd("load_save_menu",map[string]interface{}{"file":saveFile});e!=nil{return fail("host-load-save-menu:"+e.Error())}
 if _,e=host.waitState("GeoscapeState");e!=nil{return fail("host-load:"+e.Error())}
 mark,e:=host.cmd("save_markers",nil);if e!=nil{return fail("host-markers:"+e.Error())}
 expected:=float64(0);if mode=="shared"{expected=1}
 if mark["coop"]!=true||mark["campaignType"]!=expected{return fail("loaded-save-type-mismatch")}
 r.NativeSaveLoaded=true
 start,e:=host.cmd("host_tcp",map[string]interface{}{"server":"SignalFrontierFreshRestart","port":"0","player":"HostPlayer"});if e!=nil{return fail("host-rehost:"+e.Error())}
 port,ok:=start["port"].(string);if !ok||port==""{return fail("rehost-ephemeral-port-missing")}
 if _,e=host.waitState("LobbyMenu");e!=nil{return fail("host-resume-lobby:"+e.Error())}
 lobby,e:=host.cmd("lobby_state",nil);if e!=nil{return fail("host-lobby-state:"+e.Error())}
 n,ok:=lobby["lobbyMode"].(float64);if !ok{return fail("lobby-mode-untyped")}
 r.ResumeLobbyMode=int(n)
 if r.ResumeLobbyMode!=2{return fail("not-resume-lobby")}
 r.HostIdentityPreserved=true
 if _,e=client.cmd("join_tcp",map[string]interface{}{"ip":"127.0.0.1","port":port,"player":"ClientPlayer"});e!=nil{return fail("client-rejoin:"+e.Error())}
 r.ClientRejoined=true
 if _,e=client.waitState("Profile");e!=nil{return fail("client-profile:"+e.Error())}
 if _,e=client.cmd("profile_ok",nil);e!=nil{return fail("client-profile-accept:"+e.Error())}
 // A join may also generate a host profile modal; close it if present.
 if states,e:=host.cmd("get_state",nil);e==nil{
  if hostStates,ok:=states["states"].([]interface{});ok{
   for _,s:=range hostStates{
    if strings.Contains(fmt.Sprint(s),"Profile"){if _,e=host.cmd("profile_ok",nil);e!=nil{return fail("host-profile:"+e.Error())};break}
   }
  }
 }
 eligible:=false;until:=time.Now().Add(12*time.Second)
 for time.Now().Before(until){
  lobby,e=host.cmd("lobby_state",nil)
  if e==nil&&lobby["startEligible"]==true{eligible=true;break}
  time.Sleep(180*time.Millisecond)
 }
 if !eligible{return fail("resume-lobby-not-ready")}
 if _,e=host.cmd("lobby_resume_campaign",nil);e!=nil{return fail("host-resume-command:"+e.Error())}
 r.ResumeCommandAccepted=true
 // Source: LobbyMenu::resumeCampaign deliberately pushes COOP_DLG_WAIT_PLAYERS.
 // The native wait dialog (not the lobby command itself) owns final RESUME.
 // Do NOT interpret a loaded-but-paused geoscape as playable continuation.
 releaseUntil:=time.Now().Add(17*time.Second)
 for time.Now().Before(releaseUntil){
  d,ed:=host.cmd("coop_dialog_info",nil)
  if ed==nil{
   code,_:=d["code"].(float64)
   label,_:=d["backText"].(string)
   r.NativeHoldDialogCode=int(code);r.NativeHoldDialogText=label
   if d["present"]==true&&d["backVisible"]==true&&
      (code==60||code==62)&&(label=="BEGIN"||label=="RESUME"){
    if _,e=host.cmd("coop_dialog_back",nil);e!=nil{return fail("native-hold-release:"+e.Error())}
    r.NativeHoldReleaseAccepted=true
    break
   }
  }
  time.Sleep(175*time.Millisecond)
 }
 if !r.NativeHoldReleaseAccepted{return fail(fmt.Sprintf("native-hold-not-releasable: code=%d label=%q",r.NativeHoldDialogCode,r.NativeHoldDialogText))}
 until=time.Now().Add(18*time.Second)
 for time.Now().Before(until){
  hs,eh:=host.cmd("get_state",nil);cs,ec:=client.cmd("get_state",nil)
  if eh==nil&&ec==nil{
   if hstate,ok:=hs["states"].([]interface{});ok{r.LastHostStates=hstate}
   if cstate,ok:=cs["states"].([]interface{});ok{r.LastClientStates=cstate}
   hostReady,clientReady:=false,false
   if hstates,ok:=hs["states"].([]interface{});ok {
    for _,s:=range hstates{
     if strings.Contains(fmt.Sprint(s),"GeoscapeState"){hostReady=true}
     if strings.Contains(fmt.Sprint(s),"CoopState")||strings.Contains(fmt.Sprint(s),"LobbyMenu"){hostReady=false;break}
    }
   }
   if cstates,ok:=cs["states"].([]interface{});ok {
    for _,s:=range cstates{
     if strings.Contains(fmt.Sprint(s),"GeoscapeState"){clientReady=true}
     if strings.Contains(fmt.Sprint(s),"CoopState")||strings.Contains(fmt.Sprint(s),"LobbyMenu"){clientReady=false;break}
    }
   }
   r.HostWorldObserved=hostReady;r.ClientWorldObserved=clientReady
  }
  if r.HostWorldObserved&&r.ClientWorldObserved{
   hg,eh:=host.cmd("geo_state",nil);cg,ec:=client.cmd("geo_state",nil)
   if eh==nil&&ec==nil{
    r.ModeConsistent,r.SharedWorldSubsetEqual=classifyMode(hg,cg,mode)
    if r.ModeConsistent&&(mode!="shared"||r.SharedWorldSubsetEqual){
     beforeHost,_:=json.Marshal(hg["time"]);beforeClient,_:=json.Marshal(cg["time"])
     if _,e=host.cmd("geo_set_speed",map[string]interface{}{"idx":1});e!=nil{return fail("host-time-speed-after-restart:"+e.Error())}
     if _,e=client.cmd("geo_set_speed",map[string]interface{}{"idx":1});e!=nil{return fail("client-time-speed-after-restart:"+e.Error())}
     untilAdvance:=time.Now().Add(8*time.Second)
     for time.Now().Before(untilAdvance){
      h,e1:=host.cmd("geo_state",nil);c,e2:=client.cmd("geo_state",nil)
      if e1==nil&&e2==nil{
       hb,_:=json.Marshal(h["time"]);cb,_:=json.Marshal(c["time"])
       r.HostClockAdvancedAfterRestart=string(hb)!=string(beforeHost)
       r.ClientClockAdvancedAfterRestart=string(cb)!=string(beforeClient)
       r.ResumedClockEqual=string(hb)==string(cb)
       if r.HostClockAdvancedAfterRestart&&r.ClientClockAdvancedAfterRestart&&r.ResumedClockEqual{return r}
      }
      time.Sleep(250*time.Millisecond)
     }
     return fail("fresh-process-native-time-not-advanced-or-diverged")
    }
   }
  }
  time.Sleep(250*time.Millisecond)
 }
 return fail("fresh-process-world-not-consistent")
}
func main(){
 if len(os.Args)!=6{fmt.Fprintln(os.Stderr,"usage: restart <host-user> <client-user> <save-file> <separate|shared> <out.json>");os.Exit(2)}
 r:=execute(os.Args[1],os.Args[2],os.Args[3],os.Args[4])
 b,_:=json.MarshalIndent(r,"","  ");if err:=os.WriteFile(os.Args[5],append(b,'\n'),0644);err!=nil{panic(err)}
 fmt.Println(string(b))
 if r.FirstFailure!=""||!r.NewProcessHostsRespond||!r.NativeSaveLoaded||!r.HostIdentityPreserved||r.ResumeLobbyMode!=2||!r.ClientRejoined||!r.ResumeCommandAccepted||!r.NativeHoldReleaseAccepted||!r.HostWorldObserved||!r.ClientWorldObserved||!r.ModeConsistent||(r.Mode=="shared"&&!r.SharedWorldSubsetEqual)||!r.HostClockAdvancedAfterRestart||!r.ClientClockAdvancedAfterRestart||!r.ResumedClockEqual{os.Exit(1)}
}
