package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	u "github.com/ruraomsk/TLServer/internal/utils"
)

//DisplayAccInfo отображение информации об аккаунтах для администрирования
var DisplayAccInfo = func(c *gin.Context) {
	privilege := &data.Privilege{}
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := privilege.DisplayInfoForAdmin(accInfo)
	u.SendRespond(c, resp)
}
