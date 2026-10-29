package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "strings"
)

type Verdict struct {
  EvidenceClass string
  EngineExecuted bool
  GameplayQualified bool
  Missing []string
}

func block(text,key string) string {
  lines:=strings.Split(text,"\n")
  inside:=false
  out:=[]string{}
  for _,line:=range lines {
    if !inside {
      if line==key+":" {inside=true}
      continue
    }
    if line!=""&&!strings.HasPrefix(line," ")&&!strings.HasPrefix(line,"#"){break}
    out=append(out,line)
  }
  return strings.Join(out,"\n")
}

func matches(s,pattern string) bool {return regexp.MustCompile(pattern).MatchString(s)}

func Check(rules,config []byte,root string) Verdict {
  v:=Verdict{EvidenceClass:"SOURCE_DERIVED_NECESSARY_FIXTURE_CHECK_NOT_ENGINE",Missing:[]string{}}
  require:=func(ok bool,label string){if !ok{v.Missing=append(v.Missing,label)}}
  text:=string(rules)
  countries:=block(text,"countries")
  regions:=block(text,"regions")
  base:=block(text,"startingBase")
  interfaces:=block(text,"interfaces")
  globe:=block(text,"globe")
  require(matches(countries,"(?m)^\\s+- type:\\s+\\S+"),"countries:nonempty")
  require(matches(countries,"(?m)^\\s+fundingBase:\\s+[1-9]\\d*"),"countries:fundingBase")
  require(matches(countries,"(?m)^\\s+areas:\\s*$"),"countries:areas")
  require(matches(regions,"(?m)^\\s+- type:\\s+\\S+"),"regions:nonempty")
  require(matches(regions,"(?m)^\\s+areas:\\s*$"),"regions:areas")
  for _,field:=range []string{"facilities","crafts","items"} {
    require(matches(base,"(?m)^\\s+"+field+":\\s*"),"startingBase:"+field)
  }
  for _,name:=range []string{"mainMenu","newGameMenu","newBattleMenu","geoscape","pauseMenu","saveMenus"} {
    require(matches(interfaces,"(?m)^\\s+- type:\\s+"+name+"\\s*$"),"interfaces:"+name)
  }
  for _,name:=range []string{"slackingIndicator","trainingIndicator"} {
    require(matches(interfaces,"(?m)^\\s+- id:\\s+"+name+"\\s*$"),"interfaces:"+name)
  }
  require(matches(globe,"(?m)^\\s+textures:\\s*$")&&matches(globe,"(?m)^\\s+- id:\\s+0\\s*$"),"globe:nonempty-textures")
  require(bytes.Contains(config,[]byte("playIntro: false")),"config:disable-retail-intro")
  for _,name:=range []string{"base.png","avatar.png"} {
    raw,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","multiplayer",name))
    require(err==nil&&len(raw)>=8&&bytes.Equal(raw[:8],[]byte{137,80,78,71,13,10,26,10}),"original-coop-art:"+name)
  }
  for _,name:=range []string{"UFO","TFTD"} {
    _,err:=os.Stat(filepath.Join(root,"data",name))
    require(os.IsNotExist(err),"no-retail-dir:"+name)
  }
  return v
}

func main(){
  if len(os.Args)!=3{fmt.Fprintln(os.Stderr,"usage: preflight <fixture-dir> <out.json>");os.Exit(2)}
  root:=os.Args[1]
  rules,err:=os.ReadFile(filepath.Join(root,"user","mods","rdte-synthetic-master","Ruleset","10-synthetic.rul"));if err!=nil{panic(err)}
  options,err:=os.ReadFile(filepath.Join(root,"config","options.cfg"));if err!=nil{panic(err)}
  v:=Check(rules,options,root)
  b,_:=json.MarshalIndent(v,"","  ")
  if err=os.WriteFile(os.Args[2],append(b,byte(10)),0644);err!=nil{panic(err)}
  fmt.Println(string(b))
  if len(v.Missing)>0{os.Exit(1)}
}
