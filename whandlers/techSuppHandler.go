package whandlers

import (
	"encoding/json"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

var TechSupp = func(w http.ResponseWriter, r *http.Request) {
	var emailInfo data.EmailJS
	err := json.NewDecoder(r.Body).Decode(&emailInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.SendEmail(emailInfo, mapContx)
	u.Respond(w, r, resp)
}
