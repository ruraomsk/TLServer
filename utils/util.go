package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter,r *http.Request, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Println(r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	//w.Header().Add("Access-Control-Allow-Methods","POST")
	json.NewEncoder(w).Encode(data)
}
