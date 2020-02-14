package whandlers

import (
	"encoding/json"
	"net/http"

	"../data"
	u "../utils"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	_ = json.NewDecoder(r.Body).Decode(account)
	//str:= account.Privilege.ToSqlStrUpdate("accounts","MMM")
	//data.GetDB().Exec(str)
	//fmt.Println(str)
	u.Respond(w, r, u.Message(true, "Chil its ok"))
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, "MakeTest")
	if flag {
		location := &data.Locations{}
		_ = json.NewDecoder(r.Body).Decode(location)
		resp["Test"] = "OK!"
		boxPoint, _ := location.MakeBoxPoint()
		resp["boxPoint"] = boxPoint
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
