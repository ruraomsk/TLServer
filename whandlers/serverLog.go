package whandlers

import (
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
)

//DisplayServerLogFile обработчик отображения файлов лога сервера
var DisplayServerLogFile = func(w http.ResponseWriter, r *http.Request) {
	resp := data.DisplayServerLogFiles()
	u.Respond(w, r, resp)
}

//DisplayServerLogInfo обработчик выгрузки содержимого лог файла сервера
var DisplayServerLogInfo = func(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.RawQuery) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field"))
		return
	}
	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: fileName"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DisplayServerFileLog(fileName, mapContx)
	u.Respond(w, r, resp)
}
