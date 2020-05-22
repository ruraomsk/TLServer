package handlers

import (
	"github.com/JanFant/newTLServer/internal/model/data"
	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

//LoginAcc обработчик входа в систему
var LoginAcc = func(c *gin.Context) {
	account := &data.Account{}
	if err := c.ShouldBindJSON(&account); err != nil {
		resp := u.Message(http.StatusBadRequest, "Invalid request")
		resp.Obj["logLogin"] = account.Login
		u.SendRespond(c, resp)
		return
	}
	resp := data.Login(account.Login, account.Password, c.Request.RemoteAddr)
	u.SendRespond(c, resp)
}

////LoginAcc обработчик входа в систему
//var LoginAcc = func(w http.ResponseWriter, r *http.Request) {
//	data := &data.Account{}
//	err := json.NewDecoder(r.Body).Decode(data)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		resp := u.Message(false, "Invalid request")
//		resp["logLogin"] = data.login
//		u.Respond(w, r, resp)
//		return
//	}
//	resp := data.login(data.login, data.Password, r.RemoteAddr)
//	if resp["status"] == false {
//		w.WriteHeader(http.StatusUnauthorized)
//	}
//	u.Respond(w, r, resp)
//}

////LoginAccOut обработчик выхода из системы
//var LoginAccOut = func(w http.ResponseWriter, r *http.Request) {
//	mapContx := u.ParserInterface(r.Context().Value("info"))
//	resp := data.LogOut(mapContx)
//	u.Respond(w, r, resp)
//}
//
////DisplayAccInfo отображение информации об аккаунтах для администрирования
//var DisplayAccInfo = func(w http.ResponseWriter, r *http.Request) {
//	privilege := &data.Privilege{}
//	mapContx := u.ParserInterface(r.Context().Value("info"))
//	resp := privilege.DisplayInfoForAdmin(mapContx)
//	u.Respond(w, r, resp)
//}
