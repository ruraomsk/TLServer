package handlers

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/license"
	"github.com/ruraomsk/TLServer/internal/model/techSupport"
	u "github.com/ruraomsk/TLServer/internal/utils"
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
	resp := techSupport.SendEmail(emailInfo, accInfo.Login, license.LicenseFields.CompanyName, license.LicenseFields.Address)
	u.SendRespond(c, resp)
}
