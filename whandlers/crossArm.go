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
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field"))
		return
	}
	if TLight.Region.Num, err = strconv.Atoi(r.URL.Query().Get("Region")); err != nil {
		logger.Info.Println("crossArm: Blank field: Region ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: Region"))
		return
	}
	if TLight.ID, err = strconv.Atoi(r.URL.Query().Get("ID")); err != nil {
		logger.Info.Println("crossArm: Blank field: ID ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: ID"))
		return
	}
	if TLight.Area.Num, err = strconv.Atoi(r.URL.Query().Get("Area")); err != nil {
		logger.Info.Println("crossArm: Blank field: Area ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: Area"))
		return
	}
	flag, resp := FuncAccessCheak(w, r, "BuildCross")
	if flag {
		resp = data.GetCrossInfo(*TLight)
	}

	u.Respond(w, r, resp)

}
