package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//DisplayServerLogFile обработчик отображения файлов лога
var DisplayDeviceLogFile = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DisplayDeviceLog(mapContx)
	u.Respond(w, r, resp)
}

//LogDeviceInfo обработчик запроса на выгрузку информации логов устройства за определенный период
var LogDeviceInfo = func(w http.ResponseWriter, r *http.Request) {
	arm := &data.DeviceLogInfo{}
	if err := json.NewDecoder(r.Body).Decode(&arm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}

	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DisplayDeviceLogInfo(*arm, mapContx)
	u.Respond(w, r, resp)
}
