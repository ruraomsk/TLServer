package handlers

import (
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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()

	var crEdit data.PosInfo
	crEdit.Region, crEdit.Area, crEdit.Id, err = queryParser(c)
	if err != nil {
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))

	data.CrossReader(conn, crEdit, mapContx)
}

//CrossControlEngine обработчик вебсокета для работы с армом перекрестком
var CrossControlEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()

	var crEdit data.PosInfo
	crEdit.Region, crEdit.Area, crEdit.Id, err = queryParser(c)
	if err != nil {
		return
	}

	mapContx := u.ParserInterface(c.Value("info"))

	data.ControlReader(conn, crEdit, mapContx)
}

//ControlTestState обработчик проверки State
var ControlTestState = func(c *gin.Context) {
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.TestCrossStateData(mapContx)
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
