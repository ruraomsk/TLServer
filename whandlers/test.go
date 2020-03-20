package whandlers

import (
	"net/http"

	u "github.com/JanFant/TLServer/utils"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//flag, resp := FuncAccessCheck(w, r, 6)
	//if flag {
	resp := u.Message(true, "asd")
	resp["BLAA!!!"] = "Blaa!!!"
	//}
	u.Respond(w, r, resp)
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//flag, resp := FuncAccessCheck(w, r, 6)
	//if flag {
	var resp = make(map[string]interface{})
	resp["BLAA!!!"] = "Blaa!!!"
	//}
	u.Respond(w, r, resp)
})
