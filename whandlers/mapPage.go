package whandlers

import (
	"../data"
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
		u.Respond(w, r, u.Message(false, "Impossible coordinates"))
		return
	}
	if err != nil {
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.UpdateTLightInfo(*box)
	u.Respond(w, r, resp)
}
