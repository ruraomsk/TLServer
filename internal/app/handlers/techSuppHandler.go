package handlers

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/data"
	"net/http"

	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/model/techSupport"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//TechSupp обработчик отправления сообщения в тех поддержку
var TechSupp = func(c *gin.Context) {
	var emailInfo techSupport.EmailJS
	err := c.ShouldBindJSON(&emailInfo)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := techSupport.SendEmail(emailInfo, accInfo.Login, license.LicenseFields.CompanyName, license.LicenseFields.Address, data.GetDB())
	u.SendRespond(c, resp)
}
