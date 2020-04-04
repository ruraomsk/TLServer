package whandlers

import (
	"encoding/json"
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
)

//LicenseInfo обработчик сборки начальной информации
var LicenseInfo = func(w http.ResponseWriter, r *http.Request) {
	resp := data.LicenseInfo()
	u.Respond(w, r, resp)
}

//LicenseCreateToken обработчик создания токена лицензии
var LicenseCreateToken = func(w http.ResponseWriter, r *http.Request) {
	license := data.License{}
	if err := json.NewDecoder(r.Body).Decode(&license); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.CreateLicenseToken(license)
	u.Respond(w, r, resp)
}

//LicenseNewKey обработчик обработчик сохранения нового токена
var LicenseNewKey = func(w http.ResponseWriter, r *http.Request) {
	type keyStr struct {
		Key string `json:"keyStr"`
	}
	var key keyStr
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.LicenseNewKey(key.Key)
	u.Respond(w, r, resp)
}
