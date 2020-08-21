package controlCross

import (
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/internal/sockets/maps"
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

//HControlCross обработчик открытия сокета
func HControlCross(c *gin.Context, hub *HubControlCross, db *sqlx.DB) {
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

	mapContx := u.ParserInterface(c.Value("info"))

	//проверка на полномочия редактирования
	if !((crEdit.Region == mapContx["region"]) || (mapContx["region"] == "*")) {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, typeNotEdit))
		return
	}
	token, _ := c.Cookie("Authorization")
	var crossInfo = crossSock.CrossInfo{
		Login:   mapContx["login"],
		Role:    mapContx["role"],
		Edit:    false,
		Idevice: 0,
		Pos:     crEdit,
		Ip:      c.ClientIP(),
		Region:  mapContx["region"],
		Token:   token,
	}

	client := &ClientControlCr{hub: hub, conn: conn, send: make(chan ControlSokResponse, 256), crossInfo: crossInfo, regStatus: make(chan bool)}
	client.hub.register <- client
	rs := <-client.regStatus
	if rs {
		go client.writePump()
		go client.readPump(db)
	}
}
