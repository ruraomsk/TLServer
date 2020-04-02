package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//LicenseInfo обработчик сборки начальной информации
var LicenseInfo = func(w http.ResponseWriter, r *http.Request) {
	resp := u.Message(true, "Здесь чтото будет")
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
