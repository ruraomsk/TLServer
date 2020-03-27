package whandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"github.com/pkg/errors"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

//BuildCross обработчик собора данных для отображения прекрёстка
var BuildCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}
	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(w, r)
	if err != nil {
		return
	}
	resp := data.GetCrossInfo(*TLight)
	mapContx := u.ParserInterface(r.Context().Value("info"))

	controlCrossFlag, _ := data.AccessCheck(mapContx, 5)
	if (TLight.Region.Num == mapContx["region"]) || (mapContx["region"] == "*") {
		resp["controlCrossFlag"] = controlCrossFlag
	} else {
		resp["controlCrossFlag"] = false
	}
	u.Respond(w, r, resp)
}

//DevCrossInfo обработчик собора данных для отображения перекрёстка (idevice информация)
var DevCrossInfo = func(w http.ResponseWriter, r *http.Request) {
	var err error
	var idevice string
	if len(r.URL.RawQuery) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field"))
		err = errors.New("Blank field")
		return
	}
	if _, err = strconv.Atoi(r.URL.Query().Get("idevice")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: idevice"))
		return
	} else {
		idevice = r.URL.Query().Get("idevice")
	}
	resp := data.GetCrossDevInfo(idevice)
	u.Respond(w, r, resp)
}

//ControlCross обработчик данных для заполнения таблиц управления
var ControlCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}
	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(w, r)
	if err != nil {
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))

	var resp = make(map[string]interface{})
	if (TLight.Region.Num == mapContx["region"]) || (mapContx["region"] == "*") {
		resp = data.ControlGetCrossInfo(*TLight, mapContx)
	}
	u.Respond(w, r, resp)
}

//ControlEditableCross обработчик проверки редактирования перекрестка
var ControlEditableCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	arm := &data.BusyArm{}
	arm.Region, arm.Area, arm.ID, err = queryParser(w, r)
	if err != nil {
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.ControlEditableCheck(*arm, mapContx)
	u.Respond(w, r, resp)
}

//ControlCloseCross обработчик закрытия перекрестка
var ControlCloseCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	arm := &data.BusyArm{}
	arm.Region, arm.Area, arm.ID, err = queryParser(w, r)
	if err != nil {
		return
	}
	resp := data.BusyArmDelete(*arm)
	u.Respond(w, r, resp)
}

//ControlSendButton обработчик данных для отправки на устройство(сервер)
var ControlSendButton = func(w http.ResponseWriter, r *http.Request) {
	var stateData agS_pudge.Cross
	err := json.NewDecoder(r.Body).Decode(&stateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.SendCrossData(stateData, mapContx)
	u.Respond(w, r, resp)
}

//ControlCreateButton обработчик данных для создания перекрестка и отправка на устройство(сервер)
var ControlCreateButton = func(w http.ResponseWriter, r *http.Request) {
	var stateData agS_pudge.Cross
	err := json.NewDecoder(r.Body).Decode(&stateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.CreateCrossData(stateData, mapContx)
	u.Respond(w, r, resp)
}

//ControlCheckButton обработчик данных для их проверки
var ControlCheckButton = func(w http.ResponseWriter, r *http.Request) {
	var stateData agS_pudge.Cross
	err := json.NewDecoder(r.Body).Decode(&stateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	resp := data.CheckCrossData(stateData)
	u.Respond(w, r, resp)
}

//ControlDeleteButton обработчик данных для удаления перекрестка
var ControlDeleteButton = func(w http.ResponseWriter, r *http.Request) {
	var stateData agS_pudge.Cross
	err := json.NewDecoder(r.Body).Decode(&stateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Invalid request"))
		return
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.DeleteCrossData(stateData, mapContx)
	u.Respond(w, r, resp)
}

//ControlTestState обработчик проверки State
var ControlTestState = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp := data.TestCrossStateData(mapContx)
	u.Respond(w, r, resp)
}

//queryParser разбор URL строки
func queryParser(w http.ResponseWriter, r *http.Request) (region, area string, ID int, err error) {
	if len(r.URL.RawQuery) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field"))
		err = errors.New("Blank field")
		return
	}
	if _, err = strconv.Atoi(r.URL.Query().Get("Region")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: Region"))
		return
	} else {
		region = r.URL.Query().Get("Region")
	}
	if ID, err = strconv.Atoi(r.URL.Query().Get("ID")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: ID"))
		return
	}
	if _, err = strconv.Atoi(r.URL.Query().Get("Area")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Blank field: Area"))
		return
	} else {
		area = r.URL.Query().Get("Area")
	}
	return
}
