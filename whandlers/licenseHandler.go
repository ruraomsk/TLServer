package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/license"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//LicenseInfo обработчик сборки начальной информации
var LicenseInfo = func(w http.ResponseWriter, r *http.Request) {
	resp := license.LicenseInfo()
	u.Respond(w, r, resp)
}

//LicenseCreateToken обработчик создания токена лицензии
var LicenseCreateToken = func(w http.ResponseWriter, r *http.Request) {
	licInfo := license.License{}
	if err := json.NewDecoder(r.Body).Decode(&licInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := license.CreateLicenseToken(licInfo)
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
	resp := license.LicenseNewKey(key.Key)
	u.Respond(w, r, resp)
}
