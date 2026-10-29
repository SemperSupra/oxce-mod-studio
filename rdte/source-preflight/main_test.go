package main
import (
  "os"
  "path/filepath"
  "strings"
  "testing"
)
const fixture = "countries:\n  - type: STR_SF_COUNCIL\n    fundingBase: 500\n    areas:\n      - [10, 340, -65, 65]\nregions:\n  - type: STR_SF_AREA\n    areas:\n      - [10, 340, -65, 65]\nstartingBase:\n  facilities: []\n  crafts: []\n  items: {}\ninterfaces:\n  - type: mainMenu\n  - type: newGameMenu\n  - type: newBattleMenu\n  - type: geoscape\n    elements:\n      - id: slackingIndicator\n      - id: trainingIndicator\n  - type: pauseMenu\n  - type: saveMenus\nglobe:\n  textures:\n    - id: 0\n      terrain: []\n"

func validRoot(t *testing.T) string{
  t.Helper()
  root:=t.TempDir()
  path:=filepath.Join(root,"data","SYNTHETIC","multiplayer")
  if err:=os.MkdirAll(path,0700);err!=nil{t.Fatal(err)}
  png:=[]byte{137,80,78,71,13,10,26,10,0,0,0,0}
  for _,name:=range []string{"base.png","avatar.png"} {
    if err:=os.WriteFile(filepath.Join(path,name),png,0600);err!=nil{t.Fatal(err)}
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
func TestPreflightNeverClaimsNativeGameplay(t *testing.T){
  v:=Check([]byte(fixture),[]byte("playIntro: false"),validRoot(t))
  if v.EvidenceClass!="SOURCE_DERIVED_NECESSARY_FIXTURE_CHECK_NOT_ENGINE"||v.EngineExecuted||v.GameplayQualified{t.Fatal(v)}
}
