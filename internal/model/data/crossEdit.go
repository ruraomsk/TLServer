package data

import (
	u "github.com/JanFant/TLServer/internal/utils"
	"net/http"
)

func DisplayCrossEditInfo(mapContx map[string]string) u.Response {
	resp := u.Message(http.StatusOK, "Edit info")
	var temp []crossInfo
	for _, info := range crossConnect {
		temp = append(temp, info)
	}
	resp.Obj["crosses"] = temp
	temp = make([]crossInfo, 0)
	for _, info := range controlConnect {
		temp = append(temp, info)
	}
	resp.Obj["arms"] = temp
	return resp
}
