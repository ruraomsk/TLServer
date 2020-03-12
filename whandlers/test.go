package whandlers

import (
	"github.com/JanFant/TLServer/data"
	"net/http"

	u "github.com/JanFant/TLServer/utils"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, "MakeTest")
	if flag {

		resp = data.TestNewRoleSystem()
		resp["BLIA!!!"] = "Blia!!!"
	}
	u.Respond(w, r, resp)
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, "MakeTest")
	if flag {

		resp["BLIA!!!"] = "Blia!!!"
	}
	u.Respond(w, r, resp)
})

////getRegionAreaPoints создание мапы регионов с точками в облость
//func getRegionAreaPoints() (regPoint map[string]map[string]Point) {
//	regPoint = make(map[string]map[string]Point)
//	for _, nameReg := range CacheInfo.mapRegion {
//		if nameReg == "Все регионы" {
//			continue
//		}
//		for _, nameArea := range CacheInfo.mapArea[nameReg] {
//			if _, ok := regPoint[nameReg]; !ok {
//				regPoint[nameReg] = make(map[string]Point)
//			}
//			regPoint[nameReg][nameArea] = Point{2, 3}
//		}
//	}
//	return
//}
