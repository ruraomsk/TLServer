package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"net/http"
)

//MainCrossCreator сборка информации для странички создания каталогов перекрестков
var MainCrossCreator = func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, "CrossCreator")
	if flag {
		resp = data.MainCrossCreator()
	}
	u.Respond(w, r, resp)
}

//CheckAllCross обработчик проверки всех перекрестков из БД
var CheckAllCross = func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, "CrossCreator")
	if flag {
		resp = data.CheckCrossDirFromBD()
	}
	u.Respond(w, r, resp)
}

//CheckSelectedDirCross обработчик проверки регионов/районов/перекрестков выбратнных пользователем
var CheckSelectedDirCross = func(w http.ResponseWriter, r *http.Request) {
	var selectedData data.SelectedData
	err := json.NewDecoder(r.Body).Decode(&selectedData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	flag, resp := FuncAccessCheck(w, r, "CrossCreator")
	if flag {
		resp = data.CheckCrossFileSelected(selectedData.SelectedData)
	}
	u.Respond(w, r, resp)
}

//MakeSelectedDirCross обработчик проверки регионов/районов/перекрестков выбратнных пользователем
var MakeSelectedDirCross = func(w http.ResponseWriter, r *http.Request) {
	var selectedData data.SelectedData
	err := json.NewDecoder(r.Body).Decode(&selectedData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}

	flag, resp := FuncAccessCheck(w, r, "CrossCreator")
	if flag {
		resp = data.MakeSelectedDir(selectedData)
	}
	u.Respond(w, r, resp)
}
