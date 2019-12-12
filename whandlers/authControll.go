package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//LoginAcc sign in account
var LoginAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	ip := strings.Split(r.RemoteAddr, ":")
	err := json.NewDecoder(r.Body).Decode(account) //
	if err != nil {
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.Login(account.Login, account.Password, ip[0])
	u.Respond(w, r, resp)
}

//CreateAcc create new acc !!! this func only for admin
var CreateAcc = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	err := json.NewDecoder(r.Body).Decode(account) //
	if err != nil {
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := account.Create()
	u.Respond(w, r, resp)
}

//BuildMainPage собираем данные для залогиневшегося пользователя
var BuildMainPage = func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	account.Login = fmt.Sprintf("%v", r.Context().Value("user"))
	resp := account.GetInfoForUser()
	u.Respond(w, r, resp)
}
