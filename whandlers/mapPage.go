package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//BuildMainPage собираем данные для залогиневшегося пользователя
var BuildMapPage = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	account.Login = mapContx["login"]
	resp := account.GetInfoForUser()
	resp["manageFlag"], _ = data.AccessCheck(mapContx, 1)
	resp["logDeviceFlag"], _ = data.AccessCheck(mapContx, 11)
	u.Respond(w, r, resp)
}

//UpdateMapPage обновление информации о попавших в область светофорах
var UpdateMapPage = func(w http.ResponseWriter, r *http.Request) {
	box := &data.BoxPoint{}
	err := json.NewDecoder(r.Body).Decode(box)
	if box.Point0 == box.Point1 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Impossible coordinates"))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	tflight := data.GetLightsFromBD(*box)
	resp := u.Message(true, "Update map data")
	resp["DontWrite"] = "true"
	resp["tflight"] = tflight
	u.Respond(w, r, resp)
}

//LocationButtonMapPage обработка запроса на получение новых координат отрисовки рабочей области
var LocationButtonMapPage = func(w http.ResponseWriter, r *http.Request) {
	location := &data.Locations{}
	err := json.NewDecoder(r.Body).Decode(location)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	boxPoint, err := location.MakeBoxPoint()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := u.Message(true, "Jump to Location!")
	resp["DontWrite"] = "true"
	resp["boxPoint"] = boxPoint
	u.Respond(w, r, resp)
}
