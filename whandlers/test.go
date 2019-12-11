package whandlers

import (
	"../data"
	u "../utils"
	"net/http"
)

//TestHello
var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	data.GetDB().Table("accounts").Where("login = ?", "Super").First(&account)
	account.ParserPointsUser()
	tflight := data.GetLightsFromBD(account.Point0, account.Point1)
	resp := u.Message(true, "Hello")
	resp["account"] = account
	resp["tflight"] = tflight
	u.Respond(w, r, resp)
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context().Value("user")
	//fmt.Println(ctx)
	resp := make(map[string]interface{})
	resp["Test"] = "OK!"
	u.Respond(w, r, resp)
})
