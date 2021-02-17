package xctrl

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

//HXctrl обработчик открытия сокета
func HXctrl(c *gin.Context, hub *HubXctrl, db *sqlx.DB) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	client := &ClientXctrl{hub: hub, conn: conn, send: make(chan MessXctrl, 256), xInfo: accInfo}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db)
}
