package whandlers

import (
	"encoding/json"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
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
		var stateData agS_pudge.Cross
		err := json.NewDecoder(r.Body).Decode(&stateData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, "Invalid request"))
			return
		}
		mapContx := u.ParserInterface(r.Context().Value("info"))
		resp["AAA"] = data.CreateCrossData(stateData, mapContx)
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
