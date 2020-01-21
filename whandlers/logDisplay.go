package whandlers

import (
	"../data"
	u "../utils"
	"net/http"
)

var DisplayLogFile = func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheak(w, r, "LogInfo")
	if flag {
		resp = data.DisplayLogFiles()
	}
	u.Respond(w, r, resp)
}

var DisplayLogInfo = func(w http.ResponseWriter, r *http.Request) {
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

	flag, resp := FuncAccessCheak(w, r, "LogInfo")
	if flag {
		resp = data.DisplayFileLog(fileName)
	}
	u.Respond(w, r, resp)
}
