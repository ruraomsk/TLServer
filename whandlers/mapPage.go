package whandlers

import (
	"../data"
	"../logger"
	u "../utils"
	"encoding/json"
	"net/http"
)

//BuildMainPage собираем данные для залогиневшегося пользователя
var BuildMapPage = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	mapContx := data.ParserInterface(r.Context().Value("info"))
	account.Login = mapContx["login"]

	if account.Login == "Super" {
		u.Respond(w, r, nil)
		return
	}

	flag, resp := FuncAccessCheak(w, r, "BuildMapPage")
	if flag {
		resp = account.GetInfoForUser()
	}
	u.Respond(w, r, resp)
}

//UpdateMapPage обновление информации о попавших в область светофорах
var UpdateMapPage = func(w http.ResponseWriter, r *http.Request) {
	box := &data.BoxPoint{}
	err := json.NewDecoder(r.Body).Decode(box)
	if box.Point0 == box.Point1 {
		logger.Info.Println("mapPage: Impossible coordinates ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Impossible coordinates"))
		return
	}
	if err != nil {
		logger.Info.Println("Invalid request ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := u.Message(true, "Update box data")

	flag, resp := FuncAccessCheak(w, r, "UpdateMapPage")
	if flag {
		tflight := data.GetLightsFromBD(*box)
		resp["tflight"] = tflight
	}
	u.Respond(w, r, resp)
}
