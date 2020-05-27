package handlers

import (
	"github.com/JanFant/newTLServer/internal/model/license"
	"github.com/JanFant/newTLServer/internal/model/techSupport"
	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

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
