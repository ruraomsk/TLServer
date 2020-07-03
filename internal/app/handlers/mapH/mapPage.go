package mapH

import (
	"github.com/JanFant/TLServer/internal/model/data"
	mapSocket "github.com/JanFant/TLServer/internal/model/mapSock"
	"github.com/gorilla/websocket"
	"net/http"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//MapEngine обработчик вебсокета для работы с картой
var MapEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()
	mapSocket.MapReader(conn, c, data.GetDB())
}
