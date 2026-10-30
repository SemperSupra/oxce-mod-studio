package main

import(
 "bytes"
 "encoding/binary"
 "testing"
)

func TestOriginalIndexedPCKRoundtrip(t *testing.T){
 w,h:=32,40
 blank:=make([]byte,w*h)
 relay:=make([]byte,w*h)
 for y:=5;y<35;y++ {
   for x:=12;x<20;x++ {
     if x==12||x==19||y==5||y==34{relay[y*w+x]=3}else{relay[y*w+x]=1}
   }
 }
 pck,tab,err:=encodeIndexedPCK([][]byte{blank,relay},w,h)
 if err!=nil{t.Fatal(err)}
 if len(tab)!=8||binary.LittleEndian.Uint32(tab[0:4])!=0||int(binary.LittleEndian.Uint32(tab[4:8]))<=0{t.Fatalf("wrong TAB offsets: %v",tab)}
 decoded,err:=decodeIndexedPCKFixture(pck,tab,w,h)
 if err!=nil{t.Fatal(err)}
 if len(decoded)!=2||!bytes.Equal(decoded[0],blank)||!bytes.Equal(decoded[1],relay){t.Fatal("index-preserving PCK/TAB roundtrip failed")}
 if len(pck)>2*w*h+10{t.Fatalf("unexpectedly large RLE: %d",len(pck))}
}

func TestOriginalIndexedPCKMustFailClosed(t *testing.T){
 frame:=make([]byte,32*40)
 frame[0]=254
 if _,_,err:=encodeIndexedPCK([][]byte{frame},32,40);err==nil{t.Fatal("reserved control code was accepted as a palette color")}
 frame[0]=1
 if _,_,err:=encodeIndexedPCK([][]byte{frame},31,40);err==nil{t.Fatal("dimension mismatch accepted")}
 pck,tab,err:=encodeIndexedPCK([][]byte{frame},32,40)
 if err!=nil{t.Fatal(err)}
 if _,err:=decodeIndexedPCKFixture(pck[:len(pck)-1],tab,32,40);err==nil{t.Fatal("missing frame terminator accepted")}
 broken:=append([]byte(nil),tab...)
 binary.LittleEndian.PutUint32(broken[:4],2)
 if _,err:=decodeIndexedPCKFixture(pck,broken,32,40);err==nil{t.Fatal("missing TAB zero origin accepted")}
 if _,err:=decodeIndexedPCKFixture([]byte{0,254,255,255},tab,32,40);err==nil{t.Fatal("overflowing transparent run accepted")}
}
