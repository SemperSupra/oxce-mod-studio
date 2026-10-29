package main
import(
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "log"
 "net/http"
 "os"
 "strconv"
 "strings"
 "sync"
)
type Action struct{ Seat string; Kind string; Key string; X int; Y int; Expected uint64 }
type Snapshot struct{ EvidenceClass string; Multiplayer bool; Revision uint64; ActiveSeat string; ScreenClass string; Commands int }
type Broker struct{ mu sync.Mutex; mailbox string; trace string; token string; seq uint64; activeSeat string; commands int }
var keys=map[string]int{"enter":13,"escape":27,"space":32,"up":273,"down":274,"right":275,"left":276,"tab":9,"backspace":8,"f1":282,"f2":283,"f3":284,"f4":285,"f5":286}
func (b *Broker) view() Snapshot{
 b.mu.Lock();defer b.mu.Unlock()
 state:="unknown"
 if v,e:=os.ReadFile(b.trace);e==nil{
  rows:=strings.Split(strings.TrimSpace(string(v)),"\n")
  if len(rows)>0{ var s struct{Class string};if json.Unmarshal([]byte(rows[len(rows)-1]),&s)==nil&&s.Class!=""{state=s.Class}}
 }
 return Snapshot{EvidenceClass:"LOCAL_SDL_INPUT_AND_CLASS_TRACE_NOT_GAMEPLAY",Multiplayer:false,Revision:b.seq,ActiveSeat:b.activeSeat,ScreenClass:state,Commands:b.commands}
}
func (b *Broker) action(a Action) error{
 b.mu.Lock();defer b.mu.Unlock()
 if a.Seat!="human"&&a.Seat!="automation"&&a.Seat!="agent" {return errors.New("invalid-seat")}
 if a.Expected!=b.seq{return errors.New("stale-revision")}
 if b.activeSeat!=""&&b.activeSeat!=a.Seat{return errors.New("seat-not-active")}
 var line string
 switch a.Kind{
 case "key":
  k,ok:=keys[strings.ToLower(a.Key)];if !ok{return errors.New("unsupported-key")}
  line=fmt.Sprintf("KEY %d\n",k)
 case "click":
  if a.X<0||a.Y<0||a.X>4095||a.Y>4095{return errors.New("invalid-coordinates")}
  line=fmt.Sprintf("CLICK %d %d\n",a.X,a.Y)
 case "move":
  if a.X<0||a.Y<0||a.X>4095||a.Y>4095{return errors.New("invalid-coordinates")}
  line=fmt.Sprintf("MOVE %d %d\n",a.X,a.Y)
 default:return errors.New("unsupported-operation")
 }
 f,e:=os.OpenFile(b.mailbox,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600)
 if e!=nil{return e};defer f.Close()
 if _,e=f.WriteString(line);e!=nil{return e}
 b.seq++;b.commands++
 return nil
}
func (b *Broker) Handler(page []byte)http.Handler{
 m:=http.NewServeMux()
 m.HandleFunc("GET /",func(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html; charset=utf-8");w.Write(page)})
 m.HandleFunc("GET /api/capabilities",func(w http.ResponseWriter,r *http.Request){json.NewEncoder(w).Encode(map[string]interface{}{"modes":[]string{"single-engine","shared-local-control"},"realMultiplayer":false,"assistiveStatus":"screen-class-only","actions":[]string{"key","click","move"},"seats":[]string{"human","automation","agent"},"engineGameStateQualified":false})})
 m.HandleFunc("GET /api/state",func(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","application/json");json.NewEncoder(w).Encode(b.view())})
 guard:=func(w http.ResponseWriter,r *http.Request)bool{
  if b.token==""||r.Header.Get("Authorization")!="Bearer "+b.token{http.Error(w,"unauthorized",http.StatusUnauthorized);return false}
  if r.Header.Get("Origin")!=""&&r.Header.Get("Origin")!="http://"+r.Host{http.Error(w,"origin denied",http.StatusForbidden);return false}
  return true
 }
 m.HandleFunc("POST /api/action",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  r.Body=http.MaxBytesReader(w,r.Body,2048)
  var a Action
  if e:=json.NewDecoder(r.Body).Decode(&a);e!=nil{http.Error(w,"invalid JSON",400);return}
  if e:=b.action(a);e!=nil{http.Error(w,e.Error(),409);return}
  json.NewEncoder(w).Encode(b.view())
 })
 m.HandleFunc("POST /api/seat",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  r.Body=http.MaxBytesReader(w,r.Body,256)
  v,e:=io.ReadAll(r.Body);if e!=nil{http.Error(w,"invalid body",400);return}
  seat:=strings.TrimSpace(string(v))
  if seat!="human"&&seat!="automation"&&seat!="agent"&&seat!="none"{http.Error(w,"invalid seat",400);return}
  b.mu.Lock();if seat=="none"{b.activeSeat=""}else{b.activeSeat=seat};b.seq++;b.mu.Unlock()
  json.NewEncoder(w).Encode(b.view())
 })
 return m
}
func main(){
 if len(os.Args)<4{log.Fatal("usage: oxce-rdte-broker <mailbox> <trace> <index.html> [port]")}
 port:=43819
 if len(os.Args)>4{v,e:=strconv.Atoi(os.Args[4]);if e!=nil||v<1024||v>65535{log.Fatal("invalid port")};port=v}
 token:=os.Getenv("SF_RDTE_TOKEN")
 if token==""{var b [24]byte;if _,e:=rand.Read(b[:]);e!=nil{log.Fatal(e)};token=hex.EncodeToString(b[:])}
 page,e:=os.ReadFile(os.Args[3]);if e!=nil{log.Fatal(e)}
 b:=&Broker{mailbox:os.Args[1],trace:os.Args[2],token:token,activeSeat:"human"}
 fmt.Println("Local evaluation controller: http://127.0.0.1:"+strconv.Itoa(port))
 fmt.Println("Bearer token (keep private): "+token)
 log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d",port),b.Handler(page)))
}
