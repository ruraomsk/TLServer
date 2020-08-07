package mainCross

import (
	"github.com/JanFant/TLServer/internal/sockets"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HMainCross обработчик открытия сокета
func HMainCross(c *gin.Context, hub *HubCross, db *sqlx.DB) {
	var (
		crEdit sockets.PosInfo
		err    error
	)
	crEdit.Region, crEdit.Area, crEdit.Id, err = queryParser(c)
	if err != nil {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))
	var crossInfo = CrossInfo{
		Login:   mapContx["login"],
		Role:    mapContx["role"],
		Edit:    false,
		Idevice: 0,
		Pos:     crEdit,
		ip:      c.ClientIP(),
		region:  mapContx["region"],
	}

	client := &ClientCross{hub: hub, conn: conn, send: make(chan crossResponse, 256), crossInfo: crossInfo, regStatus: make(chan regStatus)}
	client.hub.register <- client
	rs := <-client.regStatus
	client.crossInfo.Edit = rs.edit
	client.crossInfo.Idevice = rs.idevice
	if rs.ok {
		go client.writePump()
		go client.readPump(db)
	}
}

//queryParser разбор URL строки
func queryParser(c *gin.Context) (region, area string, ID int, err error) {
	region = c.Query("Region")
	if region != "" {
		_, err = strconv.Atoi(region)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: Region"))
			return
		}
	}

	area = c.Query("Area")
	if area != "" {
		_, err = strconv.Atoi(area)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: Area"))
			return
		}
	}

	IDStr := c.Query("ID")
	if IDStr != "" {
		ID, err = strconv.Atoi(IDStr)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: ID"))
			return
		}
	}

	return
}
