package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

type Response struct {
	Code int
	Obj  map[string]interface{}
}

//Message создает map для ответа пользователю
func Message(code int, message string) Response {
	return Response{Code: code, Obj: map[string]interface{}{"message": message}}
}

////Message создает map для ответа пользователю
//func Message(status bool, message string) map[string]interface{} {
//	return map[string]interface{}{"status": status, "message": message}
//}

//SendRespond формирует ответ пользователю записывает
func SendRespond(c *gin.Context, resp Response) {
	if !strings.Contains(fmt.Sprint(resp.Obj["DontWrite"]), "true") {
		writeLogMessage(c.Request.RemoteAddr, c.Request.RequestURI, resp.Obj, c.Value("info"))
		delete(resp.Obj, "logLogin")
	} else {
		delete(resp.Obj, "DontWrite")
	}
	c.JSON(resp.Code, resp.Obj)
}

////Respond формирует ответ пользователю записывает необходимые хедеры и сворачивает json
//func Respond(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
//	if !strings.Contains(fmt.Sprint(data["DontWrite"]), "true") {
//		writeLogMessage(r.RemoteAddr, r.RequestURI, data, r.Context().Value("info"))
//		delete(data, "logLogin")
//	} else {
//		delete(data, "DontWrite")
//	}
//	w.Header().Add("Content-Type", "application/json")
//	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
//	json.NewEncoder(w).Encode(data)
//}
