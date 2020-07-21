package crossH

import (
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/gin-gonic/gin"

	u "github.com/JanFant/TLServer/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//CrossEngine обработчик вебсокета для работы с перекрестком
var CrossEngine = func(c *gin.Context) {
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
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()

	mapContx := u.ParserInterface(c.Value("info"))

	crossSock.CrossReader(conn, crEdit, mapContx, data.GetDB())
}

//CrossControlEngine обработчик вебсокета для работы с армом перекрестком
var CrossControlEngine = func(c *gin.Context) {
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
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()

	mapContx := u.ParserInterface(c.Value("info"))

	crossSock.ControlReader(conn, crEdit, mapContx, data.GetDB())
}

//ControlTestState обработчик проверки State
var ControlTestState = func(c *gin.Context) {
	mapContx := u.ParserInterface(c.Value("info"))
	resp := crossSock.TestCrossStateData(mapContx, data.GetDB())
	u.SendRespond(c, resp)
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
