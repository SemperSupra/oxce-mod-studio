package audience

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
)
type Audience string
const (
 Human Audience="human"
 Automation Audience="automation"
 Agent Audience="agent"
)
type Action struct {
 Audience Audience `json:"audience"`
 RequestID string `json:"requestId"`
 Operation string `json:"operation"`
 ExpectedSeq uint64 `json:"expectedSeq"`
}
type MissionState struct {
 Scenario string `json:"scenario"`
 Seed uint64 `json:"seed"`
 Sequence uint64 `json:"sequence"`
 Turn uint64 `json:"turn"`
 Scanned bool `json:"scanned"`
 Recovered bool `json:"recovered"`
}
type Receipt struct {
 EvidenceClass string `json:"evidenceClass"`
 EngineExecuted bool `json:"engineExecuted"`
 RequestID string `json:"requestId"`
 State MissionState `json:"state"`
 SHA256 string `json:"sha256"`
}
type Model struct {state MissionState}
func New(seed uint64)*Model{return &Model{state:MissionState{Scenario:"silent-relay",Seed:seed}}}
func Supported()[]string{return []string{"scan-relay","end-turn","recover-recorder"}}
func authorized(a Audience)bool{return a==Human||a==Automation||a==Agent}
func (m *Model) Snapshot(requestID string)Receipt{
 s:=m.state
 b,_:=json.Marshal(s)
 d:=sha256.Sum256(b)
 return Receipt{EvidenceClass:"CONTRACT_SIMULATOR_NOT_OXCE",EngineExecuted:false,RequestID:requestID,State:s,SHA256:hex.EncodeToString(d[:])}
}
func (m *Model) Apply(a Action)(Receipt,error){
 before:=m.Snapshot(a.RequestID)
 if !authorized(a.Audience){return before,errors.New("audience-not-authorized")}
 if a.RequestID==""{return before,errors.New("missing-request-id")}
 if a.ExpectedSeq!=m.state.Sequence{return before,fmt.Errorf("stale-sequence expected=%d actual=%d",a.ExpectedSeq,m.state.Sequence)}
 switch a.Operation{
 case "scan-relay":
  if m.state.Scanned{return before,errors.New("already-scanned")}
  m.state.Scanned=true
 case "end-turn":
  m.state.Turn++
 case "recover-recorder":
  if !m.state.Scanned{return before,errors.New("not-scanned")}
  if m.state.Recovered{return before,errors.New("already-recovered")}
  m.state.Recovered=true
 default:return before,errors.New("unknown-operation")
 }
 m.state.Sequence++
 return m.Snapshot(a.RequestID),nil
}
