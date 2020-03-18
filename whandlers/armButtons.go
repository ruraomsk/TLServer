package whandlers

import (
	"encoding/json"
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"github.com/ruraomsk/ag-server/comm"
)

//DispatchControlButton обработчик кнопок диспетчерского управления
var DispatchControlButtons = func(w http.ResponseWriter, r *http.Request) {
	arm := comm.CommandARM{}
	if err := json.NewDecoder(r.Body).Decode(&arm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DispatchControl(arm, mapContx)
	u.Respond(w, r, resp)
}
