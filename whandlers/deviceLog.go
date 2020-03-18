package whandlers

import (
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
