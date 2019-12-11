package whandlers

import (
	"../data"
	u "../utils"
	"net/http"
	"strings"
)

//TestHello
var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	ip := strings.Split(r.RemoteAddr,":")
	data.GetDB().Table("accounts").Where("login = ?", "Super").First(&account)
	account.ParserPointsUser()
	tflight := data.GetLightsFromBD(account.Point0, account.Point1)
	resp := u.Message(true, "Hello")
	resp["account"] = account
	resp["tflight"] = tflight
	u.Respond(w, resp, ip[0])
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr,":")
	resp := make(map[string]interface{})
	resp["Test"] = "OK!"
	u.Respond(w, resp, ip[0])
})
