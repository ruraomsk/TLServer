package handlers

import (
	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

////LoginAcc обработчик входа в систему
//var LoginAcc = func(c *gin.Context) {
//	account := &data.Account{}
//	if err := c.ShouldBindJSON(&account); err != nil {
//		resp := u.Message(http.StatusBadRequest, "invalid request")
//		resp.Obj["logLogin"] = account.Login
//		u.SendRespond(c, resp)
//		return
//	}
//	resp := data.Login(account.Login, account.Password, c.Request.RemoteAddr)
//	u.SendRespond(c, resp)
//}
//
////LoginAccOut обработчик выхода из системы
//var LoginAccOut = func(c *gin.Context) {
//	mapContx := u.ParserInterface(c.Value("info"))
//	resp := data.LogOut(mapContx["login"])
//	u.SendRespond(c, resp)
//}

//DisplayAccInfo отображение информации об аккаунтах для администрирования
var DisplayAccInfo = func(c *gin.Context) {
	privilege := &data.Privilege{}
	mapContx := u.ParserInterface(c.Value("info"))
	resp := privilege.DisplayInfoForAdmin(mapContx)
	u.SendRespond(c, resp)
}
