package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
	if !strings.Contains(fmt.Sprint(data["message"]), "Update map data") {
		WriteLogMessage(r.RemoteAddr, data, r.Context().Value("info"))
		delete(data, "logLogin")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	json.NewEncoder(w).Encode(data)
}
