package greenStreet

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	u "github.com/ruraomsk/TLServer/internal/utils"
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

	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	client := &ClientGS{hub: hub, conn: conn, send: make(chan gSResponse, 256), cInfo: accInfo, devices: make([]int, 0), sendPhases: false,
		db: db}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db)
}
