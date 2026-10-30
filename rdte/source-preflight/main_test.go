package main
import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)
const fixture = "extraSprites:\n  - type: GlobeMarkers\n    width: 27\n    height: 3\n    subX: 3\n    subY: 3\n    files:\n      0: multiplayer/globe-markers.png\ninvs:\n  - id: STR_GROUND\n    x: 0\n    y: 140\n    type: 2\n  - id: STR_RIGHT_HAND\n    type: 1\n  - id: STR_LEFT_HAND\n    type: 1\n  - id: STR_BELT\n    type: 0\n    slots: [[0, 0], [1, 0]]\n  - id: STR_BACK_PACK\n    type: 0\n    slots: [[0, 0], [1, 0]]\nmusics:\n  - type: GMGEO1\ncrafts:\n  - type: STR_SF_RESPONDER\nalienMissions:\n  - type: STR_SF_SIGNAL_PATROL\nalienDeployments:\n  - type: STR_SF_RELAY_DEPLOYMENT\n    data:\n      - alienRank: 0\n        itemSets: [[STR_RDTE_TOOL]]\nmapScripts:\n  - type: DEFAULT\n    commands:\n      - type: addBlock\n        size: [1, 1]\n        blocks: [0]\nterrains:\n  - name: SF_RELAY\nsoldiers:\n  - type: STR_SF_SURVEYOR\narmors:\n  - type: STR_SF_SURVEY_ARMOR\n    corpseBattle: [STR_SF_SURVEY_SHELL]\n    loftempsSet: [3]\nalienRaces:\n  - id: STR_SF_QUIET_ARRAY\nitems:\n  - type: STR_SF_SURVEY_SHELL\nunits:\n  - type: STR_SF_PROBE\nalienItemLevels:\n  - [0, 0, 0]\ncountries:\n  - type: STR_SF_COUNCIL\n    fundingBase: 500\n    areas:\n      - [0, 360, -65, 65]\nregions:\n  - type: STR_SF_AREA\n    areas:\n      - [0, 360, -65, 65]\nstartingBase:\n  facilities: []\n  crafts: []\n  items: {}\ninterfaces:\n  - type: mainMenu\n  - type: newGameMenu\n  - type: newBattleMenu\n  - type: battlescape\n    elements:\n      - id: icons\n        size: [320, 56]\n  - type: inventory\n    elements:\n      - id: weight\n        color: 61\n      - id: grid\n        color: 8\n  - type: geoscape\n    elements:\n      - id: slackingIndicator\n      - id: trainingIndicator\n  - type: pauseMenu\n  - type: saveMenus\n  - type: baseNaming\n  - type: placeFacility\n  - type: selectFacility\n  - type: basescape\nfacilities:\n  - type: STR_SF_ACCESS_LIFT\n    lift: true\n    fireSound: -1\n    hitSound: -1\n    placeSound: -1\nglobe:\n  polygons:\n    - [0, -15, -15, 15, -15, 0, 15]\n  textures:\n    - id: 0\n      terrain: []\n"

func validRoot(t *testing.T) string{
  t.Helper()
  root:=t.TempDir()
  path:=filepath.Join(root,"data","SYNTHETIC","multiplayer")
  if err:=os.MkdirAll(path,0700);err!=nil{t.Fatal(err)}
  png:=[]byte{137,80,78,71,13,10,26,10,0,0,0,0}
  markers:=make([]byte,26)
  copy(markers,[]byte{137,80,78,71,13,10,26,10})
  copy(markers[16:],[]byte{0,0,0,27,0,0,0,3,8,3})
  if err:=os.WriteFile(filepath.Join(path,"globe-markers.png"),markers,0600);err!=nil{t.Fatal(err)}
  for _,name:=range []string{"base.png","avatar.png"} {
    if err:=os.WriteFile(filepath.Join(path,name),png,0600);err!=nil{t.Fatal(err)}
  }
  geo:=filepath.Join(root,"data","SYNTHETIC","GEODATA")
  if err:=os.MkdirAll(geo,0700);err!=nil{t.Fatal(err)}
  pal:=make([]byte,5*256*3)
  for bank:=0;bank<5;bank++ {
    off:=(bank*256+1)*3
    pal[off]=61;pal[off+1]=60;pal[off+2]=52
  }
  if err:=os.WriteFile(filepath.Join(geo,"PALETTES.DAT"),pal,0600);err!=nil{t.Fatal(err)}
  sound:=filepath.Join(root,"data","SYNTHETIC","SOUND")
  if err:=os.MkdirAll(sound,0700);err!=nil{t.Fatal(err)}
  wav:=make([]byte,48);copy(wav,"RIFF");copy(wav[8:],"WAVE")
  for _,name:=range []string{"gmgeo1.wav","gminter.wav","gmtactic1.wav","gmenbase.wav"}{
    if err:=os.WriteFile(filepath.Join(sound,name),wav,0600);err!=nil{t.Fatal(err)}
  }
  unit:=filepath.Join(root,"data","SYNTHETIC","UNITS")
  if err:=os.MkdirAll(unit,0700);err!=nil{t.Fatal(err)}
  for _,ext:=range []string{".PCK",".TAB"}{
    if err:=os.WriteFile(filepath.Join(unit,"SIGNALRELAY"+ext),[]byte{0,1},0600);err!=nil{t.Fatal(err)}
  }
  for _,artifact:=range []struct{dir,name string;raw []byte}{
    {"TERRAIN","BLANKS.MCD",make([]byte,124)},
    {"MAPS","SFREL00.MAP",func()[]byte{b:=make([]byte,403);b[0]=10;b[1]=10;b[2]=1;return b}()},
    {"ROUTES","SFREL00.RMP",make([]byte,96)},
  }{
    dir:=filepath.Join(root,"data","SYNTHETIC",artifact.dir)
    if err:=os.MkdirAll(dir,0700);err!=nil{t.Fatal(err)}
    if err:=os.WriteFile(filepath.Join(dir,artifact.name),artifact.raw,0600);err!=nil{t.Fatal(err)}
  }
  langDir:=filepath.Join(root,"user","mods","rdte-synthetic-master","Language")
  if err:=os.MkdirAll(langDir,0700);err!=nil{t.Fatal(err)}
  lang:="en-US:\n"
  for _,key:=range []string{"STR_MISSION","STR_CRAFT","STR_EQUIP_CRAFT","STR_TERROR_SITE","STR_SF_RELAY_DEPLOYMENT_BRIEFING"}{
    lang+="  "+key+": \"Readable test translation\"\n"
  }
  if err:=os.WriteFile(filepath.Join(langDir,"en-US.yml"),[]byte(lang),0600);err!=nil{t.Fatal(err)}
  return root
}
func TestNecessarySourceConditions(t *testing.T){
  v:=Check([]byte(fixture),[]byte("playIntro: false"),validRoot(t))
  if len(v.Missing)!=0||v.EngineExecuted||v.GameplayQualified{t.Fatal(v)}
}
func TestCountryAbsentFailsClosed(t *testing.T){
  bad:=strings.Replace(fixture,"countries:\n","missingCountries:\n",1)
  v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  if !strings.Contains(strings.Join(v.Missing,","),"countries:nonempty"){t.Fatal(v)}
}
func TestGlobeAndGeoscapeMissingFailsClosed(t *testing.T){
  bad:=strings.Replace(fixture,"    - id: 0","    - id: 9",1)
  bad=strings.Replace(bad,"  - type: geoscape","  - type: unrelated",1)
  v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  errors:=strings.Join(v.Missing,",")
  if !strings.Contains(errors,"globe:nonempty-textures")||!strings.Contains(errors,"interfaces:geoscape"){t.Fatal(v)}
}
func TestArtAndRetailDirectoryGuards(t *testing.T){
  root:=validRoot(t)
  os.Remove(filepath.Join(root,"data","SYNTHETIC","multiplayer","base.png"))
  os.MkdirAll(filepath.Join(root,"data","UFO"),0700)
  v:=Check([]byte(fixture),[]byte("playIntro: false"),root)
  errors:=strings.Join(v.Missing,",")
  if !strings.Contains(errors,"original-coop-art:base.png")||!strings.Contains(errors,"no-retail-dir:UFO"){t.Fatal(v)}
}
func TestArmorCorpseAndUnitSheetRequired(t *testing.T){
 root:=validRoot(t)
 os.Remove(filepath.Join(root,"data","SYNTHETIC","UNITS","SIGNALRELAY.TAB"))
 bad:=strings.Replace(fixture,"    corpseBattle: [STR_SF_SURVEY_SHELL]","    corpseBattle: []",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),root)
 errors:=strings.Join(v.Missing,",")
 if !strings.Contains(errors,"armors:one-corpse")||!strings.Contains(errors,"units:signal-relay.TAB"){t.Fatal(v)}
}
func TestTacticalAndMusicMissingFailClosed(t *testing.T){
  bad:=strings.Replace(fixture,"  - type: STR_SF_RESPONDER","  - type: UNKNOWN",1)
  root:=validRoot(t)
  os.Remove(filepath.Join(root,"data","SYNTHETIC","SOUND","gmgeo1.wav"))
  v:=Check([]byte(bad),[]byte("playIntro: false"),root)
  errText:=strings.Join(v.Missing,",")
  if !strings.Contains(errText,"crafts:STR_SF_RESPONDER")||!strings.Contains(errText,"original-music:gmgeo1.wav"){t.Fatal(v)}
}
func TestPreflightNeverClaimsNativeGameplay(t *testing.T){
  v:=Check([]byte(fixture),[]byte("playIntro: false"),validRoot(t))
  if v.EvidenceClass!="SOURCE_DERIVED_NECESSARY_FIXTURE_CHECK_NOT_ENGINE"||v.EngineExecuted||v.GameplayQualified{t.Fatal(v)}
}

func TestBasePlacementNeedsOriginalLand(t *testing.T) {
 bad:=strings.Replace(fixture,"    - [0, -15, -15, 15, -15, 0, 15]","    - [0, 0, 0, 0, 0, 0, 0]",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"globe:original-land-triangle"){t.Fatal(v)}
}

func TestBaseNamingAndLiftAreNecessary(t *testing.T){
 bad:=strings.Replace(fixture,"  - type: baseNaming", "  - type: missingBaseNaming",1)
 bad=strings.Replace(bad,"    lift: true","    lift: false",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 missing:=strings.Join(v.Missing,",")
 if !strings.Contains(missing,"interfaces:baseNaming")||!strings.Contains(missing,"facilities:original-access-lift"){t.Fatal(v)}
}

func TestAccessLiftMustNotReferenceAbsentRetailSounds(t *testing.T){
 bad:=strings.Replace(fixture,"    hitSound: -1","    hitSound: 0",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"facilities:no-retail-GEO-CAT-index"){t.Fatal(v)}
}

func TestGroundInventoryKindIsRequired(t *testing.T) {
  bad:=strings.Replace(fixture,"  - id: STR_GROUND","  - id: STR_MISSING_GROUND",1)
  v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  if !strings.Contains(strings.Join(v.Missing,","),"invs:STR_GROUND:kind"){t.Fatal(v)}
  bad=strings.Replace(fixture,"    y: 140\n    type: 2","    y: 140\n    type: 1",1)
  v=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  if !strings.Contains(strings.Join(v.Missing,","),"invs:STR_GROUND:kind"){t.Fatal(v)}
}
func TestHandInventoryAndPackSlotsRequired(t *testing.T) {
  bad:=strings.Replace(fixture,"  - id: STR_RIGHT_HAND","  - id: STR_MISSING_HAND",1)
  bad=strings.Replace(bad,"    slots: [[0, 0], [1, 0]]","    slots: []",1)
  v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  missing:=strings.Join(v.Missing,",")
  if !strings.Contains(missing,"invs:STR_RIGHT_HAND:kind")||!strings.Contains(missing,"invs:STR_BELT:slots"){t.Fatal(v)}
}

func TestNewBattleRequiresNonemptyAlienMission(t *testing.T) {
  bad:=strings.Replace(fixture,"alienMissions:\n  - type: STR_SF_SIGNAL_PATROL\n","",1)
  v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  if !strings.Contains(strings.Join(v.Missing,","),"alienMissions:nonempty-synthetic-patrol"){t.Fatal(v)}
  bad=strings.Replace(fixture,"  - type: STR_SF_SIGNAL_PATROL","  - type: UNKNOWN",1)
  v=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
  if !strings.Contains(strings.Join(v.Missing,","),"alienMissions:nonempty-synthetic-patrol"){t.Fatal(v)}
}

func TestMissingTacticalDefaultScriptFailsClosed(t *testing.T){
 bad:=strings.Replace(fixture,"  - type: DEFAULT","  - type: UNKNOWN",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"mapScripts:DEFAULT-original-addBlock"){t.Fatal(v)}
 bad=strings.Replace(fixture,"      - type: addBlock","      - type: addLine",1)
 v=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"mapScripts:DEFAULT-original-addBlock"){t.Fatal(v)}
}

func TestSyntheticOriginRequiresRealCountryAndRegion(t *testing.T){
 bad:=strings.Replace(fixture,"      - [0, 360, -65, 65]","      - [10, 340, -65, 65]",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"countries:contains-origin"){t.Fatal(v)}
 bad=strings.Replace(fixture,"      - [0, 360, -65, 65]","      - [10, 340, -65, 65]",-1)
 v=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"regions:contains-origin"){t.Fatal(v)}
}

func TestMissingOriginalTerrainBinaryFailsClosed(t *testing.T){
 root:=validRoot(t)
 if err:=os.Remove(filepath.Join(root,"data","SYNTHETIC","TERRAIN","BLANKS.MCD"));err!=nil{t.Fatal(err)}
 m:=Check([]byte(fixture),[]byte("playIntro: false"),root)
 if !strings.Contains(strings.Join(m.Missing,","),"terrain:original-two-record-MCD"){t.Fatal(m)}
}
func TestBadMapAndRouteLengthsFailClosed(t *testing.T){
 root:=validRoot(t)
 if err:=os.WriteFile(filepath.Join(root,"data","SYNTHETIC","MAPS","SFREL00.MAP"),[]byte{10,11,1},0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(root,"data","SYNTHETIC","ROUTES","SFREL00.RMP"),[]byte{0},0600);err!=nil{t.Fatal(err)}
 m:=Check([]byte(fixture),[]byte("playIntro: false"),root)
 errors:=strings.Join(m.Missing,",")
 if !strings.Contains(errors,"maps:original-10x10-relay")||!strings.Contains(errors,"routes:original-four-node-relay"){t.Fatal(m)}
}

func TestAlienDeploymentRequiresNonemptyItemSet(t *testing.T){
 bad:=strings.Replace(fixture,"        itemSets: [[STR_RDTE_TOOL]]","        itemSets: []",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"alienDeployments:original-tool-itemset"){t.Fatal(v)}
}

func TestOriginalGlobeMarkersAreNecessary(t *testing.T){
 root:=validRoot(t)
 if err:=os.Remove(filepath.Join(root,"data","SYNTHETIC","multiplayer","globe-markers.png"));err!=nil{t.Fatal(err)}
 bad:=strings.Replace(fixture,"  - type: GlobeMarkers","  - type: INVALID",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),root)
 e:=strings.Join(v.Missing,",")
 if !strings.Contains(e,"extraSprites:globe-markers-nine-frames")||!strings.Contains(e,"original-coop-art:globe-markers-27x3"){t.Fatal(v)}
}

func TestGlobeMarkerRequiresIndexedPNG(t *testing.T){
 root:=validRoot(t)
 path:=filepath.Join(root,"data","SYNTHETIC","multiplayer","globe-markers.png")
 b,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 b[25]=6
 if err:=os.WriteFile(path,b,0600);err!=nil{t.Fatal(err)}
 r:=Check([]byte(fixture),[]byte("playIntro: false"),root)
 if !strings.Contains(strings.Join(r.Missing,","),"original-coop-art:globe-markers-27x3"){t.Fatal(r)}
}

func TestUnresolvedNativeBattleTranslationRejected(t *testing.T){
 root:=validRoot(t)
 path:=filepath.Join(root,"user","mods","rdte-synthetic-master","Language","en-US.yml")
 b,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
 damaged:=strings.Replace(string(b),"  STR_MISSION:","  MISSING_MISSION:",1)
 if err:=os.WriteFile(path,[]byte(damaged),0600);err!=nil{t.Fatal(err)}
 v:=Check([]byte(fixture),[]byte("playIntro: false"),root)
 if !strings.Contains(strings.Join(v.Missing,","),"language:STR_MISSION"){t.Fatal(v)}
}

func TestNativeGlyphPaletteRejectsUnreadableForeground(t *testing.T){
 root:=validRoot(t)
 path:=filepath.Join(root,"data","SYNTHETIC","GEODATA","PALETTES.DAT")
 b,e:=os.ReadFile(path);if e!=nil{t.Fatal(e)}
 b[3]=1;b[4]=3;b[5]=7
 if e=os.WriteFile(path,b,0600);e!=nil{t.Fatal(e)}
 v:=Check([]byte(fixture),[]byte("playIntro: false"),root)
 if !strings.Contains(strings.Join(v.Missing,","),"ui:original-palette-readable-foreground"){t.Fatal(v)}
}

func TestNativeTacticalHUDRequiredBeforeBriefingConfirmation(t *testing.T){
 bad:=strings.Replace(fixture,"  - type: battlescape","  - type: MISSING",1)
 v:=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"interfaces:battlescape"){t.Fatal(v)}
 bad=strings.Replace(fixture,"size: [320, 56]","size: [1, 1]",1)
 v=Check([]byte(bad),[]byte("playIntro: false"),validRoot(t))
 if !strings.Contains(strings.Join(v.Missing,","),"interfaces:battlescape-icons"){t.Fatal(v)}
}
