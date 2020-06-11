package handlers

import (
	"github.com/JanFant/TLServer/internal/model/data"
	"net/http"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//MapEngine обработчик вебсокета для работы с картой
var MapEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()
	data.MapReader(conn, c)
}
