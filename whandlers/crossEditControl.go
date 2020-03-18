package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//CrossEditInfo сбор информации о занятых перекрестках
var CrossEditInfo = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DisplayCrossEditInfo(mapContx)
	u.Respond(w, r, resp)
}

//CrossEditFree освобождение перекрестков
var CrossEditFree = func(w http.ResponseWriter, r *http.Request) {
	var busyArms data.BusyArms
	if err := json.NewDecoder(r.Body).Decode(&busyArms); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.FreeCrossEdit(busyArms)
	u.Respond(w, r, resp)
}
