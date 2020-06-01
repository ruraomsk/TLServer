package handlers

import (
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
	mapContx := u.ParserInterface(c.Value("info"))
	resp := techSupport.SendEmail(emailInfo, mapContx["login"], license.LicenseFields.CompanyName)
	u.SendRespond(c, resp)
}
