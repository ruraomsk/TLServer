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
	// fmt.Printf("X1 = %3.15f, Y1 = %3.15f, X2 = %3.15f, Y2 = %3.15f", box.Point1.X, box.Point1.Y, box.Point0.X, box.Point0.Y)
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
