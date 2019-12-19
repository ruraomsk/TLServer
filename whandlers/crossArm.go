package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"net/http"
)

//BuildCrossPage собираем данные для отображения прекрестка
var BuildCrossPage = func(w http.ResponseWriter, r *http.Request) {
	TLight := &data.TrafficLights{}
	err := json.NewDecoder(r.Body).Decode(TLight)
	if err != nil {
		u.Respond(w, r, u.Message(false, "Wrong Data"))
		return
	}
	resp := data.GetCrossInfo(*TLight)
	u.Respond(w, r, resp)

}
