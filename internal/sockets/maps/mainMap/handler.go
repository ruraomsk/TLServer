package mainMap

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

//HMainMap обработчик открытия сокета
func HMainMap(c *gin.Context, hub *HubMainMap, db *sqlx.DB) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	client := &ClientMainMap{hub: hub, conn: conn, send: make(chan mapResponse, 256), login: "", ip: c.ClientIP()}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db, c)
}
