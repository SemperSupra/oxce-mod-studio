package main
import(
 "bytes"
 "crypto/sha256"
 "encoding/binary"
 "encoding/hex"
 "encoding/json"
 "image/png"
 "os"
 "path/filepath"
 "testing"
)
func TestDeterministicGeneration(t *testing.T){
 a,b:=t.TempDir(),t.TempDir()
 if err:=generate(a);err!=nil{t.Fatal(err)}
 if err:=generate(b);err!=nil{t.Fatal(err)}
 am,_:=os.ReadFile(filepath.Join(a,"manifest.json"));bm,_:=os.ReadFile(filepath.Join(b,"manifest.json"))
 if !bytes.Equal(am,bm){t.Fatal("nondeterministic manifest")}
 var m manifest;if err:=json.Unmarshal(am,&m);err!=nil{t.Fatal(err)}
 if !m.OriginalContent||m.ActualEngineQualified||len(m.Files)!=5{t.Fatalf("provenance or count %+v",m)}
 for _,v:=range m.Files{
  x,err:=os.ReadFile(filepath.Join(a,v.Path));if err!=nil{t.Fatal(err)}
  y,err:=os.ReadFile(filepath.Join(b,v.Path));if err!=nil{t.Fatal(err)}
  if !bytes.Equal(x,y){t.Fatalf("nondeterministic %s",v.Path)}
  sum:=sha256.Sum256(x)
  if v.Bytes!=len(x)||v.SHA256!=hex.EncodeToString(sum[:]){t.Fatalf("wrong manifest %s",v.Path)}
 }
}
func TestFormatContracts(t *testing.T){
 d:=t.TempDir();if err:=generate(d);err!=nil{t.Fatal(err)}
 pal,_:=os.ReadFile(filepath.Join(d,"GEODATA/PALETTES.DAT"))
 if len(pal)!=5*256*3{t.Fatal("palette length")}
 for _,b:=range pal{if b>=64{t.Fatal("not a six-bit channel")}}
 pck,_:=os.ReadFile(filepath.Join(d,"UNITS/SIGNALRELAY.PCK"))
 tab,_:=os.ReadFile(filepath.Join(d,"UNITS/SIGNALRELAY.TAB"))
 if len(tab)!=8||binary.LittleEndian.Uint32(tab[0:4])!=0||int(binary.LittleEndian.Uint32(tab[4:8]))!=w*h+2{t.Fatal("offsets")}
 if len(pck)!=2*(w*h+2)||pck[0]!=0||pck[w*h+1]!=255{t.Fatal("pck frame")}
 pngBytes,_:=os.ReadFile(filepath.Join(d,"Art/signal-relay.png"))
 im,err:=png.Decode(bytes.NewReader(pngBytes));if err!=nil{t.Fatal(err)}
 if im.Bounds().Dx()!=w||im.Bounds().Dy()!=h{t.Fatal("png bounds")}
 wav,_:=os.ReadFile(filepath.Join(d,"Sound/signal-pulse.wav"))
 if string(wav[0:4])!="RIFF"||string(wav[8:12])!="WAVE"||binary.LittleEndian.Uint32(wav[24:28])!=22050{t.Fatal("wav header")}
}
func TestRejectReservedByte(t *testing.T){
 f:=make([]byte,w*h);f[12]=255
 if _,_,err:=encodePCK([][]uint8{f},w,h);err==nil{t.Fatal("reserved byte allowed")}
 if _,_,err:=encodePCK([][]uint8{{1,2}},w,h);err==nil{t.Fatal("wrong dimensions allowed")}
}
