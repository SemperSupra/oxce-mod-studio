package main

// All tactical artwork is originally constructed from geometric primitives.
// This is a 32x40 OXCE 8-bit indexed unit sheet, not copied from any retail
// X-COM game. We reserve 0 for transparency and 254/255 for PCK control bytes.
const battleSpriteW,battleSpriteH,battleSpriteFrames=32,40,512

func drawSurveyUnit(frame int) []uint8 {
 px:=make([]uint8,battleSpriteW*battleSpriteH)
 plot:=func(x,y int,v uint8){
  if x>=0&&x<battleSpriteW&&y>=0&&y<battleSpriteH{px[y*battleSpriteW+x]=v}
 }
 // Suit body, visor, legs and boots: original angular survey-rescue outfit.
 // Directional marker and a pulse phase vary deterministically with frame.
 facing:=(frame/8)%8
 for y:=6;y<=15;y++{
  for x:=12;x<=20;x++{
   v:=uint8(3)
   if x==12||x==20||y==6||y==15{v=1}
   if y>=9&&y<=12&&x>=14&&x<=18{v=5}
   plot(x,y,v)
  }
 }
 for y:=15;y<=25;y++{
  width:=7+(y-15)/5
  for x:=16-width;x<=16+width;x++{
   v:=uint8(3)
   if x==16-width||x==16+width||y==25{v=1}
   if y==19&&x>=12&&x<=20{v=6}
   if x==16+(facing%3)-1&&y<19{v=4}
   plot(x,y,v)
  }
 }
 for y:=25;y<36;y++{
  for x:=11;x<=14;x++{v:=uint8(3);if x==11||y==35{v=1};plot(x,y,v)}
  for x:=18;x<=21;x++{v:=uint8(3);if x==21||y==35{v=1};plot(x,y,v)}
 }
 for y:=16;y<28;y++{
  for x:=7;x<=10;x++{v:=uint8(4);if x==7||y==27{v=1};plot(x,y,v)}
  for x:=22;x<=25;x++{v:=uint8(4);if x==25||y==27{v=1};plot(x,y,v)}
 }
 if frame%8<4 {
  for y:=12;y<=14;y++{plot(15,y,6);plot(17,y,6)}
 }else{
  for y:=12;y<=14;y++{plot(14,y,4);plot(18,y,4)}
 }
 return px
}

func surveyUnitFrames()[][]uint8 {
 frames:=make([][]uint8,battleSpriteFrames)
 for i:=range frames{frames[i]=drawSurveyUnit(i)}
 return frames
}
