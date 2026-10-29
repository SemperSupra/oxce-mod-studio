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
type Observation struct{
 Name string
 ServerPort int
 Connected bool
 Pong bool
 States []string
 Error string
}
func command(conn net.Conn,cmd string)(map[string]interface{},error){
 conn.SetDeadline(time.Now().Add(4*time.Second))
 if _,err:=fmt.Fprintf(conn,"{\"cmd\":%q}\n",cmd);err!=nil{return nil,err}
 line,err:=bufio.NewReader(conn).ReadString('\n')
 if err!=nil{return nil,err}
 var out map[string]interface{}
 err=json.Unmarshal([]byte(line),&out)
 return out,err
}
func check(name,path string)Observation{
 v:=Observation{Name:name}
 until:=time.Now().Add(12*time.Second)
 var port int
 for time.Now().Before(until){
  content,err:=os.ReadFile(filepath.Join(path,"testserver_port.txt"))
  if err==nil{port,_=strconv.Atoi(strings.TrimSpace(string(content)));if port>0{break}}
  time.Sleep(300*time.Millisecond)
 }
 if port<1||port>65535{v.Error="no-ephemeral-testserver-port";return v}
 v.ServerPort=port
 conn,err:=net.DialTimeout("tcp",fmt.Sprintf("127.0.0.1:%d",port),3*time.Second)
 if err!=nil{v.Error="dial: "+err.Error();return v}
 defer conn.Close();v.Connected=true
 response,err:=command(conn,"ping")
 if err!=nil{v.Error="ping: "+err.Error();return v}
 v.Pong=response["ok"]==true && response["pong"]==true
 states,err:=command(conn,"get_state")
 if err!=nil{v.Error="get_state: "+err.Error();return v}
 if a,ok:=states["states"].([]interface{});ok{for _,x:=range a{if s,ok:=x.(string);ok{v.States=append(v.States,s)}}}
 return v
}
func main(){
 if len(os.Args)!=4{fmt.Fprintln(os.Stderr,"usage: probe <host_user> <client_user> <report.json>");os.Exit(2)}
 result:=[]Observation{check("host",os.Args[1]),check("client",os.Args[2])}
 b,_:=json.MarshalIndent(map[string]interface{}{"schemaVersion":1,"kind":"TWO_ENGINE_PROCESS_TEST_ONLY_NOT_MULTIPLAYER_SESSION","originalCommercialGameAssets":false,"playersConnected":false,"host":result[0],"client":result[1]},"","  ")
 os.WriteFile(os.Args[3],append(b,'\n'),0644)
 fmt.Println(string(b))
 if !result[0].Pong||!result[1].Pong{os.Exit(1)}
}
