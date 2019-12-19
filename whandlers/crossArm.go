package whandlers

import (
	"net/http"
	"strconv"

	"../data"
	u "../utils"
)

//BuildCross собираем данные для отображения прекрестка
var BuildCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}

	if TLight.Region.Num, err = strconv.Atoi(r.URL.Query().Get("Region")); err != nil {
		u.Respond(w, r, u.Message(false, "Blank field: Region"))
		return
	}
	if TLight.ID, err = strconv.Atoi(r.URL.Query().Get("ID")); err != nil {
		u.Respond(w, r, u.Message(false, "Blank field: ID"))
		return
	}

	resp := data.GetCrossInfo(*TLight)
	u.Respond(w, r, resp)

}
