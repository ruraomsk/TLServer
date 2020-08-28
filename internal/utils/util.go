package utils

import (
	"github.com/gin-gonic/gin"
)

//Response ответ http запроса
type Response struct {
	Code int                    //код ответа
	Obj  map[string]interface{} //данные
}

//Message создает map для ответа пользователю
func Message(code int, message string) Response {
	return Response{Code: code, Obj: map[string]interface{}{"message": message}}
}

//SendRespond формирует ответ пользователю записывает
func SendRespond(c *gin.Context, resp Response) {
	writeLogMessage(c.Request.RemoteAddr, c.Request.RequestURI, resp.Obj, c.Param("slug"))
	c.JSON(resp.Code, resp.Obj)
}
