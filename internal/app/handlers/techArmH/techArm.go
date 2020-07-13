package techArmH

import (
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//ChatEngine обработчик вебсокета для работы с чатом
var TechArmEngine = func(c *gin.Context) {
	region := c.Query("Region")
	var reg int
	if region != "" {
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

	area := c.QueryArray("Area")
	if len(area) <= 0 {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: Area"))
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()

	mapContx := u.ParserInterface(c.Value("info"))
	techArm.ArmTechReader(conn, reg, area, mapContx["login"], data.GetDB())
}
