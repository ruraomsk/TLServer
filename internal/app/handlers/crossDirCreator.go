package handlers

import (
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/gin-gonic/gin"

	u "github.com/JanFant/TLServer/internal/utils"
)

//MainCrossCreator сборка информации для странички создания каталогов перекрестков
var MainCrossCreator = func(c *gin.Context) {
	resp := data.MainCrossCreator()
	u.SendRespond(c, resp)
}

//CheckAllCross обработчик проверки всех перекрестков из БД
var CheckAllCross = func(c *gin.Context) {
	resp := data.CheckCrossDirFromBD()
	u.SendRespond(c, resp)
}

//CheckSelectedDirCross обработчик проверки регионов, районов и перекрестков, выбранных пользователем
var CheckSelectedDirCross = func(c *gin.Context) {
	var selectedData data.SelectedData
	err := c.ShouldBindJSON(&selectedData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Invalid request"))
		return
	}
	resp := data.CheckCrossFileSelected(selectedData.SelectedData)
	u.SendRespond(c, resp)
}

//MakeSelectedDirCross обработчик проверки регионов, районов и перекрестков, выбранных пользователем
var MakeSelectedDirCross = func(c *gin.Context) {
	var selectedData data.SelectedData
	err := c.ShouldBindJSON(&selectedData)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Invalid request"))
		return
	}
	resp := data.MakeSelectedDir(selectedData)
	u.SendRespond(c, resp)
}
