package main

// Original independent 8-bit PCK/TAB fixture codec for source-only bootstrap
// qualification. Format semantics are pinned to MeridianOXC/OpenXcom's
// SurfaceSet::loadPck reader. No commercial game data is copied or required.
import (
 "bytes"
 "encoding/binary"
 "errors"
 "fmt"
)
var errSpriteFormat=errors.New("invalid original indexed PCK sprite format")

// encodeIndexedPCK records 32-bit little-endian TAB offsets (first zero),
// one leading skipped-row byte per frame, 254+count transparent pixel runs,
// literal palette indices 1..253, and 255 frame terminators.
func encodeIndexedPCK(frames [][]byte,w,h int)([]byte,[]byte,error){
 if w<=0||h<=0||w>1024||h>1024||len(frames)==0||len(frames)>4096{return nil,nil,errSpriteFormat}
 area:=w*h
 if area>1024*1024{return nil,nil,errSpriteFormat}
 pck:=bytes.Buffer{}
 tab:=bytes.Buffer{}
 for i,frame:=range frames{
  if len(frame)!=area{return nil,nil,fmt.Errorf("%w: frame %d dimension",errSpriteFormat,i)}
  if pck.Len()>int(^uint32(0)){return nil,nil,errSpriteFormat}
  if err:=binary.Write(&tab,binary.LittleEndian,uint32(pck.Len()));err!=nil{return nil,nil,err}
  pck.WriteByte(0)
  for at:=0;at<len(frame); {
   pixel:=frame[at]
   if pixel>=254{return nil,nil,fmt.Errorf("%w: reserved palette index %d",errSpriteFormat,pixel)}
   if pixel!=0{pck.WriteByte(pixel);at++;continue}
   length:=0
   for at<len(frame)&&frame[at]==0&&length<255{at++;length++}
   pck.WriteByte(254);pck.WriteByte(byte(length))
  }
  pck.WriteByte(255)
 }
 return pck.Bytes(),tab.Bytes(),nil
}

// decodeIndexedPCKFixture implements a deliberately strict independent read
// of the subset produced above. This is not an engine execution/acceptance.
func decodeIndexedPCKFixture(pck,tab []byte,w,h int)([][]byte,error){
 if w<=0||h<=0||len(tab)==0||len(tab)%4!=0{return nil,errSpriteFormat}
 count:=len(tab)/4
 result:=make([][]byte,0,count)
 for f:=0;f<count;f++{
  start:=int(binary.LittleEndian.Uint32(tab[f*4:]))
  end:=len(pck)
  if f+1<count{end=int(binary.LittleEndian.Uint32(tab[(f+1)*4:]))}
  if start<0||start>=len(pck)||end<=start||end>len(pck){return nil,errSpriteFormat}
  if f==0&&start!=0{return nil,errSpriteFormat}
  data:=pck[start:end]
  pos:=int(data[0])*w
  if pos>w*h{return nil,errSpriteFormat}
  frame:=make([]byte,w*h)
  terminated:=false
  for i:=1;i<len(data);i++{
   value:=data[i]
   if value==255{if i!=len(data)-1{return nil,errSpriteFormat};terminated=true;break}
   if value==254{
    i++;if i>=len(data){return nil,errSpriteFormat}
    pos+=int(data[i]);if pos>len(frame){return nil,errSpriteFormat}
   }else{
    if pos>=len(frame){return nil,errSpriteFormat}
    frame[pos]=value;pos++
   }
  }
  if !terminated{return nil,errSpriteFormat}
  result=append(result,frame)
 }
 return result,nil
}
