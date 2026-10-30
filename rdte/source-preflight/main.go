package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "image"
  "image/png"
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


func inventoryRecord(section,id string) string {
  active:=false
  lines:=[]string{}
  for _,line:=range strings.Split(section,"\n") {
    if strings.HasPrefix(line,"  - id: ") {
      if active {break}
      active=strings.TrimSpace(strings.TrimPrefix(line,"  - id: "))==id
    }
    if active {lines=append(lines,line)}
  }
  return strings.Join(lines,"\n")
}

func matches(s,pattern string) bool {return regexp.MustCompile(pattern).MatchString(s)}

// Source-indexed required interface elements for the pinned 8.6.5 build.
// Keep the checking scope at a named interface; a duplicate id in geoscape
// cannot satisfy a missing tactical HUD element.
func interfaceRecord(interfaces,kind string)string {
  anchor:="  - type: "+kind+"\n"
  i:=strings.Index(interfaces,anchor)
  if i<0{return ""}
  body:=interfaces[i+len(anchor):]
  if n:=strings.Index(body,"\n  - type:");n>=0{body=body[:n]+"\n"}
  return body
}
// extraSpriteEntry respects each YAML sprite type boundary. A flat string
// match for an image name would accept an incorrect SurfaceSet declaration.
func extraSpriteEntry(extra,kind string)string{
  key:="  - type: "+kind+"\n"
  pos:=strings.Index(extra,key)
  if pos<0{return ""}
  following:=extra[pos+len(key):]
  if n:=strings.Index(following,"\n  - type:");n>=0{following=following[:n]}
  return following
}
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
  // Public RuleRegion::insideRegion excludes 0 when synthetic span starts at +10°.
  // Both host origin (0,0) and client +5° must be inside authored bounds.
  require(strings.Contains(regions,"- [0, 360, -65, 65]"),"regions:contains-origin")
  require(strings.Contains(countries,"- [0, 360, -65, 65]"),"countries:contains-origin")
  for _,field:=range []string{"facilities","crafts","items"} {
    require(matches(base,"(?m)^\\s+"+field+":\\s*"),"startingBase:"+field)
  }
  for _,name:=range []string{"mainMenu","newGameMenu","newBattleMenu","geoscape","pauseMenu","saveMenus","baseNaming","placeFacility","selectFacility","basescape","battlescape","inventory"} {
    require(matches(interfaces,"(?m)^\\s+- type:\\s+"+name+"\\s*$"),"interfaces:"+name)
  }
  for _,name:=range []string{"slackingIndicator","trainingIndicator"} {
    require(matches(interfaces,"(?m)^\\s+- id:\\s+"+name+"\\s*$"),"interfaces:"+name)
  }
  require(strings.Contains(interfaces,"id: icons")&&strings.Contains(interfaces,"size: [320, 56]"),"interfaces:battlescape-icons")
  require(strings.Contains(interfaces,"id: weight")&&strings.Contains(interfaces,"id: grid"),"interfaces:inventory-essentials")
  // Every element accessed through pinned Battlescape/Inventory states
  // must be present in its OWN named interface, not just anywhere in YAML.
  for _,part:=range []struct{kind string;ids []string}{
    {"battlescape",[]string{"actionMenu","barEnergy","barHealth","barMana","barMorale","barTUs","buttonAbort","buttonCenter","buttonEndTurn","buttonHelp","buttonInventory","buttonKneel","buttonLeftHand","buttonMapDown","buttonMapUp","buttonNextSoldier","buttonNextStop","buttonReserveAimed","buttonReserveAuto","buttonReserveKneel","buttonReserveNone","buttonReserveSnap","buttonRightHand","buttonShowLayers","buttonShowMap","buttonStats","buttonUnitDown","buttonUnitUp","buttonZeroTUs","disabledUserOption","grenadeBackground","grenadeMenu","icons","infoBox","infoBoxOK","infoBoxOKButton","infoBoxText","medikitBlue","medikitGreen","medikitOrange","medikitRed","messageWindowBorder","messageWindowButtons","messageWindows","numAmmoLeft","numAmmoRight","numEnergy","numHealth","numLayers","numMedikitLeft","numMedikitRight","numMorale","numTUs","numTwoHandedIndicatorLeft","numTwoHandedIndicatorRight","optionLists","passingOutUnits","pathfinding","rank","squadsightUnits","textLoad","textName","textTooltip","thinkingProgressBar","twoHandedGreen","twoHandedRed","visibleUnits","warning","woundedUnits"}},
    {"inventory",[]string{"buttonApply","buttonArmor","buttonCreate","buttonGround","buttonLinks","buttonNext","buttonOK","buttonPrev","buttonUnload","grid","numStack","rank","textAmmo","textItem","textName","textSlot","textSlots","textStatLine1","textStatLine2","textStatLine3","textStatLine4","textTUs","textWeight","weight"}},
  }{
    record:=interfaceRecord(interfaces,part.kind)
    for _,id:=range part.ids{
      require(strings.Contains(record,"      - id: "+id+"\n"),"interfaces:"+part.kind+":"+id)
    }
  }
  require(matches(globe,"(?m)^\\s+textures:\\s*$")&&matches(globe,"(?m)^\\s+- id:\\s+0\\s*$"),"globe:nonempty-textures")
  require(strings.Contains(globe,"  polygons:")&&strings.Contains(globe,"    - [0, -15, -15, 15, -15, 0, 15]"),"globe:original-land-triangle")
  require(strings.Contains(block(text,"facilities"),"lift: true")&&strings.Contains(block(text,"facilities"),"type: STR_SF_ACCESS_LIFT"),"facilities:original-access-lift")
  require(strings.Contains(block(text,"facilities"),"fireSound: -1")&&strings.Contains(block(text,"facilities"),"hitSound: -1")&&strings.Contains(block(text,"facilities"),"placeSound: -1"),"facilities:no-retail-GEO-CAT-index")
  // The authored FONT atlas must be mounted and uniformly use the visible
  // foreground palette index. Alternating palette index 2 made half of the
  // glyphs near-black in the actual native briefing screenshot.
  fontPath:=filepath.Join(root,"user","mods","rdte-synthetic-master","Language","rdte-fonts.png")
  fontHandle,fontErr:=os.Open(fontPath)
  fontValid:=false
  if fontErr==nil {
    decoded,decodeErr:=png.Decode(fontHandle)
    fontHandle.Close()
    if decodeErr==nil {
      paletted,ok:=decoded.(*image.Paletted)
      if ok&&paletted.Bounds().Dx()==128&&paletted.Bounds().Dy()==72 {
        visible:=0
        dark:=0
        for _,index:=range paletted.Pix {
          if index==1{visible++}
          if index==2{dark++}
        }
        fontValid=visible>0&&dark==0
      }
    }
  }
  require(fontValid,"font:original-128x72-visible-single-index")
  fontYAML,fontYamlErr:=os.ReadFile(filepath.Join(root,"user","mods","rdte-synthetic-master","Language","rdte-fonts.yml"))
  require(fontYamlErr==nil&&bytes.Contains(fontYAML,[]byte("FONT_SMALL"))&&bytes.Contains(fontYAML,[]byte("FONT_BIG")),"font:original-declared-atlas")
  // A native screenshot showed near-black index-1 glyphs on a deep-blue
  // checkerboard. Verify high-luminance 6-bit foreground in all palette banks.
  pal,pe:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","GEODATA","PALETTES.DAT"))
  pass:=pe==nil&&len(pal)==5*256*3
  if pass {
    for bank:=0;bank<5;bank++ {
      off:=(bank*256+1)*3
      if pal[off]<56||pal[off+1]<56||pal[off+2]<48 {pass=false;break}
    }
  }
  require(pass,"ui:original-palette-readable-foreground")
  // Pinned co-op Globe::drawTarget indexes GlobeMarkers frames; synthetic
  // atlas must provide all nine original entries before the added co-op frame 9.
  sprite:=block(text,"extraSprites")
  require(matches(sprite,"(?m)^\\s+- type:\\s+GlobeMarkers\\s*$")&&
    strings.Contains(sprite,"subX: 3")&&strings.Contains(sprite,"subY: 3")&&
    strings.Contains(sprite,"0: multiplayer/globe-markers.png"),
    "extraSprites:globe-markers-nine-frames")
  marker,me:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","multiplayer","globe-markers.png"))
  require(me==nil&&len(marker)>=24&&
    bytes.Equal(marker[:8],[]byte{137,80,78,71,13,10,26,10})&&
    bytes.Equal(marker[16:24],[]byte{0,0,0,27,0,0,0,3})&&len(marker)>=26&&marker[24]==8&&marker[25]==3,
    "original-coop-art:globe-markers-27x3")
  // Pinned BattlescapeState::BattlescapeState dereferences SPICONS frames
  // 0 and 1; Touch frames 0..9. The standalone master must create all.
  sprites:=[]struct{kind string;path string;width,height int;label string}{
    {"Touch","multiplayer/touch-controls.png",320,24,"original-touch-ten-frames"},
    {"TinyRanks","multiplayer/tiny-ranks.png",42,7,"original-ranks-six-frames"},
    {"KneelButton","multiplayer/kneel-button.png",64,16,"kneel-button-two-frames"},
  }
  for _,spec:=range sprites {
    require(strings.Contains(sprite,"- type: "+spec.kind)&&
      strings.Contains(sprite,"0: "+spec.path),
      "extraSprites:"+spec.kind)
    file,err:=os.Open(filepath.Join(root,"data","SYNTHETIC",filepath.FromSlash(spec.path)))
    valid:=false
    if err==nil {
      img,decErr:=png.Decode(file)
      file.Close()
      if decErr==nil {
        typed,ok:=img.(*image.Paletted)
        valid=ok&&typed.Bounds().Dx()==spec.width&&typed.Bounds().Dy()==spec.height
      }
    }
    require(valid,"original-coop-art:"+spec.label)
  }
  icons,iconErr:=os.Stat(filepath.Join(root,"data","SYNTHETIC","UFOGRAPH","SPICONS.DAT"))
  require(iconErr==nil&&icons.Size()>=2*32*24&&icons.Size()%(32*24)==0,
    "tactical:spicons-at-least-two-frames")
  // Pinned InventoryState::InventoryState and updateTemplateButtons dereference
  // five independent named sprite getters, whether or not oxceLinks is visible.
  for _,spec:=range []struct{kind,file string;w,h int}{
    {"oxceLinksInv","oxce-links-inventory.png",23,22},
    {"InvCopy","inv-inv-copy.png",32,22},
    {"InvPasteEmpty","inv-inv-paste-empty.png",32,22},
    {"InvCopyActive","inv-inv-copy-active.png",32,22},
    {"InvPaste","inv-inv-paste.png",32,22},
  }{
    entry:=extraSpriteEntry(sprite,spec.kind)
    require(strings.Contains(sprite,"- type: "+spec.kind+"\n")&&
      strings.Contains(entry,"singleImage: true")&&
      strings.Contains(entry,"0: multiplayer/"+spec.file),"inventory-sprite:"+spec.kind)
    p:=filepath.Join(root,"data","SYNTHETIC","multiplayer",spec.file)
    fh,err:=os.Open(p)
    valid:=false
    if err==nil {
      im,e:=png.Decode(fh);fh.Close()
      if e==nil {typed,ok:=im.(*image.Paletted);valid=ok&&typed.Bounds().Dx()==spec.w&&typed.Bounds().Dy()==spec.h}
    }
    require(valid,"inventory-original-indexed-art:"+spec.kind)
  }
  // Armor::load reads spriteInv. InventoryState::init falls through to
  // getSurface(look, true), which returns null for the default empty name
  // (and then segfaults in blitNShade). Enforce a real single-image sprite.
  doll:=extraSpriteEntry(sprite,"SF_SurveyPaperdoll")
  require(strings.Contains(doll,"singleImage: true")&&
    strings.Contains(doll,"0: multiplayer/survey-paperdoll.png"),
    "inventory:paperdoll-single-surface")
  dollFile,err:=os.Open(filepath.Join(root,"data","SYNTHETIC","multiplayer","survey-paperdoll.png"))
  dollValid:=false
  if err==nil {
    i,e:=png.Decode(dollFile);dollFile.Close()
    if e==nil {p,ok:=i.(*image.Paletted);dollValid=ok&&p.Bounds().Dx()==320&&p.Bounds().Dy()==200}
  }
  require(dollValid,"inventory:original-indexed-paperdoll")
  // Native controls revealed unresolved menu tokens and unreadable briefing text.
  // Check only the original en-US file; other languages require separate font HIL.
  locale,le:=os.ReadFile(filepath.Join(root,"user","mods","rdte-synthetic-master","Language","en-US.yml"))
  require(le==nil,"language:english-present")
  for _,key:=range []string{"STR_MISSION","STR_CRAFT","STR_EQUIP_CRAFT","STR_TERROR_SITE","STR_SF_RELAY_DEPLOYMENT_BRIEFING"} {
    require(le==nil&&strings.Contains(string(locale),"  "+key+": "),"language:"+key)
  }
  // Pinned OXCE 8.6.5 Mod::loadRule iterates "invs" with "id".
  // RuleInventory::load reads numeric type: INV_GROUND=2, INV_HAND=1.
  invs:=block(text,"invs")
  for _,spec:=range []struct{id string;kind int}{
    {"STR_GROUND",2},{"STR_RIGHT_HAND",1},{"STR_LEFT_HAND",1},
  }{
    rec:=inventoryRecord(invs,spec.id)
    require(rec!=""&&matches(rec,fmt.Sprintf("(?m)^\\s+type:\\s+%d\\s*$",spec.kind)),"invs:"+spec.id+":kind")
  }
  for _,id:=range []string{"STR_BELT","STR_BACK_PACK"} {
    rec:=inventoryRecord(invs,id)
    require(rec!=""&&matches(rec,"(?m)^\\s+type:\\s+0\\s*$")&&matches(rec,"(?m)^\\s+slots:\\s+\\[\\["),
      "invs:"+id+":slots")
  }
  // Pinned NewBattleState::btnOkClick resolves Mod::getAlienMission(getAlienMissionList().front()).
  // Absence of a mission entry was the next native GDB first failure.
  deployment:=block(text,"alienDeployments")
  require(matches(deployment,"(?m)^\\s+itemSets:\\s+\\[\\[STR_RDTE_TOOL\\]\\]\\s*$"),
    "alienDeployments:original-tool-itemset")
  mission:=block(text,"alienMissions")
  require(matches(mission,"(?m)^\\s+- type:\\s+STR_SF_SIGNAL_PATROL\\s*$"),"alienMissions:nonempty-synthetic-patrol")
  // RuleTerrain defaults to DEFAULT; pinned Mod::loadRuleset reads
  // mapScripts type/commands and MapScript::load expects addBlock.
  script:=block(text,"mapScripts")
  require(matches(script,"(?m)^\\s+- type:\\s+DEFAULT\\s*$")&&
    matches(script,"(?m)^\\s+- type:\\s+addBlock\\s*$")&&
    matches(script,"(?m)^\\s+blocks:\\s+\\[0\\]\\s*$"),
    "mapScripts:DEFAULT-original-addBlock")
  armor:=block(text,"armors")
  require(matches(armor,"(?m)^\\s+corpseBattle:\\s*\\[STR_SF_SURVEY_SHELL\\]\\s*$"),"armors:one-corpse")
  require(matches(armor,"(?m)^\\s+loftempsSet:\\s*\\[3\\]\\s*$"),"armors:one-loft-template")
  require(strings.Contains(armor,"spriteInv: SF_SurveyPaperdoll"),"armors:original-paperdoll-id")
  require(strings.Contains(block(text,"items"),"type: STR_SF_SURVEY_SHELL"),"items:survey-shell")
  for _,extension:=range []string{".PCK",".TAB"}{
    b,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","UNITS","SIGNALRELAY"+extension))
    require(err==nil&&len(b)>0,"units:signal-relay"+extension)
  }
  // Necessary OXCE binary resource invariants, not tactical/visual acceptance.
  mcd,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","TERRAIN","BLANKS.MCD"))
  require(err==nil&&len(mcd)==124,"terrain:original-two-record-MCD")
  sfMap,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","MAPS","SFREL00.MAP"))
  require(err==nil&&len(sfMap)==403&&bytes.Equal(sfMap[:3],[]byte{10,10,1}),"maps:original-10x10-relay")
  routes,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","ROUTES","SFREL00.RMP"))
  require(err==nil&&len(routes)==96,"routes:original-four-node-relay")
  require(bytes.Contains(config,[]byte("playIntro: false")),"config:disable-retail-intro")
  for _,rule:=range []struct{key,expr string}{
    {"musics:GMGEO1", "musics"},
    {"crafts:STR_SF_RESPONDER","crafts"},
    {"alienDeployments:STR_SF_RELAY_DEPLOYMENT","alienDeployments"},
    {"terrains:SF_RELAY","terrains"},
    {"soldiers:STR_SF_SURVEYOR","soldiers"},
    {"armors:STR_SF_SURVEY_ARMOR","armors"},
    {"alienRaces:STR_SF_QUIET_ARRAY","alienRaces"},
    {"units:STR_SF_PROBE","units"},
  }{
    content:=block(text,rule.expr)
    var name string
    switch rule.key{
      case "musics:GMGEO1": name="GMGEO1"
      case "crafts:STR_SF_RESPONDER": name="STR_SF_RESPONDER"
      case "alienDeployments:STR_SF_RELAY_DEPLOYMENT": name="STR_SF_RELAY_DEPLOYMENT"
      case "terrains:SF_RELAY": name="SF_RELAY"
      case "soldiers:STR_SF_SURVEYOR": name="STR_SF_SURVEYOR"
      case "armors:STR_SF_SURVEY_ARMOR": name="STR_SF_SURVEY_ARMOR"
      case "alienRaces:STR_SF_QUIET_ARRAY": name="STR_SF_QUIET_ARRAY"
      case "units:STR_SF_PROBE": name="STR_SF_PROBE"
    }
    require(strings.Contains(content,"type: "+name)||strings.Contains(content,"id: "+name)||strings.Contains(content,"name: "+name),rule.key)
  }
  require(matches(block(text,"alienItemLevels"),"(?m)^\\s+- \\["),"alienItemLevels:nonempty")
  for _,name:=range []string{"gmgeo1.wav","gminter.wav","gmtactic1.wav","gmenbase.wav"}{
    raw,err:=os.ReadFile(filepath.Join(root,"data","SYNTHETIC","SOUND",name))
    require(err==nil&&len(raw)>44&&bytes.Equal(raw[:4],[]byte("RIFF"))&&bytes.Equal(raw[8:12],[]byte("WAVE")), "original-music:"+name)
  }

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
