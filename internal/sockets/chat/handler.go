package chat

import (
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HChat обработчик открытия сокета
func HChat(c *gin.Context, hub *HubChat, db *sqlx.DB) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))
	var cInfo = clientInfo{
		login:  mapContx["login"],
		status: statusOnline,
		ip:     c.ClientIP(),
	}
	client := &ClientChat{hub: hub, conn: conn, send: make(chan chatResponse, 256), clientInfo: cInfo}

	client.hub.register <- client
	go client.writePump()
	go client.readPump(db)
}