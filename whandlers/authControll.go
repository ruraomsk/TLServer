package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"net/http"
)

//LoginAcc sign in account
var LoginAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := u.Message(false, "Invalid request")
		resp["logLogin"] = account.Login
		u.Respond(w, r, resp)
		return
	}
	resp := data.Login(account.Login, account.Password, r.RemoteAddr)
	u.Respond(w, r, resp)
}

var DisplayAccInfo = func(w http.ResponseWriter, r *http.Request) {
	privilege := &data.Privilege{}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	flag, resp := FuncAccessCheak(w, r, "DisplayAccInfo")
	if flag {
		resp = privilege.DisplayInfoForAdmin(mapContx)
	}
	u.Respond(w, r, resp)
}
