package alarm

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HTechArm обработчик открытия сокета
func HAlarm(c *gin.Context, hub *HubAlarm, db *sqlx.DB) {
	region := c.Query("Region")
	var reg int
	if len(region) != 0 {
		var err error
		reg, err = strconv.Atoi(region)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: Region"))
			return
		}
	} else {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: Region"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)

	var armInfo = Info{
		Region:  reg,
		AccInfo: accInfo,
	}
	var crossRing = CrossRing{Ring: false, CrossInfo: make(map[string]*CrossInfo, 0)}
	client := &ClientAlarm{hub: hub, conn: conn, send: make(chan alarmResponse, 256), armInfo: &armInfo, CrossRing: &crossRing}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db)
}
