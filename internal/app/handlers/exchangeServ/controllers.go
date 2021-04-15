package exchangeServ

import (
	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/exchangeData"
)

//ControllerHandler обработчик запроса контроллеров
func ControllerHandler(c *gin.Context) {
	resp := exchangeData.GetController()
	c.JSON(resp.Code, resp.Obj)
}
