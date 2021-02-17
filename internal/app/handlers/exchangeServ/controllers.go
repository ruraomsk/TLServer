package exchangeServ

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/exchangeData"
)

//ControllerHandler обработчик запроса контроллеров
func ControllerHandler(c *gin.Context, db *sqlx.DB) {
	resp := exchangeData.GetController(db)
	c.JSON(resp.Code, resp.Obj)
}
