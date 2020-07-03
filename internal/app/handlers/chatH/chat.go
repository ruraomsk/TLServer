package chatH

import (
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/chat"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//ChatEngine обработчик вебсокета для работы с чатом
var ChatEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()
	mapContx := u.ParserInterface(c.Value("info"))
	chat.Reader(conn, mapContx["login"], data.GetDB())
}
