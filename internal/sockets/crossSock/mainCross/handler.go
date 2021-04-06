package mainCross

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	"github.com/ruraomsk/TLServer/internal/sockets/maps"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HMainCross обработчик открытия сокета
func HMainCross(c *gin.Context, hub *HubCross) {
	var (
		crEdit sockets.PosInfo
		err    error
	)
	crEdit.Region, crEdit.Area, crEdit.Id, err = maps.QueryParser(c)
	if err != nil {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)

	var crossInfo = crossSock.CrossInfo{
		Edit:    false,
		Idevice: 0,
		Pos:     crEdit,
		AccInfo: accInfo,
		Login:   accInfo.Login,
	}

	client := &ClientCross{hub: hub, conn: conn, send: make(chan crossResponse, 256), crossInfo: &crossInfo, regStatus: make(chan bool)}
	client.hub.register <- client
	rs := <-client.regStatus
	if rs {
		go client.writePump()
		go client.readPump()
	}
}
