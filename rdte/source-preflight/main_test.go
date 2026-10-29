package main
import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)
const fixture = "invs:\n  - id: STR_GROUND\n    x: 0\n    y: 140\n    type: 2\n  - id: STR_RIGHT_HAND\n    type: 1\n  - id: STR_LEFT_HAND\n    type: 1\n  - id: STR_BELT\n    type: 0\n    slots: [[0, 0], [1, 0]]\n  - id: STR_BACK_PACK\n    type: 0\n    slots: [[0, 0], [1, 0]]\nmusics:\n  - type: GMGEO1\ncrafts:\n  - type: STR_SF_RESPONDER\nalienMissions:\n  - type: STR_SF_SIGNAL_PATROL\nalienDeployments:\n  - type: STR_SF_RELAY_DEPLOYMENT\nmapScripts:\n  - type: DEFAULT\n    commands:\n      - type: addBlock\n        size: [1, 1]\n        blocks: [0]\nterrains:\n  - name: SF_RELAY\nsoldiers:\n  - type: STR_SF_SURVEYOR\narmors:\n  - type: STR_SF_SURVEY_ARMOR\n    corpseBattle: [STR_SF_SURVEY_SHELL]\n    loftempsSet: [3]\nalienRaces:\n  - id: STR_SF_QUIET_ARRAY\nitems:\n  - type: STR_SF_SURVEY_SHELL\nunits:\n  - type: STR_SF_PROBE\nalienItemLevels:\n  - [0, 0, 0]\ncountries:\n  - type: STR_SF_COUNCIL\n    fundingBase: 500\n    areas:\n      - [0, 360, -65, 65]\nregions:\n  - type: STR_SF_AREA\n    areas:\n      - [0, 360, -65, 65]\nstartingBase:\n  facilities: []\n  crafts: []\n  items: {}\ninterfaces:\n  - type: mainMenu\n  - type: newGameMenu\n  - type: newBattleMenu\n  - type: geoscape\n    elements:\n      - id: slackingIndicator\n      - id: trainingIndicator\n  - type: pauseMenu\n  - type: saveMenus\n  - type: baseNaming\n  - type: placeFacility\n  - type: selectFacility\n  - type: basescape\nfacilities:\n  - type: STR_SF_ACCESS_LIFT\n    lift: true\n    fireSound: -1\n    hitSound: -1\n    placeSound: -1\nglobe:\n  polygons:\n    - [0, -15, -15, 15, -15, 0, 15]\n  textures:\n    - id: 0\n      terrain: []\n"

func validRoot(t *testing.T) string{
  t.Helper()
  root:=t.TempDir()
  path:=filepath.Join(root,"data","SYNTHETIC","multiplayer")
  if err:=os.MkdirAll(path,0700);err!=nil{t.Fatal(err)}
  png:=[]byte{137,80,78,71,13,10,26,10,0,0,0,0}
  for _,name:=range []string{"base.png","avatar.png"} {
    if err:=os.WriteFile(filepath.Join(path,name),png,0600);err!=nil{t.Fatal(err)}
  }
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
