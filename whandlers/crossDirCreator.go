package whandlers

import (
	"encoding/json"
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
)

//MainCrossCreator сборка информации для странички создания каталогов перекрестков
var MainCrossCreator = func(w http.ResponseWriter, r *http.Request) {
	resp := data.MainCrossCreator()
	u.Respond(w, r, resp)
}

//CheckAllCross обработчик проверки всех перекрестков из БД
var CheckAllCross = func(w http.ResponseWriter, r *http.Request) {
	resp := data.CheckCrossDirFromBD()
	u.Respond(w, r, resp)
}

//CheckSelectedDirCross обработчик проверки регионов, районов и перекрестков, выбранных пользователем
var CheckSelectedDirCross = func(w http.ResponseWriter, r *http.Request) {
	var selectedData data.SelectedData
	err := json.NewDecoder(r.Body).Decode(&selectedData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.CheckCrossFileSelected(selectedData.SelectedData)
	u.Respond(w, r, resp)
}

//MakeSelectedDirCross обработчик проверки регионов, районов и перекрестков, выбранных пользователем
var MakeSelectedDirCross = func(w http.ResponseWriter, r *http.Request) {
	var selectedData data.SelectedData
	err := json.NewDecoder(r.Body).Decode(&selectedData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.MakeSelectedDir(selectedData)
	u.Respond(w, r, resp)
}
