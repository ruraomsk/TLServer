package controlCross

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

//HMainCross обработчик открытия сокета
func HMainCross(c *gin.Context, hub *HubCross, db *sqlx.DB) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))
	client := &ClientCross{hub: hub, conn: conn, send: make(chan crossResponse, 256), login: mapContx["login"]}
	client.hub.register <- client

	go client.writePump(db)
	go client.readPump(db)
}
