package main

import (
 "crypto/sha256"
 "encoding/binary"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "image"
 "image/color"
 "image/png"
 "math"
 "os"
 "path/filepath"
 "sort"
)

type asset struct { Path string `json:"path"`; Bytes int `json:"bytes"`; SHA256 string `json:"sha256"` }
type manifest struct {
 SchemaVersion int `json:"schemaVersion"`
 OriginalContent bool `json:"originalContent"`
 ActualEngineQualified bool `json:"actualEngineQualified"`
 Files []asset `json:"files"`
}
const w,h=16,16
func draw() []uint8 {
 pixels:=make([]uint8,w*h)
 for y:=0;y<h;y++ { for x:=0;x<w;x++ {
  dx,dy:=x-8,y-7
  v:=uint8(1)
  if (x+y)%7==0 {v=2}
  if dx*dx+dy*dy<=36 {v=3}
  if dx*dx+dy*dy<=16 {v=4}
  if x>=7&&x<=9 && y<5 {v=5}
  if y==12&&x>=3&&x<=12 {v=6}
  pixels[y*w+x]=v
 }}
 return pixels
}
func encodePCK(frames [][]uint8,width,height int)([]byte,[]byte,error){
 if len(frames)==0 || width<=0 || height<=0{return nil,nil,fmt.Errorf("bad dimensions")}
 pck,tab:=make([]byte,0),make([]byte,len(frames)*4)
 for i,f:=range frames {
  if len(f)!=width*height{return nil,nil,fmt.Errorf("frame size")}
  for _,v:=range f {if v>=254{return nil,nil,fmt.Errorf("PCK reserved byte")}}
  binary.LittleEndian.PutUint32(tab[i*4:],uint32(len(pck)))
  pck=append(pck,0)
  pck=append(pck,f...)
  pck=append(pck,255)
 }
 return pck,tab,nil
}
func paletteDAT()[]byte{
 data:=make([]byte,5*256*3)
 for bank:=0;bank<5;bank++{for i:=0;i<256;i++{
  at:=(bank*256+i)*3
  data[at]=byte((i+bank*3)%64)
  data[at+1]=byte((i*2+bank*5)%64)
  data[at+2]=byte((i*3+bank*7)%64)
 }}
 return data
}
func spritePNG(pix []uint8)([]byte,error){
 c:=color.Palette{
  color.NRGBA{R:0,G:0,B:0,A:0},
  color.NRGBA{R:18,G:33,B:65,A:255},
  color.NRGBA{R:32,G:60,B:88,A:255},
  color.NRGBA{R:44,G:171,B:196,A:255},
  color.NRGBA{R:103,G:232,B:226,A:255},
  color.NRGBA{R:254,G:191,B:81,A:255},
  color.NRGBA{R:232,G:226,B:203,A:255},
 }
 img:=image.NewPaletted(image.Rect(0,0,w,h),c)
 copy(img.Pix,pix)
 var out byteBuffer
 if err:=png.Encode(&out,img);err!=nil{return nil,err}
 return out.buf,nil
}
type byteBuffer struct{buf []byte}
func(b *byteBuffer) Write(p []byte)(int,error){b.buf=append(b.buf,p...);return len(p),nil}
func soundWAV()[]byte{
 const sampleRate,n=22050,4410
 b:=make([]byte,44+2*n)
 copy(b,"RIFF");binary.LittleEndian.PutUint32(b[4:],uint32(len(b)-8))
 copy(b[8:],"WAVEfmt ")
 binary.LittleEndian.PutUint32(b[16:],16)
 binary.LittleEndian.PutUint16(b[20:],1);binary.LittleEndian.PutUint16(b[22:],1)
 binary.LittleEndian.PutUint32(b[24:],sampleRate);binary.LittleEndian.PutUint32(b[28:],sampleRate*2)
 binary.LittleEndian.PutUint16(b[32:],2);binary.LittleEndian.PutUint16(b[34:],16)
 copy(b[36:],"data");binary.LittleEndian.PutUint32(b[40:],uint32(n*2))
 for i:=0;i<n;i++{
  t:=float64(i)/sampleRate
  env:=math.Pow(1-float64(i)/n,2)
  wave:=math.Sin(2*math.Pi*(360+800*t)*t)
  val:=int16(math.Round(10000*env*wave))
  binary.LittleEndian.PutUint16(b[44+i*2:],uint16(val))
 }
 return b
}
// CAT indexes are little-endian (offset,size) pairs. Each embedded WAV has
// a zero-length name prefix, as required by OXCE's SoundSet::loadCatByIndex.
func encodeCAT(wav []byte,n int)[]byte{
 table:=make([]byte,n*8)
 out:=append([]byte{},table...)
 for i:=0;i<n;i++{
  offset:=len(out)
  out=append(out,byte(0))
  out=append(out,wav...)
  binary.LittleEndian.PutUint32(out[i*8:],uint32(offset))
  binary.LittleEndian.PutUint32(out[i*8+4:],uint32(1+len(wav)))
 }
 return out
}
func generate(root string) error{
 pix:=draw()
 // Render original full-size 32x40 directional unit art at all 512
 // engine-consumable frame indices, not two 16x16 technical glyphs.
 pck,tab,err:=encodePCK(surveyUnitFrames(),battleSpriteW,battleSpriteH)
 if err!=nil{return err}
 img,err:=spritePNG(pix);if err!=nil{return err}
 outputs:=map[string][]byte{
  "GEODATA/PALETTES.DAT":paletteDAT(),
  "UNITS/SIGNALRELAY.PCK":pck,
  "UNITS/SIGNALRELAY.TAB":tab,
  "Art/signal-relay.png":img,
  "Sound/signal-pulse.wav":soundWAV(),
  "SOUND/SAMPLE.CAT":encodeCAT(soundWAV(),4),
  "SOUND/SAMPLE2.CAT":encodeCAT(soundWAV(),4),
 }
 paths:=make([]string,0,len(outputs));for p:=range outputs{paths=append(paths,p)};sort.Strings(paths)
 m:=manifest{SchemaVersion:1,OriginalContent:true,ActualEngineQualified:false}
 for _,p:=range paths{
  raw:=outputs[p];dest:=filepath.Join(root,p)
  if err:=os.MkdirAll(filepath.Dir(dest),0755);err!=nil{return err}
  if err:=os.WriteFile(dest,raw,0644);err!=nil{return err}
  sum:=sha256.Sum256(raw)
  m.Files=append(m.Files,asset{Path:p,Bytes:len(raw),SHA256:hex.EncodeToString(sum[:])})
 }
 b,err:=json.MarshalIndent(m,"","  ");if err!=nil{return err}
 return os.WriteFile(filepath.Join(root,"manifest.json"),append(b,byte(10)),0644)
}
func main(){
 out:=".out/signal-frontier-assets";if len(os.Args)>1{out=os.Args[1]}
 if err:=generate(out);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 fmt.Println("generated original indexed art, palette, WAV and CAT assets; ENGINE_NOT_QUALIFIED")
}
