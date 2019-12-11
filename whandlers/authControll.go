package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"net/http"
	"strings"
)

//LoginAcc sign in account
var LoginAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	ip := strings.Split(r.RemoteAddr, ":")
	err := json.NewDecoder(r.Body).Decode(account) //
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"), ip[0])
		return
	}
	resp := data.Login(account.Login, account.Password, ip[0])
	u.Respond(w, resp, ip[0])
}

//CreateAcc create new acc !!! this func only for admin
var CreateAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	ip := strings.Split(r.RemoteAddr, ":")

	err := json.NewDecoder(r.Body).Decode(account) //
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"), ip[0])
		return
	}
	resp := account.Create()
	u.Respond(w, resp, ip[0])
}
