package handlers

import (
	"net/http"

	"github.com/JanFant/TLServer/internal/model/crossButtons"
	"github.com/gin-gonic/gin"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/ruraomsk/ag-server/comm"
)

//DispatchControlButtons обработчик кнопок диспетчерского управления
var DispatchControlButtons = func(c *gin.Context) {
	arm := comm.CommandARM{}
	if err := c.ShouldBindJSON(&arm); err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := crossButtons.DispatchControl(arm, mapContx)
	u.SendRespond(c, resp)
}
