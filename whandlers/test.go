package whandlers

import (
	u "../utils"
	"net/http"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	u.Respond(w, r, u.Message(true, "Chil its ok"))
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	resp["Test"] = "OK!"
	u.Respond(w, r, resp)
})
