package main

import "testing"

func TestCampaignModeFromPinnedCoopTestServer(t *testing.T) {
  tests:=[]struct{
    name string
    host interface{}
    client interface{}
    want string
    consistent bool
  }{
    {"separate worlds",float64(0),float64(0),"SEPARATE",true},
    {"shared worlds",float64(1),float64(1),"SHARED",true},
    {"mode disagreement",float64(0),float64(1),"UNKNOWN",false},
    {"missing mode",nil,float64(1),"UNKNOWN",false},
    {"invalid mode",float64(99),float64(99),"UNKNOWN",false},
  }
  for _,tt:=range tests{
    t.Run(tt.name,func(t *testing.T){
      got,ok:=campaignMode(map[string]interface{}{"campaignType":tt.host},map[string]interface{}{"campaignType":tt.client})
      if got!=tt.want||ok!=tt.consistent{t.Fatalf("mode=%q consistent=%v, want %q/%v",got,ok,tt.want,tt.consistent)}
    })
  }
}
