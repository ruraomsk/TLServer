package whandlers

import (
	"../data"
	"../logger"
	u "../utils"
	"encoding/json"
	"net/http"
	"strings"
)

//LoginAcc sign in account
var LoginAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	ip := strings.Split(r.RemoteAddr, ":")
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		logger.Info.Println("authControll, loginAcc: Invalid request ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.Login(account.Login, account.Password, ip[0])
	u.Respond(w, r, resp)
}

//CreateAcc create new acc
var CreateAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	err := json.NewDecoder(r.Body).Decode(account) //
	if err != nil {
		logger.Info.Println("authControll, create: Invalid request ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}

	flag, resp := FuncAccessCheak(w, r, "CreateAcc")
	if flag {
		//resp = account.Create(nil)
	}

	u.Respond(w, r, resp)
}

var DisplayAccInfo = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	mapContx := data.ParserInterface(r.Context().Value("info"))
	account.Login = mapContx["login"]

	flag, resp := FuncAccessCheak(w, r, "DisplayAccInfo")
	if flag {
		resp = account.DisplayInfoForAdmin(mapContx)
	}
	u.Respond(w, r, resp)
}
