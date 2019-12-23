package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"fmt"
	"net/http"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	_ = json.NewDecoder(r.Body).Decode(account)
	str:= account.Privilege.ToSqlStrUpdate("accounts","MMM")
	//data.GetDB().Exec(str)
	fmt.Println(str)
	u.Respond(w, r, u.Message(true, "Chil its ok"))
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	resp["Test"] = "OK!"
	u.Respond(w, r, resp)
})
