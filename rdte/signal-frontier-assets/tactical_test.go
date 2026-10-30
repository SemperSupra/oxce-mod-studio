package main
import "testing"
func TestSurveyFrames(t *testing.T){
 frames:=surveyUnitFrames()
 if len(frames)!=battleSpriteFrames{t.Fatal("count")}
 for _,frame:=range frames{
  if len(frame)!=battleSpriteW*battleSpriteH{t.Fatal("size")}
  count:=0
  for _,v:=range frame{if v>253{t.Fatal("invalid index")};if v>0{count++}}
  if count<150{t.Fatal("not visible")}
 }
}
