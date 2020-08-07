package greenStreet

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

//HGStreet обработчик открытия сокета
func HGStreet(c *gin.Context, hub *HubGStreet, db *sqlx.DB) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))
	client := &ClientGS{hub: hub, conn: conn, send: make(chan gSResponse, 256), login: mapContx["login"], ip: c.ClientIP()}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db)
}
