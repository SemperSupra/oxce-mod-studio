package main

// Original synthetic isometric tile and its two 62-byte MCD entries.
// Source interpretation: pinned public MapDataSet::loadData reads an 8-frame
// index array, movement costs, armor, and Tile_Type. Zero remains BLANKS,
// while index one is a navigable 32x40 signal-site ground tile.
const terrainW,terrainH=32,40
func drawRelayGround()[]uint8{
 px:=make([]uint8,terrainW*terrainH)
 for y:=0;y<terrainH;y++{
  for x:=0;x<terrainW;x++{
   dx:=x-16;if dx<0{dx=-dx}
   dy:=y-19;if dy<0{dy=-dy}
   if dx+2*dy>28{continue}
   v:=uint8(3)
   if dx+2*dy>=26{v=1
   }else if (x+y)%7==0{v=4
   }else if (x/4+y/3)%3==0{v=2}
   px[y*terrainW+x]=v
  }
 }
 return px
}
func relayGroundAssets()([]byte,[]byte,[]byte,error){
 blank:=make([]uint8,terrainW*terrainH)
 pck,tab,err:=encodePCK([][]uint8{blank,drawRelayGround()},terrainW,terrainH)
 if err!=nil{return nil,nil,nil,err}
 mcd:=make([]byte,2*62)
 const obj=62
 for i:=0;i<8;i++{mcd[obj+i]=1}
 // MCD is packed: TU_Walk=39, TU_Slide=40, TU_Fly=41, Armor=42,
 // Tile_Type=53. Tile_Type zero is O_FLOOR. No_Floor remains false.
 mcd[obj+39]=4;mcd[obj+40]=4;mcd[obj+41]=4;mcd[obj+42]=8
 return pck,tab,mcd,nil
}
