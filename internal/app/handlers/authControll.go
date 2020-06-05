package handlers

import (
	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//DisplayAccInfo отображение информации об аккаунтах для администрирования
var DisplayAccInfo = func(c *gin.Context) {
	privilege := &data.Privilege{}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := privilege.DisplayInfoForAdmin(mapContx)
	u.SendRespond(c, resp)
}
