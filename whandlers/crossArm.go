package whandlers

import (
	"github.com/pkg/errors"
	"net/http"
	"strconv"

	"../data"
	u "../utils"
)

//BuildCross собираем данные для отображения прекрестка
var BuildCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}

	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(w, r)
	if err != nil {
		return
	}

	flag, resp := FuncAccessCheck(w, r, "BuildCross")
	if flag {
		resp = data.GetCrossInfo(*TLight)
	}
	mapContx := u.ParserInterface(r.Context().Value("info"))
	resp["controlCrossFlag"], _ = data.RoleCheck(mapContx, "ControlCross")
	u.Respond(w, r, resp)
}

//ControlCross данные для заполнения таблиц управления
var ControlCross = func(w http.ResponseWriter, r *http.Request) {
	var err error
	TLight := &data.TrafficLights{}
	TLight.Region.Num, TLight.Area.Num, TLight.ID, err = queryParser(w, r)
	if err != nil {
		return
	}
	flag, resp := FuncAccessCheck(w, r, "ControlCross")
	if flag {
		resp = data.ControlGetCrossInfo(*TLight)
	}
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
