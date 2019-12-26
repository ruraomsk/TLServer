package whandlers

import (
	"../data"
	"../logger"
	u "../utils"
	"encoding/json"
	"fmt"
	"net/http"
)

//BuildMainPage собираем данные для залогиневшегося пользователя
var BuildMapPage = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	account.Login = fmt.Sprintf("%v", r.Context().Value("user"))
	resp := account.GetInfoForUser()
	u.Respond(w, r, resp)
}

//UpdateMapPage обновление информации о попавших в область светофорах
var UpdateMapPage = func(w http.ResponseWriter, r *http.Request) {
	box := &data.BoxPoint{}
	err := json.NewDecoder(r.Body).Decode(box)
	if box.Point0 == box.Point1 {
		logger.Info.Println("mapPage: Impossible coordinates ", r.RemoteAddr)
		u.Respond(w, r, u.Message(false, "Impossible coordinates"))
		return
	}
	if err != nil {
		logger.Info.Println("Invalid request ", r.RemoteAddr)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := u.Message(true, "Update box data")
	//resp := data.UpdateTLightInfo(*box)
	tflight := data.GetLightsFromBD(*box)
	resp["tflight"] = tflight
	u.Respond(w, r, resp)
}
