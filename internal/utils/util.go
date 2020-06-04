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
