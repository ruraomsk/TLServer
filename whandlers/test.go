package whandlers

import (
	u "../utils"
	"net/http"
)

//TestHello
var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello my friend"))
	u.Respond(w,u.Message(true,"Hello"))
	//u.Respond(w,u.Message(false,"world"))
})
