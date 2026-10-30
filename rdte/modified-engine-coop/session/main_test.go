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

func TestSourceBoundHostReleaseGuard(t *testing.T) {
 cases:=[]struct{name,mode string;code float64;label string;visible bool;pass bool}{
 {"shared real resume","shared",62,"RESUME",true,true},
 {"separate real begin","separate",60,"BEGIN",true,true},
 {"reject client wait","shared",68,"OK",false,false},
 {"reject mode confusion","separate",62,"RESUME",true,false},
 {"reject hidden control","shared",62,"RESUME",false,false},
 {"reject wrong label","shared",62,"DISCONNECT",true,false},
 {"reject invalid mode","invalid",60,"BEGIN",true,false},
 }
 for _,tt:=range cases {t.Run(tt.name,func(t *testing.T){
   d:=map[string]interface{}{"present":true,"code":tt.code,"backText":tt.label,"backVisible":tt.visible}
   if got:=hostReleaseDialog(tt.mode,d);got!=tt.pass{t.Fatalf("got=%v want=%v",got,tt.pass)}
 })}
}
func TestSourceBoundGeoscapeHoldClassifier(t *testing.T) {
 if !includesState([]string{"N8OpenXcom13GeoscapeStateE","N8OpenXcom9CoopStateE"},"CoopState"){t.Fatal("hold absent")}
 if includesState([]string{"N8OpenXcom13GeoscapeStateE"},"CoopState"){t.Fatal("invented hold")}
}
