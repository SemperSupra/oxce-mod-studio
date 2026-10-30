package main
import "testing"
func TestModeClassifierRejectsPartialOrWrongInputs(t *testing.T){
 valid:=map[string]interface{}{"campaignType":float64(1),"bases":[]interface{}{"A"},"funds":float64(100),"time":"T"}
 equal:=map[string]interface{}{"campaignType":float64(1),"bases":[]interface{}{"A"},"funds":float64(100),"time":"T"}
 ok,subset:=classifyMode(valid,equal,"shared");if !ok||!subset{t.Fatal("equal shared state rejected")}
 unequal:=map[string]interface{}{"campaignType":float64(1),"bases":[]interface{}{"B"},"funds":float64(100),"time":"T"}
 ok,subset=classifyMode(valid,unequal,"shared");if !ok||subset{t.Fatal("diverged shared bases accepted")}
 absent:=map[string]interface{}{"bases":[]interface{}{"A"},"time":"T"}
 ok,_=classifyMode(valid,absent,"shared");if ok{t.Fatal("missing mode accepted")}
 separate:=map[string]interface{}{"campaignType":float64(0),"bases":[]interface{}{"Other"},"time":"T"}
 ok,subset=classifyMode(separate,separate,"separate");if !ok||subset{t.Fatal("separate-state contract violated")}
 ok,_=classifyMode(separate,valid,"separate");if ok{t.Fatal("cross-mode state accepted")}
}
