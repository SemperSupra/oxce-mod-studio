package main
import "testing"
func TestOriginalNavigableFloor(t *testing.T){
 pck,tab,mcd,err:=relayGroundAssets();if err!=nil{t.Fatal(err)}
 if len(mcd)!=124||len(tab)!=8||len(pck)!=2*(terrainW*terrainH+2){t.Fatal("bad ground file dimensions")}
 for i:=0;i<8;i++{if mcd[62+i]!=1{t.Fatal("ground MCD frame missing")}}
 if mcd[62+39]==0||mcd[62+40]==0||mcd[62+41]==0{t.Fatal("ground is not walkable")}
 floor:=drawRelayGround();n:=0
 for _,v:=range floor{if v>=254{t.Fatal("reserved pixel value")};if v!=0{n++}}
 if n<300{t.Fatal("invisible ground")}
}
