package handlers

import (
	"net/http"
	"strconv"

	"github.com/JanFant/TLServer/internal/model/crossEdit"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/gin-gonic/gin"

	u "github.com/JanFant/TLServer/internal/utils"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

//BuildCross обработчик собора данных для отображения прекрёстка
var BuildCross = func(c *gin.Context) {
	var err error
	TLight := &data.TrafficLights{}
	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(c)
	if err != nil {
		return
	}
	resp := data.GetCrossInfo(*TLight)
	mapContx := u.ParserInterface(c.Value("info"))

	controlCrossFlag, _ := data.AccessCheck(mapContx, 5)
	if (TLight.Region.Num == mapContx["region"]) || (mapContx["region"] == "*") {
		resp.Obj["controlCrossFlag"] = controlCrossFlag
	} else {
		resp.Obj["controlCrossFlag"] = false
	}
	u.SendRespond(c, resp)
}

//DevCrossInfo обработчик собора данных для отображения перекрёстка (idevice информация)
var DevCrossInfo = func(c *gin.Context) {
	var idevice string
	idevice = c.Query("idevice")
	if idevice != "" {
		_, err := strconv.Atoi(idevice)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusBadRequest, "blank field: idevice"))
			return
		}
	}
	resp := data.GetCrossDevInfo(idevice)
	u.SendRespond(c, resp)
}

//ControlCross обработчик данных для заполнения таблиц управления
var ControlCross = func(c *gin.Context) {
	var err error
	TLight := &data.TrafficLights{}
	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(c)
	if err != nil {
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))

	var resp u.Response
	if (TLight.Region.Num == mapContx["region"]) || (mapContx["region"] == "*") {
		resp = data.ControlGetCrossInfo(*TLight)
	}
	u.SendRespond(c, resp)
}

//ControlEditableCross обработчик проверки редактирования перекрестка
var ControlEditableCross = func(c *gin.Context) {
	var err error
	arm := &crossEdit.BusyArm{}
	arm.Region, arm.Area, arm.ID, err = queryParser(c)
	if err != nil {
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.ControlEditableCheck(*arm, mapContx)
	u.SendRespond(c, resp)
}

//ControlCloseCross обработчик закрытия перекрестка
var ControlCloseCross = func(c *gin.Context) {
	var err error
	arm := &crossEdit.BusyArm{}
	arm.Region, arm.Area, arm.ID, err = queryParser(c)
	if err != nil {
		return
	}
	resp := crossEdit.BusyArmDelete(*arm)
	u.SendRespond(c, resp)
}

//ControlSendButton обработчик данных для отправки на устройство(сервер)
var ControlSendButton = func(c *gin.Context) {
	var stateData agS_pudge.Cross
	err := c.ShouldBindJSON(&stateData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.SendCrossData(stateData, mapContx)
	u.SendRespond(c, resp)
}

//ControlCreateButton обработчик данных для создания перекрестка и отправка на устройство(сервер)
var ControlCreateButton = func(c *gin.Context) {
	var stateData agS_pudge.Cross
	err := c.ShouldBindJSON(&stateData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.CreateCrossData(stateData, mapContx)
	u.SendRespond(c, resp)
}

//ControlCheckButton обработчик данных для их проверки
var ControlCheckButton = func(c *gin.Context) {
	var stateData agS_pudge.Cross
	err := c.ShouldBindJSON(&stateData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	resp := data.CheckCrossData(stateData)
	u.SendRespond(c, resp)
}

//ControlDeleteButton обработчик данных для удаления перекрестка
var ControlDeleteButton = func(c *gin.Context) {
	var stateData agS_pudge.Cross
	err := c.ShouldBindJSON(&stateData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.DeleteCrossData(stateData, mapContx)
	u.SendRespond(c, resp)
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
