package dispatchControl

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HDispatchControl обработчик открытия сокета
func HDispatchControl(c *gin.Context, hub *HubDispCtrl) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	client := &ClientDC{hub: hub, conn: conn, send: make(chan dCResponse, 256), cInfo: accInfo, devices: make([]int, 0), sendPhases: false}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
