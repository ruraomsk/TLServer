package whandlers

import (
	"../data"
	u "../utils"
	"net/http"
)

var FuncAccessCheak = func(w http.ResponseWriter, r *http.Request, act string) (flag bool, resp map[string]interface{}) {
	resp = make(map[string]interface{})
	flag, err := data.RoleCheck(u.ParserInterface(r.Context().Value("info")), act)
	if err != nil || !flag {
		resp = u.Message(false, err.Error())
		if err != nil {
			resp = u.Message(false, "Access denied")
		}
		w.WriteHeader(http.StatusForbidden)
	}
	return flag, resp
}
