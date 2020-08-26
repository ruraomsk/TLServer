package handlers

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//DisplayAccInfo отображение информации об аккаунтах для администрирования
var DisplayAccInfo = func(c *gin.Context) {
	privilege := &data.Privilege{}
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := privilege.DisplayInfoForAdmin(accInfo)
	u.SendRespond(c, resp)
}
