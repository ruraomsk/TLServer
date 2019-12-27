package whandlers

import (
	"net/http"
	"strconv"

	"../data"
	"../logger"
	u "../utils"
)

//BuildCross собираем данные для отображения прекрестка
var BuildCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}

	if len(r.URL.RawQuery) <= 0 {
		logger.Info.Println("crossArm: Blank field ", r.RemoteAddr)
		u.Respond(w, r, u.Message(false, "Blank field"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if TLight.Region.Num, err = strconv.Atoi(r.URL.Query().Get("Region")); err != nil {
		logger.Info.Println("crossArm: Blank field: Region ", r.RemoteAddr)
		u.Respond(w, r, u.Message(false, "Blank field: Region"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if TLight.ID, err = strconv.Atoi(r.URL.Query().Get("ID")); err != nil {
		logger.Info.Println("crossArm: Blank field: ID ", r.RemoteAddr)
		u.Respond(w, r, u.Message(false, "Blank field: ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	flag, resp := FuncAccessCheak(w, r, "BuildCross")
	if flag {
		resp = data.GetCrossInfo(*TLight)
	}

	u.Respond(w, r, resp)

}
