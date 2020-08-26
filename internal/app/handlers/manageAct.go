package handlers

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/accToken"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
)

//ActUpdateAccount обработчик запроса обновления (работа с пользователями)
var ActUpdateAccount = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	var shortAcc = &data.ShortAccount{}
	err := c.ShouldBindJSON(&shortAcc)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "incorrectly filled data"))
		return
	}
	err = shortAcc.ValidCreate(accInfo.Role, accInfo.Region)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, fmt.Sprintf("incorrectly filled data: %s", err.Error())))
		return
	}
	account, privilege := shortAcc.ConvertShortToAcc()
	resp := account.Update(privilege)
	u.SendRespond(c, resp)
}

//ActDeleteAccount обработчик запроса удаления (работа с пользователями)
var ActDeleteAccount = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	var shortAcc = &data.ShortAccount{}
	err := c.ShouldBindJSON(&shortAcc)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "incorrectly filled data"))
		return
	}
	account, err := shortAcc.ValidDelete(accInfo.Role, accInfo.Region)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, fmt.Sprintf("incorrectly filled data: %s", err.Error())))
		return
	}
	resp := account.Delete()
	u.SendRespond(c, resp)
}

//ActAddAccount обработчик запроса добавления (работа с пользователями)
var ActAddAccount = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	var shortAcc = &data.ShortAccount{}
	err := c.ShouldBindJSON(&shortAcc)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "incorrectly filled data"))
		return
	}
	err = shortAcc.ValidCreate(accInfo.Role, accInfo.Region)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, fmt.Sprintf("incorrectly filled data: %s", err.Error())))
		return
	}
	account, privilege := shortAcc.ConvertShortToAcc()
	resp := account.Create(privilege)
	u.SendRespond(c, resp)
}

//ActChangePw обработчик запроса смены пароля для пользователя
var ActChangePw = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	var passChange = &data.PassChange{}
	err := c.ShouldBindJSON(&passChange)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "incorrectly filled data"))
		return
	}
	account, err := passChange.ValidOldNewPW(accInfo.Login)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, err.Error()))
		return
	}
	resp := account.ChangePW()
	u.SendRespond(c, resp)
}

//ActChangePw обработчик запроса смены пароля для пользователя
var ActResetPw = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	var shortAcc = &data.ShortAccount{}
	err := c.ShouldBindJSON(&shortAcc)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Не правильно заполненные данные"))
		return
	}
	account, err := shortAcc.ValidChangePW(accInfo.Role, accInfo.Region)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, fmt.Sprintf("Не правильно заполненные данные: %s", err.Error())))
		return
	}
	//account, _ := shortAcc.ConvertShortToAcc()
	resp := account.ResetPass()
	u.SendRespond(c, resp)
}
