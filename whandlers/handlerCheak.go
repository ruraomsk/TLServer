package whandlers

import (
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
)

//FuncAccessCheck проверяет разрешение пользователя для доступа к ресурсу
var FuncAccessCheck = func(w http.ResponseWriter, r *http.Request, act int) (flag bool, resp map[string]interface{}) {
	resp = make(map[string]interface{})
	flag, err := data.NewRoleCheck(u.ParserInterface(r.Context().Value("info")), act)
	if err != nil || !flag {
		resp = u.Message(false, err.Error())
		if err != nil {
			resp = u.Message(false, "Access denied")
		}
		w.WriteHeader(http.StatusForbidden)
	}
	return flag, resp
}
