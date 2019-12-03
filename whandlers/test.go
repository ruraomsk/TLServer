package whandlers

import "net/http"

//TestHello
var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello my friend"))
})
