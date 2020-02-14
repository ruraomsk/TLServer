package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//Message создает map для ответа пользователю
func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

//Respond формирует ответ пользователю записывает необходимые хедеры и сворачивает json
func Respond(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
	if !strings.Contains(fmt.Sprint(data["DontWrite"]), "true") {
		WriteLogMessage(r.RemoteAddr, r.RequestURI, data, r.Context().Value("info"))
		delete(data, "logLogin")
	} else {
		delete(data, "DontWrite")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	json.NewEncoder(w).Encode(data)
}
