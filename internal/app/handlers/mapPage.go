package handlers

import (
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/locations"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//MapEngine обработчик вебсокета для работы с картой
var MapEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()
	//mapContx := u.ParserInterface(c.Value("info"))
	data.ReaderStrong(conn)
}

//BuildMapPage собираем данные для авторизованного пользователя
var BuildMapPage = func(c *gin.Context) {
	account := &data.Account{}
	mapContx := u.ParserInterface(c.Value("info"))
	account.Login = mapContx["login"]
	resp := account.GetInfoForUser()
	resp.Obj["manageFlag"], _ = data.AccessCheck(mapContx, 1)
	resp.Obj["logDeviceFlag"], _ = data.AccessCheck(mapContx, 11)
	u.SendRespond(c, resp)
}

//UpdateMapPage обновление информации о попавших в область светофорах
var UpdateMapPage = func(c *gin.Context) {
	box := &locations.BoxPoint{}
	if err := c.ShouldBindJSON(&box); err != nil {
		resp := u.Message(http.StatusBadRequest, "invalid request")
		u.SendRespond(c, resp)
		return
	}
	tflight := data.GetLightsFromBD(*box)
	resp := u.Message(http.StatusOK, "Update map data")
	resp.Obj["DontWrite"] = "true"
	resp.Obj["tflight"] = tflight
	u.SendRespond(c, resp)
}

//LocationButtonMapPage обработка запроса на получение новых координат рабочей области
var LocationButtonMapPage = func(c *gin.Context) {
	location := &data.Locations{}
	if err := c.ShouldBindJSON(&location); err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	boxPoint, err := location.MakeBoxPoint()
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	resp := u.Message(http.StatusOK, "jump to location!")
	resp.Obj["DontWrite"] = "true"
	resp.Obj["boxPoint"] = &boxPoint
	u.SendRespond(c, resp)
}
