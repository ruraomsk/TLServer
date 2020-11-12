package exchangeServ

import (
	"github.com/JanFant/TLServer/internal/model/exchangeData"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

//ControllerHandler обработчик запроса контроллеров
func ControllerHandler(c *gin.Context, db *sqlx.DB) {
	resp := exchangeData.GetController(db)
	c.JSON(resp.Code, resp.Obj)
}
